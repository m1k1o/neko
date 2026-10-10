import { Emitter } from '@m1k1o/neko'
import { name, type NekoApp } from './app'
import { t } from '@/i18n'

// signals between layers that must not import each other: the room's event lines (the chat
// plugin shows them) and the logout reset (every store with per-session data listens)
export type Bus = Emitter<{
  log: (id: string, name: string, content: string) => void
  logout: () => void
}>
export const createBus = (): Bus => new Emitter()

// an event line attributed to a session ("bob took the controls")
export const event = ({ bus, client }: NekoApp, id: string, content: string) =>
  bus.emit('log', id, id === client.state.session_id ? t('you') : name(client, id), content)

// an event line with a literal author ("You kicked bob")
export const line = ({ bus }: NekoApp, name: string, content: string) => bus.emit('log', '', name, content)
