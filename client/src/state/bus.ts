import { Emitter } from '../core/store'
import { client, name } from './client'
import { t } from '@/i18n'

// signals between layers that must not import each other: the room's event lines (the chat
// plugin shows them) and the logout reset (every store with per-session data listens)
export const bus = new Emitter<{
  log: (id: string, name: string, content: string) => void
  logout: () => void
}>()

// an event line attributed to a session ("bob took the controls")
export const event = (id: string, content: string) =>
  bus.emit('log', id, id === client.state.session_id ? t('you') : name(id), content)

// an event line with a literal author ("You kicked bob")
export const line = (name: string, content: string) => bus.emit('log', '', name, content)
