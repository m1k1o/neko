import EventEmitter from 'eventemitter3'
import { ReconnecterConfig } from '../types/reconnecter'

export interface ReconnecterAbstractEvents {
  connect: () => void
  disconnect: (error?: Error) => void
}

export abstract class ReconnecterAbstract extends EventEmitter<ReconnecterAbstractEvents> {
  constructor() {
    super()

    if (this.constructor == ReconnecterAbstract) {
      throw new Error("Abstract classes can't be instantiated.")
    }
  }

  public get connected(): boolean {
    throw new Error("Getter'connected()' must be implemented.")
  }

  public async connect() {
    throw new Error("Method 'connect()' must be implemented.")
  }

  public async disconnect() {
    throw new Error("Method 'disconnect()' must be implemented.")
  }
}

export interface ReconnecterEvents {
  open: () => void
  connect: () => void
  disconnect: () => void
  close: (error?: Error) => void
}

export class Reconnecter extends EventEmitter<ReconnecterEvents> {
  private _conn: ReconnecterAbstract
  private _config: ReconnecterConfig
  private _timeout?: number

  private _open = false
  private _total_reconnects = 0
  private _last_connected?: Date

  constructor(conn: ReconnecterAbstract, config?: ReconnecterConfig) {
    super()

    this._conn = conn
    this._config = {
      maxReconnects: 10,
      timeoutMs: 1500,
      backoffMs: 750,
      ...config,
    }

    this._conn.on('connect', this.onConnect.bind(this))
    this._conn.on('disconnect', this.onDisconnect.bind(this))
  }

  private onConnect() {
    if (this._timeout) {
      window.clearTimeout(this._timeout)
      this._timeout = undefined
    }

    if (this._open) {
      this._last_connected = new Date()
      this.emit('connect')
    } else {
      this._conn.disconnect()
    }
  }

  private onDisconnect() {
    if (this._timeout) {
      window.clearTimeout(this._timeout)
      this._timeout = undefined
    }

    if (this._open) {
      this.emit('disconnect')
      this.reconnect()
    }
  }

  public get isOpen(): boolean {
    return this._open
  }

  public get isConnected(): boolean {
    return this._conn.connected
  }

  public get totalReconnects(): number {
    return this._total_reconnects
  }

  public get lastConnected(): Date | undefined {
    return this._last_connected
  }

  public get config(): ReconnecterConfig {
    return { ...this._config }
  }

  public set config(conf: ReconnecterConfig) {
    this._config = { ...conf }

    if (this._config.maxReconnects > this._total_reconnects) {
      this.close(new Error('reconnection config changed'))
    }
  }

  public open(deferredConnection = false): void {
    if (this._open) {
      throw new Error('connection is already open')
    }

    this._open = true
    this.emit('open')

    if (!deferredConnection) {
      this.connect()
    }
  }

  public close(error?: Error): void {
    if (this._timeout) {
      window.clearTimeout(this._timeout)
      this._timeout = undefined
    }

    if (!this._open) {
      throw new Error('connection is already closed')
    }

    this._open = false
    this._last_connected = undefined
    this.emit('close', error)

    if (this._conn.connected) {
      this._conn.disconnect()
    }
  }

  public connect(): void {
    if (this._timeout) {
      window.clearTimeout(this._timeout)
      this._timeout = undefined
    }

    this._conn.connect()
    this._timeout = window.setTimeout(this.onDisconnect.bind(this), this._config.timeoutMs)
  }

  public reconnect(): void {
    if (this._conn.connected) {
      throw new Error('connection is already connected')
    }

    this._total_reconnects++

    if (this._config.maxReconnects > this._total_reconnects || this._total_reconnects < 0) {
      setTimeout(this.connect.bind(this), this._config.backoffMs)
    } else {
      this.close(new Error('reconnection failed'))
    }
  }

  public destroy() {
    this._conn.off('connect', this.onConnect.bind(this))
    this._conn.off('disconnect', this.onDisconnect.bind(this))
  }
}
