import { useStore } from 'zustand'
import { selectIsAdmin, type Settings } from '@m1k1o/neko'
import { client } from '@/state/client'
import { api } from '@/state/api'
import { nextId } from '@/state/dialogs'
import { store, type Upload } from './store'

export const refresh = () => client.send('filetransfer/update')

// the admin's lock: users may not transfer files while it is set
export const locked = (settings: Settings = client.state.settings) =>
  settings.plugins?.['filetransfer.enabled'] === false
export const toggleLock = () => api('POST', '/room/settings', { plugins: { 'filetransfer.enabled': locked() } })

// whether this viewer gets the Files tab
export function useAllowed() {
  const f = useStore(store, (s) => s.files)
  const admin = useStore(client.store, selectIsAdmin)
  const lock = useStore(client.store, (s) => locked(s.settings))
  return !!f?.enabled && (admin || !lock) && (admin || f.user_download || f.user_upload)
}

// plain link so the browser streams the download; the token only goes in the URL when the
// server runs without cookies (otherwise the session cookie authenticates it)
export const fileUrl = (name: string) =>
  `${client.api.url}/api/filetransfer?filename=${encodeURIComponent(name)}` +
  (client.api.token ? `&token=${encodeURIComponent(client.api.token)}` : '')

export const fileDelete = (name: string) => api('DELETE', `/filetransfer?filename=${encodeURIComponent(name)}`)

export function upload(files: FileList | File[]) {
  for (const file of files) {
    const u: Upload = { id: nextId(), name: file.name, size: file.size, progress: 0, status: 'inprogress' }
    store.setState((s) => ({ uploads: [...s.uploads, u] }))
    const update = (part: Partial<Upload>) =>
      store.setState((s) => ({ uploads: s.uploads.map((x) => (x.id === u.id ? { ...x, ...part } : x)) }))
    const form = new FormData()
    form.append('files', file)
    client.api
      .upload('/filetransfer', form, (p) => update({ progress: p.loaded }))
      .then(() => update({ progress: file.size, status: 'completed' }))
      .catch((err) => update({ status: 'failed', error: err.message }))
  }
}
