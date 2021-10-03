import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import * as EVENT from '../types/events'

import { NekoWebSocket } from './websocket'
import { NekoLoggerFactory } from './logger'
import { NekoWebRTC } from './webrtc'
import { Connection, WebRTCStats } from '../types/state'

import { Reconnector } from './reconnector'
import { WebsocketReconnector } from './reconnector/websocket'
import { WebrtcReconnector } from './reconnector/webrtc'
import { Logger } from '../utils/logger'

const WEBRTC_RECONN_MAX_LOSS = 25
const WEBRTC_RECONN_FAILED_ATTEMPTS = 5

const WEBRTC_FALLBACK_TIMEOUT_MS = 750

export interface NekoConnectionEvents {
  close: (error?: Error) => void
}

export class NekoConnection extends EventEmitter<NekoConnectionEvents> {
  private _open = false

  public websocket = new NekoWebSocket()
  public logger = new NekoLoggerFactory(this.websocket)
  public webrtc = new NekoWebRTC(this.logger.new('webrtc'))

  private _reconnector: {
    websocket: Reconnector
    webrtc: Reconnector
  }

  private _onConnectHandle: () => void
  private _onDisconnectHandle: () => void
  private _onCloseHandle: (error?: Error) => void

  private _webrtcCongestionControlHandle: (stats: WebRTCStats) => void

  // eslint-disable-next-line
  constructor(
    private readonly _state: Connection,
  ) {
    super()

    this._reconnector = {
      websocket: new Reconnector(new WebsocketReconnector(_state, this.websocket), _state.websocket.config),
      webrtc: new Reconnector(new WebrtcReconnector(_state, this.websocket, this.webrtc), _state.webrtc.config),
    }

    this._onConnectHandle = () => {
      Vue.set(this._state.websocket, 'connected', this.websocket.connected)
      Vue.set(this._state.webrtc, 'connected', this.webrtc.connected)

      if (this._state.status !== 'connected' && this.websocket.connected && this.webrtc.connected) {
        Vue.set(this._state, 'status', 'connected')
      }

      if (this.websocket.connected && !this.webrtc.connected) {
        this._reconnector.webrtc.connect()
      }
    }

    this._onDisconnectHandle = () => {
      Vue.set(this._state.websocket, 'connected', this.websocket.connected)
      Vue.set(this._state.webrtc, 'connected', this.webrtc.connected)

      if (this._state.status === 'connected' && this.activated) {
        Vue.set(this._state, 'status', 'connecting')
      }
    }

    this._onCloseHandle = this.close.bind(this)

    // bind events to all reconnecters
    Object.values(this._reconnector).forEach((r) => {
      r.on('connect', this._onConnectHandle)
      r.on('disconnect', this._onDisconnectHandle)
      r.on('close', this._onCloseHandle)
    })

    //
    // TODO: Use server side congestion control.
    //

    let webrtcCongestion: number = 0
    let webrtcFallbackTimeout: number

    this._webrtcCongestionControlHandle = (stats: WebRTCStats) => {
      Vue.set(this._state.webrtc, 'stats', stats)

      // if automatic quality adjusting is turned off
      if (!this._reconnector.webrtc.isOpen) return

      // if there are no or just one quality, no switching can be done
      if (this._state.webrtc.videos.length <= 1) return

      // current quality is not known
      if (this._state.webrtc.video == null) return

      // check if video is not playing smoothly
      if (stats.fps && stats.packetLoss < WEBRTC_RECONN_MAX_LOSS && !stats.muted) {
        if (webrtcFallbackTimeout) {
          window.clearTimeout(webrtcFallbackTimeout)
        }

        Vue.set(this._state.webrtc, 'connected', true)
        webrtcCongestion = 0
        return
      }

      // try to downgrade quality if it happend many times
      if (++webrtcCongestion >= WEBRTC_RECONN_FAILED_ATTEMPTS) {
        webrtcFallbackTimeout = window.setTimeout(() => {
          Vue.set(this._state.webrtc, 'connected', false)
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
    }

    this.webrtc.on('stats', this._webrtcCongestionControlHandle)
  }

  public get activated() {
    // check if every reconnecter is open
    return Object.values(this._reconnector).every((r) => r.isOpen)
  }

  public reloadConfigs() {
    this._reconnector.websocket.config = this._state.websocket.config
    this._reconnector.webrtc.config = this._state.webrtc.config
  }

  public setVideo(video: string) {
    if (!this._state.webrtc.videos.includes(video)) {
      throw new Error('video id not found')
    }

    this.websocket.send(EVENT.SIGNAL_VIDEO, { video })
  }

  public getLogger(scope?: string): Logger {
    return this.logger.new(scope)
  }

  public open(video?: string) {
    if (this._open) {
      throw new Error('connection already open')
    }

    this._open = true

    if (video) {
      if (!this._state.webrtc.videos.includes(video)) {
        throw new Error('video id not found')
      }

      Vue.set(this._state.webrtc, 'video', video)
    }

    Vue.set(this._state, 'status', 'connecting')

    // open all reconnecters
    Object.values(this._reconnector).forEach((r) => r.open(true))

    this._reconnector.websocket.connect()
  }

  public close(error?: Error) {
    if (this._open) {
      this._open = false

      Vue.set(this._state.websocket, 'connected', false)
      Vue.set(this._state.webrtc, 'connected', false)
      Vue.set(this._state, 'status', 'disconnected')

      this.emit('close', error)
    }

    // close all reconnecters
    Object.values(this._reconnector).forEach((r) => r.close())
  }

  public destroy() {
    this.logger.destroy()

    // TODO: Use server side congestion control.
    this.webrtc.off('stats', this._webrtcCongestionControlHandle)

    // unbind events from all reconnecters
    Object.values(this._reconnector).forEach((r) => {
      r.off('connect', this._onConnectHandle)
      r.off('disconnect', this._onDisconnectHandle)
      r.off('close', this._onCloseHandle)
    })

    // destroy all reconnecters
    Object.values(this._reconnector).forEach((r) => r.destroy())

    Vue.set(this._state.websocket, 'connected', false)
    Vue.set(this._state.webrtc, 'connected', false)
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
