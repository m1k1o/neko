import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { get, set } from '~/utils/localstorage'

export const namespaced = true

export const state = () => ({
  side: get<boolean>('side', false),
  tab: get<string>('tab', 'chat'),
  about: false,
  about_page: '',
})

export const getters = getterTree(state, {})

export const mutations = mutationTree(state, {
  setTab(state, tab: string) {
    state.tab = tab
    set('tab', tab)
  },
  setAbout(state, page: string) {
    state.about_page = page
  },
  toggleAbout(state) {
    state.about = !state.about
  },
  toggleSide(state) {
    state.side = !state.side
    set('side', state.side)
  },
  setSide(state, side: boolean) {
    state.side = side
    set('side', side)
  },
})

export const actions = actionTree({ state, getters, mutations }, {})
