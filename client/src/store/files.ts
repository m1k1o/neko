import { actionTree, getterTree, mutationTree } from 'typed-vuex'
import { FileListItem } from '~/neko/types'
import { accessor } from '~/store'

export const state = () => ({
  cwd: '',
  files: [] as FileListItem[]
})

export const getters = getterTree(state, {
  //
})

export const mutations = mutationTree(state, {
  _setCwd(state, cwd: string) {
    state.cwd = cwd
  },

  _setFileList(state, files: FileListItem[]) {
    state.files = files
  }
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    setCwd(store, cwd: string) {
      accessor.files._setCwd(cwd)
    },

    setFileList(store, files: FileListItem[]) {
      accessor.files._setFileList(files)
    }
  }
)
