import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { Member } from '~/neko/types'
import { EVENT } from '~/neko/events'

import md from 'simple-markdown'
import { accessor } from '~/store'

export const namespaced = true

interface Members {
  [id: string]: Member
}

export const state = () => ({
  id: '',
  members: {} as Members,
})

export const getters = getterTree(state, {
  member: (state) => state.members[state.id] || null,
  admin: (state) => (state.members[state.id] ? state.members[state.id].admin : false),
  muted: (state) => (state.members[state.id] ? state.members[state.id].muted : false),
})

export const mutations = mutationTree(state, {
  setIgnored(state, { id, ignored }: { id: string; ignored: boolean }) {
    state.members[id] = {
      ...state.members[id],
      ignored,
    }
  },
  setMuted(state, { id, muted }: { id: string; muted: boolean }) {
    state.members[id] = {
      ...state.members[id],
      muted,
    }
  },
  setMembers(state, members: Member[]) {
    const data: Members = {}
    for (const member of members) {
      data[member.id] = {
        connected: true,
        ...member,
        displayname: md.sanitizeText(member.displayname),
      }
    }
    state.members = data
  },
  setMember(state, id: string) {
    state.id = id
  },
  addMember(state, member: Member) {
    state.members = {
      ...state.members,
      [member.id]: {
        connected: true,
        ...member,
        displayname: md.sanitizeText(member.displayname),
      },
    }
  },
  delMember(state, id: string) {
    state.members[id] = {
      ...state.members[id],
      connected: false,
    }
  },
  reset(state) {
    state.members = {}
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    ban({ state }, member: string | Member) {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      if (typeof member === 'string') {
        member = state.members[member]
      }

      if (!member) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.BAN, { id: member.id })
    },

    kick({ state }, member: string | Member) {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      if (typeof member === 'string') {
        member = state.members[member]
      }

      if (!member) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.KICK, { id: member.id })
    },

    mute({ state }, member: string | Member) {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      if (typeof member === 'string') {
        member = state.members[member]
      }

      if (!member) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.MUTE, { id: member.id })
    },

    unmute({ state }, member: string | Member) {
      if (!accessor.connected || !accessor.user.admin) {
        return
      }

      if (typeof member === 'string') {
        member = state.members[member]
      }

      if (!member) {
        return
      }

      $client.sendMessage(EVENT.ADMIN.UNMUTE, { id: member.id })
    },
  },
)
