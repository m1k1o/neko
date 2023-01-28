import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { get, set } from '~/utils/localstorage'
import { EVENT } from '~/neko/events'
import { ScreenConfigurations, ScreenResolution } from '~/neko/types'
import { accessor } from '~/store'

export const namespaced = true

export const state = () => ({
  index: -1,
  tracks: [] as MediaStreamTrack[],
  streams: [] as MediaStream[],
  configurations: [] as ScreenResolution[],
  width: 1280,
  height: 720,
  rate: 30,
  horizontal: 16,
  vertical: 9,
  volume: get<number>('volume', 100),
  muted: get<boolean>('muted', false),
  playing: false,
  playable: false,
})

export const getters = getterTree(state, {
  stream: (state) => state.streams[state.index],
  track: (state) => state.tracks[state.index],
  resolution: (state) => ({ w: state.width, h: state.height }),
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

  setMuted(state, muted: boolean) {
    state.muted = muted
    set('mute', muted)
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

  setResolution(state, { width, height, rate }: { width: number; height: number; rate: number }) {
    state.width = width
    state.height = height
    state.rate = rate

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

  setConfigurations(state, configurations: ScreenConfigurations) {
    const data: ScreenResolution[] = []

    for (const i of Object.keys(configurations)) {
      const { width, height, rates } = configurations[i]
      if (width >= 600 && height >= 300) {
        for (const j of Object.keys(rates)) {
          const rate = rates[j]
          if (rate === 30 || rate === 60) {
            data.push({
              width,
              height,
              rate,
            })
          }
        }
      }
    }

    state.configurations = data.sort((a, b) => {
      if (b.width === a.width && b.height == a.height) {
        return b.rate - a.rate
      } else if (b.width === a.width) {
        return b.height - a.height
      }
      return b.width - a.width
    })
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

  reset(state) {
    state.index = -1
    state.tracks = []
    state.streams = []
    state.configurations = []
    state.width = 1280
    state.height = 720
    state.rate = 30
    state.horizontal = 16
    state.vertical = 9
    state.playing = false
    state.playable = false
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    screenConfiguations() {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      $client.sendMessage(EVENT.SCREEN.CONFIGURATIONS)
    },

    screenGet() {
      if (!accessor.connected) {
        return
      }

      $client.sendMessage(EVENT.SCREEN.RESOLUTION)
    },

    screenSet(store, resolution: ScreenResolution) {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      $client.sendMessage(EVENT.SCREEN.SET, resolution)
    },
  },
)
