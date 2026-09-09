import { getterTree, mutationTree, actionTree } from 'typed-vuex'

export const namespaced = true

export type ConnectionState = 'disconnected' | 'connecting' | 'reconnecting' | 'connected'
export type NetworkQuality = 'unknown' | 'good' | 'fair' | 'poor'

export const state = () => ({
  state: 'disconnected' as ConnectionState,
  error: '',
  quality: 'unknown' as NetworkQuality,
  rtt: null as number | null,
})

export const getters = getterTree(state, {
  connected: (state) => state.state === 'connected',
  connecting: (state) => state.state === 'connecting' || state.state === 'reconnecting',
})

export const mutations = mutationTree(state, {
  setConnecting(state) {
    state.state = 'connecting'
    state.error = ''
    state.quality = 'unknown'
    state.rtt = null
  },

  setConnected(state, connected: boolean) {
    state.state = connected ? 'connected' : 'disconnected'
    if (!connected) {
      state.quality = 'unknown'
      state.rtt = null
    }
  },

  setState(state, connectionState: ConnectionState) {
    state.state = connectionState
    if (connectionState === 'disconnected') {
      state.quality = 'unknown'
      state.rtt = null
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
