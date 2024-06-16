export interface Settings {
  enabled: boolean
}

export const FILETRANSFER_UPDATE = "filetransfer/update"

export interface Message {
	enabled: boolean
	root_dir: string
	files: Item[]
}

type ItemType = "file" | "dir"

export interface Item {
  name: string
  type: ItemType
  size: number
}

export interface FileTransfer {
  id: number
  name: string
  direction: 'upload' | 'download'
  size: number
  progress: number
  status: 'pending' | 'inprogress' | 'completed' | 'failed'
  error?: string
  abortController?: AbortController
}
