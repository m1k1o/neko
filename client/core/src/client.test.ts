// Unit tests of the connection state machine: `npm test` (vitest). The browser is faked just far
// enough for the core to load and connect, and the clock is vitest's: `tick(ms)` fires the timers
// that fall due, in order, so nothing waits.
import { test, vi } from 'vitest'
import assert from 'node:assert/strict'

const g = globalThis as any
g.window = g
g.location = { href: 'http://neko.test/' }
Object.defineProperty(g, 'navigator', {
  value: { platform: 'Linux', userAgent: 'test', maxTouchPoints: 0 },
  configurable: true,
})
const storage = new Map<string, string>()
g.localStorage = {
  getItem: (k: string) => storage.get(k) ?? null,
  setItem: (k: string, v: string) => storage.set(k, v),
  removeItem: (k: string) => storage.delete(k),
}
g.matchMedia = () => ({ matches: false })
g.ResizeObserver = class {
  observe() {}
  disconnect() {}
}

// the clock: timers and Date are fake, setImmediate stays real so promise chains can be drained
vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout', 'setInterval', 'clearInterval', 'Date'] })
// one macrotask turn drains every pending promise chain (the handshake, a fetch)
const settle = () => new Promise<void>((r) => setImmediate(r))
const tick = async (ms: number) => {
  await settle() // whatever is in flight may still schedule a timer
  await vi.advanceTimersByTimeAsync(ms) // fires the due timers in order, settling after each
}

// the network
const sent: string[] = []
const sockets: FakeSocket[] = []
class FakeSocket {
  static OPEN = 1
  readyState = 1
  url: string
  closed = false
  code?: number
  onmessage?: (e: { data: string }) => void
  onclose?: () => void
  constructor(url: string) {
    this.url = url
    sockets.push(this)
  }
  send(data: string) {
    sent.push(JSON.parse(data).event)
  }
  close(code?: number) {
    this.closed = true
    this.code = code
  }
}
g.WebSocket = FakeSocket
const OFFER = 'v=0\r\nm=audio 9 RTP/AVP 0\r\na=rtpmap:0 opus/48000\r\nm=video 9 RTP/AVP 96\r\na=rtpmap:96 VP8/90000\r\n'
const ANSWER = 'v=0\r\nm=audio 9 RTP/AVP 0\r\nm=video 9 RTP/AVP 96\r\n'
class FakePeer {
  static all: FakePeer[] = []
  static answerSdp = ANSWER
  connectionState = 'new'
  signalingState = 'stable'
  remoteDescription: unknown = null
  localDescription: { sdp: string } | null = null
  candidates: unknown[] = []
  closed = false
  onconnectionstatechange?: () => void
  constructor() {
    FakePeer.all.push(this)
  }
  async setRemoteDescription(d: unknown) {
    this.remoteDescription = d
  }
  async createAnswer() {
    return { type: 'answer', sdp: FakePeer.answerSdp }
  }
  async setLocalDescription(d: { sdp: string }) {
    this.localDescription = d
  }
  async addIceCandidate(c: unknown) {
    this.candidates.push(c)
  }
  getSenders() {
    return []
  }
  close() {
    this.closed = true
  }
  // what the browser's ICE agent would report
  transition(s: string) {
    this.connectionState = s
    this.onconnectionstatechange?.()
  }
}
g.RTCPeerConnection = FakePeer
const fetches: { url: string; init: RequestInit }[] = []
let respond: (url: string, init: RequestInit) => { status: number; body?: unknown } = () => ({ status: 200 })
g.fetch = async (url: string, init: RequestInit) => {
  fetches.push({ url, init })
  const { status, body } = respond(url, init)
  const text = body === undefined ? '' : JSON.stringify(body)
  return {
    ok: status < 300,
    status,
    statusText: 'status ' + status,
    json: async () => JSON.parse(text),
    text: async () => text,
  }
}

const { NekoClient, ApiError } = await import('./index.ts')
type Client = InstanceType<typeof NekoClient>
type Options = ConstructorParameters<typeof NekoClient>[0]

