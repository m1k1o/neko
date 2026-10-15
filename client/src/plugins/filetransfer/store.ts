import { createStore } from 'zustand/vanilla'
import { scoped } from '@/state/app'
import type { PluginEvents } from '@/plugins/types'

// filetransfer/update: the shared folder and what the viewer may do with it
export type FileTransfer = PluginEvents['filetransfer/update']
export type FileItem = FileTransfer['files'][number]

export interface Upload {
  id: string // uid()
  name: string
  size: number
  progress: number
  status: 'inprogress' | 'completed' | 'failed'
  error?: string
}

// the store of an instance (`store(neko)`, `store(useNeko())`)
export const store = scoped(() =>
  createStore(() => ({
    files: null as FileTransfer | null,
    uploads: [] as Upload[],
  })),
)
