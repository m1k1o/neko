import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { Member } from '~/neko/types'
import { EVENT } from '~/neko/events'
import { accessor } from '~/store'

const keyboardModifierState = (capsLock: boolean, numLock: boolean, scrollLock: boolean) =>
  Number(capsLock) + 2 * Number(numLock) + 4 * Number(scrollLock)

export const namespaced = true

export const state = () => ({
  id: '',
  clipboard: '',
  locked: false,
  implicitHosting: true,
  fileTransfer: true,
  keyboardModifierState: -1,
})

export const getters = getterTree(state, {
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
    state.clipboard = ''
    state.locked = false
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    sendClipboard({ getters }, clipboard: string) {
      if (!accessor.connected || !getters.hosting) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.CLIPBOARD, { text: clipboard })
    },

    toggle({ getters }) {
      if (!accessor.connected) {
        return
      }

      if (!getters.hosting) {
        $client.sendMessage(EVENT.CONTROL.REQUEST)
      } else {
        $client.sendMessage(EVENT.CONTROL.RELEASE)
      }
    },

    request({ getters }) {
      if (!accessor.connected || getters.hosting) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.REQUEST)
    },

    release({ getters }) {
      if (!accessor.connected || !getters.hosting) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.RELEASE)
    },

    give({ getters }, member: string | Member) {
      if (!accessor.connected || !getters.hosting) {
        return
      }

      if (typeof member === 'string') {
        member = accessor.user.members[member]
      }

      if (!member) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.GIVE, { id: member.id })
    },

    adminControl() {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.CONTROL)
    },

    adminRelease() {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.RELEASE)
    },

    adminGive(store, member: string | Member) {
      if (!accessor.connected) {
        return
      }

      if (typeof member === 'string') {
        member = accessor.user.members[member]
      }

      if (!member) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.GIVE, { id: member.id })
    },

    changeKeyboard({ getters }) {
      if (!accessor.connected || !getters.hosting) {
        return
      }

      $client.sendMessage(EVENT.CONTROL.KEYBOARD, { layout: accessor.settings.keyboard_layout })
    },

    syncKeyboardModifierState({ state }, { capsLock, numLock, scrollLock }) {
      if (state.keyboardModifierState === keyboardModifierState(capsLock, numLock, scrollLock)) {
        return
      }

      accessor.remote.setKeyboardModifierState({ capsLock, numLock, scrollLock })
      $client.sendMessage(EVENT.CONTROL.KEYBOARD, { capsLock, numLock, scrollLock })
    },
  },
)
