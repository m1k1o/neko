import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { EVENT } from '~/neko/events'
import { accessor } from '~/store'

export const namespaced = true

export const state = () => ({
  enabled: false,
})

export const getters = getterTree(state, {
  //
})

export const mutations = mutationTree(state, {
  setEnabled(state, enabled: boolean) {
    state.enabled = enabled
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    sendOpenLink(store, url: string) {
      if (!accessor.connected) return
      $client.sendMessage(EVENT.OPENINAPP.OPENLINK, { text: url })
    },
  },
)
