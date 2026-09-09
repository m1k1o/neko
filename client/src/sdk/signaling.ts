export type SignalingState = 'idle' | 'connecting' | 'open' | 'closing' | 'closed'

export interface SignalingMessage {
  event: string
  /** Canonical server envelope. Flat messages remain accepted on receive. */
  payload?: Record<string, unknown>
  [key: string]: unknown
}

export interface SignalingTransportOptions {
  onMessage: (message: SignalingMessage) => void | Promise<void>
  onOpen?: () => void
  onError?: (error: Error) => void
  onClose?: (event: CloseEvent) => void
  webSocketFactory?: (url: string) => WebSocket
}

function toError(value: unknown, fallback: string): Error {
  if (value instanceof Error) {
    return value
  }
  if (typeof value === 'string' && value !== '') {
    return new Error(value)
  }
  return new Error(fallback)
}

/**
 * Owns the WebSocket lifecycle and JSON envelope validation. A generation
 * token makes stale socket callbacks harmless when a reconnect replaces the
 * previous socket.
 */
export class SignalingTransport {
  private readonly options: SignalingTransportOptions
  private readonly createWebSocket: (url: string) => WebSocket
  private socket?: WebSocket
  private generation = 0
  private _state: SignalingState = 'idle'
  private _url = ''

  constructor(options: SignalingTransportOptions) {
    this.options = options
    this.createWebSocket = options.webSocketFactory || ((url) => new WebSocket(url))
  }

  get state() {
    return this._state
  }

  get url() {
    return this._url
  }

  get open() {
    return this.socket !== undefined && this.socket.readyState === 1
  }

  connect(url: string) {
    this.close()
    const generation = ++this.generation
    this._url = url
    this.setState('connecting')

    let socket: WebSocket
    try {
      socket = this.createWebSocket(url)
    } catch (error) {
      this.setState('closed')
      this.options.onError?.(toError(error, 'unable to create signaling socket'))
      return
    }

    this.socket = socket
    socket.onopen = () => {
      if (!this.isCurrent(socket, generation)) return
      this.setState('open')
      this.options.onOpen?.()
    }
    socket.onmessage = (event: MessageEvent) => {
      if (!this.isCurrent(socket, generation)) return
      this.handleMessage(event.data)
    }
    socket.onerror = () => {
      if (!this.isCurrent(socket, generation)) return
      this.options.onError?.(new Error('signaling socket error'))
    }
    socket.onclose = (event: CloseEvent) => {
      if (!this.isCurrent(socket, generation)) return
      this.socket = undefined
      this.setState('closed')
      this.options.onClose?.(event)
    }
  }

  send(message: SignalingMessage) {
    if (!this.open || !this.socket) {
      return false
    }

    try {
      this.socket.send(JSON.stringify(message))
      return true
    } catch (error) {
      this.options.onError?.(toError(error, 'unable to send signaling message'))
      return false
    }
  }

  close(code?: number, reason?: string) {
    const socket = this.socket
    this.generation++
    this.socket = undefined

    if (!socket) {
      if (this._state !== 'idle') this.setState('closed')
      return
    }

    this.setState('closing')
    socket.onopen = null
    socket.onmessage = null
    socket.onerror = null
    socket.onclose = null
    try {
      socket.close(code, reason)
    } catch (error) {
      this.options.onError?.(toError(error, 'unable to close signaling socket'))
    }
    this.setState('closed')
  }

  private setState(state: SignalingState) {
    this._state = state
  }

  private isCurrent(socket: WebSocket, generation: number) {
    return this.socket === socket && this.generation === generation
  }

  private handleMessage(data: unknown) {
    if (typeof data !== 'string') {
      this.options.onError?.(new Error('signaling message must be a JSON string'))
      return
    }

    let value: unknown
    try {
      value = JSON.parse(data)
    } catch (error) {
      this.options.onError?.(toError(error, 'invalid signaling JSON'))
      return
    }

    if (value === null || typeof value !== 'object' || typeof (value as { event?: unknown }).event !== 'string') {
      this.options.onError?.(new Error('signaling message is missing an event'))
      return
    }

    const payload = (value as { payload?: unknown }).payload
    if (payload !== undefined && payload !== null && (typeof payload !== 'object' || Array.isArray(payload))) {
      this.options.onError?.(new Error('signaling payload must be an object'))
      return
    }

    try {
      const result = this.options.onMessage(value as SignalingMessage)
      if (result instanceof Promise) {
        result.catch((error) => this.options.onError?.(toError(error, 'signaling message handler failed')))
      }
    } catch (error) {
      this.options.onError?.(toError(error, 'signaling message handler failed'))
    }
  }
}
