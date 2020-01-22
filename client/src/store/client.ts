import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { accessor } from '~/store'

export const namespaced = true

export const state = () => {
  let side = false
  let _side = localStorage.getItem('side')
  if (_side) {
    side = _side === '1'
  }

  let tab = 'chat'
  let _tab = localStorage.getItem('tab')
  if (_tab) {
    tab = _tab
  }

  return {
    side,
    about: false,
    about_page: '',
    tab,
  }
}

export const getters = getterTree(state, {})

export const mutations = mutationTree(state, {
  setTab(state, tab: string) {
    state.tab = tab
    localStorage.setItem('tab', tab)
  },
  setAbout(state, page: string) {
    state.about_page = page
  },
  toggleAbout(state) {
    state.about = !state.about
  },
  toggleSide(state) {
    state.side = !state.side
    localStorage.setItem('side', state.side ? '1' : '0')
  },
})

export const actions = actionTree({ state, getters, mutations }, {})
