import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { Member } from '~/neko/types'
import { EVENT } from '~/neko/events'
import { accessor } from '~/store'

const keyboardModifierState = (capsLock: boolean, numLock: boolean, scrollLock: boolean) =>
  Number(capsLock) + 2 * Number(numLock) + 4 * Number(scrollLock)

export const namespaced = true

export const state = () => ({
  id: '',
  epoch: 0,
  clipboard: '',
  locked: false,
  implicitHosting: true,
  fileTransfer: true,
  keyboardModifierState: -1,
})

export const getters = getterTree(state, {
  controlling: (state, getters, root) => {
    return root.user.id === state.id
  },
  hosting: (state, getters, root) => {
    return root.user.id === state.id || state.implicitHosting
  },
  hosted: (state) => {
    return state.id !== '' || state.implicitHosting
  },
  host: (state, getters, root) => {
    return root.user.members[state.id] || (state.implicitHosting && root.user.id) || null
  },
})

export const mutations = mutationTree(state, {
  setHost(state, host: string | Member) {
    if (typeof host === 'string') {
      state.id = host
    } else {
      state.id = host.id
    }
  },

  setEpoch(state, epoch: number) {
    state.epoch = epoch
  },

  setClipboard(state, clipboard: string) {
    state.clipboard = clipboard
  },

  setKeyboardModifierState(state, { capsLock, numLock, scrollLock }) {
    state.keyboardModifierState = keyboardModifierState(capsLock, numLock, scrollLock)
  },

  setLocked(state, locked: boolean) {
    state.locked = locked
  },

  setImplicitHosting(state, val: boolean) {
    state.implicitHosting = val
  },

  setFileTransfer(state, val: boolean) {
    state.fileTransfer = val
  },

  reset(state) {
    state.id = ''
    state.epoch = 0
    state.clipboard = ''
    state.locked = false
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    sendClipboard({ getters }, clipboard: string) {
      if (!accessor.connection.connected || !getters.hosting) {
        return
      }

      $client.sendMessage(EVENT.CLIPBOARD.SET, { text: clipboard })
    },

    toggle({ getters }) {
      if (!accessor.connection.connected) {
        return
      }

      if (!getters.hosting) {
        $client.sendMessage(EVENT.CONTROL.REQUEST)
      } else {
        $client.sendMessage(EVENT.CONTROL.RELEASE)
      }
    },

    request({ getters }) {
      if (!accessor.connection.connected || getters.controlling) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.REQUEST)
    },

    release({ getters }) {
      if (!accessor.connection.connected || !getters.hosting) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.RELEASE)
    },

    async give({ getters }, member: string | Member) {
      if (!accessor.connection.connected || !getters.hosting) {
        return
      }

      if (typeof member === 'string') {
        member = accessor.user.members[member]
      }

      if (!member) {
        return
      }

      await $client.room.giveControl(member.id)
    },

    adminControl() {
      if (!accessor.connection.connected || !accessor.user.admin) {
        return
      }

      $client.room.takeControl()
    },

    adminRelease() {
      if (!accessor.connection.connected || !accessor.user.admin) {
        return
      }

      $client.room.resetControl()
    },

    adminGive(store, member: string | Member) {
      if (!accessor.connection.connected) {
        return
      }

      if (typeof member === 'string') {
        member = accessor.user.members[member]
      }

      if (!member) {
        return
      }

      $client.room.giveControl(member.id)
    },

    changeKeyboard({ getters }) {
      if (!accessor.connection.connected || !getters.hosting) {
        return
      }

      $client.sendMessage(EVENT.KEYBOARD.MAP, { layout: accessor.settings.keyboard_layout })
    },

    syncKeyboardModifierState({ state }, { capsLock, numLock, scrollLock }) {
      if (state.keyboardModifierState === keyboardModifierState(capsLock, numLock, scrollLock)) {
        return
      }

      accessor.remote.setKeyboardModifierState({ capsLock, numLock, scrollLock })
      $client.sendMessage(EVENT.KEYBOARD.MODIFIERS, {
        capslock: capsLock,
        numlock: numLock,
      })
    },
  },
)
