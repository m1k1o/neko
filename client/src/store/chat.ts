import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { EVENT } from '~/client/events'
import { accessor } from '~/store'

export const namespaced = true

interface Message {
  id: string
  content: string
  created: Date
  type: 'text' | 'event'
}

export const state = () => ({
  history: [] as Message[],
})

export const getters = getterTree(state, {
  //
})

export const mutations = mutationTree(state, {
  addMessage(state, message: Message) {
    state.history = state.history.concat([message])
  },
  clear(state) {
    state.history = []
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    sendMessage(store, content: string) {
      if (!accessor.connected || accessor.user.muted) {
        return
      }

      $client.sendMessage(EVENT.CHAT.MESSAGE, { content })
    },

    sendEmoji(store, emoji: string) {
      if (!accessor.connected || !accessor.user.muted) {
        return
      }

      $client.sendMessage(EVENT.CHAT.EMOJI, { emoji })
    },
  },
)
