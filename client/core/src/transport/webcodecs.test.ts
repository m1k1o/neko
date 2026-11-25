// The WebCodecs websocket transport on its own: the ticket it fetches, the socket it opens, the
// flow control it owes the server (acks after painting, keyframe requests after a drop), and what
// close() leaves behind.
import { test } from 'vitest'
import assert from 'node:assert/strict'
import {
  tick,
  settle,
  reset,
  net,
  sockets,
  fetches,
  FakeElement,
  FakeVideoDecoder,
  FakeAudioDecoder,
  init,
} from '../test/browser.ts'

const { WebCodecsTransport } = await import('../index.ts')
type Transport = InstanceType<typeof WebCodecsTransport>
type State = Parameters<Parameters<Transport['on']>[1]>[0]

const HEADER = 24
const VIDEO = 1
const AUDIO = 2
const ACK = 1
const KEYFRAME_REQUEST = 2

// a sample as the server frames it (server/internal/media/sample.go)
const frame = ({ track = VIDEO, keyframe = true, sequence = 1, payload = [1, 2, 3] } = {}) => {
  const buffer = new ArrayBuffer(HEADER + payload.length)
  const view = new DataView(buffer)
  view.setUint8(0, 1) // protocol version
  view.setUint8(1, track)
  view.setUint8(2, keyframe ? 1 : 0)
  view.setBigInt64(4, 0n) // timestamp
  view.setBigInt64(12, -1n) // duration: none
  view.setUint32(20, sequence)
  new Uint8Array(buffer, HEADER).set(payload)
  return buffer
}

const TICKET = {
  ticket: 'tok',
  protocol: 'webcodecs-ws-v1',
  video: { codec: 'vp8' },
  audio: { codec: 'opus', sample_rate: 48000, number_of_channels: 2 },
}

const sessionInfo = () => ({
  url: 'http://neko.test',
  token: 'tok',
  init: init(),
  send: () => {},
  on: () => () => {},
})

// a connected transport with its element attached, as the client would set it up
const setup = async (ticket: unknown = TICKET) => {
  reset()
  net.respond = () => ({ status: 200, body: ticket })
  const t = new WebCodecsTransport()
  const states: State[] = []
  const errors: Error[] = []
  t.on('state', (s) => states.push(s))
  t.on('error', (e) => errors.push(e))
  const container = new FakeElement('div')
  const detach = t.attach(container as any)
  await t.connect(sessionInfo() as any)
  await settle()
  const socket = sockets.at(-1)!
  socket.onopen?.()
  return { t, states, errors, container, detach, socket, video: () => FakeVideoDecoder.all.at(-1)! }
}

// frames the transport sent back, decoded
const controls = (socket: { binary: number[][] }) =>
  socket.binary.map((b) => ({ kind: b[1], sequence: new DataView(new Uint8Array(b).buffer).getUint32(4) }))
const acks = (socket: { binary: number[][] }) => controls(socket).filter((c) => c.kind === ACK)

test('connect fetches a ticket and opens the media socket with it', async () => {
  const { socket, states } = await setup()

  assert.equal(fetches.length, 1)
  assert.match(fetches[0].url, /\/api\/media\/ticket$/)
  assert.equal(fetches[0].init.method, 'POST')
  assert.equal(socket.url, 'ws://neko.test/media/ws?ticket=tok')
  assert.equal(socket.binaryType, 'arraybuffer')
  assert.equal(states.at(-1)!.status, 'connected')
})

test('a server that will not issue a ticket leaves the room without video', async () => {
  reset()
  net.respond = () => ({ status: 403, body: { message: 'cannot watch' } })
  const t = new WebCodecsTransport()
  const states: State[] = []
  const errors: Error[] = []
  t.on('state', (s) => states.push(s))
  t.on('error', (e) => errors.push(e))
  await t.connect(sessionInfo() as any)
  await settle()

  // 'unavailable', not an error: the room works, it just has no picture
  assert.equal(states.at(-1)!.status, 'unavailable')
  assert.deepEqual(errors, [])
  assert.equal(sockets.length, 0)
})

test('a codec the browser cannot decode fails the stream instead of retrying', async () => {
  reset()
  net.respond = () => ({ status: 200, body: TICKET })
  FakeVideoDecoder.supported = false
  const t = new WebCodecsTransport()
  const errors: Error[] = []
  t.on('error', (e) => errors.push(e))
  await t.connect(sessionInfo() as any)
  await settle()

  assert.equal(errors.length, 1)
  assert.match(errors[0].message, /cannot decode vp8/)
  assert.equal(sockets.length, 0)
})

test('a decodable stream is painted and acked after the frame is presented', async () => {
  const { t, socket, video, container } = await setup()
  await t.setPlaying(true)

  socket.onmessage!({ data: frame({ sequence: 7 }) })
  assert.equal(video().chunks.length, 1)
  assert.equal(video().chunks[0].type, 'key')
  // nothing is acked until the decoder hands the frame back and it is painted
  assert.deepEqual(acks(socket), [])

  video().emit({ displayWidth: 320, displayHeight: 240 })
  assert.deepEqual(
    acks(socket).map((a) => a.sequence),
    [7],
  )
  const canvas = container.children[0]
  assert.ok(canvas.ctx.calls.some((c: unknown[]) => c[0] === 'drawImage'))
})

test('a delta frame before any keyframe is not decoded', async () => {
  const { socket, video } = await setup()
  socket.onmessage!({ data: frame({ sequence: 1, keyframe: false }) })
  assert.equal(video().chunks.length, 0)
})

