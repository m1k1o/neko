import Vue from 'vue'
import Vuex from 'vuex'
import { useAccessor, mutationTree, actionTree } from 'typed-vuex'
import { EVENT } from '~/neko/events'
import { get, set } from '~/utils/localstorage'

import * as video from './video'
import * as chat from './chat'
import * as remote from './remote'
import * as user from './user'
import * as settings from './settings'
import * as client from './client'
import * as emoji from './emoji'

export const state = () => ({
  username: get<string>('username', ''),
  password: get<string>('password', ''),
  connecting: false,
  connected: false,
  locked: false,
})

export const mutations = mutationTree(state, {
  setLogin(state, { username, password }: { username: string; password: string }) {
    state.username = username
    state.password = password
  },

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
    if (connected) {
      set('username', state.username)
      set('password', state.password)
    }
  },
})

export const actions = actionTree(
  { state, mutations },
  {
    initialise(store) {
      accessor.emoji.initialise()
    },

    lock() {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.LOCK)
    },

    unlock() {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.UNLOCK)
    },

    login({ state }, { username, password }: { username: string; password: string }) {
      accessor.setLogin({ username, password })
      $client.login(password, username)
    },

    logout({ state }) {
      accessor.setLogin({ username: '', password: '' })
      set('username', '')
      set('password', '')
      $client.logout()
    },
  },
)

export const storePattern = {
  state,
  mutations,
  actions,
  modules: { video, chat, user, remote, settings, client, emoji },
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
