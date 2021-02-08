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

  // reconnection
  private _reconTimer?: NodeJS.Timeout
  private _reconnTimeout: number = reconnTimeout

  constructor() {
    super()

    this._log = new Logger('websocket')

    this._url = ''
    this.setUrl(location.href)
  }

  public setUrl(url: string) {
    this._url = url.replace(/^http/, 'ws').replace(/\/+$/, '') + '/api/ws'
  }

  get supported() {
    return typeof WebSocket !== 'undefined' && WebSocket.OPEN === 1
  }

  get connected() {
    return typeof this._ws !== 'undefined' && this._ws.readyState === WebSocket.OPEN
  }

  public connect() {
    if (this.connected) {
      throw new Error('attempting to create websocket while connection open')
    }

    if (typeof this._ws !== 'undefined') {
      this._log.debug(`previous websocket connection needs to be closed`)
      this.disconnect(new Error('connection replaced'))
    }

    this.emit('connecting')

    this._ws = new WebSocket(this._url)
    this._log.info(`connecting`)

    this._ws.onopen = this.onConnected.bind(this)
    this._ws.onclose = this.onClose.bind(this)
    this._ws.onerror = this.onError.bind(this)
    this._ws.onmessage = this.onMessage.bind(this)

    this._connTimer = setTimeout(this.onTimeout.bind(this), connTimeout)
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

  private onTimeout() {
    this._log.info(`connection timeout`)
    this.disconnect(new Error('connection timeout'))
    this.tryReconnect()
  }

  private onError() {
    this._log.info(`connection error`)
    this.disconnect(new Error('connection error'))
    this.tryReconnect()
  }

  private onClose() {
    this._log.info(`connection closed`)
    this.disconnect(new Error('connection closed'))
    this.tryReconnect()
  }

  private onReconnect() {
    this._log.info(`reconnecting after ${this._reconnTimeout}ms`)
    this._reconnTimeout *= reconnMultiplier

    this.connect()
  }
}
