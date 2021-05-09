import EventEmitter from 'eventemitter3'
import { Logger } from '../utils/logger'

export const connTimeout = 15000
export const reconnTimeout = 1000
export const reconnMultiplier = 1.2

export interface NekoWebSocketEvents {
  connecting: () => void
  connected: () => void
  disconnected: (error?: Error) => void
  message: (event: string, payload: any) => void
}

export class NekoWebSocket extends EventEmitter<NekoWebSocketEvents> {
  private _ws?: WebSocket
  private _connTimer?: NodeJS.Timeout
  private _log: Logger
  private _url: string
  private _token: string

  // reconnection
  private _reconTimer?: NodeJS.Timeout
  private _reconnTimeout: number = reconnTimeout

  constructor() {
    super()

    this._log = new Logger('websocket')

    this._url = ''
    this._token = ''
    this.setUrl(location.href)
  }

  public setUrl(url: string) {
    this._url = url.replace(/^http/, 'ws').replace(/\/+$/, '') + '/api/ws'
  }

  public setToken(token: string) {
    this._token = token
  }

  get supported() {
    return typeof WebSocket !== 'undefined' && WebSocket.OPEN === 1
  }

  get connected() {
    return typeof this._ws !== 'undefined' && this._ws.readyState === WebSocket.OPEN
  }

  public async connect() {
    if (this.connected) {
      throw new Error('attempting to create websocket while connection open')
    }

    if (typeof this._ws !== 'undefined') {
      this._log.debug(`previous websocket connection needs to be closed`)
      this.disconnect(new Error('connection replaced'))
    }

    this.emit('connecting')

    let url = this._url
    if (this._token) {
      url += '?token=' + encodeURIComponent(this._token)
    }

    await new Promise<void>((res, rej) => {
      this._ws = new WebSocket(url)
      this._log.info(`connecting`)

      this._ws.onclose = rej.bind(this, new Error('connection close'))
      this._ws.onerror = rej.bind(this, new Error('connection error'))
      this._ws.onmessage = this.onMessage.bind(this)

      this._ws.onopen = () => {
        this._ws!.onclose = this.onClose.bind(this, 'close')
        this._ws!.onerror = this.onClose.bind(this, 'error')

        this.onConnected()
        res()
      }

      this._connTimer = setTimeout(rej.bind(this, new Error('connection timeout')), connTimeout)
    })
  }

  public disconnect(reason?: Error) {
    this.emit('disconnected', reason)

    if (this._connTimer) {
      clearTimeout(this._connTimer)
    }

    if (typeof this._ws !== 'undefined') {
      // unmount all events
      this._ws.onopen = () => {}
      this._ws.onclose = () => {}
      this._ws.onerror = () => {}
      this._ws.onmessage = () => {}

      try {
        this._ws.close()
      } catch (err) {}

      this._ws = undefined
    }
  }

  public send(event: string, payload?: any) {
    if (!this.connected) {
      this._log.warn(`attempting to send message while disconnected`)
      return
    }

    this._log.debug(`sending event '${event}' ${payload ? `with payload: ` : ''}`, payload)
    this._ws!.send(JSON.stringify({ event, ...payload }))
  }

  private tryReconnect() {
    if (this._reconTimer) {
      clearTimeout(this._reconTimer)
    }

    this._reconTimer = setTimeout(this.onReconnect.bind(this), this._reconnTimeout)
  }

  private onMessage(e: MessageEvent) {
    const { event, ...payload } = JSON.parse(e.data)

    this._log.debug(`received websocket event ${event} ${payload ? `with payload: ` : ''}`, payload)
    this.emit('message', event, payload)
  }

  private onConnected() {
    if (this._connTimer) {
      clearTimeout(this._connTimer)
    }

    if (this._reconTimer) {
      clearTimeout(this._reconTimer)
    }

    if (!this.connected) {
      this._log.warn(`onConnected called while being disconnected`)
      return
    }

    this._log.info(`connected`)
    this.emit('connected')

    // reset reconnect timeout
    this._reconnTimeout = reconnTimeout
  }

  private onClose(reason: string) {
    this._log.info(`connection ${reason}`)
    this.disconnect(new Error(`connection ${reason}`))
    this.tryReconnect()
  }

  private onReconnect() {
    this._log.info(`reconnecting after ${this._reconnTimeout}ms`)
    this._reconnTimeout *= reconnMultiplier

    this.connect()
  }
}
