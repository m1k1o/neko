import { Connection } from '../../types/state'

import { NekoWebSocket } from '../websocket'

import { ReconnectorAbstract } from '.'

export class WebsocketReconnector extends ReconnectorAbstract {
  private _state: Connection
  private _websocket: NekoWebSocket

  private _onConnectHandle: () => void
  private _onDisconnectHandle: (error?: Error) => void

  constructor(state: Connection, websocket: NekoWebSocket) {
    super()

    this._state = state
    this._websocket = websocket

    this._onConnectHandle = () => this.emit('connect')
    this._websocket.on('connected', this._onConnectHandle)

    this._onDisconnectHandle = (error?: Error) => this.emit('disconnect', error)
    this._websocket.on('disconnected', this._onDisconnectHandle)
  }

  public get connected() {
    return this._websocket.connected
  }

  public connect() {
    if (this._websocket.connected) {
      this._websocket.disconnect()
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
    this._websocket.disconnect()
  }

  public destroy() {
    this._websocket.off('connected', this._onConnectHandle)
    this._websocket.off('disconnected', this._onDisconnectHandle)
  }
}
