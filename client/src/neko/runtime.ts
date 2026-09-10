import { AuthHttpClient } from '~/sdk/auth'
import { RoomHttpClient } from '~/sdk/room'

/**
 * Store and UI ports used by the browser client. The protocol client only
 * knows these ports; Vue, Vuex and notification libraries stay in the adapter.
 */
export interface NekoStatePort {
  [module: string]: any
}

export interface NekoUIAdapter {
  translate(key: string, params?: Record<string, unknown>): string
  notify(options: Record<string, unknown>): void
  alert(options: Record<string, unknown>): void
  log: {
    warn(...args: any[]): void
    error(...args: any[]): void
    info(...args: any[]): void
    debug(...args: any[]): void
  }
}

export interface NekoClientRuntime {
  http: AuthHttpClient &
    RoomHttpClient & {
      defaults: {
        baseURL?: string
        withCredentials?: boolean
        headers: {
          common: Record<string, unknown>
        }
      }
    }
  state: NekoStatePort
  ui: NekoUIAdapter
}
