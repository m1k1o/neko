// The one NekoClient of the GUI with its keyboard/mouse/touch device, and the small readers of
// its state that components share.
import { NekoClient, Overlay } from '@m1k1o/neko'
import type { Settings, State } from '@m1k1o/neko'
import { app } from './app'
import { t } from '@/i18n'

export const client = new NekoClient({
  autologin: true,
  autoconnect: true,
  // read when a track arrives, so the setting applies without a reload
  get autoplay() {
    return app.getState().settings.autoplay
  },
})
export const overlay = new Overlay(client)

export const name = (id?: string | null) => (id && client.state.sessions[id]?.profile.name) || t('somebody')
// the server's chat plugin took this member's right to send (chat.can_send); emotes follow it too
export const selectMuted = (s: State, id = s.session_id) =>
  !!id && s.sessions[id]?.profile.plugins?.['chat.can_send'] === false
export const isMuted = (id?: string | null) => selectMuted(client.state, id)
export const hostId = () => client.state.control.host_id

export type LockResource = 'login' | 'control'
export const isLocked = (r: LockResource, settings: Settings = client.state.settings) =>
  r === 'login' ? settings.locked_logins : settings.locked_controls
