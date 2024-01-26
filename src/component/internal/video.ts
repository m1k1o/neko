import Vue from 'vue'
import { Video } from '../types/state'

export function register(el: HTMLVideoElement, state: Video) {
  el.addEventListener('canplaythrough', () => {
    Vue.set(state, 'playable', true)
  })
  el.addEventListener('playing', () => {
    Vue.set(state, 'playing', true)
  })
  el.addEventListener('pause', () => {
    Vue.set(state, 'playing', false)
  })
  el.addEventListener('emptied', () => {
    Vue.set(state, 'playable', false)
    Vue.set(state, 'playing', false)
  })
  el.addEventListener('error', () => {
    Vue.set(state, 'playable', false)
    Vue.set(state, 'playing', false)
  })
  el.addEventListener('volumechange', () => {
    Vue.set(state, 'muted', el.muted)
    Vue.set(state, 'volume', el.volume)
  })

  // Initial state
  Vue.set(state, 'muted', el.muted)
  Vue.set(state, 'volume', el.volume)
  Vue.set(state, 'playing', !el.paused)
}
