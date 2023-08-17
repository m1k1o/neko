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

  get useWebrtc() {
    // we want to use webrtc if we're connected and we're the host
    // because webrtc is faster and it doesn't request control
    // in contrast to the websocket
    return this._connection.webrtc.connected && this._state.is_host
  }

  get hasTouchEvents() {
    return this._state.touch_events
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
    if (this.useWebrtc) {
      this._connection.webrtc.send('mousemove', pos)
    } else {
      this._connection.websocket.send(EVENT.CONTROL_MOVE, pos as message.ControlPos)
    }
  }

  // TODO: rename pos to delta, and add a new pos parameter
  public scroll(pos: ControlPos) {
    if (this.useWebrtc) {
      this._connection.webrtc.send('wheel', pos)
    } else {
      this._connection.websocket.send(EVENT.CONTROL_SCROLL, pos as message.ControlPos)
    }
  }

  // buttonpress ensures that only one button is pressed at a time
  public buttonPress(code: number, pos?: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_BUTTONPRESS, { code, ...pos } as message.ControlButton)
  }

  public buttonDown(code: number, pos?: ControlPos) {
    if (this.useWebrtc) {
      if (pos) this._connection.webrtc.send('mousemove', pos)
      this._connection.webrtc.send('mousedown', { key: code })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_BUTTONDOWN, { code, ...pos } as message.ControlButton)
    }
  }

  public buttonUp(code: number, pos?: ControlPos) {
    if (this.useWebrtc) {
      if (pos) this._connection.webrtc.send('mousemove', pos)
      this._connection.webrtc.send('mouseup', { key: code })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_BUTTONUP, { code, ...pos } as message.ControlButton)
    }
  }

  // keypress ensures that only one key is pressed at a time
  public keyPress(keysym: number, pos?: ControlPos) {
    this._connection.websocket.send(EVENT.CONTROL_KEYPRESS, { keysym, ...pos } as message.ControlKey)
  }

  public keyDown(keysym: number, pos?: ControlPos) {
    if (this.useWebrtc) {
      if (pos) this._connection.webrtc.send('mousemove', pos)
      this._connection.webrtc.send('keydown', { key: keysym })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_KEYDOWN, { keysym, ...pos } as message.ControlKey)
    }
  }

  public keyUp(keysym: number, pos?: ControlPos) {
    if (this.useWebrtc) {
      if (pos) this._connection.webrtc.send('mousemove', pos)
      this._connection.webrtc.send('keyup', { key: keysym })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_KEYUP, { keysym, ...pos } as message.ControlKey)
    }
  }

  public touchBegin(touchId: number, pos: ControlPos, pressure: number) {
    if (this.useWebrtc) {
      this._connection.webrtc.send('touchbegin', { touchId, ...pos, pressure })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_TOUCHBEGIN, { touchId, ...pos, pressure } as message.ControlTouch)
    }
  }

  public touchUpdate(touchId: number, pos: ControlPos, pressure: number) {
    if (this.useWebrtc) {
      this._connection.webrtc.send('touchupdate', { touchId, ...pos, pressure })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_TOUCHUPDATE, { touchId, ...pos, pressure } as message.ControlTouch)
    }
  }

  public touchEnd(touchId: number, pos: ControlPos, pressure: number) {
    if (this.useWebrtc) {
      this._connection.webrtc.send('touchend', { touchId, ...pos, pressure })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_TOUCHEND, { touchId, ...pos, pressure } as message.ControlTouch)
    }
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
