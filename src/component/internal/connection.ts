import EventEmitter from 'eventemitter3'

import { NekoWebSocket } from './websocket'
import { NekoWebRTC, WebRTCStats } from './webrtc'

export interface NekoConnectionEvents {
  connecting: () => void
  connected: () => void
  disconnected: (error?: Error) => void
}

export class NekoConnection extends EventEmitter<NekoConnectionEvents> {
  private _url: string
  private _token: string

  public websocket = new NekoWebSocket()
  public webrtc = new NekoWebRTC()

  constructor() {
    super()

    this._url = ''
    this._token = ''
  }

  public setUrl(url: string) {
    this._url = url.replace(/^http/, 'ws').replace(/\/+$/, '') + '/api/ws'
  }

  public setToken(token: string) {
    this._token = token
  }

  public async connect(): Promise<void> {
    let url = this._url
    if (this._token) {
      url += '?token=' + encodeURIComponent(this._token)
    }

    await this.websocket.connect(url)

    // TODO: connect to WebRTC
    //this.websocket.send(EVENT.SIGNAL_REQUEST, { video: video })
  }

  public disconnect() {
    this.websocket.disconnect()
  }
}
