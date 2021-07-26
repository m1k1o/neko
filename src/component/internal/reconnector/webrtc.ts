import * as EVENT from '../../types/events'
import { Connection } from '../../types/state'

import { NekoWebSocket } from '../websocket'
import { NekoWebRTC } from '../webrtc'

import { ReconnectorAbstract } from '.'

export class WebrtcReconnector extends ReconnectorAbstract {
  private _state: Connection
  private _websocket: NekoWebSocket
  private _webrtc: NekoWebRTC

  private _onConnectHandle: () => void
  private _onDisconnectHandle: (error?: Error) => void

  constructor(state: Connection, websocket: NekoWebSocket, webrtc: NekoWebRTC) {
    super()

    this._state = state
    this._websocket = websocket
    this._webrtc = webrtc

    this._onConnectHandle = () => this.emit('connect')
    this._webrtc.on('connected', this._onConnectHandle)

    this._onDisconnectHandle = (error?: Error) => this.emit('disconnect', error)
    this._webrtc.on('disconnected', this._onDisconnectHandle)
  }

  public get connected() {
    return this._webrtc.connected
  }

  public connect() {
    if (this._webrtc.connected) {
      this._webrtc.disconnect()
    }

    if (this._websocket.connected) {
      this._websocket.send(EVENT.SIGNAL_REQUEST, { video: this._state.webrtc.video })
    }
  }

  public disconnect() {
    this._webrtc.disconnect()
  }

  public destroy() {
    this._webrtc.off('connected', this._onConnectHandle)
    this._webrtc.off('disconnected', this._onDisconnectHandle)
  }
}
