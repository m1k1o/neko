// The seam between NekoClient and a StreamTransport, with a fake transport: what the client hands
// over, how the connection status follows the transport, and the media calls it delegates.
import { test } from 'vitest'
import assert from 'node:assert/strict'
import { tick, settle, reset, sent, sentPayloads, sockets, init, receive, FakeElement } from './test/browser.ts'

const { NekoClient, Emitter, WebSocketInput } = await import('./index.ts')
type Options = ConstructorParameters<typeof NekoClient>[0]
type Transport = NonNullable<Exclude<NonNullable<Options>['transport'], () => unknown>>
type SessionInfo = Parameters<Transport['connect']>[0]
type TransportState = Parameters<Parameters<Transport['on']>[1]>[0]

class FakeTransport implements Transport {
  readonly kind = 'ws-mse' as const
  element = null
  input?: Transport['input']
  shareMedia?: Transport['shareMedia']
  calls: string[] = []
  session: SessionInfo | null = null
  startError: Error | null = null
  readonly events = new Emitter<any>()
  video = { playable: false, playing: false, volume: 1, muted: false, mutedByAutoplay: false }
  async connect(session: SessionInfo) {
    this.calls.push('connect')
    this.session = session
    if (this.startError) throw this.startError
  }
  close() {
    this.calls.push('close')
  }
  attach() {
    this.calls.push('attach')
    return () => this.calls.push('detach')
  }
  async setPlaying(on: boolean) {
    this.calls.push(`setPlaying ${on}`)
  }
  setVolume(v: number) {
    this.calls.push(`setVolume ${v}`)
  }
  setMuted(on: boolean) {
    this.calls.push(`setMuted ${on}`)
  }
  on = (event: string, cb: (...a: any[]) => void) => this.events.on(event, cb)
  // what the transport reports
  state(status: TransportState['status'], video = this.video) {
    this.events.emit('state', { status, size: { width: 0, height: 0 }, video })
  }
  fail(err: Error) {
    this.events.emit('error', err)
  }
}

const setup = (opts: Options = {}) => {
  reset()
  const t = new FakeTransport()
  const client = new NekoClient({ transport: t, ...opts })
  client.setUrl('http://neko.test/') // closes whatever was there, as always
  t.calls.length = 0
  const closed: (Error | undefined)[] = []
  client.events.on('connection.closed', (e) => closed.push(e))
  const statuses: string[] = []
  client.events.on('connection.status', (s) => statuses.push(s))
  client.state.authenticated = true
  client.state.connection.token = 'tok'
  return { t, client, closed, statuses }
}

test('connect: the transport gets the session (url, token, init, send, on) and nothing else', async () => {
  const { t, client } = setup()
  client.connect()
  assert.deepEqual(t.calls, [], 'not before the server introduced the session')
  receive('system/init', init())
  assert.deepEqual(t.calls, ['connect'])
  const s = t.session!
  assert.deepEqual(Object.keys(s).sort(), ['init', 'on', 'send', 'token', 'url'])
  assert.equal(s.url, 'http://neko.test')
  assert.equal(s.token, 'tok')
  assert.deepEqual(s.init, init())
  s.send('signal/request', { video: {} })
  assert.deepEqual(sentPayloads.at(-1), { event: 'signal/request', payload: { video: {} } })

  const got: unknown[] = []
  const messages: string[] = []
  client.events.on('message', (e) => messages.push(e))
  const off = s.on('signal/provide', (p) => got.push(p))
  receive('signal/provide', { sdp: 'x' })
  receive('signal/video', { video: 'hd' })
  assert.deepEqual(got, [{ sdp: 'x' }])
  assert.deepEqual(messages, [], 'signal/* is the transport’s, not a client message')
  receive('chat/message', { text: 'hi' })
  assert.deepEqual(messages, ['chat/message'])
  off()
  receive('signal/provide', { sdp: 'y' })
  assert.equal(got.length, 1)
})

