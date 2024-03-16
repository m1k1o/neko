import EventEmitter from 'eventemitter3'

import type { ReconnectorConfig } from '../../types/reconnector'

export interface ReconnectorAbstractEvents {
  connect: () => void
  disconnect: (error?: Error) => void
}

export abstract class ReconnectorAbstract extends EventEmitter<ReconnectorAbstractEvents> {
  constructor() {
    super()

    if (this.constructor == ReconnectorAbstract) {
      throw new Error("Abstract classes can't be instantiated.")
    }
  }

  public abstract get connected(): boolean

  public abstract connect(): void
  public abstract disconnect(): void
  public abstract destroy(): void
}

/*

Reconnector handles reconnection logic according to supplied config for an abstract class. It can reconnect anything that:
- can be connected to
- can send event once it is connected to
- can be disconnected from
- can send event once it is disconnected from
- can provide information at any moment if it is connected to or not

Reconnector creates one additional abstract layer for a user. User can open and close a connection. If the connection is open,
when connection will be disconnected, reconnector will attempt to connect to it again. Once connection is closed, no further
events will be emitted and connection will be disconnected.
- When using deferred connection in opening function, reconnector does not try to connect when opening a connection. This is
the initial state, when reconnector is not connected but no reconnect attempts are in progress, since there has not been
any disconnect even. It is up to user to call initial connect attempt.
- Events 'open' and 'close' will be fired exactly once, no matter how many times open() and close() funxtions were called.
- Events 'connecŧ' and 'disconnect' can fire throughout open connection at any time.

*/
export interface ReconnectorEvents {
  open: () => void
  connect: () => void
  disconnect: () => void
  close: (error?: Error) => void
}

export class Reconnector extends EventEmitter<ReconnectorEvents> {
  private _config: ReconnectorConfig
  private _timeout?: number

  private _open = false
  private _total_reconnects = 0
  private _last_connected?: Date

  private _onConnectHandle: () => void
  private _onDisconnectHandle: (error?: Error) => void

  // eslint-disable-next-line
  constructor(
    private readonly _conn: ReconnectorAbstract,
    config?: ReconnectorConfig,
  ) {
    super()

    // setup default config values
    this._config = {
      max_reconnects: 10,
      timeout_ms: 1500,
      backoff_ms: 750,
      ...config,
    }

    // register connect and disconnect handlers with current class
    // as 'this' context, store them to a variable so that they
    // can be later unregistered

    this._onConnectHandle = this.onConnect.bind(this)
    this._conn.on('connect', this._onConnectHandle)

    this._onDisconnectHandle = this.onDisconnect.bind(this)
    this._conn.on('disconnect', this._onDisconnectHandle)
  }

  private clearTimeout() {
    if (this._timeout) {
      window.clearTimeout(this._timeout)
      this._timeout = undefined
    }
  }

  private onConnect() {
    this.clearTimeout()

    // only if connection is open, fire connect event
    if (this._open) {
      this._last_connected = new Date()
      this._total_reconnects = 0
      this.emit('connect')
    } else {
      // in this case we are connected but this connection
      // has been closed, so we simply disconnect again
      this._conn.disconnect()
    }
  }

  private onDisconnect() {
    this.clearTimeout()

    // only if connection is open, fire disconnect event
    // and start reconnecteing logic
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

  public get config(): ReconnectorConfig {
    return { ...this._config }
  }

  public set config(conf: ReconnectorConfig) {
    this._config = { ...this._config, ...conf }

    if (this._config.max_reconnects <= this._total_reconnects) {
      this.close(new Error('reconnection config changed'))
    }
  }

  // allows future reconnect attempts and connects if not set
  // deferred connection to true
  public open(deferredConnection = false): void {
    this.clearTimeout()

    // assuming open event can fire multiple times, we need to
    // ensure, that open event get fired only once along with
    // resetting total reconnects counter

    if (!this._open) {
      this._open = true
      this._total_reconnects = 0
      this.emit('open')
    }

    if (!deferredConnection && !this._conn.connected) {
      this.connect()
    }
  }

  // disconnects and forbids future reconnect attempts
  public close(error?: Error): void {
    this.clearTimeout()

    // assuming close event can fire multiple times, the same
    // precautions need to be taken as in open event, so that
    // close event fires only once

    if (this._open) {
      this._open = false
      this._last_connected = undefined
      this.emit('close', error)
    }

    // if connected, tries to disconnect even if it has been
    // called multiple times by user

    if (this._conn.connected) {
      this._conn.disconnect()
    }
  }

  // tries to connect and calls on disconnected if it could not
  // connect within specified timeout according to config
  public connect(): void {
    this.clearTimeout()

    this._conn.connect()
    this._timeout = window.setTimeout(this.onDisconnect.bind(this), this._config.timeout_ms)
  }

  // tries to connect with specified backoff time if
  // maximum reconnect theshold was not exceeded, otherwise
  // closes the connection with an error message
  public reconnect(): void {
    this.clearTimeout()

    if (this._config.max_reconnects > ++this._total_reconnects || this._config.max_reconnects == -1) {
      this._timeout = window.setTimeout(this.connect.bind(this), this._config.backoff_ms)
    } else {
      this.close(new Error('reconnection failed'))
    }
  }

  // closes connection and unregisters all events
  public destroy() {
    this.close(new Error('connection destroyed'))

    this._conn.off('connect', this._onConnectHandle)
    this._conn.off('disconnect', this._onDisconnectHandle)
    this._conn.destroy()
  }
}
