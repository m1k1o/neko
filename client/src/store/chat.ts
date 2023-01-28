import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { makeid } from '~/utils'
import { EVENT } from '~/neko/events'
import { accessor } from '~/store'

export const namespaced = true

interface Emote {
  type: string
}

interface Emotes {
  [id: string]: Emote
}

interface Message {
  id: string
  content: string
  created: Date
  type: 'text' | 'event'
}

export const state = () => ({
  history: [] as Message[],
  emotes: {} as Emotes,
  texts: 0,
})

export const getters = getterTree(state, {
  //
})

export const mutations = mutationTree(state, {
  addMessage(state, message: Message) {
    if (message.type == 'text') {
      state.texts++
    }

    state.history = state.history.concat([message])
  },

  addEmote(state, { id, emote }: { id: string; emote: Emote }) {
    state.emotes = {
      ...state.emotes,
      [id]: emote,
    }
  },

  delEmote(state, id: string) {
    const emotes = {
      ...state.emotes,
    }
    delete emotes[id]
    state.emotes = emotes
  },

  reset(state) {
    state.emotes = {}
    state.history = []
    state.texts = 0
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    newEmote(store, emote: Emote) {
      if (accessor.settings.ignore_emotes || document.visibilityState === 'hidden') {
        return
      }

      const id = makeid(10)
      accessor.chat.addEmote({ id, emote })
    },

    newMessage(store, message: Message) {
      if (accessor.settings.chat_sound) {
        new Audio('chat.mp3').play().catch(console.error)
      }
      accessor.chat.addMessage(message)
    },

    sendMessage(store, content: string) {
      if (!accessor.connected || accessor.user.muted) {
        return
      }
      $client.sendMessage(EVENT.CHAT.MESSAGE, { content })
    },

    sendEmote(store, emote: string) {
      if (!accessor.connected || accessor.user.muted) {
        return
      }
      $client.sendMessage(EVENT.CHAT.EMOTE, { emote })
    },
  },
)
