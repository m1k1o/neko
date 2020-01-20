import { getterTree, mutationTree, actionTree } from 'typed-vuex'

export const namespaced = true

export const state = () => ({
  index: -1,
  streams: [] as MediaStream[],
  width: 1280,
  height: 720,
  volume: 0,
  playing: false,
})

export const getters = getterTree(state, {
  stream: state => state.streams[state.index],
  resolution: state => ({ w: state.width, h: state.height }),
  aspect: state => {
    const { width, height } = state

    if ((height == 0 && width == 0) || (height == 0 && width != 0) || (height != 0 && width == 0)) {
      return null
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

    return {
      horizontal: width / gcd,
      vertical: height / gcd,
    }
  },
})

export const mutations = mutationTree(state, {
  setResolution(state, { width, height }: { width: number; height: number }) {
    state.width = width
    state.height = height
  },

  setVolume(state, volume: number) {
    state.volume = volume
  },

  setStream(state, index: number) {
    state.index = index
  },

  addStream(state, stream: MediaStream) {
    state.streams = state.streams.concat([stream])
  },

  delStream(state, index: number) {
    state.streams = state.streams.filter((_, i) => i !== index)
  },

  clearStream(state) {
    state.streams = []
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    //
  },
)
