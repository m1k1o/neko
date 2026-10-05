// Runnable self-check: `node src/app/markdown.check.ts`
import assert from 'node:assert/strict'
import { parse } from './markdown.ts'

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

console.log('markdown ok')
