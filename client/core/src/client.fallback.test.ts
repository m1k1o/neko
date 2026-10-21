// The fallback transport: when the first transport gives up on the stream, the client switches to
// the fallback in place instead of ending the session. The point is a user whose WebRTC cannot
// connect still gets video, without having to reach a setting they cannot reach while stuck.
import { test } from 'vitest'
import assert from 'node:assert/strict'
import { reset, init, receive, FakeElement } from './test/browser.ts'

const { NekoClient, Emitter, WebSocketInput } = await import('./index.ts')
type Options = ConstructorParameters<typeof NekoClient>[0]
type Transport = NonNullable<Exclude<NonNullable<Options>['transport'], () => unknown>>
type SessionInfo = Parameters<Transport['connect']>[0]

class FakeTransport implements Transport {
  element = null
  input?: Transport['input']
  calls: string[] = []
  session: SessionInfo | null = null
  readonly events = new Emitter<any>()
  constructor(readonly kind: Transport['kind']) {}
  async connect(session: SessionInfo) {
    this.calls.push('connect')
    this.session = session
  }
  close() {
    this.calls.push('close')
  }
  attach(box: HTMLElement) {
    ;(box as unknown as FakeElement).append(new FakeElement('video'))
    return () => {}
  }
  async setPlaying() {}
  setVolume() {}
  setMuted() {}
  on = (event: string, cb: (...a: any[]) => void) => this.events.on(event, cb)
  fail(err: Error) {
    this.events.emit('error', err)
  }
}

// a connected session whose first transport is `first`, falling back to `fallback` if given.
// `firstInput` gives the first transport its own channel, as WebRTC's data channel is, so the
// client binds input to it at construction.
const setup = (fallback?: () => FakeTransport, firstInput?: Transport['input']) => {
  reset()
  const first = new FakeTransport('webrtc')
  first.input = firstInput
  const client = new NekoClient({ transport: first, fallbackTransport: fallback })
  client.setUrl('http://neko.test/')
  first.calls.length = 0
  const closed: (Error | undefined)[] = []
  client.events.on('connection.closed', (e) => closed.push(e))
  const changed: string[] = []
  client.events.on('transport.changed', (kind) => changed.push(kind))
  client.store.setState({ authenticated: true, connection: { ...client.state.connection, token: 'tok' } })
  client.connect()
  receive('system/init', init())
  return { client, first, closed, changed }
}

test('a stream failure switches to the fallback instead of closing the session', () => {
  const second = new FakeTransport('webcodecs-ws')
  const { client, first, closed, changed } = setup(() => second)

  first.fail(new Error('video connection failed (WebRTC)'))

  assert.deepEqual(closed, [], 'the session must survive: the fallback takes over the stream')
  assert.equal(client.transport, second, 'the client now streams through the fallback')
  assert.ok(first.calls.includes('close'), 'the failed transport is closed')
  assert.deepEqual(changed, ['webcodecs-ws'], 'the GUI is told so it attaches the new element')
})

test('the fallback starts at once with the session it already has, without a reconnect', () => {
  const second = new FakeTransport('webcodecs-ws')
  const { first } = setup(() => second)
  const firstSession = first.session

  first.fail(new Error('video connection failed (WebRTC)'))

  assert.deepEqual(second.calls, ['connect'], 'connected straight away')
  assert.deepEqual(second.session!.init, firstSession!.init, 'with the session the server already introduced')
  assert.equal(second.session!.token, 'tok')
})

test('input moves to the websocket when the fallback has no channel of its own', () => {
  const second = new FakeTransport('webcodecs-ws') // no input, like WebCodecsTransport
  const dataChannel = { move() {}, scroll() {}, button() {}, key() {}, touch() {} }
  const { client, first } = setup(() => second, dataChannel)
  assert.equal(client.input, dataChannel, 'before: input goes through the first transport’s channel')

  first.fail(new Error('video connection failed (WebRTC)'))

  assert.notEqual(client.input, dataChannel, 'input must not stay bound to the dead transport')
  assert.ok(client.input instanceof WebSocketInput, 'it falls back to the main websocket')
})

test('the fallback is used at most once: a second failure closes the session', () => {
  const second = new FakeTransport('webcodecs-ws')
  const { first, closed, changed } = setup(() => second)

  first.fail(new Error('video connection failed (WebRTC)'))
  second.fail(new Error('media websocket failed'))

  assert.deepEqual(changed, ['webcodecs-ws'], 'no second switch')
  assert.equal(closed.length, 1, 'with nothing left to try, the session closes as before')
  assert.equal(closed[0]!.message, 'media websocket failed')
})

test('without a fallback, a stream failure closes the session exactly as before', () => {
  const { first, closed, changed } = setup()

  first.fail(new Error('video connection failed (WebRTC)'))

  assert.deepEqual(changed, [])
  assert.equal(closed.length, 1)
  assert.equal(closed[0]!.message, 'video connection failed (WebRTC)')
})

test('a fallback that cannot be built leaves the original error standing', () => {
  const { first, closed, changed } = setup(() => {
    throw new Error('browser has no WebCodecs')
  })

  first.fail(new Error('video connection failed (WebRTC)'))

  assert.deepEqual(changed, [])
  assert.equal(closed.length, 1)
  assert.equal(closed[0]!.message, 'video connection failed (WebRTC)', 'the real cause, not the fallback’s')
})

test('a transport error after the user disconnected does not resurrect the stream', () => {
  let built = 0
  const { client, first, changed } = setup(() => {
    built++
    return new FakeTransport('webcodecs-ws')
  })

  client.disconnect()
  first.fail(new Error('video connection failed (WebRTC)'))

  assert.equal(built, 0, 'no fallback is built for a session the user ended')
  assert.deepEqual(changed, [])
})