// server payloads
const profile = (over = {}) => ({
  name: 'Alice',
  is_admin: false,
  can_login: true,
  can_connect: true,
  can_watch: true,
  can_host: true,
  can_share_media: false,
  can_access_clipboard: true,
  sends_inactive_cursor: false,
  can_see_inactive_cursors: false,
  ...over,
})
const session = (id: string, over = {}) => ({
  id,
  profile: profile(over),
  state: { is_connected: true, is_watching: false },
})
const settings = (over = {}) => ({
  private_mode: false,
  locked_logins: false,
  locked_controls: false,
  control_protection: false,
  implicit_hosting: true,
  inactive_cursors: false,
  merciful_reconnect: true,
  ...over,
})
const init = () => ({
  session_id: 's1',
  sessions: { s1: session('s1'), s2: session('s2', { name: 'Bob' }) },
  settings: settings(),
  screen_size: { width: 1920, height: 1080, rate: 60 },
  control_host: { has_host: true, host_id: 's2' },
  touch_events: true,
})
const receive = (event: string, payload?: unknown, ws = sockets.at(-1)!) =>
  ws.onmessage!({ data: JSON.stringify({ event, payload }) })

const setup = (opts?: Options) => {
  sent.length = sockets.length = fetches.length = FakePeer.all.length = 0
  vi.clearAllTimers()
  storage.clear()
  respond = () => ({ status: 200 })
  FakePeer.answerSdp = ANSWER
  const client = new NekoClient(opts)
  const closed: (Error | undefined)[] = [] // every connection.closed, with its error
  client.events.on('connection.closed', (e) => closed.push(e))
  return { client, closed }
}
// a client through the handshake, with `sent` cleared
const connected = async (opts?: Options) => {
  const r = setup(opts)
  r.client.state.authenticated = true
  r.client.connect()
  receive('system/init', init())
  receive('signal/provide', { sdp: OFFER })
  await settle()
  FakePeer.all[0].transition('connected')
  await settle()
  assert.equal(r.client.state.connection.status, 'connected')
  sent.length = 0
  return r
}
const messages = (closed: (Error | undefined)[]) => closed.map((e) => e?.message)

test('handshake: init -> signal/request -> provide -> signal/answer -> peer connected', async () => {
  const { client } = setup()
  client.setUrl('http://neko.test/') // as the GUI does at boot
  assert.throws(() => client.connect(), /not authenticated/)
  client.state.authenticated = true
  client.state.connection.token = 'tok'
  const statuses: string[] = []
  client.events.on('connection.status', (s) => statuses.push(s))
  client.connect()
  client.connect() // already wanted: not a second socket
  assert.equal(sockets.length, 1)
  assert.equal(sockets[0].url, 'ws://neko.test/api/ws?token=tok')
  assert.equal(client.state.connection.status, 'connecting')

  receive('system/init', init())
  assert.deepEqual(sent, ['signal/request'])
  assert.equal(client.state.session_id, 's1')
  assert.deepEqual(client.state.screen.size, { width: 1920, height: 1080, rate: 60 })
  assert.deepEqual(Object.keys(client.state.sessions), ['s1', 's2'])
  assert.equal(client.session?.profile.name, 'Alice')
  assert.equal(client.state.control.host_id, 's2')
  assert.equal(client.state.control.touch, true)
  assert.equal(client.implicitControl, true)

  receive('signal/candidate', { candidate: 'early' }) // trickled before the offer: held back
  receive('signal/provide', { sdp: OFFER, iceservers: [] })
  await settle()
  assert.deepEqual(sent, ['signal/request', 'signal/answer'])
  const pc = FakePeer.all[0]
  assert.deepEqual(pc.remoteDescription, { type: 'offer', sdp: OFFER })
  assert.deepEqual(pc.candidates, [{ candidate: 'early' }])
  receive('signal/candidate', { candidate: 'late' })
  assert.deepEqual(pc.candidates, [{ candidate: 'early' }, { candidate: 'late' }])
  assert.equal(client.state.connection.status, 'connecting', 'not connected before the media path is')

  pc.transition('connected')
  await settle()
  assert.equal(client.state.connection.status, 'connected')
  assert.equal(client.connected, true)
  assert.deepEqual(statuses, ['connecting', 'connected'])
})

