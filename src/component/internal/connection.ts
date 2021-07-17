import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import * as EVENT from '../types/events'

import { NekoWebSocket } from './websocket'
import { NekoWebRTC } from './webrtc'
import { Connection, WebRTCStats } from '../types/state'
import { Reconnecter, ReconnecterAbstract } from '../utils/reconnecter'

const WEBRTC_RECONN_MAX_LOSS = 25
const WEBRTC_RECONN_FAILED_ATTEMPTS = 5

export interface NekoConnectionEvents {
  disconnect: (error?: Error) => void
}

class WebsocketReconnecter extends ReconnecterAbstract {
  private _state: Connection
  private _websocket: NekoWebSocket

  constructor(state: Connection, websocket: NekoWebSocket) {
    super()

    this._state = state

    // TODO: Unmount.
    this._websocket = websocket
    this._websocket.on('connected', () => this.emit('connect'))
    this._websocket.on('disconnected', (error) => this.emit('disconnect', error))
  }

  public get connected() {
    return this._websocket.connected
  }

  public connect() {
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
}

class WebrtcReconnecter extends ReconnecterAbstract {
  private _state: Connection
  private _websocket: NekoWebSocket
  private _webrtc: NekoWebRTC

  constructor(state: Connection, websocket: NekoWebSocket, webrtc: NekoWebRTC) {
    super()

    this._state = state
    this._websocket = websocket

    // TODO: Unmount.
    this._webrtc = webrtc
    this._webrtc.on('connected', () => this.emit('connect'))
    this._webrtc.on('disconnected', (error) => this.emit('disconnect', error))
  }

  public get connected() {
    return this._webrtc.connected
  }

  public connect() {
    if (this._websocket.connected) {
      this._websocket.send(EVENT.SIGNAL_REQUEST, { video: this._state.webrtc.video })
    }
  }

  public disconnect() {
    this._webrtc.disconnect()
  }
}

export class NekoConnection extends EventEmitter<NekoConnectionEvents> {
  private _state: Connection

  public websocket = new NekoWebSocket()
  public _websocket_reconn: Reconnecter

  public webrtc = new NekoWebRTC()
  public _webrtc_reconn: Reconnecter

  constructor(state: Connection) {
    super()

    this._state = state
    this._websocket_reconn = new Reconnecter(new WebsocketReconnecter(state, this.websocket), state.websocket.config)
    this._webrtc_reconn = new Reconnecter(
      new WebrtcReconnecter(state, this.websocket, this.webrtc),
      state.webrtc.config,
    )

    // initial state
    Vue.set(this._state, 'type', 'screencast')

    // websocket
    this._websocket_reconn.on('connect', () => {
      if (this.websocket.connected && this.webrtc.connected) {
        Vue.set(this._state, 'status', 'connected')
      }

      if (!this.webrtc.connected) {
        this._webrtc_reconn.connect()
      }
    })
    this._websocket_reconn.on('disconnect', () => {
      if (this._state.status === 'connected' && this.activated) {
        Vue.set(this._state, 'status', 'connecting')
      }
    })
    this._websocket_reconn.on('close', (error) => {
      this.disconnect(error)
    })

    // webrtc
    this._webrtc_reconn.on('connect', () => {
      if (this.websocket.connected && this.webrtc.connected) {
        Vue.set(this._state, 'status', 'connected')
      }

      Vue.set(this._state, 'type', 'webrtc')
    })
    this._webrtc_reconn.on('disconnect', () => {
      if (this._state.status === 'connected' && this.activated) {
        Vue.set(this._state, 'status', 'connecting')
      }

      Vue.set(this._state, 'type', 'screencast')
    })
    this._webrtc_reconn.on('close', (error) => {
      this.disconnect(error)
    })

    let webrtcCongestion: number = 0
    this.webrtc.on('stats', (stats: WebRTCStats) => {
      Vue.set(this._state.webrtc, 'stats', stats)

      // if automatic quality adjusting is turned off
      if (!this._state.webrtc.auto || !this._webrtc_reconn.isOpen) return

      // if there are no or just one quality, no switching can be done
      if (this._state.webrtc.videos.length <= 1) return

      // current quality is not known
      if (this._state.webrtc.video == null) return

      // check if video is not playing smoothly
      if (stats.fps && stats.packetLoss < WEBRTC_RECONN_MAX_LOSS && !stats.muted) {
        webrtcCongestion = 0
        return
      }

      // try to downgrade quality if it happend many times
      if (++webrtcCongestion >= WEBRTC_RECONN_FAILED_ATTEMPTS) {
        webrtcCongestion = 0

        const quality = this._webrtcQualityDowngrade(this._state.webrtc.video)

        // downgrade if lower video quality exists
        if (quality && this.webrtc.connected) {
          this.setVideo(quality)
        }

        // try to perform ice restart, if available
        if (this.webrtc.open) {
          this.websocket.send(EVENT.SIGNAL_RESTART)
          return
        }

        // try to reconnect
        this._webrtc_reconn.reconnect()
      }
    })
  }

  public get activated() {
    return this._websocket_reconn.isOpen && this._webrtc_reconn.isOpen
  }

  public setVideo(video: string) {
    if (!this._state.webrtc.videos.includes(video)) {
      throw new Error('video id not found')
    }

    this.websocket.send(EVENT.SIGNAL_VIDEO, { video })
  }

  public connect(video?: string) {
    if (video) {
      if (!this._state.webrtc.videos.includes(video)) {
        throw new Error('video id not found')
      }

      Vue.set(this._state.webrtc, 'video', video)
    }

    Vue.set(this._state, 'status', 'connecting')
    this._webrtc_reconn.open(true)
    this._websocket_reconn.open()
  }

  public disconnect(error?: Error) {
    this._websocket_reconn.close()
    this._webrtc_reconn.close()

    Vue.set(this._state, 'status', 'disconnected')
    this.emit('disconnect', error)
  }

  _webrtcQualityDowngrade(quality: string): string | undefined {
    // get index of selected or surrent quality
    const index = this._state.webrtc.videos.indexOf(quality)

    // edge case: current quality is not in qualities list
    if (index === -1) return

    // current quality is the lowest one
    if (index + 1 == this._state.webrtc.videos.length) return

    // downgrade video quality
    return this._state.webrtc.videos[index + 1]
  }
}
