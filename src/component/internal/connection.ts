import EventEmitter from 'eventemitter3'
import * as EVENT from '../types/events'
import type * as webrtcTypes from '../types/webrtc'

import { NekoWebSocket } from './websocket'
import { NekoLoggerFactory } from './logger'
import { NekoWebRTC } from './webrtc'
import type { Connection, WebRTCStats } from '../types/state'

import { Reconnector } from './reconnector'
import { WebsocketReconnector } from './reconnector/websocket'
import { WebrtcReconnector } from './reconnector/webrtc'
import type { Logger } from '../utils/logger'

const WEBRTC_RECONN_MAX_LOSS = 25
const WEBRTC_RECONN_FAILED_ATTEMPTS = 5

const WEBRTC_FALLBACK_TIMEOUT_MS = 750

export interface NekoConnectionEvents {
  close: (error?: Error) => void
}

export class NekoConnection extends EventEmitter<NekoConnectionEvents> {
  private _open = false
  private _closing = false
  private _peerRequest?: webrtcTypes.PeerRequest

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

  private _webrtcStatsHandle: (stats: WebRTCStats) => void
  private _webrtcStableHandle: (isStable: boolean) => void
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
      this._state.websocket.connected = this.websocket.connected // TODO: Vue.Set
      this._state.webrtc.connected = this.webrtc.connected // TODO: Vue.Set

      if (this._state.status !== 'connected' && this.websocket.connected && this.webrtc.connected) {
        this._state.status = 'connected' // TODO: Vue.Set
      }

      if (this.websocket.connected && !this.webrtc.connected) {
        // if custom peer request is set, send custom peer request
        if (this._peerRequest) {
          this.websocket.send(EVENT.SIGNAL_REQUEST, this._peerRequest)
          this._peerRequest = undefined
        } else {
          // otherwise use reconnectors connect method
          this._reconnector.webrtc.connect()
        }
      }
    }

    this._onDisconnectHandle = () => {
      this._state.websocket.connected = this.websocket.connected // TODO: Vue.Set
      this._state.webrtc.connected = this.webrtc.connected // TODO: Vue.Set

      if (this._state.webrtc.stable && !this.webrtc.connected) {
        this._state.webrtc.stable = false // TODO: Vue.Set
      }

      if (this._state.status === 'connected' && this.activated) {
        this._state.status = 'connecting' // TODO: Vue.Set
      }
    }

    this._onCloseHandle = this.close.bind(this)

    // bind events to all reconnectors
    Object.values(this._reconnector).forEach((r) => {
      r.on('connect', this._onConnectHandle)
      r.on('disconnect', this._onDisconnectHandle)
      r.on('close', this._onCloseHandle)
    })

    // synchronize webrtc stats with global state
    this._webrtcStatsHandle = (stats: WebRTCStats) => {
      this._state.webrtc.stats = stats // TODO: Vue.Set
    }
    this.webrtc.on('stats', this._webrtcStatsHandle)

    // synchronize webrtc stable with global state
    this._webrtcStableHandle = (isStable: boolean) => {
      this._state.webrtc.stable = isStable // TODO: Vue.Set
    }
    this.webrtc.on('stable', this._webrtcStableHandle)

    //
    // TODO: Use server side congestion control.
    //

    let webrtcCongestion: number = 0
    let webrtcFallbackTimeout: number

    this._webrtcCongestionControlHandle = (stats: WebRTCStats) => {
      // if automatic quality adjusting is turned off
      if (this._state.webrtc.video.auto) return

      // when connection is paused or video disabled, 0fps and muted track is expected
      if (stats.paused || this._state.webrtc.video.disabled) return

      // if automatic quality adjusting is turned off
      if (!this._reconnector.webrtc.isOpen) return

      // if there are no or just one quality, no switching can be done
      if (this._state.webrtc.videos.length <= 1) return

      // current quality is not known
      if (this._state.webrtc.video.id == '') return

      // check if video is not playing smoothly
      if (stats.fps && stats.packetLoss < WEBRTC_RECONN_MAX_LOSS && !stats.muted) {
        if (webrtcFallbackTimeout) {
          window.clearTimeout(webrtcFallbackTimeout)
        }

        this._state.webrtc.connected = true // TODO: Vue.Set
        webrtcCongestion = 0
        return
      }

      // try to downgrade quality if it happend many times
      if (++webrtcCongestion >= WEBRTC_RECONN_FAILED_ATTEMPTS) {
        webrtcFallbackTimeout = window.setTimeout(() => {
          this._state.webrtc.connected = false // TODO: Vue.Set
        }, WEBRTC_FALLBACK_TIMEOUT_MS)

        webrtcCongestion = 0

        const quality = this._webrtcQualityDowngrade(this._state.webrtc.video.id)

        // downgrade if lower video quality exists
        if (quality && this.webrtc.connected) {
          this.websocket.send(EVENT.SIGNAL_VIDEO, { video: quality })
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

  public getLogger(scope?: string): Logger {
    return this.logger.new(scope)
  }

  public open(peerRequest?: webrtcTypes.PeerRequest) {
    if (this._open) {
      throw new Error('connection already open')
    }

    this._open = true
    this._peerRequest = peerRequest

    this._state.status = 'connecting' // TODO: Vue.Set

    // open all reconnectors with deferred connection
    Object.values(this._reconnector).forEach((r) => r.open(true))

    this._reconnector.websocket.connect()
  }

  public close(error?: Error) {
    // we want to make sure that close event is only emitted once
    // and is not intercepted by any other close event
    const active = this._open && !this._closing

    if (active) {
      // set state to disconnected
      this._state.websocket.connected = false // TODO: Vue.Set
      this._state.webrtc.connected = false // TODO: Vue.Set
      this._state.status = 'disconnected' // TODO: Vue.Set
      this._closing = true
    }

    // close all reconnectors
    Object.values(this._reconnector).forEach((r) => r.close())

    if (active) {
      this._open = false
      this._closing = false
      this.emit('close', error)
    }
  }

  public destroy() {
    this.logger.destroy()

    this.webrtc.off('stats', this._webrtcStatsHandle)
    this.webrtc.off('stable', this._webrtcStableHandle)
    // TODO: Use server side congestion control.
    this.webrtc.off('stats', this._webrtcCongestionControlHandle)

    // unbind events from all reconnectors
    Object.values(this._reconnector).forEach((r) => {
      r.off('connect', this._onConnectHandle)
      r.off('disconnect', this._onDisconnectHandle)
      r.off('close', this._onCloseHandle)
    })

    // destroy all reconnectors
    Object.values(this._reconnector).forEach((r) => r.destroy())

    // set state to disconnected
    this._state.websocket.connected = false // TODO: Vue.Set
    this._state.webrtc.connected = false // TODO: Vue.Set
    this._state.status = 'disconnected' // TODO: Vue.Set
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
