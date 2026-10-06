// Runnable self-check: `node src/app/markdown.check.ts`
import assert from 'node:assert/strict'
import { parse, parseSafe, FORMAT_LIMIT } from './markdown.ts'

const j = (s: string) => JSON.stringify(parse(s))
const eq = (src: string, expected: unknown) => assert.deepEqual(parse(src), expected, src)
const T = (v: string) => ({ t: 'text', v })

eq('hello world', [T('hello world')])
eq('**b** *e* _e_ __u__ ~~s~~', [
  { t: 'strong', c: [T('b')] },
  T(' '),
  { t: 'em', c: [T('e')] },
  T(' '),
  { t: 'em', c: [T('e')] },
  T(' '),
  { t: 'u', c: [T('u')] },
  T(' '),
  { t: 's', c: [T('s')] },
])
eq('snake_case_name stays', [T('snake_case_name stays')])
eq('`a*b*` x', [{ t: 'code', v: 'a*b*' }, T(' x')])
eq('```js\nlet a = **1**\n```', [{ t: 'pre', v: 'let a = **1**' }])
eq('||secret||', [{ t: 'spoiler', c: [T('secret')] }])
eq('a\nb', [T('a'), { t: 'br' }, T('b')])
eq('hi :smile: :not an emoji', [T('hi '), { t: 'emoji', v: 'smile' }, T(' :not an emoji')])
eq('\\*not em\\*', [T('*not em*')])

// links: http(s)/mailto only, the rest stays text
eq('see https://neko.m1k1o.net/docs.', [
  T('see '),
  { t: 'link', href: 'https://neko.m1k1o.net/docs', c: [T('https://neko.m1k1o.net/docs')] },
  T('.'),
])
eq('[docs](https://x.io/a)', [{ t: 'link', href: 'https://x.io/a', c: [T('docs')] }])
eq('[https://a.io](https://b.io)', [{ t: 'link', href: 'https://b.io/', c: [T('https://a.io')] }])
eq('<https://x.io>', [{ t: 'link', href: 'https://x.io/', c: [T('https://x.io')] }])
eq('[click](javascript:alert(1))', [T('[click](javascript:alert(1))')])
assert.ok(!j('[x](data:text/html,<script>)').includes('"link"'))
// markup is just text
eq('<script>alert(1)</script>', [T('<script>alert(1)</script>')])
eq('<img src=x onerror=alert(1)>', [T('<img src=x onerror=alert(1)>')])

// quotes: line start only, no nesting
eq('> quoted\nplain', [{ t: 'quote', c: [T('quoted')] }, T('plain')])
eq('>>> all\nof it', [{ t: 'quote', c: [T('all'), { t: 'br' }, T('of it')] }])
eq('a > b', [T('a > b')])
eq('> > nested', [{ t: 'quote', c: [T('> nested')] }])

// hostile input: a fence with no closing fence and hundreds of newlines fits the 512-character
// chat box and used to take seconds; deep nesting used to overflow the stack
const timed = (src: string) => {
  const t0 = performance.now()
  parseSafe(src)
  return performance.now() - t0
}
assert.ok(timed('```a' + '\n'.repeat(508)) < 50, 'unclosed fence must parse in milliseconds')
eq('```a' + '\n'.repeat(3), [T('```a'), { t: 'br' }, { t: 'br' }, { t: 'br' }])
eq('```\ncode\n```', [{ t: 'pre', v: 'code' }])
eq('``````', [T('``````')])
assert.equal(parseSafe('*'.repeat(40_000)).length, 1) // falls back to text instead of throwing
assert.deepEqual(parseSafe('x'.repeat(FORMAT_LIMIT + 1)), [T('x'.repeat(FORMAT_LIMIT + 1))])
assert.ok(timed('\n '.repeat(16_000)) < 200, 'line-start detection must not rescan the whole prefix')

console.log('markdown ok')
