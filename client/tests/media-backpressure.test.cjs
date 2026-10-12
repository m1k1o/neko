const { test } = require('node:test')
const assert = require('node:assert/strict')
require('ts-node').register({ transpileOnly: true, compilerOptions: { module: 'CommonJS' } })
const { WebCodecsPlayer } = require('../src/neko/webcodecs')
const { parseMediaSample, encodeAck } = require('../src/neko/media-protocol')

const HEADER_SIZE = 24

// frame builds a server sample message for the given track and sequence.
function frame({ track = 1, keyframe = true, sequence = 1, payload = [1, 2, 3] } = {}) {
  const buffer = new ArrayBuffer(HEADER_SIZE + payload.length)
  const view = new DataView(buffer)
  view.setUint8(0, 1)
  view.setUint8(1, track)
  view.setUint8(2, keyframe ? 1 : 0)
  view.setBigInt64(4, 0n)
  view.setBigInt64(12, -1n)
  view.setUint32(20, sequence)
  new Uint8Array(buffer, HEADER_SIZE).set(payload)
  return buffer
}

function setup({ videoQueueSize = 0 } = {}) {
  const sent = []
  const decoded = []
  global.window = { addEventListener() {}, removeEventListener() {} }
  global.EncodedVideoChunk = class {
    constructor(init) {
      Object.assign(this, init)
      decoded.push(init)
    }
  }
  global.EncodedAudioChunk = global.EncodedVideoChunk
  const canvas = { width: 0, height: 0, getContext: () => ({ drawImage() {} }) }
  const player = new WebCodecsPlayer(
    canvas,
    { protocol: 'webcodecs-ws-v1', url: 'ws://localhost/media/ws', video: { codec: 'vp8' }, audio: { codec: 'opus' } },
    () => {},
    () => {},
  )
  // Stand in for the decoder and transport, which are created in start().
  player.videoDecoder = { decodeQueueSize: videoQueueSize, decode() {} }
  player.audioDecoder = { decodeQueueSize: 0, decode() {} }
  player.transport = {
    send: (data) => sent.push(parseAck(data)),
    start() {},
    stop() {},
    bufferedAmount: () => 4096,
  }
  return { player, sent, decoded }
}

function parseAck(buffer) {
  const view = new DataView(buffer)
  return { version: view.getUint8(0), kind: view.getUint8(1), sequence: view.getUint32(4) }
}

test('stats expose the signals needed to tune the window', () => {
  const { player } = setup()
  player.onSample(parseMediaSample(frame({ sequence: 1, keyframe: true })))
  player.onVideoFrame({ displayWidth: 0, displayHeight: 0, close() {} })

  // A healthy stream reports no drops and a known ack position.
  assert.deepEqual(player.stats(), {
    droppedFrames: 0,
    keyframeRequests: 0,
    videoDecodeQueue: 0,
    audioDecodeQueue: 0,
    bufferedAmount: 4096,
    ackedSequence: 1,
  })

  // A drop is attributed, which is what makes the window tunable.
  player.videoDecoder.decodeQueueSize = 99
  player.onSample(parseMediaSample(frame({ sequence: 2, keyframe: false })))
  const stats = player.stats()
  assert.equal(stats.droppedFrames, 1)
  assert.equal(stats.keyframeRequests, 1)
})

test('a client side drop asks the server for a keyframe', () => {
  const { player, sent } = setup()
  player.onSample(parseMediaSample(frame({ sequence: 1, keyframe: true })))

  player.videoDecoder.decodeQueueSize = 99
  player.onSample(parseMediaSample(frame({ sequence: 2, keyframe: false })))

  // Without a keyframe request the stream stays frozen until the encoder's
  // next scheduled keyframe, which can be a second or more away.
  assert.ok(
    sent.some((message) => message.kind === 2),
    'a keyframe request must follow a client side drop',
  )
})

test('the ack round trip preserves the sequence number', () => {
  const sample = parseMediaSample(frame({ sequence: 1234 }))
  assert.equal(sample.sequence, 1234)
  assert.deepEqual(parseAck(encodeAck(sample.sequence)), { version: 1, kind: 1, sequence: 1234 })
})

test('a painted frame is acked so the server can advance its window', () => {
  const { player, sent } = setup()
  player.onSample(parseMediaSample(frame({ sequence: 7 })))
  // Nothing is acked until the frame is actually presented.
  assert.deepEqual(sent, [])

  player.onVideoFrame({ displayWidth: 0, displayHeight: 0, close() {} })
  assert.deepEqual(
    sent.map((ack) => ack.sequence),
    [7],
  )
})

test('a frame dropped for decoder pressure is acked and does not fail the connection', () => {
  const errors = []
  const { player, sent, decoded } = setup({ videoQueueSize: 99 })
  player.onError = (error) => errors.push(error)

  player.onSample(parseMediaSample(frame({ sequence: 3 })))

  assert.deepEqual(decoded, [], 'the frame should not reach the decoder')
  assert.deepEqual(
    sent.filter((message) => message.kind === 1).map((ack) => ack.sequence),
    [3],
    'the dropped frame must be acked so the window does not stall',
  )
  assert.equal(player.stopped, false, 'decoder pressure must not tear down the player')
})

test('after a drop the client waits for a keyframe before decoding deltas', () => {
  const { player, decoded } = setup()

  // Establish a healthy stream so the decoder has a reference frame.
  player.onSample(parseMediaSample(frame({ sequence: 1, keyframe: true })))
  player.onSample(parseMediaSample(frame({ sequence: 2, keyframe: false })))
  assert.deepEqual(
    decoded.map((chunk) => chunk.type),
    ['key', 'delta'],
    'a healthy stream should decode both frame types',
  )

  // The decoder falls behind and a frame is dropped, breaking the chain.
  player.videoDecoder.decodeQueueSize = 99
  player.onSample(parseMediaSample(frame({ sequence: 3, keyframe: false })))
  assert.equal(decoded.length, 2, 'the dropped frame must not reach the decoder')

  // The decoder drains, but deltas are still undecodable after the gap.
  player.videoDecoder.decodeQueueSize = 0
  player.onSample(parseMediaSample(frame({ sequence: 4, keyframe: false })))
  assert.equal(decoded.length, 2, 'a delta after a drop is undecodable')

  player.onSample(parseMediaSample(frame({ sequence: 5, keyframe: true })))
  assert.deepEqual(
    decoded.map((chunk) => chunk.type),
    ['key', 'delta', 'key'],
    'a keyframe should restore decoding',
  )
})

test('acks are cumulative and never move backwards', () => {
  const { player, sent } = setup()
  player.onSample(parseMediaSample(frame({ sequence: 5 })))
  player.onVideoFrame({ displayWidth: 0, displayHeight: 0, close() {} })
  // A late frame with an older sequence must not un-ack newer progress.
  player.pendingVideoSequence = 2
  player.onVideoFrame({ displayWidth: 0, displayHeight: 0, close() {} })

  assert.deepEqual(
    sent.map((ack) => ack.sequence),
    [5],
  )
})
