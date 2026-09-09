import Vue from 'vue'
import Vuex from 'vuex'
import { useAccessor, mutationTree, getterTree, actionTree } from 'typed-vuex'
import { AdminLockResource } from '~/neko/messages'
import { get, set } from '~/utils/localstorage'

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

export const state = () => ({
  displayname: get<string>('displayname', ''),
  password: get<string>('password', ''),
  active: false,
  locked: {} as Record<string, boolean>,
})

export const mutations = mutationTree(state, {
  setActive(state) {
    state.active = true
  },

  setLogin(state, { displayname, password }: { displayname: string; password: string }) {
    state.displayname = displayname
    state.password = password
  },

  setLocked(state, resource: string) {
    Vue.set(state.locked, resource, true)
  },

  setUnlocked(state, resource: string) {
    Vue.set(state.locked, resource, false)
  },
})

export const getters = getterTree(state, {
  isLocked: (state) => (resource: AdminLockResource) => resource in state.locked && state.locked[resource],
})

async function setFileTransferEnabled(enabled: boolean) {
  const response = await $http.get<{ plugins?: Record<string, unknown> }>('/api/room/settings')
  const plugins = {
    ...(response.data.plugins || {}),
    'filetransfer.enabled': enabled,
  }
  await $http.post('/api/room/settings', { plugins })
}

export const actions = actionTree(
  { state, getters, mutations },
  {
    initialise() {
      accessor.emoji.initialise()
      accessor.settings.initialise()
    },

    async lock(_, resource: AdminLockResource) {
      if (!accessor.connection.connected || !accessor.user.admin) {
        return
      }

      const setting = resource === 'login' ? 'locked_logins' : resource === 'control' ? 'locked_controls' : null
      if (setting) {
        await $http.post('/api/room/settings', { [setting]: true })
      } else {
        await setFileTransferEnabled(false)
      }
      accessor.setLocked(resource)
    },

    async unlock(_, resource: AdminLockResource) {
      if (!accessor.connection.connected || !accessor.user.admin) {
        return
      }

      const setting = resource === 'login' ? 'locked_logins' : resource === 'control' ? 'locked_controls' : null
      if (setting) {
        await $http.post('/api/room/settings', { [setting]: false })
      } else {
        await setFileTransferEnabled(true)
      }
      accessor.setUnlocked(resource)
    },

    toggleLock(_, resource: AdminLockResource) {
      if (accessor.isLocked(resource)) {
        accessor.unlock(resource)
      } else {
        accessor.lock(resource)
      }
    },

    login(store, { displayname, password }: { displayname: string; password: string }) {
      accessor.setLogin({ displayname, password })
      $client.login(password, displayname)
    },

    logout() {
      accessor.setLogin({ displayname: '', password: '' })
      accessor.connection.setError('')
      set('displayname', '')
      set('password', '')
      $client.logout()
    },
  },
)

export const storePattern = {
  state,
  mutations,
  actions,
  getters,
  modules: { connection, video, chat, files, openinapp, user, remote, settings, client, emoji },
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
