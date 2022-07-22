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

  public move(x: number, y: number) {
    this._connection.websocket.send(EVENT.CONTROL_MOVE, { x, y } as message.ControlPos)
  }

  public scroll(x: number, y: number) {
    this._connection.websocket.send(EVENT.CONTROL_SCROLL, { x, y } as message.ControlPos)
  }

  public buttonPress(code: number, x?: number, y?: number) {
    this._connection.websocket.send(EVENT.CONTROL_BUTTONPRESS, { code, x, y } as message.ControlButton)
  }

  public buttonDown(code: number, x?: number, y?: number) {
    this._connection.websocket.send(EVENT.CONTROL_BUTTONDOWN, { code, x, y } as message.ControlButton)
  }

  public buttonUp(code: number, x?: number, y?: number) {
    this._connection.websocket.send(EVENT.CONTROL_BUTTONUP, { code, x, y } as message.ControlButton)
  }

  public keyPress(keysym: number, x?: number, y?: number) {
    this._connection.websocket.send(EVENT.CONTROL_KEYPRESS, { keysym, x, y } as message.ControlKey)
  }

  public keyDown(keysym: number, x?: number, y?: number) {
    this._connection.websocket.send(EVENT.CONTROL_KEYDOWN, { keysym, x, y } as message.ControlKey)
  }

  public keyUp(keysym: number, x?: number, y?: number) {
    this._connection.websocket.send(EVENT.CONTROL_KEYUP, { keysym, x, y } as message.ControlKey)
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
