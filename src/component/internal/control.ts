import Vue from 'vue'

import * as EVENT from '../types/events'
import * as message from '../types/messages'

import EventEmitter from 'eventemitter3'
import { NekoConnection } from './connection'
import { Control } from '../types/state'

export interface NekoControlEvents {
  ['overlay.click']: () => void
  ['overlay.contextmenu']: () => void
}

export class NekoControl extends EventEmitter<NekoControlEvents> {
  // eslint-disable-next-line
  constructor(
    private readonly _connection: NekoConnection,
    private readonly _state: Control,
  ) {
    super()
  }

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

  public keyPress(keysym: number) {
    this._connection.websocket.send(EVENT.CONTROL_KEYPRESS, { keysym } as message.ControlKey)
  }

  public keyDown(keysym: number) {
    this._connection.websocket.send(EVENT.CONTROL_KEYDOWN, { keysym } as message.ControlKey)
  }

  public keyUp(keysym: number) {
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
