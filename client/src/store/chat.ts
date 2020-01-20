import { getterTree, mutationTree, actionTree } from 'typed-vuex'

export const namespaced = true

interface Message {
  id: string
  content: string
  created: Date
}

export const state = () => ({
  messages: [] as Message[],
})

export const getters = getterTree(state, {
  //
})

export const mutations = mutationTree(state, {
  addMessage(state, message: Message) {
    state.messages = state.messages.concat([message])
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    //
  },
)
