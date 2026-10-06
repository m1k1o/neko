// Runnable self-check of the connection state machine: `node src/core/client.check.ts`
import assert from 'node:assert/strict'

// just enough of the browser for the core to load and connect
const g = globalThis as any
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
const sent: string[] = []
const sockets: FakeSocket[] = []
class FakeSocket {
  static OPEN = 1
  readyState = 1
  onmessage?: (e: { data: string }) => void
  onclose?: () => void
  constructor() {
    sockets.push(this)
  }
  send(data: string) {
    sent.push(JSON.parse(data).event)
  }
  close() {}
}
g.WebSocket = FakeSocket
class FakePeer {
  static all: FakePeer[] = []
  connectionState = 'new'
  signalingState = 'stable'
  remoteDescription: unknown = null
  localDescription = { sdp: 'v=0\r\nm=audio 9 RTP/AVP 0\r\nm=video 9 RTP/AVP 96\r\n' }
  onconnectionstatechange?: () => void
  closed = false
  constructor() {
    FakePeer.all.push(this)
  }
  async setRemoteDescription(d: unknown) {
    this.remoteDescription = d
  }
  async createAnswer() {
    return {}
  }
  async setLocalDescription() {}
  async addIceCandidate() {}
  getSenders() {
    return []
  }
  close() {
    this.closed = true
  }
}
g.RTCPeerConnection = FakePeer

const { NekoClient } = await import('./client.ts')
const settle = (ms = 20) => new Promise((r) => setTimeout(r, ms))

const client = new NekoClient()
client.state.authenticated = true
client.connect()
const ws = sockets[0]
const receive = (event: string, payload?: unknown) => ws.onmessage!({ data: JSON.stringify({ event, payload }) })

receive('system/init', {
  session_id: 's1',
  sessions: {},
  settings: {},
  screen_size: { width: 1280, height: 720, rate: 30 },
})
assert.deepEqual(sent, ['signal/request'])
receive('signal/provide', { sdp: 'v=0\r\nm=video 9 RTP/AVP 96\r\n', iceservers: [] })
await settle()
assert.deepEqual(sent, ['signal/request', 'signal/answer'])
const pc = FakePeer.all[0]
pc.connectionState = 'connected'
pc.onconnectionstatechange!()
await settle()
assert.equal(client.state.connection.status, 'connected')

// the server dropped its peer (it noticed the media path failing first): the client must ask
// for a new one instead of sitting on a closed connection with status "connected"
sent.length = 0
receive('signal/close')
await settle()
assert.ok(pc.closed)
assert.equal(client.state.connection.status, 'connecting')
await settle(1600) // RECONNECT_BACKOFF_MS
assert.deepEqual(sent, ['signal/request'])

// a sender that belonged to the previous peer connection is ignored, not thrown on
receive('signal/provide', { sdp: 'v=0\r\nm=video 9 RTP/AVP 96\r\n' })
await settle()
assert.equal(FakePeer.all.length, 2)
client.removeTrack({} as RTCRtpSender)

// before a paste keystroke: the side that copied most recently wins
receive('clipboard/updated', { text: 'remote' })
sent.length = 0
await client.preparePaste('remote') // same text as the remote clipboard: nothing to do
assert.deepEqual(sent, [])
let settled = false
client.preparePaste('local').then(() => (settled = true)) // newer than anything the remote reported
await settle()
assert.deepEqual([sent, settled], [['clipboard/set'], false])
receive('clipboard/updated', { text: 'local' }) // resolves once the server confirms
await settle()
assert.equal(settled, true)
await settle(5)
receive('clipboard/updated', { text: 'remote2' }) // the remote copied since: local text is taken as older
sent.length = 0
await client.preparePaste('local')
assert.deepEqual(sent, [])
const t0 = Date.now()
await client.preparePaste('other') // nothing changed remotely since that look: local wins again
assert.deepEqual(sent, ['clipboard/set'])
assert.ok(Date.now() - t0 >= 450, 'waits for the confirmation, or 500 ms')

// no offer after a request: the room is shown without video only while the server is talking.
// After a lost network the media fails first and the request goes into a dead socket; that must
// not become "connected" over a frozen picture (the stale check will reconnect)
const realTimeout = setTimeout
const timers: { fn: () => void; ms: number }[] = []
g.setTimeout = (fn: () => void, ms: number) => timers.push({ fn, ms })
const fire = (ms: number) =>
  timers
    .splice(
      timers.findIndex((t) => t.ms === ms),
      1,
    )[0]
    .fn()
sent.length = 0
receive('signal/close')
fire(1500) // RECONNECT_BACKOFF_MS -> signal/request
assert.deepEqual(sent, ['signal/request'])
;(client as any).lastMessage = Date.now() - 30_000 // the server has been silent since
fire(8000) // OFFER_TIMEOUT_MS
assert.equal(client.state.connection.status, 'connecting', 'a silent server is not "connected" without video')
receive('system/heartbeat') // alive after all
fire(8000)
assert.equal(client.state.connection.status, 'connected')
g.setTimeout = realTimeout

client.disconnect()
console.log('client ok')
