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

client.disconnect()
console.log('client ok')
