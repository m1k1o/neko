import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { get, set } from '~/utils/localstorage'

export const namespaced = true

export const state = () => ({
  index: -1,
  tracks: [] as MediaStreamTrack[],
  streams: [] as MediaStream[],
  width: 1280,
  height: 720,
  horizontal: 16,
  vertical: 9,
  volume: get<number>('volume', 100),
  muted: get<boolean>('muted', false),
  playing: false,
  playable: false,
})

export const getters = getterTree(state, {
  stream: state => state.streams[state.index],
  track: state => state.tracks[state.index],
  resolution: state => ({ w: state.width, h: state.height }),
})

export const mutations = mutationTree(state, {
  play(state) {
    if (state.playable) {
      state.playing = true
    }
  },

  pause(state) {
    if (state.playable) {
      state.playing = false
    }
  },

  togglePlay(state) {
    if (state.playable) {
      state.playing = !state.playing
    }
  },

  toggleMute(state) {
    state.muted = !state.muted
    set('mute', state.muted)
  },

  setPlayable(state, playable: boolean) {
    if (!playable && state.playing) {
      state.playing = false
    }
    state.playable = playable
  },

  setResolution(state, { width, height }: { width: number; height: number }) {
    state.width = width
    state.height = height

    if ((height == 0 && width == 0) || (height == 0 && width != 0) || (height != 0 && width == 0)) {
      return
    }

    if (height == width) {
      return {
        horizontal: 1,
        vertical: 1,
      }
    }

    let dividend = width
    let divisor = height
    let gcd = -1

    if (height > width) {
      dividend = height
      divisor = width
    }

    while (gcd == -1) {
      const remainder = dividend % divisor
      if (remainder == 0) {
        gcd = divisor
      } else {
        dividend = divisor
        divisor = remainder
      }
    }

    state.horizontal = width / gcd
    state.vertical = height / gcd
  },

  setVolume(state, volume: number) {
    state.volume = volume
    set('volume', volume)
  },

  setStream(state, index: number) {
    state.index = index
  },

  addTrack(state, [track, stream]: [MediaStreamTrack, MediaStream]) {
    state.tracks = state.tracks.concat([track])
    state.streams = state.streams.concat([stream])
  },

  delTrack(state, index: number) {
    state.streams = state.streams.filter((_, i) => i !== index)
    state.tracks = state.tracks.filter((_, i) => i !== index)
  },

  clear(state) {
    state.index = -1
    state.tracks = []
    state.streams = []
  },
})
