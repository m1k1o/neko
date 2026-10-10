// Unit tests of the chat's markdown: `npm test` (vitest, react-dom/server). The Markdown component
// is rendered under a <NekoProvider> with the browser faked (src/test/browser.ts) and the emoji
// data set by hand; the HTML of each message is asserted.
import { expect, test, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: {} }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
fakeBrowser()
const { createNekoApp } = await import('@/state/neko')
const { NekoProvider } = await import('@/state/provider')
const { emoji } = await import('./emoji.ts')
const { FORMAT_LIMIT, safeUrl } = await import('./markdown.ts')
const { Markdown } = await import('./Markdown.tsx')

Object.assign(emoji, { names: new Set(['smile', 'neko']), chars: { smile: '😄' } })
const neko = createNekoApp({ transport: fakeTransport() })

// the message's HTML, inside the chat-body div
const html = (source: string, app = neko) => {
  const out = renderToStaticMarkup(
    <NekoProvider neko={app}>
      <Markdown source={source} />
    </NekoProvider>,
  )
  // React 19 hoists a <link rel="preload"> before the div for an image (the custom emoji)
  const m = /^(?:<link [^>]*\/>)*<div class="[^"]*" data-testid="chat-body">([\s\S]*)<\/div>$/.exec(out)
  if (!m) throw new Error('not a chat body: ' + out)
  return m[1]
}
const p = (inner: string) => `<p>${inner}</p>`
const anchors = (out: string) => out.match(/<a /g)?.length ?? 0
const SPOILER =
  '<button type="button" class="cursor-pointer rounded bg-background-tertiary px-0.5 [&amp;&gt;span]:opacity-0" aria-label="Spoiler" title="Spoiler" data-testid="spoiler"><span>secret</span></button>'
const EMOJI =
  '<span class="inline-block h-[22px] w-[22px] text-center align-bottom text-[18px] leading-[22px]" data-emoji="smile" title=":smile:">😄</span>'

test('inline formatting, code, spoilers, line breaks, emoji, escapes', () => {
  expect(html('hello world')).toBe(p('hello world'))
  expect(html('**b** *e* _e_ __u__ ~~s~~')).toBe(p('<strong>b</strong> <em>e</em> <em>e</em> <u>u</u> <del>s</del>'))
  expect(html('~one tilde~ snake_case_name stays')).toBe(p('~one tilde~ snake_case_name stays'))
  expect(html('`a*b*` x')).toMatch(/^<p><code class="[^"]*font-mono[^"]*">a\*b\*<\/code> x<\/p>$/)
  expect(html('```js\nlet a = **1**\n```')).toMatch(
    /^<pre class="[^"]*"><code class="[^"]*font-mono[^"]*">let a = \*\*1\*\*\n<\/code><\/pre>$/,
  )
  expect(html('||secret||')).toBe(p(SPOILER))
  expect(html('||**x**||')).toBe(p('||<strong>x</strong>||')) // not supported: formatting inside
  expect(html('||a||b||c||')).toBe(p(SPOILER.replace('secret', 'a') + 'b' + SPOILER.replace('secret', 'c')))
  expect(html('a\nb')).toBe(p('a<br/>\nb'))
  expect(html('a\n\nb')).toBe(p('a') + '\n' + p('b'))
  expect(html('hi :smile: :not an emoji :nope: 12:30:45')).toBe(p(`hi ${EMOJI} :not an emoji :nope: 12:30:45`))
  expect(html('||:smile:||')).toBe(p(SPOILER.replace('<span>secret</span>', `<span>${EMOJI}</span>`)))
  expect(html('`:smile:` and `||x||`')).not.toContain('data-')
  expect(html(':neko:')).toMatch(/^<p><img class="[^"]*" src="[^"]+" alt=":neko:" data-emoji="neko"\/><\/p>$/) // the custom one is an image of ours
  expect(html('\\*not em\\*')).toBe(p('*not em*'))
})

