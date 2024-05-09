import type { NekoWebSocket } from './websocket'
import * as EVENT from '../types/events'
import type * as message from '../types/messages'
import { Logger } from '../utils/logger'

const MAX_LOG_MESSAGES = 25
const FLUSH_TIMEOUT_MS = 250
const RETRY_INTERVAL_MS = 2500

export class NekoLoggerFactory {
  private _logs: message.SystemLog[] = []
  private _timeout: number | null = null
  private _interval: number | null = null

  // eslint-disable-next-line
  constructor(
    private readonly _ws: NekoWebSocket,
  ) {}

  private _flush() {
    if (this._logs.length > 0) {
      this._ws.send(EVENT.SYSTEM_LOGS, this._logs)
      this._logs = []
    }
  }

  private _send(level: string, message: string, fields?: Record<string, any>) {
    for (const key in fields) {
      const field = fields[key]

      if (field instanceof Error) {
        fields[key] = (field as Error).message
      }
    }

    const payload = { level, message, fields } as message.SystemLog
    this._logs.push(payload)

    // rotate if exceeded maximum
    if (this._logs.length > MAX_LOG_MESSAGES) {
      this._logs.shift()
    }

    // postpone logs sending
    if (!this._timeout && !this._interval) {
      this._timeout = window.setTimeout(() => {
        if (!this._timeout) {
          return
        }

        if (this._ws.connected) {
          this._flush()
        } else {
          this._interval = window.setInterval(() => {
            if (!this._ws.connected || !this._interval) {
              return
            }

            this._flush()
            window.clearInterval(this._interval)
            this._interval = null
          }, RETRY_INTERVAL_MS)
        }

        window.clearTimeout(this._timeout)
        this._timeout = null
      }, FLUSH_TIMEOUT_MS)
    }
  }

  public new(submodule?: string): Logger {
    return new NekoLogger((level: string, message: string, fields?: Record<string, any>) => {
      if (!fields) {
        fields = { submodule }
      } else {
        fields['submodule'] = submodule
      }

      this._send(level, message, fields)
    }, submodule)
  }

  public destroy() {
    if (this._ws.connected) {
      this._flush()
    }

    if (this._interval) {
      window.clearInterval(this._interval)
      this._interval = null
    }

    if (this._timeout) {
      window.clearTimeout(this._timeout)
      this._timeout = null
    }
  }
}

type NekoLoggerMessage = (level: string, message: string, fields?: Record<string, any>) => void

export class NekoLogger extends Logger {
  private _on_send: NekoLoggerMessage

  constructor(onSend: NekoLoggerMessage, scope?: string) {
    super(scope)

    this._on_send = onSend
  }

  public error(message: string, fields?: Record<string, any>) {
    this._console('error', message, fields)
    this._on_send('error', message, fields)
  }

  public warn(message: string, fields?: Record<string, any>) {
    this._console('warn', message, fields)
    this._on_send('warn', message, fields)
  }

  public info(message: string, fields?: Record<string, any>) {
    this._console('info', message, fields)
    this._on_send('info', message, fields)
  }

  public debug(message: string, fields?: Record<string, any>) {
    this._console('debug', message, fields)
    this._on_send('debug', message, fields)
  }
}