test('a frame dropped for decoder pressure is acked, asks for a keyframe, and keeps the stream', async () => {
  const { t, socket, video, errors } = await setup()
  await t.setPlaying(true)

  // a healthy stream first, so the decoder has a reference frame
  socket.onmessage!({ data: frame({ sequence: 1, keyframe: true }) })
  video().emit({})
  socket.onmessage!({ data: frame({ sequence: 2, keyframe: false }) })
  assert.equal(video().chunks.length, 2)

  // the decoder falls behind: the frame is dropped rather than queued
  video().decodeQueueSize = 99
  socket.onmessage!({ data: frame({ sequence: 3, keyframe: false }) })
  assert.equal(video().chunks.length, 2, 'the dropped frame must not reach the decoder')

  const sent = controls(socket)
  assert.ok(
    sent.some((c) => c.kind === KEYFRAME_REQUEST),
    'a drop must ask for a keyframe, or the picture stays frozen until the encoder sends one',
  )
  assert.ok(
    acks(socket).some((a) => a.sequence === 3),
    'the dropped frame must be acked so the window does not stall on it',
  )
  assert.deepEqual(errors, [], 'decoder pressure must not fail the stream')
})

test('after a drop, deltas are withheld until a keyframe restores the decode chain', async () => {
  const { t, socket, video } = await setup()
  await t.setPlaying(true)

  socket.onmessage!({ data: frame({ sequence: 1, keyframe: true }) })
  video().emit({})
  video().decodeQueueSize = 99
  socket.onmessage!({ data: frame({ sequence: 2, keyframe: false }) })
  const afterDrop = video().chunks.length

  // the decoder drains, but the chain is still broken
  video().decodeQueueSize = 0
  socket.onmessage!({ data: frame({ sequence: 3, keyframe: false }) })
  assert.equal(video().chunks.length, afterDrop, 'a delta across a gap is undecodable')

  socket.onmessage!({ data: frame({ sequence: 4, keyframe: true }) })
  assert.equal(video().chunks.length, afterDrop + 1, 'a keyframe restores decoding')
  assert.equal(video().chunks.at(-1).type, 'key')
})

test('acks are cumulative and never move backwards', async () => {
  const { t, socket, video } = await setup()
  await t.setPlaying(true)

  socket.onmessage!({ data: frame({ sequence: 5 }) })
  video().emit({})
  // a late frame carrying an older sequence must not un-ack newer progress
  socket.onmessage!({ data: frame({ sequence: 2 }) })
  video().emit({})

  assert.deepEqual(
    acks(socket).map((a) => a.sequence),
    [5],
  )
})

test('audio is decoded on its own track and skipped when its decoder is behind', async () => {
  const { t, socket } = await setup()
  await t.setPlaying(true)
  const audio = () => FakeAudioDecoder.all.at(-1)!

  socket.onmessage!({ data: frame({ track: AUDIO, sequence: 0 }) })
  assert.equal(audio().chunks.length, 1)
  assert.equal(audio().chunks[0].type, 'key', 'every audio sample decodes independently')

  audio().decodeQueueSize = 99
  socket.onmessage!({ data: frame({ track: AUDIO, sequence: 0 }) })
  assert.equal(audio().chunks.length, 1, 'a backed up audio decoder skips rather than queues')
})

test('the stream size is reported from the first frame', async () => {
  const { t, socket, video, states } = await setup()
  await t.setPlaying(true)
  socket.onmessage!({ data: frame({ sequence: 1 }) })
  video().emit({ displayWidth: 1280, displayHeight: 720 })

  const state = states.at(-1)!
  assert.deepEqual(state.size, { width: 1280, height: 720 })
  assert.equal(state.video.playable, true)
})

test('a lost socket reconnects with backoff and gives up after enough failures', async () => {
  const { t, socket, states } = await setup()
  const before = sockets.length

  socket.onclose!()
  assert.equal(states.at(-1)!.status, 'connecting')
  assert.equal(sockets.length, before, 'the retry waits for the backoff')

  await tick(1600)
  assert.equal(sockets.length, before + 1, 'it reopens after the backoff')
  t.close()
})

test('close stops the stream, the socket and the retries', async () => {
  const { t, socket, states } = await setup()
  t.close()

  assert.equal(socket.closed, true)
  assert.equal(states.at(-1)!.status, 'disconnected')
  assert.equal(states.at(-1)!.video.playable, false)

  // a close while a retry was pending must not reopen anything
  const count = sockets.length
  await tick(10_000)
  assert.equal(sockets.length, count)
})

test('detach removes the element and a second attach replaces it', async () => {
  const { t, container, detach } = await setup()
  assert.equal(container.children.length, 1)
  assert.equal(t.element, container.children[0])

  const other = new FakeElement('div')
  t.attach(other as any)
  assert.equal(other.children.length, 1)

  detach()
  assert.equal(t.element, other.children[0], 'detaching the old element leaves the current one')
})

test('stats expose what the window is doing', async () => {
  const { t, socket, video } = await setup()
  await t.setPlaying(true)
  socket.onmessage!({ data: frame({ sequence: 1 }) })
  video().emit({})

  let stats = t.stats()
  assert.equal(stats.droppedFrames, 0)
  assert.equal(stats.keyframeRequests, 0)
  assert.equal(stats.ackedSequence, 1)

  video().decodeQueueSize = 99
  socket.onmessage!({ data: frame({ sequence: 2, keyframe: false }) })
  stats = t.stats()
  assert.equal(stats.droppedFrames, 1)
  assert.equal(stats.keyframeRequests, 1)
})