test('signal/close: the peer is dropped and a new one requested after the backoff', async () => {
  const { client } = await connected()
  const pc = FakePeer.all[0]
  receive('signal/close')
  assert.ok(pc.closed)
  assert.equal(client.state.connection.status, 'connecting')
  receive('signal/close') // no peer any more: nothing to drop, no second request
  await tick(1499)
  assert.deepEqual(sent, [], 'waits RECONNECT_BACKOFF_MS before asking again')
  await tick(1)
  assert.deepEqual(sent, ['signal/request'])
  receive('signal/provide', { sdp: OFFER })
  await settle()
  assert.equal(FakePeer.all.length, 2)
  assert.deepEqual(sent, ['signal/request', 'signal/answer'])
  assert.doesNotThrow(() => client.removeTrack({} as RTCRtpSender), 'a sender of the dropped peer is ignored')
  FakePeer.all[1].transition('connected')
  await settle()
  assert.equal(client.state.connection.status, 'connected')
})

test('peer failed: re-requested with growing backoff, given up after RECONNECT_MAX', async () => {
  const { client, closed } = await connected()
  FakePeer.all[0].transition('disconnected') // ICE may still recover
  await settle()
  assert.equal(client.state.connection.status, 'connecting')
  assert.equal(FakePeer.all[0].closed, false)
  for (let n = 1; n <= 10; n++) {
    sent.length = 0
    FakePeer.all.at(-1)!.transition('failed')
    assert.ok(FakePeer.all.at(-1)!.closed)
    const backoff = 1500 * Math.min(n, 4)
    await tick(backoff - 1)
    assert.deepEqual(sent, [], `failure ${n}: waits ${backoff} ms`)
    await tick(1)
    assert.deepEqual(sent, ['signal/request'], `failure ${n}`)
    receive('signal/provide', { sdp: OFFER })
    await settle()
  }
  assert.deepEqual(closed, [])
  FakePeer.all.at(-1)!.transition('failed')
  await settle()
  assert.deepEqual(messages(closed), ['video connection failed (WebRTC)'])
  assert.equal(client.state.connection.status, 'disconnected')
  assert.equal(sockets[0].code, 1000)
  await tick(60_000)
  assert.equal(sockets.length, 1, 'no reconnect after giving up')
})

test('peer connected: the failure count starts over', async () => {
  await connected()
  for (const backoff of [1500, 3000]) {
    FakePeer.all.at(-1)!.transition('failed')
    await tick(backoff)
    receive('signal/provide', { sdp: OFFER })
    await settle()
  }
  FakePeer.all.at(-1)!.transition('connected')
  await settle()
  sent.length = 0
  FakePeer.all.at(-1)!.transition('failed')
  await tick(1500)
  assert.deepEqual(sent, ['signal/request'])
})

test('no offer: shown without video only while the server keeps talking', async () => {
  const { client } = await connected()
  await tick(10_000) // silence since the offer (the stale check tolerates 25 s)
  FakePeer.all[0].transition('failed')
  await tick(1500) // -> signal/request, 11.5 s into the silence
  assert.deepEqual(sent, ['signal/request'])
  await tick(8000) // OFFER_TIMEOUT_MS: 19.5 s of silence
  assert.equal(client.state.connection.status, 'connecting', 'a silent server is not "connected" without video')
  receive('system/heartbeat') // alive after all (can_watch is off)
  await tick(8000)
  assert.equal(client.state.connection.status, 'connected')
  assert.equal(FakePeer.all.length, 1)
})

