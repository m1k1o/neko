import { createStore } from 'zustand/vanilla'
import type { PluginEvents } from '@/plugins/types'

// filetransfer/update: the shared folder and what the viewer may do with it
export type FileTransfer = PluginEvents['filetransfer/update']
export type FileItem = FileTransfer['files'][number]

export interface Upload {
  id: number
  name: string
  size: number
  progress: number
  status: 'inprogress' | 'completed' | 'failed'
  error?: string
}

export const store = createStore(() => ({
  files: null as FileTransfer | null,
  uploads: [] as Upload[],
}))
