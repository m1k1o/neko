// file transfer: the shared folder and the viewer's uploads into it
import { app } from './app'
import { client } from './client'
import { api } from './api'
import { nextId } from './dialogs'

export interface FileItem {
  name: string
  type: 'file' | 'dir'
  size?: number
}

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

const s = app.state

export const filesRefresh = () => client.send('filetransfer/update')

// plain link so the browser streams the download; the token only goes in the URL when the
// server runs without cookies (otherwise the session cookie authenticates it)
export const fileUrl = (name: string) =>
  `${client.api.url}/api/filetransfer?filename=${encodeURIComponent(name)}` +
  (client.api.token ? `&token=${encodeURIComponent(client.api.token)}` : '')

export const fileDelete = (name: string) => api('DELETE', `/filetransfer?filename=${encodeURIComponent(name)}`)

export function upload(files: FileList | File[]) {
  for (const file of files) {
    const u: Upload = { id: nextId(), name: file.name, size: file.size, progress: 0, status: 'inprogress' }
    s.uploads.push(u)
    const live = () => s.uploads.find((x) => x.id === u.id)!
    const form = new FormData()
    form.append('files', file)
    client.api
      .upload('/filetransfer', form, (p) => (live().progress = p.loaded))
      .then(() => Object.assign(live(), { progress: file.size, status: 'completed' }))
      .catch((err) => Object.assign(live(), { status: 'failed', error: err.message }))
  }
}