test('stale socket: silent for STALE_MS it is replaced; a message keeps it', async () => {
  const { client } = await connected()
  await tick(20_000)
  receive('system/heartbeat')
  await tick(20_000)
  assert.equal(sockets[0].closed, false, '20 s since the last message is not stale')
  assert.equal(sockets.length, 1)
  await tick(10_000) // 30 s since
  const old = sockets[0]
  assert.ok(old.closed)
  assert.ok(FakePeer.all[0].closed)
  assert.equal(client.state.connection.status, 'connecting')
  assert.equal(fetches.at(-1)?.url, 'http://neko.test/api/whoami')
  await tick(1500)
  assert.equal(sockets.length, 2)
  receive('system/init', init()) // the handshake starts over on the new socket
  assert.deepEqual(sent, ['client/heartbeat', 'signal/request'])
  old.onclose!() // the browser reports the closure of the replaced socket: not a second loss
  await tick(20_000)
  assert.deepEqual([sockets.length, fetches.length], [2, 1])
})

test('socket closed: reconnects after the backoff once the session is confirmed', async () => {
  const { client } = await connected()
  sockets[0].onclose!()
  assert.equal(client.state.connection.status, 'connecting')
  assert.ok(FakePeer.all[0].closed)
  await settle()
  assert.equal(fetches.at(-1)?.url, 'http://neko.test/api/whoami')
  await tick(1499)
  assert.equal(sockets.length, 1)
  await tick(1)
  assert.equal(sockets.length, 2)
  sockets[1].onclose!() // a second loss in a row waits longer
  await tick(2999)
  assert.equal(sockets.length, 2)
  await tick(1)
  assert.equal(sockets.length, 3)
  respond = () => {
    throw new TypeError('network') // a lookup that fails for any other reason is still worth a retry
  }
  sockets[2].onclose!()
  await tick(4500)
  assert.equal(sockets.length, 4)
  respond = () => ({ status: 200 })
  receive('system/init', init())
  receive('signal/provide', { sdp: OFFER })
  await settle()
  FakePeer.all.at(-1)!.transition('connected') // the attempt count starts over
  await settle()
  sockets[3].onclose!()
  await tick(1500)
  assert.equal(sockets.length, 5)
})

test('socket closed: a deleted session (401) is not retried', async () => {
  const { client, closed } = await connected()
  respond = () => ({ status: 401, body: { message: 'session not found' } })
  sockets[0].onclose!()
  await settle()
  assert.deepEqual(messages(closed), ['session expired'])
  assert.equal(client.state.connection.status, 'disconnected')
  await tick(60_000)
  assert.equal(sockets.length, 1)
})

test('socket closed: given up after RECONNECT_MAX attempts', async () => {
  const { client, closed } = await connected()
  for (let n = 1; n <= 10; n++) {
    sockets.at(-1)!.onclose!()
    await tick(6000)
    assert.equal(sockets.length, n + 1, `attempt ${n}`)
  }
  sockets.at(-1)!.onclose!()
  await tick(60_000)
  assert.equal(sockets.length, 11)
  assert.deepEqual(messages(closed), ['connection lost'])
  assert.equal(client.state.connection.status, 'disconnected')
})

test('disconnect: closes socket and peer, clears the room state, cancels every timer', async () => {
  const { client, closed } = await connected()
  client.disconnect()
  assert.equal(sockets[0].code, 1000, 'a normal closure, or the server keeps the session for 5 s')
  assert.ok(FakePeer.all[0].closed)
  assert.equal(client.state.connection.status, 'disconnected')
  assert.equal(client.state.session_id, null)
  assert.deepEqual(client.state.sessions, {})
  assert.equal(client.state.control.host_id, null)
  assert.deepEqual(closed, [undefined])
  assert.equal(vi.getTimerCount(), 0)
  sockets[0].onclose!() // the server's side of the closure
  await tick(60_000)
  assert.deepEqual([sockets.length, fetches.length], [1, 0])

  client.connect() // and connecting again works
  assert.equal(sockets.length, 2)
  sockets[1].onclose!()
  await settle() // the session lookup came back: a reconnect is scheduled
  client.disconnect()
  await tick(60_000)
  assert.equal(sockets.length, 2, 'a disconnect cancels the pending reconnect')

  client.connect()
  sockets[2].onclose!()
  client.disconnect() // while the lookup is still running
  await tick(60_000)
  assert.equal(sockets.length, 3)

  client.connect()
  sockets[3].onclose!()
  client.disconnect()
  client.connect() // reconnected by hand while the old lookup is still running
  assert.equal(sockets.length, 5)
  await tick(10_000) // the old lookup comes back and would reconnect after 1.5 s
  assert.equal(sockets.length, 5, 'the old lookup must not replace the new connection')
  assert.equal(sockets[4].closed, false)
  assert.deepEqual(closed, [undefined, undefined, undefined, undefined])
})

