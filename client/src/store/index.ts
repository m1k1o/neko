import Vue from 'vue'
import Vuex from 'vuex'
import { useAccessor, mutationTree, getterTree, actionTree } from 'typed-vuex'

import * as video from './video'
import * as chat from './chat'
import * as files from './files'
import * as openinapp from './openinapp'
import * as remote from './remote'
import * as user from './user'
import * as settings from './settings'
import * as client from './client'
import * as emoji from './emoji'
import * as connection from './connection'
import * as session from './session'

export const state = () => ({})

export const mutations = mutationTree(state, {})

export const getters = getterTree(state, {})

export const actions = actionTree(
  { state, getters, mutations },
  {
    initialise() {
      accessor.emoji.initialise()
      accessor.settings.initialise()
    },
  },
)

export const storePattern = {
  state,
  mutations,
  actions,
  getters,
  modules: { connection, video, chat, files, openinapp, user, remote, settings, client, emoji, session },
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
