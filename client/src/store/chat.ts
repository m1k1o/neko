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
  name?: string
  avatar?: string
}

const bubbleTimers: Record<string, number> = {}

export const state = () => ({
  history: [] as Message[],
  emotes: {} as Emotes,
  texts: 0,
  bubbles: {} as Record<string, string>,
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

  setHistory(state, messages: Message[]) {
    state.history = messages
    state.texts = messages.filter((message) => message.type === 'text').length
  },

  setBubble(state, { id, content }: { id: string; content: string }) {
    state.bubbles = { ...state.bubbles, [id]: content }
  },

  clearBubble(state, id: string) {
    const bubbles = { ...state.bubbles }
    delete bubbles[id]
    state.bubbles = bubbles
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
    state.bubbles = {}
    for (const id of Object.keys(bubbleTimers)) {
      window.clearTimeout(bubbleTimers[id])
      delete bubbleTimers[id]
    }
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
      accessor.chat.setBubble({ id: message.id, content: message.content })
      if (bubbleTimers[message.id]) {
        window.clearTimeout(bubbleTimers[message.id])
      }
      bubbleTimers[message.id] = window.setTimeout(() => {
        accessor.chat.clearBubble(message.id)
        delete bubbleTimers[message.id]
      }, 6500)
    },

    restoreHistory(store, messages: Message[]) {
      accessor.chat.setHistory(messages)
    },

    sendMessage(store, content: string) {
      if (!accessor.connection.connected || accessor.user.muted) {
        return
      }
      $client.sendMessage(EVENT.CHAT.MESSAGE, { text: content })
    },

    sendEmote(store, emote: string) {
      if (!accessor.connection.connected || accessor.user.muted) {
        return
      }
      $client.sendMessage(EVENT.CHAT.EMOTE, { emote })
    },
  },
)