test('system/heartbeat is answered with client/heartbeat', async () => {
  await connected()
  receive('system/heartbeat')
  assert.deepEqual(sent, ['client/heartbeat'])
})

test('codec mismatch: the offer is refused naming the codec, no retry', async () => {
  const { client, closed } = setup()
  client.state.authenticated = true
  client.connect()
  receive('system/init', init())
  FakePeer.answerSdp = 'v=0\r\nm=audio 9 RTP/AVP 0\r\nm=video 0 RTP/AVP 96\r\n' // the browser rejected the video section
  receive('signal/provide', { sdp: OFFER.replace('VP8', 'H264') })
  await settle()
  assert.deepEqual(sent, ['signal/request'], 'no answer')
  assert.deepEqual(messages(closed), ['this browser cannot play H264 video'])
  assert.equal(client.state.connection.status, 'disconnected')
  assert.equal(sockets[0].code, 1000)
  await tick(60_000)
  assert.deepEqual([sockets.length, FakePeer.all.length], [1, 1])
})

test('system/disconnect (the session opened in another tab): closed with the message, no reconnect', async () => {
  const { client, closed } = await connected()
  receive('system/disconnect', { message: 'connection replaced' })
  assert.deepEqual(messages(closed), ['connection replaced'])
  assert.equal(client.state.connection.status, 'disconnected')
  assert.equal(client.state.session_id, null)
  assert.equal(sockets[0].code, 1000)
  assert.ok(FakePeer.all[0].closed)
  sockets[0].onclose!() // the server closes the socket right after
  await tick(60_000)
  assert.deepEqual([sockets.length, fetches.length], [1, 0])
})

