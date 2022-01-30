import Vue from 'vue'

import * as EVENT from '../types/events'
import * as message from '../types/messages'

import { NekoConnection } from './connection'
import { Control } from '../types/state'

export class NekoControl {
  // eslint-disable-next-line
  constructor(
    private readonly _connection: NekoConnection,
    private readonly _state: Control,
  ) {}

  public lock() {
    Vue.set(this._state, 'locked', true)
  }

  public unlock() {
    Vue.set(this._state, 'locked', false)
  }

  public request() {
    this._connection.websocket.send(EVENT.CONTROL_REQUEST)
  }

  public release() {
    this._connection.websocket.send(EVENT.CONTROL_RELEASE)
  }

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
