import * as EVENT from '../types/events'
import type * as message from '../types/messages'

import EventEmitter from 'eventemitter3'
import type { NekoConnection } from './connection'
import type { Control } from '../types/state'

export interface NekoControlEvents {
  ['overlay.click']: (e: MouseEvent) => void
  ['overlay.contextmenu']: (e: MouseEvent) => void
}

export interface ControlPos {
  x: number
  y: number
}

export interface ControlScroll {
  delta_x: number
  delta_y: number
  control_key?: boolean
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

  get enabledTouchEvents() {
    return this._state.touch.enabled
  }

  get supportedTouchEvents() {
    return this._state.touch.supported
  }

  public lock() {
    this._state.locked = true // TODO: Vue.Set
  }

  public unlock() {
    this._state.locked = false // TODO: Vue.Set
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

  // TODO: add pos parameter
  public scroll(scroll: ControlScroll) {
    if (this.useWebrtc) {
      this._connection.webrtc.send('wheel', scroll)
    } else {
      this._connection.websocket.send(EVENT.CONTROL_SCROLL, scroll as message.ControlScroll)
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

  public touchBegin(touch_id: number, pos: ControlPos, pressure: number) {
    if (this.useWebrtc) {
      this._connection.webrtc.send('touchbegin', { touch_id, ...pos, pressure })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_TOUCHBEGIN, { touch_id, ...pos, pressure } as message.ControlTouch)
    }
  }

  public touchUpdate(touch_id: number, pos: ControlPos, pressure: number) {
    if (this.useWebrtc) {
      this._connection.webrtc.send('touchupdate', { touch_id, ...pos, pressure })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_TOUCHUPDATE, { touch_id, ...pos, pressure } as message.ControlTouch)
    }
  }

  public touchEnd(touch_id: number, pos: ControlPos, pressure: number) {
    if (this.useWebrtc) {
      this._connection.webrtc.send('touchend', { touch_id, ...pos, pressure })
    } else {
      this._connection.websocket.send(EVENT.CONTROL_TOUCHEND, { touch_id, ...pos, pressure } as message.ControlTouch)
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
