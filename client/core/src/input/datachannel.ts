// Input over the WebRTC data channel: big-endian binary messages (server/internal/webrtc/payload),
// and the host's cursor coming back the same way.
import type { InputChannel } from '../transport.ts'
import type { CursorImage } from '../types.ts'

// opcodes (server/internal/webrtc/payload)
const OP = {
  // client -> server
  MOVE: 1,
  SCROLL: 2,
  KEY_DOWN: 3,
  KEY_UP: 4,
  BTN_DOWN: 5,
  BTN_UP: 6,
  TOUCH_BEGIN: 8,
  TOUCH_UPDATE: 9,
  TOUCH_END: 10,
  // server -> client
  CURSOR_POSITION: 1,
  CURSOR_IMAGE: 2,
}
const TOUCH_OP = { begin: OP.TOUCH_BEGIN, update: OP.TOUCH_UPDATE, end: OP.TOUCH_END }

// a field of a message: its width in bytes (2 = u16, 4 = u32, -2 = i16, -4 = i32, 1 = u8) and value
type Field = [bytes: number, value: number]

export class DataChannelInput implements InputChannel {
  private dc: RTCDataChannel | null = null

  onCursorPosition?: (pos: { x: number; y: number }) => void
  onCursorImage?: (image: CursorImage) => void

  // the channel the server opened on the peer connection; null once the peer is gone
  bind(dc: RTCDataChannel | null) {
    if (!dc) {
      this.dc = null
      return
    }
    dc.binaryType = 'arraybuffer'
    dc.onmessage = (e) => {
      const v = new DataView(e.data)
      const op = v.getUint8(0)
      if (op === OP.CURSOR_POSITION) {
        this.onCursorPosition?.({ x: v.getUint16(3), y: v.getUint16(5) })
      } else if (op === OP.CURSOR_IMAGE) {
        const uri = URL.createObjectURL(new Blob([e.data.slice(11)], { type: 'image/png' }))
        this.onCursorImage?.({
          width: v.getUint16(3),
          height: v.getUint16(5),
          x: v.getUint16(7),
          y: v.getUint16(9),
          uri,
        })
      }
    }
    this.dc = dc
  }

  move(x: number, y: number) {
    this.send(OP.MOVE, [2, x], [2, y])
  }

  scroll(deltaX: number, deltaY: number, controlKey: boolean) {
    this.send(OP.SCROLL, [-2, deltaX], [-2, deltaY], [1, controlKey ? 1 : 0])
  }

  button(code: number, down: boolean) {
    this.send(down ? OP.BTN_DOWN : OP.BTN_UP, [4, code])
  }

  key(keysym: number, down: boolean) {
    this.send(down ? OP.KEY_DOWN : OP.KEY_UP, [4, keysym])
  }

  touch(phase: 'begin' | 'update' | 'end', id: number, x: number, y: number, pressure: number) {
    this.send(TOUCH_OP[phase], [4, id], [-4, x], [-4, y], [1, pressure])
  }

  // opcode, u16 payload length, the fields; dropped while the channel is not open
  private send(op: number, ...fields: Field[]) {
    if (this.dc?.readyState !== 'open') return
    const len = fields.reduce((n, [b]) => n + Math.abs(b), 0)
    const v = new DataView(new ArrayBuffer(3 + len))
    v.setUint8(0, op)
    v.setUint16(1, len)
    let o = 3
    for (const [b, val] of fields) {
      if (b === 1) v.setUint8(o, val)
      else if (b === 2) v.setUint16(o, val)
      else if (b === -2) v.setInt16(o, val)
      else if (b === 4) v.setUint32(o, val)
      else v.setInt32(o, val)
      o += Math.abs(b)
    }
    this.dc.send(v.buffer)
  }
}
