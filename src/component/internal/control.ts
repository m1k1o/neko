import * as EVENT from '../types/events'
import * as message from '../types/messages'

import { NekoConnection } from './connection'

export class NekoControl {
  // eslint-disable-next-line
  constructor(
    private readonly _connection: NekoConnection,
  ) {}

  public keypress(keysym: number) {
    this._connection.websocket.send(EVENT.CONTROL_KEYPRESS, { keysym } as message.ControlKey)
  }

  public keydown(keysym: number) {
    this._connection.websocket.send(EVENT.CONTROL_KEYDOWN, { keysym } as message.ControlKey)
  }

  public keyup(keysym: number) {
    this._connection.websocket.send(EVENT.CONTROL_KEYUP, { keysym } as message.ControlKey)
  }

  public cut() {
    this._connection.websocket.send(EVENT.CONTROL_CUT)
  }

  public copy() {
    this._connection.websocket.send(EVENT.CONTROL_COPY)
  }

  public paste() {
    this._connection.websocket.send(EVENT.CONTROL_PASTE)
  }

  public selectAll() {
    this._connection.websocket.send(EVENT.CONTROL_SELECT_ALL)
  }
}