test('session, control, screen, clipboard and settings events update the state', async () => {
  const { client } = await connected()
  const got: string[] = []
  const names = [
    'session.created',
    'session.updated',
    'session.deleted',
    'room.control.host',
    'room.screen.updated',
    'room.clipboard.updated',
    'room.settings.updated',
    'room.broadcast.status',
    'receive.unicast',
    'receive.broadcast',
    'message',
  ] as const
  for (const n of names) client.events.on(n, (...a: unknown[]) => got.push(n + ' ' + JSON.stringify(a)))

  receive('session/created', session('s3', { name: 'Carol' }))
  assert.equal(client.state.sessions.s3.profile.name, 'Carol')
  receive('session/profile', { id: 's3', ...profile({ name: 'Caroline' }) })
  assert.equal(client.state.sessions.s3.profile.name, 'Caroline')
  receive('session/state', { id: 's3', is_connected: true, is_watching: true })
  assert.equal(client.state.sessions.s3.state.is_watching, true)
  receive('session/profile', { id: 'nobody', ...profile() }) // unknown session: ignored
  receive('session/state', { id: 'nobody', is_connected: true, is_watching: true })
  assert.equal('nobody' in client.state.sessions, false)
  receive('session/deleted', { id: 's3' })
  assert.equal('s3' in client.state.sessions, false)
  receive('session/profile', { id: 's1', ...profile({ is_admin: true }) }) // our own profile
  assert.equal(client.isAdmin, true)

  assert.equal(client.controlling, false)
  receive('control/host', { id: 's1', has_host: true, host_id: 's1' })
  assert.equal(client.state.control.host_id, 's1')
  assert.equal(client.controlling, true)
  await settle()
  assert.deepEqual(sent, ['keyboard/map'], 'the keyboard layout is sent on taking control')
  receive('control/host', { id: 's1', has_host: false })
  assert.equal(client.state.control.host_id, null)
  receive('control/request', { id: 's2' })

  receive('screen/updated', { id: 's2', width: 1280, height: 720, rate: 30 })
  assert.deepEqual(client.state.screen.size, { width: 1280, height: 720, rate: 30 })
  receive('clipboard/updated', { text: 'hi' })
  assert.deepEqual(client.state.control.clipboard, { text: 'hi' })
  receive('system/settings', { id: 's2', ...settings({ private_mode: true }) })
  assert.equal(client.state.settings.private_mode, true)
  assert.equal(client.privateModeEnabled, false, 'admins see through private mode')
  receive('system/admin', {
    screen_sizes_list: [
      { width: 1280, height: 720, rate: 30 },
      { width: 1920, height: 1080, rate: 30 },
      { width: 1920, height: 1080, rate: 60 },
    ],
    broadcast_status: { is_active: true, url: 'rtmp://x' },
  })
  assert.deepEqual(
    client.state.screen.configurations.map((c) => `${c.width}x${c.height}@${c.rate}`),
    ['1920x1080@60', '1920x1080@30', '1280x720@30'],
  )
  receive('send/unicast', { sender: 's2', subject: 'hey', body: 1 })
  receive('send/broadcast', { sender: 's2', subject: 'all', body: [2] })
  receive('chat/message', { text: 'x' }) // a plugin event
  assert.deepEqual(got, [
    'session.created ["s3"]',
    'session.updated ["s3"]',
    'session.updated ["s3"]',
    'session.deleted ["s3"]',
    'session.updated ["s1"]',
    'room.control.host [true,"s1","s1"]',
    'room.control.host [false,null,"s1"]',
    'room.screen.updated [1280,720,30,"s2"]',
    'room.clipboard.updated ["hi"]',
    `room.settings.updated [${JSON.stringify(settings({ private_mode: true }))},"s2"]`,
    'room.broadcast.status [true,"rtmp://x"]',
    'receive.unicast ["s2","hey",1]',
    'receive.broadcast ["s2","all",[2]]',
    'message ["chat/message",{"text":"x"}]',
  ])
})

test('preparePaste: the side that copied most recently wins', async () => {
  const { client } = await connected()
  receive('clipboard/updated', { text: 'remote' })
  for (const text of ['remote', '']) {
    const p = client.preparePaste(text) // the remote clipboard's own text, or nothing: nothing to do
    await settle()
    assert.deepEqual(sent, [], `"${text}" is not sent`)
    await p
  }

  let settled = false
  const p = client.preparePaste('local').then(() => (settled = true)) // newer than anything the remote reported
  await settle()
  assert.deepEqual([sent, settled], [['clipboard/set'], false])
  receive('clipboard/updated', { text: 'local' }) // resolves once the server confirms
  await p

  await tick(1)
  receive('clipboard/updated', { text: 'remote2' }) // the remote copied since: the local text is older
  sent.length = 0
  const older = client.preparePaste('local')
  await settle()
  assert.deepEqual(sent, [])
  await older
  const same = client.preparePaste('remote2') // on https the remote text is in the local clipboard too: nothing to send
  await settle()
  assert.deepEqual(sent, [])
  await same

  settled = false
  client.preparePaste('other').then(() => (settled = true)) // nothing changed remotely since that look
  await tick(499)
  assert.deepEqual([sent, settled], [['clipboard/set'], false])
  await tick(1) // no confirmation: the wait ends after 500 ms
  assert.equal(settled, true)
})

