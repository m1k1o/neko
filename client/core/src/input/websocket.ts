// Input over the main websocket, for transports without an input path of their own: the v3
// control/* events the server handles (server/internal/websocket/handler/control.go, payloads in
// server/pkg/types/message).
import type { InputChannel } from '../transport.ts'

const MOVE_COALESCE_MS = 16

export class WebSocketInput implements InputChannel {
  private readonly socket: { send(event: string, payload?: unknown): void }
  private pending: { x: number; y: number } | null = null
  private moveTimer = 0

  constructor(socket: { send(event: string, payload?: unknown): void }) {
    this.socket = socket
  }

  // one move per frame, the latest wins
  move(x: number, y: number) {
    this.pending = { x, y }
    if (!this.moveTimer) this.moveTimer = window.setTimeout(() => this.flush(), MOVE_COALESCE_MS)
  }

  scroll(deltaX: number, deltaY: number, controlKey: boolean) {
    this.flush()
    this.socket.send('control/scroll', { delta_x: deltaX, delta_y: deltaY, control_key: controlKey })
  }

  button(code: number, down: boolean) {
    this.flush() // a click lands where the pointer is now
    this.socket.send(down ? 'control/buttondown' : 'control/buttonup', { code })
  }

  key(keysym: number, down: boolean) {
    this.flush()
    this.socket.send(down ? 'control/keydown' : 'control/keyup', { keysym })
  }

  touch(phase: 'begin' | 'update' | 'end', id: number, x: number, y: number, pressure: number) {
    this.flush()
    this.socket.send(`control/touch${phase}`, { touch_id: id, x, y, pressure })
  }

  private flush() {
    clearTimeout(this.moveTimer)
    this.moveTimer = 0
    if (!this.pending) return
    this.socket.send('control/move', this.pending)
    this.pending = null
  }
}
