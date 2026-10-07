// The one NekoClient of the GUI, and the small readers of its state that components share.
import { NekoClient } from '../core/client'
import type { Settings } from '../core/types'
import { app } from './app'
import { t } from '@/i18n'

export const client = new NekoClient({
  autologin: true,
  autoconnect: true,
  // read when a track arrives, so the setting applies without a reload
  get autoplay() {
    return app.state.settings.autoplay
  },
})

export const name = (id?: string | null) => (id && client.state.sessions[id]?.profile.name) || t('somebody')
export const isMuted = (id = client.state.session_id) =>
  !!id && client.state.sessions[id]?.profile.plugins?.['chat.can_send'] === false
export const hostId = () => client.state.control.host_id

export type LockResource = 'login' | 'control' | 'file_transfer'
export function isLocked(r: LockResource, settings: Settings = client.state.settings) {
  if (r === 'login') return settings.locked_logins
  if (r === 'control') return settings.locked_controls
  return settings.plugins?.['filetransfer.enabled'] === false
}
