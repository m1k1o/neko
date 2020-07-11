import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { get, set } from '~/utils/localstorage'
import { accessor } from '~/store'

export const namespaced = true

interface KeyboardLayouts {
  [code: string]: string
}

export const state = () => {
  return {
    scroll: get<number>('scroll', 10),
    scroll_invert: get<boolean>('scroll_invert', true),
    autoplay: get<boolean>('autoplay', true),
    ignore_emotes: get<boolean>('ignore_emotes', false),
    chat_sound: get<boolean>('chat_sound', true),
    keyboard_layout: get<string>('keyboard_layout', 'us'),

    keyboard_layouts_list: {} as KeyboardLayouts,
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

  setKeyboardLayout(state, value: string) {
    state.keyboard_layout = value
    set('keyboard_layout', value)
  },

  setKeyboardLayoutsList(state, value: KeyboardLayouts) {
    state.keyboard_layouts_list = value
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    initialise() {
      $http
        .get<KeyboardLayouts>('/keyboard_layouts.json')
        .then((req) => {
          accessor.settings.setKeyboardLayoutsList(req.data)
          console.log(req.data)
        })
        .catch(console.error)
    },
  },
)
