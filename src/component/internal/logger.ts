import { NekoWebSocket } from './websocket'
import * as EVENT from '../types/events'
import * as message from '../types/messages'
import { Logger } from '../utils/logger'

const RETRY_INTERVAL = 2500

export class NekoLogger extends Logger {
  private _ws: NekoWebSocket
  private _logs: message.SystemLog[] = []
  private _interval: number | null = null

  constructor(websocket: NekoWebSocket, scope?: string) {
    super(scope)

    this._ws = websocket
  }

  protected _send(level: string, message: string, fields?: Record<string, any>) {
    if (!fields) {
      fields = { scope: this._scope }
    } else {
      fields['scope'] = this._scope
    }

    const payload = { level, message, fields } as message.SystemLog
    if (!this._ws.connected) {
      this._logs.push(payload)

      // postpone logs sending
      if (this._interval == null) {
        this._interval = window.setInterval(() => {
          if (!this._ws.connected || !this._interval) {
            return
          }

          if (this._logs.length > 0) {
            this._ws.send(EVENT.SYSTEM_LOGS, this._logs)
          }

          window.clearInterval(this._interval)
          this._interval = null
        }, RETRY_INTERVAL)
      }

      return
    }

    // abort postponed logs sending
    if (this._interval != null) {
      window.clearInterval(this._interval)
      this._interval = null
    }

    this._ws.send(EVENT.SYSTEM_LOGS, [...this._logs, payload])
    this._logs = []
  }

  public error(message: string, fields?: Record<string, any>) {
    this._console('error', message, fields)
    this._send('error', message, fields)
  }

  public warn(message: string, fields?: Record<string, any>) {
    this._console('warn', message, fields)
    this._send('warn', message, fields)
  }

  public info(message: string, fields?: Record<string, any>) {
    this._console('info', message, fields)
    this._send('info', message, fields)
  }

  public debug(message: string, fields?: Record<string, any>) {
    this._console('debug', message, fields)
    this._send('debug', message, fields)
  }
}
