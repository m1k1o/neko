import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import * as EVENT from '../types/events'

import { NekoWebSocket } from './websocket'
import { NekoWebRTC, WebRTCStats } from './webrtc'
import { Connection } from '../types/state'

export interface NekoConnectionEvents {
  disconnect: (error?: Error) => void
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

    let webrtcCongestion: number = 0
    this.webrtc.on('stats', (stats: WebRTCStats) => {
      Vue.set(this._state.webrtc, 'stats', stats)

      // if automatic quality adjusting is turned off
      if (!this._state.webrtc.auto) return

      // if there are no or just one quality, no switching can be done
      if (this._state.webrtc.videos.length <= 1) return

      // current quality is not known
      if (this._state.webrtc.video == null) return

      // check if video is not playing
      if (stats.fps) {
        webrtcCongestion = 0
        return
      }

      // try to downgrade quality if it happend many times
      if (++webrtcCongestion >= 3) {
        const index = this._state.webrtc.videos.indexOf(this._state.webrtc.video)

        // edge case: current quality is not in qualities list
        if (index === -1) return

        // current quality is the lowest one
        if (index + 1 == this._state.webrtc.videos.length) return

        // downgrade video quality
        this.setVideo(this._state.webrtc.videos[index + 1])
        webrtcCongestion = 0
      }
    })
  }

  public setUrl(url: string) {
    this._url = url.replace(/^http/, 'ws').replace(/\/+$/, '') + '/api/ws'
  }

  public setToken(token: string) {
    this._token = token
  }

  public setVideo(video: string) {
    if (!this._state.webrtc.videos.includes(video)) {
      throw new Error('video id not found')
    }

    this.websocket.send(EVENT.SIGNAL_VIDEO, { video: video })
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
    this.webrtc.disconnect()
    this.websocket.disconnect()
    Vue.set(this._state, 'status', 'disconnected')
    this.emit('disconnect')
  }
}
