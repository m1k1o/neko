import Vue from 'vue'

import * as EVENT from '../types/events'
import * as message from '../types/messages'

import EventEmitter from 'eventemitter3'
import { NekoConnection } from './connection'
import { Control } from '../types/state'

export interface NekoControlEvents {
  ['overlay.click']: (e: MouseEvent) => void
  ['overlay.contextmenu']: (e: MouseEvent) => void
}

export interface ControlPos {
  x: number
  y: number
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

  public move(pos: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_MOVE, pos as message.ControlPos)
  }

  public scroll(pos: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_SCROLL, pos as message.ControlPos)
  }

  public buttonPress(code: number, pos?: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_BUTTONPRESS, { code, ...pos } as message.ControlButton)
  }

  public buttonDown(code: number, pos?: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_BUTTONDOWN, { code, ...pos } as message.ControlButton)
  }

  public buttonUp(code: number, pos?: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_BUTTONUP, { code, ...pos } as message.ControlButton)
  }

  public keyPress(keysym: number, pos?: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_KEYPRESS, { keysym, ...pos } as message.ControlKey)
  }

  public keyDown(keysym: number, pos?: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_KEYDOWN, { keysym, ...pos } as message.ControlKey)
  }

  public keyUp(keysym: number, pos?: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_KEYUP, { keysym, ...pos } as message.ControlKey)
  }

  public cut() {
    this._connection.websocket.send(EVENT.CONTROL_CUT)
  }

  public copy() {
    this._connection.websocket.send(EVENT.CONTROL_COPY)
  }

  public paste(text?: string) {
    this._connection.websocket.send(EVENT.CONTROL_PASTE, { text } as message.ClipboardData)
  }

  public selectAll() {
    this._connection.websocket.send(EVENT.CONTROL_SELECT_ALL)
  }
}
