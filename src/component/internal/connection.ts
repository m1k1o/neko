import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import { Logger } from '../utils/logger'
import * as EVENT from '../types/events'

import { NekoWebSocket } from './websocket'
import { NekoWebRTC, WebRTCStats } from './webrtc'
import { Connection } from '../types/state'

const websocketTimer = 1000
const webrtcTimer = 10000

export interface NekoConnectionEvents {
  disconnect: (error?: Error) => void
}

export class NekoConnection extends EventEmitter<NekoConnectionEvents> {
  private _shouldReconnect = false
  private _url: string
  private _token: string
  private _state: Connection
  private _log: Logger

  public websocket = new NekoWebSocket()
  public webrtc = new NekoWebRTC()

  constructor(state: Connection) {
    super()

    this._url = ''
    this._token = ''
    this._log = new Logger('connection')
    this._state = state

    // initial state
    Vue.set(this._state, 'type', 'webrtc')

    // websocket
    this.websocket.on('connected', () => {
      if (this.websocket.connected && this.webrtc.connected) {
        Vue.set(this._state, 'status', 'connected')
      }
    })
    this.websocket.on('disconnected', () => {
      if (this._state.status === 'connected') {
        Vue.set(this._state, 'status', 'disconnected')
      }

      this._websocketReconnect()
    })

    // webrtc
    this.webrtc.on('connected', () => {
      if (this.websocket.connected && this.webrtc.connected) {
        Vue.set(this._state, 'status', 'connected')
      }
    })
    this.webrtc.on('disconnected', () => {
      if (this._state.status === 'connected') {
        Vue.set(this._state, 'status', 'disconnected')
      }

      this._webrtcReconnect()
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

  public async connect(video?: string): Promise<void> {
    try {
      await this._websocketConnect()

      if (video && !this._state.webrtc.videos.includes(video)) {
        throw new Error('video id not found')
      }

      await this._webrtcConnect(video)

      this._shouldReconnect = true
    } catch (e) {
      this.disconnect()
      throw e
    }
  }

  public disconnect() {
    this._shouldReconnect = false

    this.webrtc.disconnect()
    this.websocket.disconnect()

    Vue.set(this._state, 'status', 'disconnected')
    this.emit('disconnect')
  }

  async _websocketConnect() {
    Vue.set(this._state, 'status', 'connecting')

    let url = this._url
    if (this._token) {
      url += '?token=' + encodeURIComponent(this._token)
    }

    this.websocket.connect(url)

    await new Promise<void>((res, rej) => {
      const timeout = window.setTimeout(() => {
        this.websocket.disconnect()
        rej(new Error('timeouted'))
      }, websocketTimer)
      this.websocket.once('connected', () => {
        window.clearTimeout(timeout)
        res()
      })
      this.websocket.once('disconnected', (reason) => {
        window.clearTimeout(timeout)
        rej(reason)
      })
    })
  }

  _websocketIsReconnecting = false
  _websocketReconnect() {
    if (this._websocketIsReconnecting) {
      this._log.debug(`websocket reconnection already in progress`)
      return
    }

    this._log.debug(`starting websocket reconnection`)

    setTimeout(async () => {
      while (this._shouldReconnect) {
        try {
          await this._websocketConnect()
          this._webrtcReconnect()
          break
        } catch (e) {
          this._log.debug(`websocket reconnection failed`, e)
        }
      }

      this._websocketIsReconnecting = false
      this._log.debug(`websocket reconnection finished`)
    }, 0)
  }

  async _webrtcConnect(video?: string) {
    if (video && !this._state.webrtc.videos.includes(video)) {
      throw new Error('video id not found')
    }

    Vue.set(this._state, 'status', 'connecting')
    this.websocket.send(EVENT.SIGNAL_REQUEST, { video: video })

    await new Promise<void>((res, rej) => {
      const timeout = window.setTimeout(() => {
        this.webrtc.disconnect()
        rej(new Error('timeouted'))
      }, webrtcTimer)
      this.webrtc.once('connected', () => {
        window.clearTimeout(timeout)
        res()
      })
      this.webrtc.once('disconnected', (reason) => {
        window.clearTimeout(timeout)
        rej(reason)
      })
    })
  }

  _webrtcIsReconnecting = false
  _webrtcReconnect() {
    if (this._webrtcIsReconnecting) {
      this._log.debug(`webrtc reconnection already in progress`)
      return
    }

    const lastVideo = this._state.webrtc.video ?? undefined
    this._log.debug(`starting webrtc reconnection`)

    setTimeout(async () => {
      while (this._shouldReconnect && this.websocket.connected) {
        try {
          await this._webrtcConnect(lastVideo)
          break
        } catch (e) {
          this._log.debug(`webrtc reconnection failed`, e)
        }
      }

      this._webrtcIsReconnecting = false
      this._log.debug(`webrtc reconnection finished`)
    }, 0)
  }
}