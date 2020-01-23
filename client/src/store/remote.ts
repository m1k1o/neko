import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { Member } from '~/neko/types'
import { EVENT } from '~/neko/events'
import { accessor } from '~/store'

export const namespaced = true

export const state = () => ({
  id: '',
})

export const getters = getterTree(state, {
  hosting: (state, getters, root) => {
    return root.user.id === state.id
  },
  hosted: (state, getters, root) => {
    return state.id !== ''
  },
  host: (state, getters, root) => {
    return root.user.member[state.id] || null
  },
})

export const mutations = mutationTree(state, {
  setHost(state, host: string | Member) {
    if (typeof host === 'string') {
      state.id = host
    } else {
      state.id = host.id
    }
  },

  clear(state) {
    state.id = ''
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    initialise({ commit }) {
      //
    },
    toggle({ getters }) {
      if (!accessor.connected) {
        return
      }

      if (!getters.hosting) {
        $client.sendMessage(EVENT.CONTROL.REQUEST)
      } else {
        $client.sendMessage(EVENT.CONTROL.RELEASE)
      }
    },

    request({ getters }) {
      if (!accessor.connected || !getters.hosting) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.REQUEST)
    },

    release({ getters }) {
      if (!accessor.connected || getters.hosting) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.RELEASE)
    },

    give({ getters }, member: string | Member) {
      if (!accessor.connected || !getters.hosting) {
        return
      }

      if (typeof member === 'string') {
        member = accessor.user.members[member]
      }

      if (!member) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.GIVE, { id: member.id })
    },

    adminControl() {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.CONTROL)
    },

    adminRelease() {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.RELEASE)
    },

    adminGive({ getters }, member: string | Member) {
      if (!accessor.connected) {
        return
      }

      if (typeof member === 'string') {
        member = accessor.user.members[member]
      }

      if (!member) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.GIVE, { id: member.id })
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
  },
)
