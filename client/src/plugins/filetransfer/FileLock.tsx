import { useNeko } from '@/state/hooks'
import { LockButton } from '@/components/LockButton'
import { t } from '@/i18n'
import { store } from './store'
import { locked, toggleLock } from './actions'

// the file-transfer lock in the header, once the server says file transfer is on
export function FileLock() {
  const { client } = useNeko()
  if (!store.state.files?.enabled) return null
  const admin = client.isAdmin
  const isLocked = locked()
  const tip = admin
    ? t(`locks.file_transfer.${isLocked ? 'unlock' : 'lock'}`)
    : t(`locks.file_transfer.${isLocked ? 'locked' : 'unlocked'}`)
  return <LockButton icon="fa-file" locked={isLocked} admin={admin} tip={tip} onToggle={toggleLock} />
}
