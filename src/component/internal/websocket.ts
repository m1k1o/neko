import EventEmitter from 'eventemitter3'
import { Logger } from '../utils/logger'

export const connTimeout = 15000
export const reconnInterval = 1000

export interface NekoWebSocketEvents {
  connecting: () => void
  connected: () => void
  disconnected: (error?: Error) => void
  message: (event: string, payload: any) => void
}

export class NekoWebSocket extends EventEmitter<NekoWebSocketEvents> {
  private _ws?: WebSocket
  private _log: Logger

  constructor() {
    super()

    this._log = new Logger('websocket')
  }

  get supported() {
    return typeof WebSocket !== 'undefined' && WebSocket.OPEN === 1
  }

  get connected() {
    return typeof this._ws !== 'undefined' && this._ws.readyState === WebSocket.OPEN
  }

  public async connect(url: string) {
    if (!this.supported) {
      throw new Error('browser does not support websockets')
    }

    if (this.connected) {
      throw new Error('attempting to create websocket while connection open')
    }

    if (typeof this._ws !== 'undefined') {
      this._log.debug(`previous websocket connection needs to be closed`)
      this.disconnect()
    }

    await new Promise<void>((res, rej) => {
      this._ws = new WebSocket(url)

      this._log.info(`connecting`)
      this.emit('connecting')

      this._ws.onclose = rej.bind(this, new Error('connection close'))
      this._ws.onerror = rej.bind(this, new Error('connection error'))
      this._ws.onmessage = this.onMessage.bind(this)

      const timeout = window.setTimeout(rej.bind(this, new Error('connection timeout')), connTimeout)
      this._ws.onopen = () => {
        window.clearTimeout(timeout)

        this._ws!.onclose = this.onDisconnected.bind(this, 'close')
        this._ws!.onerror = this.onDisconnected.bind(this, 'error')

        this.onConnected()
        res()
      }
    })
  }

  public disconnect() {
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

  private onMessage(e: MessageEvent) {
    const { event, ...payload } = JSON.parse(e.data)

    this._log.debug(`received websocket event ${event} ${payload ? `with payload: ` : ''}`, payload)
    this.emit('message', event, payload)
  }

  private onConnected() {
    if (!this.connected) {
      this._log.warn(`onConnected called while being disconnected`)
      return
    }

    this._log.info(`connected`)
    this.emit('connected')
  }

  private onDisconnected(reason: string) {
    this.disconnect()

    this._log.info(`connection ${reason}`)
    this.emit('disconnected', new Error(`connection ${reason}`))
  }
}