test('status: connected only when the socket and the transport are; a transport error closes', async () => {
  const { t, client, closed, statuses } = setup()
  client.connect()
  receive('system/init', init())
  assert.equal(client.state.connection.status, 'connecting')
  t.state('connecting')
  assert.equal(client.state.connection.status, 'connecting')
  t.state('connected')
  assert.equal(client.state.connection.status, 'connected')
  await settle()
  t.state('connecting') // the media path hiccups
  assert.equal(client.state.connection.status, 'connecting')
  await settle()
  t.state('unavailable') // the server will not stream: the room without video
  assert.equal(client.state.connection.status, 'connected')
  await settle()

  sockets[0].onclose!()
  assert.deepEqual(t.calls, ['connect', 'close'], 'the lost socket closes the transport')
  assert.equal(client.state.connection.status, 'connecting')
  t.state('disconnected')
  assert.equal(client.state.connection.status, 'connecting')
  await tick(1500)
  receive('system/init', init(), sockets[1])
  assert.deepEqual(t.calls, ['connect', 'close', 'connect'], 'the new socket reconnects the transport')
  t.state('connected')
  await settle()
  // connect() and the first 'connected' fell into one store flush
  assert.deepEqual(statuses, ['connected', 'connecting', 'connected', 'connecting', 'connected'])

  t.fail(new Error('video connection failed (WebRTC)'))
  assert.deepEqual(
    closed.map((e) => e?.message),
    ['video connection failed (WebRTC)'],
  )
  assert.equal(client.state.connection.status, 'disconnected')
  assert.equal(sockets[1].code, 1000)
  assert.deepEqual(t.calls.at(-1), 'close')
  t.state('connected') // a late report from a closed transport changes nothing
  assert.equal(client.state.connection.status, 'disconnected')
  await tick(60_000)
  assert.equal(sockets.length, 2, 'no reconnect after an error')
})

test('a transport that cannot start closes the connection with its error', async () => {
  const { t, client, closed } = setup()
  t.startError = new Error('WebCodecs not supported')
  client.connect()
  receive('system/init', init())
  await settle()
  assert.deepEqual(
    closed.map((e) => e?.message),
    ['WebCodecs not supported'],
  )
  assert.equal(client.state.connection.status, 'disconnected')
})

test('socket attempts start over when the transport connects, not when the server declines media', async () => {
  const { t, client } = setup()
  client.connect()
  for (let n = 1; n <= 10; n++) {
    sockets.at(-1)!.onclose!()
    await tick(6000)
    assert.equal(sockets.length, n + 1, `attempt ${n}`)
  }
  receive('system/init', init())
  t.state('unavailable')
  assert.equal(client.state.connection.status, 'connected')
  sockets.at(-1)!.onclose!()
  await tick(60_000)
  assert.equal(client.state.connection.status, 'disconnected', 'the 11th loss: given up, as without a transport')

  const second = setup()
  second.client.connect()
  for (let n = 1; n <= 10; n++) {
    sockets.at(-1)!.onclose!()
    await tick(6000)
  }
  receive('system/init', init())
  second.t.state('connected')
  sockets.at(-1)!.onclose!()
  await tick(6000)
  assert.equal(sockets.length, 12, 'a connected transport reset the count')
})

test('disconnect closes the transport; the overlay channel is the transport’s input or the websocket', async () => {
  const { t, client } = setup()
  client.connect()
  receive('system/init', init())
  client.disconnect()
  assert.deepEqual(t.calls, ['connect', 'close'])
  assert.ok(client.input instanceof WebSocketInput)
  client.connect()
  receive('system/init', init())
  assert.deepEqual(t.calls, ['connect', 'close', 'connect'])

  reset()
  const own = new FakeTransport()
  const keys: [number, boolean][] = []
  own.input = { move() {}, scroll() {}, button() {}, key: (keysym, down) => keys.push([keysym, down]), touch() {} }
  const c2 = new NekoClient({ transport: () => own })
  assert.equal(c2.input, own.input, 'the transport’s own channel wins')
  assert.equal(c2.transport, own, 'a factory is called once')
  c2.input.key(0xff0d, true)
  assert.deepEqual(keys, [[0xff0d, true]])
  assert.deepEqual(sent, [])
})

