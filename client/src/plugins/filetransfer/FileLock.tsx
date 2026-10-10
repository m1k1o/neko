import { useStore } from 'zustand'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { File } from 'lucide-react'
import { cn } from '@/lib/utils'
import { client } from '@/state/client'
import { IconButton } from '@/components/IconButton'
import { store } from './store'
import { locked, toggleLock } from './actions'

// the file-transfer lock in the header, once the server says file transfer is on: admins toggle it,
// users only see its state
export function FileLock() {
  const { t } = useTranslation()
  const enabled = useStore(store, (s) => !!s.files?.enabled)
  const admin = useStore(client.store, selectIsAdmin)
  const isLocked = useStore(client.store, (s) => locked(s.settings))
  if (!enabled) return null
  const tip = admin
    ? t(`files:locks.${isLocked ? 'unlock' : 'lock'}`)
    : t(`files:locks.${isLocked ? 'locked' : 'unlocked'}`)
  return (
    <li className="mr-2.5 inline-block">
      <IconButton
        variant="header"
        label={tip}
        className={cn(isLocked && 'text-style-error/50')}
        data-testid="lock-files"
        data-locked={isLocked || undefined}
        aria-disabled={!admin}
        onClick={() => admin && toggleLock()}
      >
        <File className="size-4" />
      </IconButton>
    </li>
  )
}
