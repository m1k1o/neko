const { test } = require('node:test')
const assert = require('node:assert/strict')
require('ts-node').register({ transpileOnly: true, compilerOptions: { module: 'CommonJS' } })
const { BaseClient } = require('../src/neko/base')

function setup() {
  const timers = new Map()
  const frames = new Map()
  let id = 0
  global.window = {
    setTimeout: (callback) => (timers.set(++id, callback), id),
    clearTimeout: (handle) => timers.delete(handle),
  }
  global.requestAnimationFrame = (callback) => (frames.set(++id, callback), id)
  global.cancelAnimationFrame = (handle) => frames.delete(handle)
  global.WebSocket = { OPEN: 1 }
  const sent = []
  const client = new BaseClient()
  const socket = () => ({ readyState: 1, close() {}, send: (data) => sent.push(JSON.parse(data)) })
  client._ws = socket()
  client._webCodecsMode = true
  const flush = (callbacks) => {
    for (const [handle, callback] of [...callbacks]) {
      callbacks.delete(handle)
      callback()
    }
  }
  return { client, sent, timers, frames, socket, flush }
}

test('pointer updates keep working when animation frames are suspended', () => {
  const { client, sent, timers, flush } = setup()
  client.sendData('mousemove', { x: 10, y: 20 })
  client.sendData('mousemove', { x: 30, y: 40 })
  flush(timers)
  assert.deepEqual(sent, [{ event: 'control/move', x: 30, y: 40 }])
  client.sendData('mousemove', { x: 50, y: 60 })
  flush(timers)
  assert.equal(sent.at(-1).x, 50)
})

test('clicks and scrolling follow the latest pointer position without duplicates', () => {
  for (const [event, data, expected] of [
    ['mousedown', { key: 1 }, 'control/buttondown'],
    ['mouseup', { key: 1 }, 'control/buttonup'],
    ['wheel', { x: 0, y: 10 }, 'control/scroll'],
  ]) {
    const { client, sent, timers, frames, flush } = setup()
    client.sendData('mousemove', { x: 25, y: 35 })
    client.sendData(event, data)
    flush(timers)
    flush(frames)
    assert.deepEqual(
      sent.map((message) => message.event),
      ['control/move', expected],
    )
  }
})

test('disconnect discards pending movement instead of sending it on the new connection', () => {
  const { client, sent, timers, frames, socket, flush } = setup()
  client.sendData('mousemove', { x: 1, y: 2 })
  client.disconnect()
  client._ws = socket()
  flush(timers)
  flush(frames)
  assert.deepEqual(sent, [])
  client.sendData('mousemove', { x: 3, y: 4 })
  flush(timers)
  assert.equal(sent.at(-1).x, 3)
})

test('a failed send does not leave pointer batching permanently locked', () => {
  const { client, sent, timers, frames, socket, flush } = setup()
  client._ws.send = () => {
    throw new Error('send failed')
  }
  client.sendData('mousemove', { x: 1, y: 2 })
  assert.throws(() => {
    flush(timers)
    flush(frames)
  }, /send failed/)
  client._ws = socket()
  client.sendData('mousemove', { x: 3, y: 4 })
  flush(timers)
  flush(frames)
  assert.equal(sent.at(-1)?.x, 3)
})

test('WebRTC still sends its binary mouse packet immediately', () => {
  const { client, timers } = setup()
  client._webCodecsMode = false
  client._peer = {}
  client._state = 'connected'
  let packet
  client._channel = {
    send: (data) => {
      packet = new DataView(data)
    },
  }
  client.sendData('mousemove', { x: 300, y: 400 })
  assert.equal(packet.byteLength, 7)
  assert.equal(packet.getUint16(3, true), 300)
  assert.equal(packet.getUint16(5, true), 400)
  assert.equal(timers.size, 0)
})
