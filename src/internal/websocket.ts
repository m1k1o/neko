import EventEmitter from 'eventemitter3'
import { Logger } from '../utils/logger'

export const timeout = 15000

export interface NekoWebSocketEvents {
  connecting: () => void
  connected: () => void
  disconnected: (error?: Error) => void
}

export class NekoWebSocket extends EventEmitter<NekoWebSocketEvents> {
  private _ws?: WebSocket
  private _timeout?: NodeJS.Timeout
  private _log: Logger

  constructor() {
    super()
  
    this._log = new Logger('websocket')
  }

  get connected() {
    return typeof this._ws !== 'undefined' && this._ws.readyState === WebSocket.OPEN
  }

  public connect(url: string, password: string) {
    if (this.connected) {
      throw new Error('attempting to create websocket while connection open')
    }

    this.emit('connecting')

    this._ws = new WebSocket(`${url}ws?password=${password}`)
    this._log.debug(`connecting to ${this._ws.url}`)

    this._ws.onopen = this.onConnected.bind(this)
    this._ws.onclose = this.onDisconnected.bind(this, new Error('websocket closed'))
    this._ws.onerror = this.onError.bind(this)
    this._ws.onmessage = this.onMessage.bind(this)

    this._timeout = setTimeout(this.onTimeout.bind(this), timeout)
  }

  public disconnect() {
    if (this._timeout) {
      clearTimeout(this._timeout)
    }

    if (this.connected) {
      try {
        this._ws!.close()
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

  private onMessage(e: MessageEvent) {
    const { event, ...payload } = JSON.parse(e.data)
    this._log.debug(`received websocket event ${event} ${payload ? `with payload: ` : ''}`, payload)

    // @ts-ignore
    if (typeof this[event] === 'function') {
      // @ts-ignore
      // TODO: REFACTOR
      this[event](payload)
    } else {
      this._log.warn(`unhandled websocket event '${event}':`, payload)
    }
  }

  private onError(event: Event) {
    this._log.error((event as ErrorEvent).error)
  }

  private onConnected() {
    if (this._timeout) {
      clearTimeout(this._timeout)
    }

    if (!this.connected) {
      this._log.warn(`onConnected called while being disconnected`)
      return
    }

    this._log.debug(`connected`)
    this.emit('connected')
  }

  private onTimeout() {
    this._log.debug(`connection timeout`)
    this.onDisconnected(new Error('connection timeout'))
  }

  private onDisconnected(reason?: Error) {
    this._log.debug(`disconnected:`, reason)
  
    this.disconnect()
    this.emit('disconnected', reason)
  }
}
