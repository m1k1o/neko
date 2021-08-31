import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { get, set } from '~/utils/localstorage'
import { accessor } from '~/store'

export const namespaced = true

interface Group {
  name: string
  id: string
  list: string[]
}

interface Keywords {
  [name: string]: string[]
}

interface Emojis {
  groups: Group[]
  keywords: Keywords
  list: string[]
}

export const state = () => ({
  groups: [
    {
      id: 'recent',
      name: 'Recent',
      list: JSON.parse(get('emoji_recent', '[]')) as string[],
    },
  ] as Group[],
  keywords: {} as Keywords,
  list: [] as string[],
})

export const getters = getterTree(state, {})

export const mutations = mutationTree(state, {
  setRecent(state, emoji: string) {
    if (!state.groups[0].list.includes(emoji)) {
      if (state.groups[0].list.length > 30) {
        state.groups[0].list.shift()
      }
      state.groups[0].list.push(emoji)
      set('emoji_recent', JSON.stringify(state.groups[0].list))
    }
  },
  addGroup(state, group: Group) {
    state.groups.push(group)
  },
  setKeywords(state, keywords: Keywords) {
    state.keywords = keywords
  },
  setList(state, list: string[]) {
    state.list = list
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    async initialise() {
      try {
        const req = await $http.get<Emojis>('emoji.json')
        for (const group of req.data.groups) {
          accessor.emoji.addGroup(group)
        }
        accessor.emoji.setList(req.data.list)
        accessor.emoji.setKeywords(req.data.keywords)
      } catch (err: any) {
        console.error(err)
      }
    },
  },
)
