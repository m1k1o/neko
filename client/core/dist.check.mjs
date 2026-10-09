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
g.ResizeObserver = class {
  observe() {}
  disconnect() {}
}

const dist = join(import.meta.dirname, 'dist')
const pkg = await import('./dist/index.js')
const { WebRTCTransport } = pkg

// the public surface
for (const name of [
  'NekoClient',
  'Overlay',
  'Store',
  'Emitter',
  'NekoApi',
  'ApiError',
  'WebRTCTransport',
  'WebSocketInput',
  'DataChannelInput',
])
  assert.equal(typeof pkg[name], 'function', name)
assert.equal(WebRTCTransport.supported(), typeof RTCPeerConnection !== 'undefined')
assert.equal(typeof pkg.OP, 'object')
assert.equal(pkg.OP.MOVE, 1)

// Store: a nested write bumps the version and reaches a watcher one microtask later
const s = new pkg.Store({ a: { b: 1 } })
const seen = []
s.watch(
  () => s.state.a.b,
  (v, old) => seen.push([v, old]),
)
s.state.a.b = 2
await Promise.resolve()
assert.equal(s.version, 1)
assert.deepEqual(seen, [[2, 1]])

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
  suspend: noop,
  close: noop,
  attach: noop,
  detach: noop,
  setPlaying: async () => {},
  setVolume: noop,
  setMuted: noop,
  on: () => noop,
}
const other = new pkg.NekoClient({ transport: custom })
assert.equal(other.transport, custom)
assert.ok(other.input instanceof pkg.WebSocketInput, 'no input channel on the transport: the websocket')

// build.mjs rewrote the `.ts` specifiers the declarations inherit from the sources
for (const f of readdirSync(dist, { recursive: true }).filter((f) => f.endsWith('.d.ts'))) {
  const src = readFileSync(join(dist, f), 'utf8')
  assert.ok(!/from '[^']*\.ts'/.test(src) && !/import\("[^"]*\.ts"\)/.test(src), `${f} still imports a .ts path`)
}

console.log('dist ok')
