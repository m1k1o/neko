// Input over the WebRTC data channel: big-endian binary messages (server/internal/webrtc/payload),
// and the host's cursor coming back the same way.
import type { InputChannel } from '../transport.ts'
import { OP, type CursorImage } from '../types.ts'

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

  // dropped while the channel is not open
  send(op: number, ...fields: [number, number][]) {
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