test('login, authenticate, logout: the token and the session', async () => {
  const { client } = setup({ autologin: true })
  client.setUrl('http://neko.test/')
  respond = (url, init) =>
    url.endsWith('/api/login') && JSON.parse(init.body as string).password === 'pw'
      ? { status: 200, body: { ...session('s1'), token: 'tok' } }
      : { status: 401, body: { message: 'invalid password' } }
  await assert.rejects(
    client.login('admin', 'wrong'),
    (e: unknown) => e instanceof ApiError && e.status === 401 && e.message === 'invalid password',
  )
  assert.equal(client.state.authenticated, false)
  await client.login('admin', 'pw')
  assert.equal(fetches.at(-1)?.url, 'http://neko.test/api/login')
  assert.equal(client.state.authenticated, true)
  assert.equal(client.api.token, 'tok')
  assert.equal(storage.get('neko_session'), 'tok')

  respond = () => ({ status: 200, body: session('s1') })
  await client.authenticate() // whoami carries the token
  assert.equal(fetches.at(-1)?.url, 'http://neko.test/api/whoami')
  assert.equal((fetches.at(-1)?.init.headers as Record<string, string>).Authorization, 'Bearer tok')
  client.connect()
  assert.equal(sockets[0].url, 'ws://neko.test/api/ws?token=tok')

  await client.logout()
  assert.equal(fetches.at(-1)?.url, 'http://neko.test/api/logout')
  assert.equal(client.state.authenticated, false)
  assert.equal(client.api.token, '')
  assert.equal(storage.has('neko_session'), false)
  assert.equal(sockets[0].code, 1000, 'logging out closes the connection')
  assert.equal(client.state.connection.status, 'disconnected')
})

test('authenticate: resumes the saved session, or fails with the ApiError', async () => {
  const { client } = setup({ autologin: true, autoconnect: true })
  storage.set('neko_session', 'saved')
  client.setUrl('http://neko.test/some/path/')
  await settle()
  assert.equal(fetches[0].url, 'http://neko.test/some/path/api/whoami')
  assert.equal((fetches[0].init.headers as Record<string, string>).Authorization, 'Bearer saved')
  assert.equal(client.state.authenticated, true)
  assert.equal(sockets[0]?.url, 'ws://neko.test/some/path/api/ws?token=saved', 'autoconnect')

  const expired = setup({ autologin: true }).client
  storage.set('neko_session', 'expired')
  respond = () => ({ status: 401, body: { message: 'unauthorized' } })
  await assert.rejects(expired.authenticate(), (e: unknown) => e instanceof ApiError && e.status === 401)
  assert.equal(expired.state.authenticated, false)
  assert.throws(() => expired.connect(), /not authenticated/)
})

test('send: dropped while the socket is not open', async () => {
  const { client } = setup()
  client.send('chat/message', { text: 'x' }) // no socket at all
  client.state.authenticated = true
  client.connect()
  sockets[0].readyState = 0 // still connecting
  client.send('chat/message', { text: 'x' })
  assert.deepEqual(sent, [])
  sockets[0].readyState = 1
  client.send('chat/message', { text: 'x' })
  assert.deepEqual(sent, ['chat/message'])
  client.disconnect()
  client.send('chat/message', { text: 'x' })
  assert.deepEqual(sent, ['chat/message'])
})

test('sendData: big-endian input messages on the data channel, dropped while it is not open', async () => {
  const { client } = await connected()
  const frames: number[][] = []
  const dc = { readyState: 'connecting', binaryType: '', send: (b: ArrayBuffer) => frames.push([...new Uint8Array(b)]) }
  ;(FakePeer.all[0] as any).ondatachannel({ channel: dc })
  client.sendData(1, [2, 100], [2, 200])
  assert.deepEqual(frames, [])
  dc.readyState = 'open'
  client.sendData(1, [2, 100], [2, 200]) // MOVE x=100 y=200
  client.sendData(2, [-2, -1], [-2, 1]) // SCROLL
  client.sendData(3, [4, 0xffff], [1, 1]) // KEY_DOWN with a u32 keysym, u8 flag
  assert.deepEqual(frames, [
    [1, 0, 4, 0, 100, 0, 200],
    [2, 0, 4, 255, 255, 0, 1],
    [3, 0, 5, 0, 0, 255, 255, 1],
  ])
})
