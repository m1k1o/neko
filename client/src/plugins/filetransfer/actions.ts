import { useStore } from 'zustand'
import { selectIsAdmin, type Settings } from '@m1k1o/neko'
import type { NekoApp } from '@/state/app'
import { api } from '@/state/api'
import { useNeko } from '@/state/provider'
import { store, type Upload } from './store'

export const refresh = ({ client }: NekoApp) => client.send('filetransfer/update')

// the admin's lock: users may not transfer files while it is set
export const locked = (settings: Settings) => settings.plugins?.['filetransfer.enabled'] === false
export const toggleLock = (neko: NekoApp) =>
  api(neko, 'POST', '/room/settings', { plugins: { 'filetransfer.enabled': locked(neko.client.state.settings) } })

// whether this viewer gets the Files tab
export function useAllowed() {
  const neko = useNeko()
  const f = useStore(store(neko), (s) => s.files)
  const admin = useStore(neko.client.store, selectIsAdmin)
  const lock = useStore(neko.client.store, (s) => locked(s.settings))
  return !!f?.enabled && (admin || !lock) && (admin || f.user_download || f.user_upload)
}

// plain link so the browser streams the download; the token only goes in the URL when the
// server runs without cookies (otherwise the session cookie authenticates it)
export const fileUrl = ({ client }: NekoApp, name: string) =>
  `${client.api.url}/api/filetransfer?filename=${encodeURIComponent(name)}` +
  (client.api.token ? `&token=${encodeURIComponent(client.api.token)}` : '')

export const fileDelete = (neko: NekoApp, name: string) =>
  api(neko, 'DELETE', `/filetransfer?filename=${encodeURIComponent(name)}`)

export function upload(neko: NekoApp, files: FileList | File[]) {
  const uploads = store(neko)
  for (const file of files) {
    const u: Upload = { id: crypto.randomUUID(), name: file.name, size: file.size, progress: 0, status: 'inprogress' }
    uploads.setState((s) => ({ uploads: [...s.uploads, u] }))
    const update = (part: Partial<Upload>) =>
      uploads.setState((s) => ({ uploads: s.uploads.map((x) => (x.id === u.id ? { ...x, ...part } : x)) }))
    const form = new FormData()
    form.append('files', file)
    neko.client.api
      .upload('/filetransfer', form, (p) => update({ progress: p.loaded }))
      .then(() => update({ progress: file.size, status: 'completed' }))
      .catch((err) => update({ status: 'failed', error: err.message }))
  }
}
