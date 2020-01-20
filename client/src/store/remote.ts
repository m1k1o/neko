import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { Member } from '~/client/types'

export const namespaced = true

export const state = () => ({
  id: '',
})

export const getters = getterTree(state, {
  hosting: (state, getters, root) => {
    return root.user.id === state.id
  },
  host: (state, getters, root) => {
    return root.user.member[state.id] || null
  },
})

export const mutations = mutationTree(state, {
  clearHost(state) {
    state.id = ''
  },
  setHost(state, host: string | Member) {
    if (typeof host === 'string') {
      state.id = host
    } else {
      state.id = host.id
    }
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    //
  },
)
