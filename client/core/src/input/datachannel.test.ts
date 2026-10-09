// The data channel input: big-endian frames out, the host's cursor in.
import { test } from 'vitest'
import assert from 'node:assert/strict'
import { FakeChannel } from '../test/browser.ts'

const { DataChannelInput, OP } = await import('../index.ts')

test('send: big-endian input messages on the data channel, dropped while it is not open', () => {
  const input = new DataChannelInput()
  input.send(1, [2, 100], [2, 200]) // no channel at all
  const dc = new FakeChannel()
  input.bind(dc as any)
  assert.equal(dc.binaryType, 'arraybuffer')
  input.send(1, [2, 100], [2, 200])
  assert.deepEqual(dc.frames, [])
  dc.readyState = 'open'
  input.send(1, [2, 100], [2, 200]) // MOVE x=100 y=200
  input.send(2, [-2, -1], [-2, 1]) // SCROLL
  input.send(3, [4, 0xffff], [1, 1]) // KEY_DOWN with a u32 keysym, u8 flag
  input.send(8, [4, 7], [-4, -5], [-4, 6], [1, 255]) // TOUCH_BEGIN with i32 coordinates
  assert.deepEqual(dc.frames, [
    [1, 0, 4, 0, 100, 0, 200],
    [2, 0, 4, 255, 255, 0, 1],
    [3, 0, 5, 0, 0, 255, 255, 1],
    [8, 0, 13, 0, 0, 0, 7, 255, 255, 255, 251, 0, 0, 0, 6, 255],
  ])
  input.bind(null) // the peer is gone
  input.send(1, [2, 1], [2, 1])
  assert.equal(dc.frames.length, 4)
})

test('receive: cursor position and image frames reach the callbacks', () => {
  const input = new DataChannelInput()
  const got: unknown[] = []
  input.onCursorPosition = (p) => got.push(p)
  input.onCursorImage = (img) => got.push(img)
  const dc = new FakeChannel()
  input.bind(dc as any)
  const frame = (bytes: number[]) => dc.onmessage!({ data: new Uint8Array(bytes).buffer })
  frame([OP.CURSOR_POSITION, 0, 4, 0x01, 0x02, 0x03, 0x04]) // x=258 y=772
  frame([OP.CURSOR_IMAGE, 0, 12, 0, 16, 0, 20, 0, 3, 0, 5, 0x89, 0x50, 0x4e, 0x47]) // 16x20 hotspot 3,5 + png bytes
  frame([9, 0, 0]) // unknown: ignored
  assert.deepEqual(got[0], { x: 258, y: 772 })
  const img = got[1] as { width: number; height: number; x: number; y: number; uri: string }
  assert.deepEqual([img.width, img.height, img.x, img.y], [16, 20, 3, 5])
  assert.match(img.uri, /^blob:/)
  assert.equal(got.length, 2)
})
