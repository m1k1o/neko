import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { Member } from '~/client/types'

export const namespaced = true

interface Members {
  [id: string]: Member
}

export const state = () => ({
  id: '',
  members: {} as Members,
})

export const getters = getterTree(state, {
  member: state => state.members[state.id] || null,
  admin: state => (state.members[state.id] ? state.members[state.id].admin : false),
})

export const mutations = mutationTree(state, {
  setMembers(state, members: Member[]) {
    const data: Members = {}
    for (const member of members) {
      data[member.id] = member
    }
    state.members = data
  },
  setMember(state, id: string) {
    state.id = id
  },
  addMember(state, member: Member) {
    state.members = {
      ...state.members,
      [member.id]: member,
    }
  },
  delMember(state, id: string) {
    const data = { ...state.members }
    delete data[id]
    state.members = data
  },
  clearMembers(state) {
    state.members = {}
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    //
  },
)