test('links: http(s)/mailto only, the rest stays the text it was typed as; markup is just text', () => {
  expect(html('see https://neko.m1k1o.net/docs.')).toBe(
    p(
      'see <a class="text-text-link underline" href="https://neko.m1k1o.net/docs" target="_blank" rel="noopener noreferrer">https://neko.m1k1o.net/docs</a>.',
    ),
  )
  expect(html('[docs](https://x.io/a)')).toContain(
    'href="https://x.io/a" target="_blank" rel="noopener noreferrer">docs</a>',
  )
  const nested = html('[https://a.io](https://b.io)')
  expect(anchors(nested)).toBe(1)
  expect(nested).toContain('href="https://b.io/" target="_blank" rel="noopener noreferrer">https://a.io</a>')
  expect(html('<https://x.io>')).toContain(
    'href="https://x.io/" target="_blank" rel="noopener noreferrer">https://x.io</a>',
  )
  expect(html('mail me@x.io or www.x.io')).toMatch(/href="mailto:me@x.io".*href="http:\/\/www.x.io\/"/)
  for (const src of [
    '[click](javascript:alert(1))',
    '[x](data:text/html,<script>)',
    '<javascript:alert(1)>',
    '[x](/relative/path)',
    '[x](#fragment)',
    '[x][1]\n\n[1]: javascript:alert(1)',
  ]) {
    const out = html(src)
    expect(out, src).not.toContain('<a')
    expect(out, src).not.toContain('href')
    expect(out, src).toContain(src.split('\n')[0].replace(/</g, '&lt;').replace(/>/g, '&gt;'))
  }
  expect(html('<script>alert(1)</script>')).toBe('&lt;script&gt;alert(1)&lt;/script&gt;') // an HTML block: no paragraph
  expect(html('<img src=x onerror=alert(1)>')).toBe('&lt;img src=x onerror=alert(1)&gt;') // an HTML block too
  expect(html('<b>b</b> <span data-emoji="smile">x</span>')).toBe(
    p('&lt;b&gt;b&lt;/b&gt; &lt;span data-emoji=&quot;smile&quot;&gt;x&lt;/span&gt;'),
  )
  for (const src of [
    '![a](https://x.io/y.png)',
    '![a][i]\n\n[i]: https://x.io/y.png',
    '<img src="https://x.io/y.png">',
  ]) {
    const out = html(src)
    expect(out, src).not.toContain('<img')
    expect(out, src).toContain(src.split('\n')[0].replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;'))
  }
  expect(safeUrl('HTTP://X.IO')).toBe('http://x.io/')
  expect(safeUrl('ftp://x.io')).toBeNull()
})

test('quotes and the CommonMark extras: lists, headings, tables; the `>>>` block quote is gone', () => {
  expect(html('> quoted\n\nplain')).toBe('<blockquote>\n' + p('quoted') + '\n</blockquote>\n' + p('plain'))
  expect(html('> a\n> b')).toBe('<blockquote>\n' + p('a<br/>\nb') + '\n</blockquote>')
  expect(html('a > b')).toBe(p('a &gt; b'))
  expect(html('> > nested')).toBe('<blockquote>\n<blockquote>\n' + p('nested') + '\n</blockquote>\n</blockquote>')
  expect(html('>>> all\nof it')).toBe(
    '<blockquote>\n<blockquote>\n<blockquote>\n' +
      p('all<br/>\nof it') +
      '\n</blockquote>\n</blockquote>\n</blockquote>',
  )
  expect(html('- a\n- b')).toBe('<ul>\n<li>a</li>\n<li>b</li>\n</ul>')
  expect(html('# title')).toBe('<h1>title</h1>')
  expect(html('#hashtag')).toBe(p('#hashtag'))
  expect(html('| a | b |\n| - | - |\n| 1 | 2 |')).toMatch(
    /^<table>[\s\S]*<th>a<\/th>[\s\S]*<td>2<\/td>[\s\S]*<\/table>$/,
  )
})

test('open in app: the icon after a link when the plugin is there and we host', () => {
  const app = createNekoApp({ transport: fakeTransport() })
  // react-dom/server reads a zustand store's initial state (useSyncExternalStore): make it the current one
  app.app.getInitialState = app.app.getState
  app.client.store.getInitialState = app.client.store.getState
  const link = '[x](https://x.io)'
  expect(html(link, app)).not.toContain('Open in app')
  app.app.setState({ openInApp: true })
  expect(html(link, app)).not.toContain('Open in app') // not hosting
  const s = app.client.state
  app.client.store.setState({ session_id: 'me', control: { ...s.control, host_id: 'me' } })
  expect(html(link, app)).toMatch(/<\/a><button type="button" aria-label="Open in app"/)
})

// a fence with no closing fence and hundreds of newlines fits the 512-character chat box and used
// to take seconds in the old parser; deep nesting used to overflow its stack
test('hostile input: renders in milliseconds or falls back to the text, never throws', () => {
  const timed = (src: string) => {
    const t0 = performance.now()
    const out = html(src)
    return [performance.now() - t0, out] as const
  }
  const [fence] = timed('```a' + '\n'.repeat(508))
  expect(fence).toBeLessThan(50)
  expect(html('```a' + '\n'.repeat(3))).toMatch(/^<pre class="[^"]*"><code class="[^"]*">\n*<\/code><\/pre>$/)
  expect(html('```\ncode\n```')).toMatch(/^<pre class="[^"]*"><code class="[^"]*">code\n<\/code><\/pre>$/)
  expect(html('``````')).toMatch(/^<pre/) // CommonMark: an unclosed fence of six backticks, empty
  for (const src of [
    '*'.repeat(FORMAT_LIMIT),
    '* '.repeat(FORMAT_LIMIT / 2),
    '> '.repeat(FORMAT_LIMIT / 2),
    '>'.repeat(FORMAT_LIMIT), // nested quotes: micromark's cost grows with depth²
    '- '.repeat(FORMAT_LIMIT / 2),
    '['.repeat(FORMAT_LIMIT),
    '<'.repeat(FORMAT_LIMIT),
    '`'.repeat(FORMAT_LIMIT),
    '|'.repeat(FORMAT_LIMIT),
    '_a'.repeat(FORMAT_LIMIT / 2),
    '\n '.repeat(FORMAT_LIMIT / 2),
    ':'.repeat(FORMAT_LIMIT),
  ]) {
    const [ms, out] = timed(src)
    expect(ms, JSON.stringify(src.slice(0, 4))).toBeLessThan(1000)
    expect(typeof out, JSON.stringify(src.slice(0, 4))).toBe('string')
  }
  // above FORMAT_LIMIT: the text as it is
  expect(html('*'.repeat(40_000))).toBe('*'.repeat(40_000))
  expect(html('x'.repeat(FORMAT_LIMIT + 1))).toBe('x'.repeat(FORMAT_LIMIT + 1))
  expect(html('**b**' + 'x'.repeat(FORMAT_LIMIT))).toBe('**b**' + 'x'.repeat(FORMAT_LIMIT))
  expect(html('**b**' + 'x'.repeat(FORMAT_LIMIT - 5))).toBe(p('<strong>b</strong>' + 'x'.repeat(FORMAT_LIMIT - 5)))
  const [spaces] = timed('\n '.repeat(16_000))
  expect(spaces).toBeLessThan(200)
})
