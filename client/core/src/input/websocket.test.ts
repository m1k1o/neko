// Input over the websocket: the data channel opcodes as the server's control/* events.
import { test } from 'vitest'
import assert from 'node:assert/strict'
import { tick, reset, sent, sentPayloads, sockets, init, receive } from '../test/browser.ts'

const { WebSocketInput, NekoClient, OP } = await import('../index.ts')

const recorder = () => {
  const out: { event: string; payload: unknown }[] = []
  return { out, send: (event: string, payload?: unknown) => out.push({ event, payload }) }
}

test('every opcode becomes the matching control/* event with the server payload fields', async () => {
  reset()
  const r = recorder()
  const input = new WebSocketInput(r)
  input.send(OP.SCROLL, [-2, -1], [-2, 2], [1, 1])
  input.send(OP.SCROLL, [-2, 3], [-2, 0], [1, 0])
  input.send(OP.KEY_DOWN, [4, 0xff0d])
  input.send(OP.KEY_UP, [4, 0xff0d])
  input.send(OP.BTN_DOWN, [4, 1])
  input.send(OP.BTN_UP, [4, 1])
  input.send(OP.TOUCH_BEGIN, [4, 7], [-4, 10], [-4, 20], [1, 128])
  input.send(OP.TOUCH_UPDATE, [4, 7], [-4, 11], [-4, 21], [1, 129])
  input.send(OP.TOUCH_END, [4, 7], [-4, 12], [-4, 22], [1, 0])
  input.send(99, [1, 1]) // not an input opcode
  assert.deepEqual(r.out, [
    { event: 'control/scroll', payload: { delta_x: -1, delta_y: 2, control_key: true } },
    { event: 'control/scroll', payload: { delta_x: 3, delta_y: 0, control_key: false } },
    { event: 'control/keydown', payload: { keysym: 0xff0d } },
    { event: 'control/keyup', payload: { keysym: 0xff0d } },
    { event: 'control/buttondown', payload: { code: 1 } },
    { event: 'control/buttonup', payload: { code: 1 } },
    { event: 'control/touchbegin', payload: { touch_id: 7, x: 10, y: 20, pressure: 128 } },
    { event: 'control/touchupdate', payload: { touch_id: 7, x: 11, y: 21, pressure: 129 } },
    { event: 'control/touchend', payload: { touch_id: 7, x: 12, y: 22, pressure: 0 } },
  ])
  await tick(100)
  assert.equal(r.out.length, 9)
})

test('moves within 16 ms are coalesced, the last one wins; a click flushes the move first', async () => {
  reset()
  const r = recorder()
  const input = new WebSocketInput(r)
  input.send(OP.MOVE, [2, 10], [2, 20])
  input.send(OP.MOVE, [2, 11], [2, 21])
  assert.deepEqual(r.out, [])
  await tick(15)
  assert.deepEqual(r.out, [])
  await tick(1)
  assert.deepEqual(r.out, [{ event: 'control/move', payload: { x: 11, y: 21 } }])
  await tick(100)
  assert.equal(r.out.length, 1, 'a flushed move is not sent again')

  input.send(OP.MOVE, [2, 30], [2, 40])
  input.send(OP.BTN_DOWN, [4, 1]) // the press lands where the pointer is now
  assert.deepEqual(r.out.slice(1), [
    { event: 'control/move', payload: { x: 30, y: 40 } },
    { event: 'control/buttondown', payload: { code: 1 } },
  ])
  await tick(16)
  assert.equal(r.out.length, 3)

  input.send(OP.MOVE, [2, 50], [2, 60]) // and the timer works again after a flush
  await tick(16)
  assert.deepEqual(r.out.at(-1), { event: 'control/move', payload: { x: 50, y: 60 } })
})

test('nothing reaches the server while the socket is not open', async () => {
  reset()
  const transport = {
    kind: 'hls',
    element: null,
    connect: async () => {},
    suspend() {},
    close() {},
    attach() {},
    detach() {},
    setPlaying: async () => {},
    setVolume() {},
    setMuted() {},
    on: () => () => {},
  } as const
  const client = new NekoClient({ transport })
  assert.ok(client.input instanceof WebSocketInput, 'no input channel on the transport: the websocket')
  client.input.send(OP.KEY_DOWN, [4, 1]) // no socket at all
  client.state.authenticated = true
  client.connect()
  sockets[0].readyState = 0 // still connecting
  client.input.send(OP.KEY_DOWN, [4, 1])
  client.input.send(OP.MOVE, [2, 1], [2, 2])
  await tick(16)
  assert.deepEqual(sent, [])
  sockets[0].readyState = 1
  receive('system/init', init())
  client.input.send(OP.KEY_DOWN, [4, 1])
  assert.deepEqual(sent, ['control/keydown'])
  assert.deepEqual(sentPayloads.at(-1), { event: 'control/keydown', payload: { keysym: 1 } })
  client.disconnect()
  client.input.send(OP.KEY_UP, [4, 1])
  assert.deepEqual(sent, ['control/keydown'])
})
