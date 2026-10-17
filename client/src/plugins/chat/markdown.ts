// The chat's markdown: react-markdown (CommonMark; GFM strikethrough, tables and bare links;
// a newline is a line break) with the chat's own dialect on top as rehype plugins (`:name:` emoji,
// `||spoiler||`, `__underline__`), and the rules for text that comes from other users: links are
// http(s)/mailto only and anything else stays the text it was typed as, so do images; raw HTML is
// text too (react-markdown's default). Markdown.tsx renders the tree; nothing is an HTML string.
import type { Element, ElementContent, Root } from 'hast'
import type { Plugin, PluggableList } from 'unified'
import remarkBreaks from 'remark-breaks'
import remarkGfm from 'remark-gfm'
import { emoji } from './emoji'

// Above FORMAT_LIMIT characters a message is shown as plain text; so is one the renderer fails on
// (Markdown.tsx). The limit is the server's default chat.max_length, so only a server configured
// for longer messages ever hits it, and a hostile message stays cheap (nested quotes cost micromark
// about depth²: 2000 bare `>` took 130 ms, 5000 `> ` 420 ms)
export const FORMAT_LIMIT = 512

// only these schemes become links; everything else stays text (javascript:, data:, relative, ...)
export function safeUrl(href: string): string | null {
  try {
    const u = new URL(href)
    return ['http:', 'https:', 'mailto:'].includes(u.protocol) ? u.href : null
  } catch {
    return null
  }
}

// f sees every element of the tree but code (its text is verbatim), parents first; it may replace
// the children
function walk(node: Root | Element, f: (el: Element) => void) {
  if (node.type === 'element') f(node)
  for (const c of node.children) if (c.type === 'element' && c.tagName !== 'code') walk(c, f)
}

// the text children of el split on re: each match becomes make(match), or stays text when that is null
function splitText(el: Element, re: RegExp, make: (m: RegExpExecArray) => ElementContent | null) {
  el.children = el.children.flatMap((c) => {
    if (c.type !== 'text') return [c]
    const out: ElementContent[] = []
    let last = 0
    for (const m of c.value.matchAll(re)) {
      const node = make(m)
      if (!node) continue
      if (m.index > last) out.push({ type: 'text', value: c.value.slice(last, m.index) })
      out.push(node)
      last = m.index + m[0].length
    }
    if (last < c.value.length) out.push({ type: 'text', value: c.value.slice(last) })
    return out
  })
}

// a link to anything but http(s)/mailto ([x](javascript:...), <data:...>, a relative path) and an
// image stay the text they were typed as; a generated one (a GFM footnote's) keeps its text only
const rehypeSafe: Plugin<[], Root> = () => (tree, file) => {
  const src = String(file)
  walk(tree, (el) => {
    el.children = el.children.flatMap((c) => {
      if (c.type !== 'element') return [c]
      const unsafe = c.tagName === 'img' || (c.tagName === 'a' && !safeUrl(String(c.properties.href ?? '')))
      if (!unsafe) return [c]
      return c.position
        ? [{ type: 'text', value: src.slice(c.position.start.offset, c.position.end.offset) }]
        : c.children
    })
  })
}

// __underline__: CommonMark reads it as strong; the marker in the source tells them apart
const rehypeUnderline: Plugin<[], Root> = () => (tree, file) => {
  const src = String(file)
  walk(tree, (el) => {
    if (el.tagName === 'strong' && src[el.position?.start.offset ?? -1] === '_') el.tagName = 'u'
  })
}

// ||spoiler||: a span Markdown.tsx renders click-to-reveal. Plain text only: a spoiler with
// formatting inside (||**x**||) is split by the parser first and stays as typed
const rehypeSpoiler: Plugin<[], Root> = () => (tree) => {
  walk(tree, (el) =>
    splitText(el, /\|\|(.+?)\|\|/g, (m) => ({
      type: 'element',
      tagName: 'span',
      properties: { dataSpoiler: true },
      children: [{ type: 'text', value: m[1] }],
    })),
  )
}

// :name: of a name the emoji data knows (emoji.ts): an empty span with the name, which Markdown.tsx
// renders as the emoji (Emoji.tsx: the character, or the image of a custom one)
const rehypeEmoji: Plugin<[], Root> = () => (tree) => {
  walk(tree, (el) =>
    splitText(el, /:([^:\s]+):/g, (m) =>
      emoji.names.has(m[1])
        ? {
            type: 'element',
            tagName: 'span',
            properties: { dataEmoji: m[1], title: m[0] },
            children: [],
          }
        : null,
    ),
  )
}

// react-markdown's options (Markdown.tsx adds the components and urlTransform: safeUrl)
export const remarkPlugins: PluggableList = [[remarkGfm, { singleTilde: false }], remarkBreaks]
export const rehypePlugins: PluggableList = [rehypeSafe, rehypeUnderline, rehypeSpoiler, rehypeEmoji]
// the plugins above leave no image in the tree; should one get through, it is dropped here
export const disallowedElements = ['img']
