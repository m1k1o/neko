import Vue from 'vue'
import EventEmitter from 'eventemitter3'

import { NekoWebSocket } from './websocket'
import { NekoWebRTC, WebRTCStats } from './webrtc'
import { Connection } from '../types/state'

export interface NekoConnectionEvents {
  connecting: () => void
  connected: () => void
  disconnected: (error?: Error) => void
}

export class NekoConnection extends EventEmitter<NekoConnectionEvents> {
  private _url: string
  private _token: string
  private _state: Connection

  public websocket = new NekoWebSocket()
  public webrtc = new NekoWebRTC()

  constructor(state: Connection) {
    super()

    this._url = ''
    this._token = ''
    this._state = state

    // initial state
    Vue.set(this._state, 'type', 'webrtc')
    Vue.set(this._state, 'websocket', this.websocket.supported ? 'disconnected' : 'unavailable')
    Vue.set(this._state.webrtc, 'status', this.webrtc.supported ? 'disconnected' : 'unavailable')

    // websocket
    this.websocket.on('connecting', () => {
      Vue.set(this._state, 'websocket', 'connecting')
    })
    this.websocket.on('connected', () => {
      Vue.set(this._state, 'websocket', 'connected')
    })
    this.websocket.on('disconnected', () => {
      Vue.set(this._state, 'websocket', 'disconnected')
    })

    // webrtc
    this.webrtc.on('connecting', () => {
      Vue.set(this._state.webrtc, 'status', 'connecting')
    })
    this.webrtc.on('connected', () => {
      Vue.set(this._state.webrtc, 'status', 'connected')
    })
    this.webrtc.on('disconnected', () => {
      Vue.set(this._state.webrtc, 'status', 'disconnected')
    })
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
