import { getterTree, mutationTree } from 'typed-vuex'
import { get, set } from '~/utils/localstorage'

export const namespaced = true

export const state = () => {
  return {
    scroll: get<number>('scroll', 10),
    scroll_invert: get<boolean>('scroll_invert', true),
    autoplay: get<boolean>('autoplay', true),
    ignore_emotes: get<boolean>('ignore_emotes', false),
    chat_sound: get<boolean>('chat_sound', true),
  }
}

export const getters = getterTree(state, {})

export const mutations = mutationTree(state, {
  setScroll(state, scroll: number) {
    state.scroll = scroll
    set('scroll', scroll)
  },

  setInvert(state, value: boolean) {
    state.scroll_invert = value
    set('scroll_invert', value)
  },

  setAutoplay(state, value: boolean) {
    state.autoplay = value
    set('autoplay', value)
  },

  setIgnore(state, value: boolean) {
    state.ignore_emotes = value
    set('ignore_emotes', value)
  },

  setSound(state, value: boolean) {
    state.chat_sound = value
    set('chat_sound', value)
  },
})
