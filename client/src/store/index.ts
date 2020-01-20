import Vue from 'vue'
import Vuex from 'vuex'
import { useAccessor, mutationTree, actionTree } from 'typed-vuex'

import * as video from './video'
import * as remote from './remote'
import * as user from './user'

export const state = () => ({
  connecting: false,
  connected: false,
})

// type RootState = ReturnType<typeof state>

export const getters = {
  // connected: (state: RootState) => state.connected
}

export const mutations = mutationTree(state, {
  initialiseStore() {
    // TODO: init with localstorage to retrieve save settings
  },
  setConnnecting(state, connecting: boolean) {
    state.connecting = connecting
  },
  setConnected(state, connected: boolean) {
    state.connected = connected
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    //
  },
)

export const storePattern = {
  state,
  mutations,
  actions,
  modules: { video, user, remote },
}

Vue.use(Vuex)

const store = new Vuex.Store(storePattern)
export const accessor = useAccessor(store, storePattern)

Vue.prototype.$accessor = accessor

export default store
