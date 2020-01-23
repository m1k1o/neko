import Vue from 'vue'
import Vuex from 'vuex'
import { useAccessor, mutationTree, actionTree } from 'typed-vuex'

import * as video from './video'
import * as chat from './chat'
import * as remote from './remote'
import * as user from './user'
import * as settings from './settings'
import * as client from './client'

export const state = () => ({
  connecting: false,
  connected: false,
  locked: false,
})

export const mutations = mutationTree(state, {
  setLocked(state, locked: boolean) {
    state.locked = locked
  },

  setConnnecting(state) {
    state.connected = false
    state.connecting = true
  },

  setConnected(state, connected: boolean) {
    state.connected = connected
    state.connecting = false
  },
})

export const actions = actionTree(
  { state, mutations },
  {
    //
    connect(store, { username, password }: { username: string; password: string }) {
      $client.connect(password, username)
    },
  },
)

export const storePattern = {
  state,
  mutations,
  actions,
  modules: { video, chat, user, remote, settings, client },
}

Vue.use(Vuex)

const store = new Vuex.Store(storePattern)
export const accessor = useAccessor(store, storePattern)

Vue.prototype.$accessor = accessor

declare module 'vue/types/vue' {
  interface Vue {
    $accessor: typeof accessor
  }
}

export default store