test('play, pause, mute, unmute, volume and shareMedia go through the transport; the video state mirrors it', async () => {
  const { t, client } = setup()
  await client.play()
  client.pause()
  client.mute()
  client.unmute()
  client.setVolume(1.5)
  client.setVolume(-1)
  client.setVolume(0.3)
  assert.deepEqual(t.calls, [
    'setPlaying true',
    'setPlaying false',
    'setMuted true',
    'setMuted false',
    'setVolume 1',
    'setVolume 0',
    'setVolume 0.3',
  ])
  const stream = {} as MediaStream
  assert.throws(() => client.shareMedia(stream), /transport cannot send media/)
  t.shareMedia = (s) => {
    t.calls.push(`shareMedia ${s === stream}`)
    return () => t.calls.push('stop')
  }
  client.shareMedia(stream)()
  assert.deepEqual(t.calls.slice(7), ['shareMedia true', 'stop'])

  t.state('disconnected', { playable: true, playing: true, volume: 0.3, muted: true, mutedByAutoplay: true })
  assert.deepEqual(client.state.video, {
    playable: true,
    playing: true,
    volume: 0.3,
    muted: true,
    mutedByAutoplay: true,
  })
  t.state('disconnected', { playable: true, playing: false, volume: 0.3, muted: false, mutedByAutoplay: false })
  assert.deepEqual(client.state.video, {
    playable: true,
    playing: false,
    volume: 0.3,
    muted: false,
    mutedByAutoplay: false,
  })
})

test('autoplay: the client starts playback whenever the transport’s stream becomes playable', async () => {
  const { t, client } = setup({ autoplay: true })
  client.connect()
  receive('system/init', init())
  const video = { playable: false, playing: false, volume: 1, muted: false, mutedByAutoplay: false }
  t.state('connected', video)
  assert.deepEqual(t.calls, ['connect'], 'not before a frame can be shown')
  t.state('connected', { ...video, playable: true })
  assert.deepEqual(t.calls, ['connect', 'setPlaying true'])
  t.state('connected', { ...video, playable: true, playing: true }) // still playable: not again
  t.state('connected', { ...video, playable: true, playing: false }) // paused by the user: not again
  assert.deepEqual(t.calls, ['connect', 'setPlaying true'])
  t.state('connected', video) // a new stream is loading (reconnect)...
  t.state('connected', { ...video, playable: true })
  assert.deepEqual(t.calls, ['connect', 'setPlaying true', 'setPlaying true'])

  const off = setup()
  off.t.state('connected', { ...video, playable: true })
  assert.deepEqual(off.t.calls, [], 'without the option the user presses play')
})

test('socket lost: playback that was running resumes on the new stream, autoplay or not', async () => {
  const { t, client } = setup()
  const video = { playable: true, playing: true, volume: 1, muted: false, mutedByAutoplay: false }
  const loading = { ...video, playable: false, playing: false }
  client.connect()
  receive('system/init', init())
  t.state('connected', video)
  sockets[0].onclose!()
  t.state('disconnected', loading) // close(): the picture stays, nothing plays
  await tick(1500)
  receive('system/init', init(), sockets[1])
  t.state('connected', loading)
  t.state('connected', { ...video, playing: false })
  assert.deepEqual(t.calls, ['connect', 'close', 'connect', 'setPlaying true'])
  t.state('connected', video)

  sockets[1].onclose!() // and again
  t.state('disconnected', loading)
  await tick(1500)
  receive('system/init', init(), sockets[2])
  t.state('connected', loading)
  t.state('connected', { ...video, playing: false })
  assert.deepEqual(t.calls.slice(4), ['close', 'connect', 'setPlaying true'], 'the first stream after the loss')
  t.state('connected', loading) // a replaced peer brings another stream: the transport's business
  t.state('connected', { ...video, playing: false })
  assert.deepEqual(t.calls.slice(7), [], 'a later stream does not')

  t.state('connected', video)
  sockets[2].onclose!() // lost while playing...
  t.state('disconnected', loading)
  client.disconnect() // ...and given up on before the stream came back: the next one waits for play (or autoplay)
  client.connect()
  receive('system/init', init(), sockets[3])
  t.state('connected', loading)
  t.state('connected', { ...video, playing: false })
  assert.deepEqual(t.calls.slice(7), ['close', 'close', 'connect'])
})

test('mount: the transport draws into the client’s container; unmount removes it again', () => {
  const { t, client } = setup()
  const el = new FakeElement('div')
  client.mount(el as any)
  assert.deepEqual(t.calls, ['attach'])
  assert.equal(el.children.length, 1, 'the letterboxed container')
  assert.throws(() => client.mount(el as any), /already mounted/)
  client.unmount()
  assert.deepEqual(t.calls, ['attach', 'detach'])
  assert.deepEqual(el.children, [])
  client.mount(el as any) // and again
  assert.deepEqual(t.calls, ['attach', 'detach', 'attach'])
  client.unmount()
})
