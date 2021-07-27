import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import * as EVENT from '../types/events'

import { NekoWebSocket } from './websocket'
import { NekoWebRTC } from './webrtc'
import { Connection, WebRTCStats } from '../types/state'

import { Reconnector } from './reconnector'
import { WebsocketReconnector } from './reconnector/websocket'
import { WebrtcReconnector } from './reconnector/webrtc'

const WEBRTC_RECONN_MAX_LOSS = 25
const WEBRTC_RECONN_FAILED_ATTEMPTS = 5

const WEBRTC_FALLBACK_TIMEOUT_MS = 750

export interface NekoConnectionEvents {
  close: (error?: Error) => void
}

export class NekoConnection extends EventEmitter<NekoConnectionEvents> {
  private _state: Connection

  public websocket = new NekoWebSocket()
  public webrtc = new NekoWebRTC()

  private _reconnector: {
    websocket: Reconnector
    webrtc: Reconnector
  }

  constructor(state: Connection) {
    super()

    this._state = state
    this._reconnector = {
      websocket: new Reconnector(new WebsocketReconnector(state, this.websocket), state.websocket.config),
      webrtc: new Reconnector(new WebrtcReconnector(state, this.websocket, this.webrtc), state.webrtc.config),
    }

    // websocket
    this._reconnector.websocket.on('connect', () => {
      if (this.websocket.connected && this.webrtc.connected) {
        Vue.set(this._state, 'status', 'connected')
      }

      if (!this.webrtc.connected) {
        this._reconnector.webrtc.connect()
      }
    })
    this._reconnector.websocket.on('disconnect', () => {
      if (this._state.status === 'connected' && this.activated) {
        Vue.set(this._state, 'status', 'connecting')
      }
    })
    this._reconnector.websocket.on('close', this.close.bind(this))

    // webrtc
    this._reconnector.webrtc.on('connect', () => {
      if (this.websocket.connected && this.webrtc.connected) {
        Vue.set(this._state, 'status', 'connected')
      }

      Vue.set(this._state, 'type', 'webrtc')
    })
    this._reconnector.webrtc.on('disconnect', () => {
      if (this._state.status === 'connected' && this.activated) {
        Vue.set(this._state, 'status', 'connecting')
      }

      Vue.set(this._state, 'type', 'fallback')
    })
    this._reconnector.webrtc.on('close', this.close.bind(this))

    let webrtcCongestion: number = 0
    let webrtcFallbackTimeout: number
    this.webrtc.on('stats', (stats: WebRTCStats) => {
      Vue.set(this._state.webrtc, 'stats', stats)

      // if automatic quality adjusting is turned off
      if (!this._state.webrtc.auto || !this._reconnector.webrtc.isOpen) return

      // if there are no or just one quality, no switching can be done
      if (this._state.webrtc.videos.length <= 1) return

      // current quality is not known
      if (this._state.webrtc.video == null) return

      // check if video is not playing smoothly
      if (stats.fps && stats.packetLoss < WEBRTC_RECONN_MAX_LOSS && !stats.muted) {
        if (webrtcFallbackTimeout) {
          window.clearTimeout(webrtcFallbackTimeout)
        }

        if (this._state.type === 'fallback') {
          Vue.set(this._state, 'type', 'webrtc')
        }

        webrtcCongestion = 0
        return
      }

      // try to downgrade quality if it happend many times
      if (++webrtcCongestion >= WEBRTC_RECONN_FAILED_ATTEMPTS) {
        webrtcFallbackTimeout = window.setTimeout(() => {
          if (this._state.type === 'webrtc') {
            Vue.set(this._state, 'type', 'fallback')
          }
        }, WEBRTC_FALLBACK_TIMEOUT_MS)

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

        // try to reconnect webrtc
        this._reconnector.webrtc.reconnect()
      }
    })
  }

  public get activated() {
    // check if every reconnecter is open
    return Object.values(this._reconnector).every((r) => r.isOpen)
  }

  public setVideo(video: string) {
    if (!this._state.webrtc.videos.includes(video)) {
      throw new Error('video id not found')
    }

    this.websocket.send(EVENT.SIGNAL_VIDEO, { video })
  }

  public open(video?: string) {
    if (video) {
      if (!this._state.webrtc.videos.includes(video)) {
        throw new Error('video id not found')
      }

      Vue.set(this._state.webrtc, 'video', video)
    }

    Vue.set(this._state, 'type', 'fallback')
    Vue.set(this._state, 'status', 'connecting')

    // open all reconnecters
    Object.values(this._reconnector).forEach((r) => r.open(true))

    this._reconnector.websocket.connect()
  }

  public close(error?: Error) {
    if (this.activated) {
      Vue.set(this._state, 'type', 'none')
      Vue.set(this._state, 'status', 'disconnected')

      this.emit('close', error)
    }

    // close all reconnecters
    Object.values(this._reconnector).forEach((r) => r.close())
  }

  public destroy() {
    // destroy all reconnecters
    Object.values(this._reconnector).forEach((r) => r.destroy())

    Vue.set(this._state, 'type', 'none')
    Vue.set(this._state, 'status', 'disconnected')
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
