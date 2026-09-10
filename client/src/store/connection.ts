import { getterTree, mutationTree, actionTree } from 'typed-vuex'

export const namespaced = true

export type ConnectionState = 'disconnected' | 'connecting' | 'reconnecting' | 'connected'
export type NetworkQuality = 'unknown' | 'good' | 'fair' | 'poor'

export const state = () => ({
  state: 'disconnected' as ConnectionState,
  // Authentication completes as soon as the authenticated websocket sends
  // system/init. Media transport may still be negotiating at that point.
  authenticated: false,
  // A reconnecting ICE transport still owns a live session. Keep this
  // separate from the lifecycle label so the UI does not tear down the
  // session during a transient network interruption.
  connected: false,
  error: '',
  quality: 'unknown' as NetworkQuality,
  rtt: null as number | null,
})

export const getters = getterTree(state, {
  connected: (state) => state.connected,
  authenticated: (state) => state.authenticated,
  connecting: (state) => state.state === 'connecting' || state.state === 'reconnecting',
})

export const mutations = mutationTree(state, {
  setConnecting(state) {
    state.state = 'connecting'
    state.authenticated = false
    state.connected = false
    state.error = ''
    state.quality = 'unknown'
    state.rtt = null
  },

  setConnected(state, connected: boolean) {
    state.state = connected ? 'connected' : 'disconnected'
    state.authenticated = connected
    state.connected = connected
    if (!connected) {
      state.quality = 'unknown'
      state.rtt = null
    }
  },

  setState(state, connectionState: ConnectionState) {
    state.state = connectionState
    if (connectionState === 'connected') {
      state.connected = true
    } else if (connectionState === 'connecting' || connectionState === 'disconnected') {
      state.connected = false
    }
    if (connectionState === 'disconnected') {
      state.quality = 'unknown'
      state.rtt = null
    }
  },

  setAuthenticated(state, authenticated: boolean) {
    state.authenticated = authenticated
    if (!authenticated) {
      state.connected = false
    }
  },

  setError(state, message: string) {
    state.error = message
  },

  setNetworkQuality(state, { quality, rtt }: { quality: NetworkQuality; rtt: number | null }) {
    state.quality = quality
    state.rtt = rtt
  },
})

export const actions = actionTree({ state, getters, mutations }, {})
