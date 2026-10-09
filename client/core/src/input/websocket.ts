// Input over the main websocket, for transports without an input path of their own: the same
// opcodes as the data channel, sent as the v3 control/* events the server handles
// (server/internal/websocket/handler/control.go, payloads in server/pkg/types/message).
import type { InputChannel } from '../transport.ts'
import { OP } from '../types.ts'

const MOVE_COALESCE_MS = 16

const TOUCH: Record<number, string> = {
  [OP.TOUCH_BEGIN]: 'control/touchbegin',
  [OP.TOUCH_UPDATE]: 'control/touchupdate',
  [OP.TOUCH_END]: 'control/touchend',
}

export class WebSocketInput implements InputChannel {
  private readonly socket: { send(event: string, payload?: unknown): void }
  private move: { x: number; y: number } | null = null
  private moveTimer = 0

  constructor(socket: { send(event: string, payload?: unknown): void }) {
    this.socket = socket
  }

  send(op: number, ...fields: [number, number][]) {
    const [a, b, c, d] = fields.map(([, v]) => v)
    if (op === OP.MOVE) {
      // one move per frame, the latest wins
      this.move = { x: a, y: b }
      if (!this.moveTimer) this.moveTimer = window.setTimeout(() => this.flush(), MOVE_COALESCE_MS)
      return
    }
    this.flush() // a click lands where the pointer is now
    if (op === OP.SCROLL) this.socket.send('control/scroll', { delta_x: a, delta_y: b, control_key: !!c })
    else if (op === OP.KEY_DOWN) this.socket.send('control/keydown', { keysym: a })
    else if (op === OP.KEY_UP) this.socket.send('control/keyup', { keysym: a })
    else if (op === OP.BTN_DOWN) this.socket.send('control/buttondown', { code: a })
    else if (op === OP.BTN_UP) this.socket.send('control/buttonup', { code: a })
    else if (TOUCH[op]) this.socket.send(TOUCH[op], { touch_id: a, x: b, y: c, pressure: d })
  }

  private flush() {
    clearTimeout(this.moveTimer)
    this.moveTimer = 0
    if (!this.move) return
    this.socket.send('control/move', this.move)
    this.move = null
  }
}
