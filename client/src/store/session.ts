import Vue from 'vue'
import { actionTree, getterTree, mutationTree } from 'typed-vuex'
import { AdminLockResource } from '~/neko/messages'
import { accessor } from '~/store'
import { get, set } from '~/utils/localstorage'

export const namespaced = true

export const state = () => ({
  displayname: get<string>('displayname', ''),
  password: get<string>('password', ''),
  locked: {} as Record<string, boolean>,
})

export const mutations = mutationTree(state, {
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
    async lock(_, resource: AdminLockResource) {
      if (!accessor.connection.connected || !accessor.user.admin) return

      const setting = resource === 'login' ? 'locked_logins' : resource === 'control' ? 'locked_controls' : null
      if (setting) {
        await $http.post('/api/room/settings', { [setting]: true })
      } else {
        await setFileTransferEnabled(false)
      }
      accessor.session.setLocked(resource)
    },

    async unlock(_, resource: AdminLockResource) {
      if (!accessor.connection.connected || !accessor.user.admin) return

      const setting = resource === 'login' ? 'locked_logins' : resource === 'control' ? 'locked_controls' : null
      if (setting) {
        await $http.post('/api/room/settings', { [setting]: false })
      } else {
        await setFileTransferEnabled(true)
      }
      accessor.session.setUnlocked(resource)
    },

    toggleLock(_, resource: AdminLockResource) {
      if (accessor.session.isLocked(resource)) {
        accessor.session.unlock(resource)
      } else {
        accessor.session.lock(resource)
      }
    },

    login(_, { displayname, password }: { displayname: string; password: string }) {
      accessor.session.setLogin({ displayname, password })
      $client.login(password, displayname)
    },

    logout() {
      accessor.session.setLogin({ displayname: '', password: '' })
      accessor.connection.setError('')
      set('displayname', '')
      set('password', '')
      $client.logout()
    },
  },
)
