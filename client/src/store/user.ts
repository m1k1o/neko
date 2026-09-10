import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { Member } from '~/neko/types'

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
    async ban({ state }, member: string | Member) {
      if (!accessor.connection.connected || !accessor.user.admin) {
        return
      }

      if (typeof member === 'string') {
        member = state.members[member]
      }

      if (!member) {
        return
      }

      const endpoint = `/api/members/${encodeURIComponent(member.id)}`
      const response = await $http.get(endpoint)
      await $http.post(endpoint, { ...response.data, can_login: false })
    },

    async kick({ state }, member: string | Member) {
      if (!accessor.connection.connected || !accessor.user.admin) {
        return
      }

      if (typeof member === 'string') {
        member = state.members[member]
      }

      if (!member) {
        return
      }

      await $http.post(`/api/sessions/${encodeURIComponent(member.id)}/disconnect`)
    },

    mute({ state }, member: string | Member) {
      if (!accessor.connection.connected || !accessor.user.admin) {
        return
      }

      if (typeof member === 'string') {
        member = state.members[member]
      }

      if (!member) {
        return
      }

      accessor.user.setMuted({ id: member.id, muted: true })
    },

    unmute({ state }, member: string | Member) {
      if (!accessor.connection.connected || !accessor.user.admin) {
        return
      }

      if (typeof member === 'string') {
        member = state.members[member]
      }

      if (!member) {
        return
      }

      accessor.user.setMuted({ id: member.id, muted: false })
    },
  },
)
