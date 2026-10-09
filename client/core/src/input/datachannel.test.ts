// The data channel input: the overlay's input as big-endian frames out, the host's cursor in.
import { test } from 'vitest'
import assert from 'node:assert/strict'
import { FakeChannel } from '../test/browser.ts'

const { DataChannelInput } = await import('../index.ts')

test('every method is one frame: opcode, u16 payload length, big-endian fields of the server’s widths', () => {
  const input = new DataChannelInput()
  const dc = new FakeChannel()
  input.bind(dc as any)
  assert.equal(dc.binaryType, 'arraybuffer')
  dc.readyState = 'open'
  input.move(100, 200)
  input.move(65535, 1)
  input.scroll(-1, 1, true)
  input.scroll(3, 0, false)
  input.button(1, true)
  input.button(3, false)
  input.key(0xff0d, true)
  input.key(0xff0d, false)
  input.key(0x1008ff11, true) // XF86AudioLowerVolume: a keysym above u16
  input.touch('begin', 7, -5, 6, 255)
  input.touch('update', 7, 10, 20, 128)
  input.touch('end', 7, 10, 20, 0)
  assert.deepEqual(dc.frames, [
    [1, 0, 4, 0, 100, 0, 200], // MOVE: u16 x, u16 y
    [1, 0, 4, 255, 255, 0, 1],
    [2, 0, 5, 255, 255, 0, 1, 1], // SCROLL: i16 delta_x, i16 delta_y, u8 control_key
    [2, 0, 5, 0, 3, 0, 0, 0],
    [5, 0, 4, 0, 0, 0, 1], // BTN_DOWN: u32 code
    [6, 0, 4, 0, 0, 0, 3], // BTN_UP
    [3, 0, 4, 0, 0, 0xff, 0x0d], // KEY_DOWN: u32 keysym
    [4, 0, 4, 0, 0, 0xff, 0x0d], // KEY_UP
    [3, 0, 4, 0x10, 0x08, 0xff, 0x11],
    [8, 0, 13, 0, 0, 0, 7, 255, 255, 255, 251, 0, 0, 0, 6, 255], // TOUCH_BEGIN: u32 id, i32 x, i32 y, u8 pressure
    [9, 0, 13, 0, 0, 0, 7, 0, 0, 0, 10, 0, 0, 0, 20, 128], // TOUCH_UPDATE
    [10, 0, 13, 0, 0, 0, 7, 0, 0, 0, 10, 0, 0, 0, 20, 0], // TOUCH_END
  ])
})

test('nothing is sent while the channel is not open: unbound, connecting, or gone', () => {
  const input = new DataChannelInput()
  input.move(1, 2) // no channel at all
  const dc = new FakeChannel()
  input.bind(dc as any)
  input.move(1, 2) // still connecting
  input.key(1, true)
  assert.deepEqual(dc.frames, [])
  dc.readyState = 'open'
  input.move(1, 2)
  assert.equal(dc.frames.length, 1)
  input.bind(null) // the peer is gone
  input.move(1, 2)
  input.button(1, true)
  assert.equal(dc.frames.length, 1)
})

test('receive: cursor position and image frames reach the callbacks', () => {
  const input = new DataChannelInput()
  const got: unknown[] = []
  input.onCursorPosition = (p) => got.push(p)
  input.onCursorImage = (img) => got.push(img)
  const dc = new FakeChannel()
  input.bind(dc as any)
  const frame = (bytes: number[]) => dc.onmessage!({ data: new Uint8Array(bytes).buffer })
  frame([1, 0, 4, 0x01, 0x02, 0x03, 0x04]) // CURSOR_POSITION x=258 y=772
  frame([2, 0, 12, 0, 16, 0, 20, 0, 3, 0, 5, 0x89, 0x50, 0x4e, 0x47]) // CURSOR_IMAGE 16x20 hotspot 3,5 + png bytes
  frame([9, 0, 0]) // unknown: ignored
  assert.deepEqual(got[0], { x: 258, y: 772 })
  const img = got[1] as { width: number; height: number; x: number; y: number; uri: string }
  assert.deepEqual([img.width, img.height, img.x, img.y], [16, 20, 3, 5])
  assert.match(img.uri, /^blob:/)
  assert.equal(got.length, 2)
})
