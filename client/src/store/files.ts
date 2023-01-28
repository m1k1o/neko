import { actionTree, getterTree, mutationTree } from 'typed-vuex'
import { FileListItem, FileTransfer } from '~/neko/types'
import { EVENT } from '~/neko/events'
import { accessor } from '~/store'

export const state = () => ({
  cwd: '',
  files: [] as FileListItem[],
  transfers: [] as FileTransfer[],
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
  },

  _addTransfer(state, transfer: FileTransfer) {
    state.transfers = [...state.transfers, transfer]
  },

  _removeTransfer(state, transfer: FileTransfer) {
    state.transfers = state.transfers.filter((t) => t.id !== transfer.id)
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    setCwd(store, cwd: string) {
      accessor.files._setCwd(cwd)
    },

    setFileList(store, files: FileListItem[]) {
      accessor.files._setFileList(files)
    },

    addTransfer(store, transfer: FileTransfer) {
      if (transfer.status !== 'pending') {
        return
      }
      accessor.files._addTransfer(transfer)
    },

    removeTransfer(store, transfer: FileTransfer) {
      accessor.files._removeTransfer(transfer)
    },

    cancelAllTransfers() {
      for (const t of accessor.files.transfers) {
        if (t.status !== 'completed') {
          t.abortController?.abort()
        }
        accessor.files.removeTransfer(t)
      }
    },

    refresh() {
      if (!accessor.connected) {
        return
      }
      $client.sendMessage(EVENT.FILETRANSFER.REFRESH)
    },
  },
)
