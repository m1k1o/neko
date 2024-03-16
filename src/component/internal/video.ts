import type { Video } from '../types/state'

export function register(el: HTMLVideoElement, state: Video) {
  el.addEventListener('canplaythrough', () => {
    state.playable = true
  })
  el.addEventListener('playing', () => {
    state.playing = true
  })
  el.addEventListener('pause', () => {
    state.playing = false
  })
  el.addEventListener('emptied', () => {
    state.playable = false
    state.playing = false
  })
  el.addEventListener('error', () => {
    state.playable = false
    state.playing = false
  })
  el.addEventListener('volumechange', () => {
    state.muted = el.muted
    state.volume = el.volume
  })

  // Initial state
  state.muted = el.muted
  state.volume = el.volume
  state.playing = !el.paused
}
