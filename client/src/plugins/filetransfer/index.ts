import { File } from 'lucide-react'
import { event } from '@/state/bus'
import { t } from '@/i18n'
import type { Plugin } from '@/plugins/types'
import { Files } from './Files'
import { FileLock } from './FileLock'
import { store } from './store'
import { useAllowed, locked, refresh } from './actions'

export const filetransfer: Plugin = {
  id: 'filetransfer',
  ns: 'files',
  tab: { id: 'files', icon: File, label: 'files:tab', component: Files, useVisible: useAllowed },
  topBar: [{ id: 'filetransfer-lock', component: FileLock }],
  onEvent(neko, ...[event, payload]) {
    if (event === 'filetransfer/update') store(neko).setState({ files: payload })
  },
  init(neko) {
    const { client, bus } = neko
    // the list is requested once the session is known; a lock change becomes an event line
    let last: boolean | null = null
    client.store.subscribe(
      (s) => s.session_id,
      (id) => {
        if (!id) return
        last = locked(client.state.settings)
        refresh(neko)
      },
    )
    client.events.on('room.settings.updated', (next, id) => {
      const now = locked(next)
      if (last !== null && id && now !== last) event(neko, id, t(`files:locks.notif_${now ? 'locked' : 'unlocked'}`))
      last = now
    })
    bus.on('logout', () => store(neko).setState({ uploads: [] }))
  },
}
