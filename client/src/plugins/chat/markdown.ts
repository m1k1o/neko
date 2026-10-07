// Chat markdown, same rules as the legacy client (Discord-like subset):
// ```code```, `code`, **bold**, *em* / _em_, __underline__, ~~strike~~, ||spoiler||,
// > quote / >>> quote, [text](url), <url>, bare urls, :emoji:, \escapes.
// Parses to a plain tree that React renders as elements: no HTML strings, no innerHTML,
// so message text can never become markup.

export type Node =
  | { t: 'text'; v: string }
  | { t: 'br' }
  | { t: 'code' | 'pre'; v: string }
  | { t: 'emoji'; v: string }
  | { t: 'link'; href: string; c: Node[] }
  | { t: 'strong' | 'em' | 'u' | 's' | 'spoiler' | 'quote'; c: Node[] }

// only these schemes become links; everything else stays text (javascript:, data:, ...)
export function safeUrl(href: string): string | null {
  try {
    const u = new URL(href)
    return ['http:', 'https:', 'mailto:'].includes(u.protocol) ? u.href : null
  } catch {
    return null
  }
}

// link text is not linked again: [https://a](https://b) would nest two anchors
const unlink = (nodes: Node[]): Node[] =>
  nodes.flatMap((n) => (n.t === 'link' ? unlink(n.c) : 'c' in n ? [{ ...n, c: unlink(n.c) }] : [n]))

const URL_RE = /^(https?:\/\/[^\s<]+[^<.,:;"')\]\s])/

// what a rule may look at: the text from here on, and two facts about what came before
interface Before {
  lineStart: boolean // only spaces since the start of the line
  word: boolean // the previous character is a word character
}
type Rule = (src: string, before: Before, inQuote: boolean) => { len: number; node: Node } | null

const wrap =
  (re: RegExp, t: 'strong' | 'em' | 'u' | 's' | 'spoiler'): Rule =>
  (src, _before, inQuote) => {
    const m = re.exec(src)
    return m ? { len: m[0].length, node: { t, c: parse(m[1], inQuote) } } : null
  }

const rules: Rule[] = [
  // \* -> literal *
  (src) => {
    const m = /^\\([^0-9A-Za-z\s])/.exec(src)
    return m ? { len: 2, node: { t: 'text', v: m[1] } } : null
  },
  // ```lang\ncode``` (language is ignored, like before). Found with indexOf: the regex this
  // replaces backtracked for seconds on a fence followed by many newlines and no closing fence
  (src) => {
    if (!src.startsWith('```')) return null
    const end = src.indexOf('```', 3)
    if (end < 4) return null
    const body = src.slice(3, end).replace(/^[a-z0-9-]+\n/i, '')
    return { len: end + 3, node: { t: 'pre', v: body.replace(/^\n+|\n+$/g, '') } }
  },
  // quotes only start a line and do not nest; >>> quotes the rest of the message
  (src, before, inQuote) => {
    if (inQuote || !before.lineStart) return null
    const quote = (len: number, body: string) => ({ len, node: { t: 'quote', c: parse(body, true) } as Node })
    const block = /^ *>>> ?([\s\S]*)/.exec(src)
    if (block) return quote(block[0].length, block[1])
    const lines = /^ *> [^\n]+(?:\n *> [^\n]+)*\n?/.exec(src)
    return lines ? quote(lines[0].length, lines[0].replace(/^ *> ?/gm, '').replace(/\n$/, '')) : null
  },
  wrap(/^\|\|([\s\S]+?)\|\|/, 'spoiler'),
  // `code`: closed by the next run of exactly as many backticks (the regex with a backreference
  // this replaces was quadratic per position on long runs of backticks)
  (src) => {
    if (src[0] !== '`') return null
    let k = 1
    while (src[k] === '`') k++
    const ticks = src.slice(0, k)
    for (let from = k; ; from = src.indexOf(ticks, from) + 1) {
      const end = src.indexOf(ticks, from)
      if (end < 0) return null
      if (src[end - 1] === '`' || src[end + k] === '`') continue
      const v = src.slice(k, end).trim()
      return v ? { len: end + k, node: { t: 'code', v } } : null
    }
  },
  // [text](url)
  (src, _prev, inQuote) => {
    const m = /^\[([^\]\n]+)\]\(([^)\s]+)\)/.exec(src)
    if (!m) return null
    const href = safeUrl(m[2])
    return {
      len: m[0].length,
      node: href ? { t: 'link', href, c: unlink(parse(m[1], inQuote)) } : { t: 'text', v: m[0] },
    }
  },
  // <url>
  (src) => {
    const m = /^<([^: >]+:\/[^ >]+)>/.exec(src)
    const href = m && safeUrl(m[1])
    return m && href ? { len: m[0].length, node: { t: 'link', href, c: [{ t: 'text', v: m[1] }] } } : null
  },
  (src) => {
    const m = URL_RE.exec(src)
    const href = m && safeUrl(m[1])
    return m && href ? { len: m[0].length, node: { t: 'link', href, c: [{ t: 'text', v: m[1] }] } } : null
  },
  wrap(/^\*\*([\s\S]+?)\*\*(?!\*)/, 'strong'),
  wrap(/^__([\s\S]+?)__(?!_)/, 'u'),
  wrap(/^~~([\s\S]+?)~~/, 's'),
  wrap(/^\*(?=\S)([\s\S]*?\S)\*(?!\*)/, 'em'),
  // _em_ only at word boundaries, so snake_case_words stay text
  (src, before, inQuote) => {
    if (before.word) return null
    const m = /^_(?=\S)([\s\S]*?\S)_(?!\w)/.exec(src)
    return m ? { len: m[0].length, node: { t: 'em', c: parse(m[1], inQuote) } } : null
  },
  (src) => {
    const m = /^:([^:\s]+):/.exec(src)
    return m ? { len: m[0].length, node: { t: 'emoji', v: m[1] } } : null
  },
  (src) => (src[0] === '\n' ? { len: 1, node: { t: 'br' } } : null),
]

// plain text runs up to the next character that could start a rule
const TEXT_RE = /^[\s\S]+?(?=[\\`*_~|<[:>\n]|https?:\/\/|$)/

export function parse(src: string, inQuote = false): Node[] {
  const out: Node[] = []
  let i = 0
  while (i < src.length) {
    const rest = src.slice(i)
    let j = i
    while (j > 0 && src[j - 1] === ' ') j--
    const before: Before = { lineStart: j === 0 || src[j - 1] === '\n', word: /\w/.test(src[i - 1] ?? '') }
    let hit = null
    for (const rule of rules) if ((hit = rule(rest, before, inQuote))) break
    if (!hit) {
      const len = (TEXT_RE.exec(rest)?.[0] ?? rest[0]).length || 1
      hit = { len, node: { t: 'text', v: rest.slice(0, len) } as Node }
    }
    // merge neighbouring text so the tree stays small
    const last = out[out.length - 1]
    if (hit.node.t === 'text' && last?.t === 'text') last.v += hit.node.v
    else out.push(hit.node)
    i += hit.len
  }
  return out
}

// Messages come from other users. Above FORMAT_LIMIT characters they are shown as plain text
// (formatting a message that long is not worth a stalled renderer), and a parser failure of
// any kind, such as a stack overflow from pathological nesting, also falls back to plain text.
export const FORMAT_LIMIT = 10_000
export function parseSafe(src: string): Node[] {
  if (src.length > FORMAT_LIMIT) return [{ t: 'text', v: src }]
  try {
    return parse(src)
  } catch {
    return [{ t: 'text', v: src }]
  }
}
