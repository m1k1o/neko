export interface Member {
  id: string
  displayname: string
  avatar?: string
  admin: boolean
  muted: boolean
  connected?: boolean
  ignored?: boolean
}

/** Screen modes returned by GET /api/room/screen/configurations. */
export type ScreenConfigurations = ScreenResolution[]

export interface ScreenResolution {
  width: number
  height: number
  rate: number
}

export interface FileListItem {
  name: string
  type: 'file' | 'dir'
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
