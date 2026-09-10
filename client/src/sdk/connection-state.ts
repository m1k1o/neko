export type ConnectionPhase = 'disconnected' | 'connecting' | 'reconnecting' | 'connected'

export type ConnectionEvent = 'connect' | 'reconnect' | 'connected' | 'disconnect'

export interface ConnectionTransition {
  state: ConnectionPhase
  changed: boolean
}

const transitions: Record<ConnectionPhase, Partial<Record<ConnectionEvent, ConnectionPhase>>> = {
  disconnected: {
    connect: 'connecting',
  },
  connecting: {
    connect: 'connecting',
    reconnect: 'reconnecting',
    connected: 'connected',
    disconnect: 'disconnected',
  },
  reconnecting: {
    connect: 'connecting',
    reconnect: 'reconnecting',
    connected: 'connected',
    disconnect: 'disconnected',
  },
  connected: {
    reconnect: 'reconnecting',
    connected: 'connected',
    disconnect: 'disconnected',
  },
}

/**
 * Small, deterministic lifecycle machine shared by the signaling and media
 * layers. Keeping transitions here prevents duplicate connected/disconnected
 * callbacks when WebSocket and ICE emit events for the same failure.
 */
export class ConnectionStateMachine {
  private _state: ConnectionPhase = 'disconnected'

  get state() {
    return this._state
  }

  transition(event: ConnectionEvent): ConnectionTransition {
    const next = transitions[this._state][event]
    if (!next || next === this._state) {
      return { state: this._state, changed: false }
    }

    this._state = next
    return { state: this._state, changed: true }
  }

  reset() {
    this._state = 'disconnected'
  }
}
