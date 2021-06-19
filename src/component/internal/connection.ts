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

    let webSocketStatus = 'disconnected'
    let webRTCStatus = 'disconnected'

    // websocket
    this.websocket.on('connecting', () => {
      webSocketStatus = 'connecting'
      if (this._state.status !== 'connecting') {
        Vue.set(this._state, 'status', 'connecting')
      }
    })
    this.websocket.on('connected', () => {
      webSocketStatus = 'connected'
      if (webSocketStatus == 'connected' && webRTCStatus == 'connected') {
        Vue.set(this._state, 'status', 'connected')
      }
    })
    this.websocket.on('disconnected', () => {
      webSocketStatus = 'disconnected'
      if (this._state.status !== 'disconnected') {
        Vue.set(this._state, 'status', 'disconnected')
      }
    })

    // webrtc
    this.webrtc.on('connecting', () => {
      webRTCStatus = 'connecting'
      if (this._state.status !== 'connecting') {
        Vue.set(this._state, 'status', 'connecting')
      }
    })
    this.webrtc.on('connected', () => {
      webRTCStatus = 'connected'
      if (webSocketStatus == 'connected' && webRTCStatus == 'connected') {
        Vue.set(this._state, 'status', 'connected')
      }
    })
    this.webrtc.on('disconnected', () => {
      webRTCStatus = 'disconnected'
      if (this._state.status !== 'disconnected') {
        Vue.set(this._state, 'status', 'disconnected')
      }
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
    Vue.set(this._state, 'status', 'disconnected')
  }
}
