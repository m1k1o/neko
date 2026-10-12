const { test } = require('node:test')
const assert = require('node:assert/strict')
require('ts-node').register({ transpileOnly: true, compilerOptions: { module: 'CommonJS' } })
const { WebCodecsPlayer } = require('../src/neko/webcodecs')

function setup() {
  const contexts = []
  const listeners = new Map()
  global.window = {
    addEventListener: (name, callback) => listeners.set(name, callback),
    removeEventListener: (name) => listeners.delete(name),
  }
  global.AudioContext = class {
    state = 'suspended'
    resumes = 0
    destination = {}
    constructor() {
      contexts.push(this)
    }
    createGain() {
      return { gain: { value: 1 }, connect() {}, disconnect() {} }
    }
    resume() {
      this.resumes++
      return new Promise(() => {})
    }
    close() {
      this.state = 'closed'
      return Promise.resolve()
    }
  }
  global.VideoDecoder = global.AudioDecoder = class {
    static async isConfigSupported() {
      return { supported: true }
    }
    configure() {}
    close() {}
  }
  global.WebSocket = class {
    close() {}
  }
  const errors = []
  const player = new WebCodecsPlayer(
    {},
    {
      protocol: 'webcodecs-ws-v1',
      url: 'ws://localhost/media/ws',
      video: { codec: 'vp8' },
      audio: { codec: 'opus', sample_rate: 48000, number_of_channels: 2 },
    },
    () => {},
    (error) => errors.push(error),
  )
  return { player, contexts, listeners, errors }
}

test('play initializes audio without a second gesture or waiting for autoplay permission', async () => {
  const { player, contexts } = setup()
  player.setVolume(0.4)
  player.setMuted(true)
  const play = player.setPlaying(true)
  assert.equal(contexts.length, 1)
  assert.equal(contexts[0].resumes, 1)
  assert.equal(player.gain.gain.value, 0)
  await play // resume intentionally never resolves, but startup must proceed.
  player.setMuted(false)
  assert.equal(player.gain.gain.value, 0.4)
  assert.equal(contexts[0].resumes, 2)
  player.setVolume(0.7)
  assert.equal(player.gain.gain.value, 0.7)
  assert.equal(contexts[0].resumes, 3)
  player.stop()
})

test('activation listeners remain available after suspension and are removed on stop', async () => {
  const { player, contexts, listeners } = setup()
  await player.start()
  await player.setPlaying(true)
  contexts[0].state = 'running'
  listeners.get('pointerdown')()
  assert.equal(contexts[0].resumes, 1)
  assert.equal(listeners.size, 2)
  contexts[0].state = 'suspended'
  listeners.get('keydown')()
  assert.equal(contexts[0].resumes, 2)
  player.stop()
  assert.equal(listeners.size, 0)
  assert.equal(contexts[0].state, 'closed')
  await player.setPlaying(true)
  player.setMuted(false)
  assert.equal(contexts.length, 1)
})

test('stopping during codec support checks does not restart decoders or listeners', async () => {
  const { player, listeners, contexts } = setup()
  const start = player.start()
  player.stop()
  await start
  assert.equal(listeners.size, 0)
  assert.equal(contexts.length, 0)
  assert.equal(player.transport, undefined)
})
