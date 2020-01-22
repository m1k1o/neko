import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { accessor } from '~/store'

export const namespaced = true

export const state = () => ({
  scroll: 10,
  scroll_invert: true,
})

export const getters = getterTree(state, {})

export const mutations = mutationTree(state, {
  setScroll(state, scroll: number) {
    state.scroll = scroll
    localStorage.setItem('scroll', `${scroll}`)
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    initialise() {
      const scroll = localStorage.getItem('scroll')
      if (scroll) {
        accessor.settings.setScroll(parseInt(scroll))
      }
    },
  },
)
