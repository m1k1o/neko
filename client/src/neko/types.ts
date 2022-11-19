export interface Member {
  id: string
  displayname: string
  admin: boolean
  muted: boolean
  connected?: boolean
  ignored?: boolean
}

export interface ScreenConfigurations {
  [index: string]: ScreenConfiguration
}

export interface ScreenConfiguration {
  width: number
  height: number
  rates: { [index: string]: number }
}

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
