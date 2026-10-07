import { client } from '@/state/client'
import { bus, event } from '@/state/bus'
import { t } from '@/i18n'
import type { Plugin } from '../types'
import { Files } from './Files'
import { FileLock } from './FileLock'
import { store } from './store'
import { allowed, locked, refresh } from './actions'
import { locale } from './locale'

export const filetransfer: Plugin = {
  id: 'filetransfer',
  locale,
  tab: { id: 'files', icon: 'fa-file', component: Files, visible: allowed },
  topBar: [{ id: 'filetransfer-lock', component: FileLock }],
  onEvent(event, payload) {
    if (event === 'filetransfer/update') store.state.files = payload
  },
  init() {
    // the list is requested once the session is known; a lock change becomes an event line
    let last: boolean | null = null
    client.store.watch(
      () => client.state.session_id,
      (id) => {
        if (!id) return
        last = locked()
        refresh()
      },
    )
    client.events.on('room.settings.updated', (next, id) => {
      const now = locked(next)
      if (last !== null && id && now !== last) event(id, t(`locks.file_transfer.notif_${now ? 'locked' : 'unlocked'}`))
      last = now
    })
    bus.on('logout', () => (store.state.uploads = []))
  },
}
