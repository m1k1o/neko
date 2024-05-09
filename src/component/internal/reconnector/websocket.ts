import type { Connection } from '../../types/state'

import type { NekoWebSocket } from '../websocket'

import { ReconnectorAbstract } from '.'

export class WebsocketReconnector extends ReconnectorAbstract {
  private _onConnectHandle: () => void
  private _onDisconnectHandle: (error?: Error) => void

  // eslint-disable-next-line
  constructor(
    private readonly _state: Connection,
    private readonly _websocket: NekoWebSocket,
  ) {
    super()

    this._onConnectHandle = () => this.emit('connect')
    this._websocket.on('connected', this._onConnectHandle)

    this._onDisconnectHandle = (error?: Error) => this.emit('disconnect', error)
    this._websocket.on('disconnected', this._onDisconnectHandle)
  }

  public get connected() {
    return this._websocket.connected
  }

  public connect() {
    if (!this._websocket.supported) return

    if (this._websocket.connected) {
      this._websocket.disconnect('connection replaced')
    }

    let url = this._state.url
    url = url.replace(/^http/, 'ws').replace(/\/+$/, '') + '/api/ws'

    const token = this._state.token
    if (token) {
      url += '?token=' + encodeURIComponent(token)
    }

    this._websocket.connect(url)
  }

  public disconnect() {
    this._websocket.disconnect('manual disconnect')
  }

  public destroy() {
    this._websocket.off('connected', this._onConnectHandle)
    this._websocket.off('disconnected', this._onDisconnectHandle)
  }
}
