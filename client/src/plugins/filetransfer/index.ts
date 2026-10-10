import { client } from '@/state/client'
import { bus, event } from '@/state/bus'
import { t } from '@/i18n'
import type { Plugin } from '@/plugins/types'
import { Files } from './Files'
import { FileLock } from './FileLock'
import { store } from './store'
import { useAllowed, locked, refresh } from './actions'

export const filetransfer: Plugin = {
  id: 'filetransfer',
  ns: 'files',
  tab: { id: 'files', icon: 'fa-file', label: 'files:tab', component: Files, useVisible: useAllowed },
  topBar: [{ id: 'filetransfer-lock', component: FileLock }],
  onEvent(...[event, payload]) {
    if (event === 'filetransfer/update') store.setState({ files: payload })
  },
  init() {
    // the list is requested once the session is known; a lock change becomes an event line
    let last: boolean | null = null
    client.store.subscribe(
      (s) => s.session_id,
      (id) => {
        if (!id) return
        last = locked()
        refresh()
      },
    )
    client.events.on('room.settings.updated', (next, id) => {
      const now = locked(next)
      if (last !== null && id && now !== last) event(id, t(`files:locks.notif_${now ? 'locked' : 'unlocked'}`))
      last = now
    })
    bus.on('logout', () => store.setState({ uploads: [] }))
  },
}
