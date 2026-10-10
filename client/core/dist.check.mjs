// Runnable self-check of the built package in dist/ (after `npm run build`): `node dist.check.mjs`
import assert from 'node:assert/strict'
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

// just enough of the browser for the core to load and construct (as in src/client.test.ts)
const g = globalThis
g.window = g
g.location = { href: 'http://neko.test/' }
Object.defineProperty(g, 'navigator', {
  value: { platform: 'Linux', userAgent: 'check', maxTouchPoints: 0 },
  configurable: true,
})
g.localStorage = { getItem: () => null, setItem() {}, removeItem() {} }
g.matchMedia = () => ({ matches: false })

const dist = join(import.meta.dirname, 'dist')
const pkg = await import('./dist/index.js')
const { WebRTCTransport } = pkg

// the public surface
for (const name of [
  'NekoClient',
  'Overlay',
  'mount',
  'Emitter',
  'selectSession',
  'selectControlling',
  'selectIsAdmin',
  'NekoApi',
  'ApiError',
  'WebRTCTransport',
  'WebSocketInput',
  'DataChannelInput',
])
  assert.equal(typeof pkg[name], 'function', name)

// the client is free of the keyboard/mouse overlay and the Guacamole keyboard: a stream-only
// consumer that imports NekoClient (and tree-shakes the rest, sideEffects: false) never loads them
const imports = (file, seen = new Set()) => {
  if (seen.has(file)) return seen
  seen.add(file)
  for (const [, spec] of readFileSync(join(dist, file), 'utf8').matchAll(/(?:from|import) ["'](\.[^"']+)["']/g))
    imports(join(file, '..', spec), seen)
  return seen
}
const clientGraph = [...imports('client.js')]
assert.ok(
  !clientGraph.some((f) => /overlay|keyboard/.test(f)),
  'client.js must not reach the overlay or the keyboard: ' + clientGraph.join(', '),
)
assert.ok(
  [...imports('overlay.js')].some((f) => f.startsWith('keyboard/')),
  'the keyboard is still used by the overlay',
)
assert.equal(WebRTCTransport.supported(), typeof RTCPeerConnection !== 'undefined')
assert.ok(!('OP' in pkg), 'the data channel opcodes are not part of the API')

// the store: zustand, the one runtime dependency, resolvable from the built package (client.js keeps
// the bare specifiers); a selector subscription fires on its value only, updates are immutable, and
// the proxy store's API (Store, watch, version) is gone
assert.ok(!('Store' in pkg), 'the proxy Store is gone')
await import('zustand/vanilla')
assert.ok(readFileSync(join(dist, 'client.js'), 'utf8').includes("from 'zustand/vanilla'"))
const probe = new pkg.NekoClient({ autologin: false, autoconnect: false })
assert.equal(probe.state, probe.store.getState())
assert.ok(!('watch' in probe.store) && !('version' in probe.store))
const seen = []
probe.store.subscribe(
  (s) => s.control.host_id,
  (v, old) => seen.push([v, old]),
)
probe.store.setState({ mobile_keyboard_open: true })
assert.deepEqual(seen, [])
const before = probe.state
probe.store.setState({ session_id: 'h', control: { ...before.control, host_id: 'h' } })
assert.deepEqual(seen, [['h', null]])
assert.notEqual(probe.state, before)
assert.equal(probe.state.video, before.video, 'untouched slices keep their identity')
assert.equal(pkg.selectControlling(probe.state), true)
assert.equal(probe.controlling, true)

// Emitter: on() returns the unsubscribe
const e = new pkg.Emitter()
const got = []
const off = e.on('x', (v) => got.push(v))
e.emit('x', 1)
off()
e.emit('x', 2)
assert.deepEqual(got, [1])

// NekoClient: constructs without a connection, setUrl normalises the server URL, and the
// transport seam: WebRTC by default (its data channel as the input), or the one given
const client = new pkg.NekoClient({ autologin: false, autoconnect: false })
assert.equal(client.state.connection.status, 'disconnected')
client.setUrl('http://neko.test/some/path/')
assert.equal(client.api.url, 'http://neko.test/some/path')
assert.ok(client.transport instanceof pkg.WebRTCTransport)
assert.equal(client.transport.kind, 'webrtc')
assert.ok(client.input instanceof pkg.DataChannelInput)
const noop = () => {}
const custom = {
  kind: 'hls',
  element: null,
  connect: async () => {},
  close: noop,
  attach: () => noop,
  setPlaying: async () => {},
  setVolume: noop,
  setMuted: noop,
  on: () => noop,
}
const other = new pkg.NekoClient({ transport: custom })
assert.equal(other.transport, custom)
assert.ok(other.input instanceof pkg.WebSocketInput, 'no input channel on the transport: the websocket')
for (const m of ['move', 'scroll', 'button', 'key', 'touch']) assert.equal(typeof other.input[m], 'function', m)
assert.throws(() => other.shareMedia({}), /transport cannot send media/)

// a stream-only consumer: the transport's element in a box, no overlay (no input, no cursor), and
// input from a device of its own straight through client.input
const box = { children: [], append: (el) => box.children.push(el) }
custom.attach = (el) => (el.append('media'), () => el.children.pop())
const detach = other.transport.attach(box)
assert.deepEqual(box.children, ['media'])
detach()
assert.deepEqual(box.children, [])
other.input.key(0xff0d, true) // a gamepad mapped to Return, say; dropped here: no socket

// build.mjs rewrote the `.ts` specifiers the declarations inherit from the sources
for (const f of readdirSync(dist, { recursive: true }).filter((f) => f.endsWith('.d.ts'))) {
  const src = readFileSync(join(dist, f), 'utf8')
  assert.ok(!/from '[^']*\.ts'/.test(src) && !/import\("[^"]*\.ts"\)/.test(src), `${f} still imports a .ts path`)
}

console.log('dist ok')
