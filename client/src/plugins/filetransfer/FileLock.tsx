import { useStore } from 'zustand'
import { selectIsAdmin } from '@m1k1o/neko'
import { client } from '@/state/client'
import { LockButton } from '@/components/LockButton'
import { t } from '@/i18n'
import { store } from './store'
import { locked, toggleLock } from './actions'

// the file-transfer lock in the header, once the server says file transfer is on
export function FileLock() {
  const enabled = useStore(store, (s) => !!s.files?.enabled)
  const admin = useStore(client.store, selectIsAdmin)
  const isLocked = useStore(client.store, (s) => locked(s.settings))
  if (!enabled) return null
  const tip = admin
    ? t(`locks.file_transfer.${isLocked ? 'unlock' : 'lock'}`)
    : t(`locks.file_transfer.${isLocked ? 'locked' : 'unlocked'}`)
  return <LockButton icon="fa-file" locked={isLocked} admin={admin} tip={tip} onToggle={toggleLock} />
}
