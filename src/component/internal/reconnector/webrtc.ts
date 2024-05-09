import * as EVENT from '../../types/events'
import type { Connection } from '../../types/state'

import type { NekoWebSocket } from '../websocket'
import type { NekoWebRTC } from '../webrtc'

import { ReconnectorAbstract } from '.'

export class WebrtcReconnector extends ReconnectorAbstract {
  private _onConnectHandle: () => void
  private _onDisconnectHandle: (error?: Error) => void

  // eslint-disable-next-line
  constructor(
    private readonly _state: Connection,
    private readonly _websocket: NekoWebSocket,
    private readonly _webrtc: NekoWebRTC,
  ) {
    super()

    this._onConnectHandle = () => this.emit('connect')
    this._webrtc.on('connected', this._onConnectHandle)

    this._onDisconnectHandle = (error?: Error) => this.emit('disconnect', error)
    this._webrtc.on('disconnected', this._onDisconnectHandle)
  }

  public get connected() {
    return this._webrtc.connected
  }

  public connect() {
    if (!this._webrtc.supported) return

    if (this._webrtc.connected) {
      this._webrtc.disconnect()
    }

    if (this._websocket.connected) {
      // use requests from state to connect with selected values

      let selector = null
      if (this._state.webrtc.video.id) {
        selector = {
          id: this._state.webrtc.video.id,
          type: 'exact',
        }
      }

      this._websocket.send(EVENT.SIGNAL_REQUEST, {
        video: {
          disabled: this._state.webrtc.video.disabled,
          selector,
          auto: this._state.webrtc.video.auto,
        },
        audio: {
          disabled: this._state.webrtc.audio.disabled,
        },
      })
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
