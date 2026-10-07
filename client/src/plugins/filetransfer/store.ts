import { createStore } from '@/state/stores'

export interface FileItem {
  name: string
  type: 'file' | 'dir'
  size?: number
}

// filetransfer/update: the shared folder and what the viewer may do with it
export interface FileTransfer {
  enabled: boolean
  root_dir: string
  user_download: boolean
  user_upload: boolean
  user_delete: boolean
  files: FileItem[]
}

export interface Upload {
  id: number
  name: string
  size: number
  progress: number
  status: 'inprogress' | 'completed' | 'failed'
  error?: string
}

export const store = createStore({
  files: null as FileTransfer | null,
  uploads: [] as Upload[],
})
