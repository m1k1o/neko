// the chat log: messages and the event lines the room writes into it
import { app } from './app'
import { client, name } from './client'
import { api } from './api'
import { t } from '@/i18n'

export interface ChatLine {
  seq: number // unique and increasing: the React key, and what scrolling follows
  id: string
  name: string
  type: 'text' | 'event'
  content: string
  created: Date
}

const s = app.state

// chat keeps the last CHAT_LIMIT lines in memory; nothing is persisted. Trimming in place keeps
// the proxies the store handed out (see store.ts wrap) and the row keys stable.
const CHAT_LIMIT = 1000
let chatSeq = 0
export function pushChat(line: Omit<ChatLine, 'seq'>) {
  s.chat.push({ ...line, seq: ++chatSeq })
  if (s.chat.length > CHAT_LIMIT) s.chat.splice(0, s.chat.length - CHAT_LIMIT)
}

// an event line attributed to a session ("bob took the controls")
export function event(id: string, content: string) {
  pushChat({
    id,
    name: id === client.state.session_id ? t('you') : name(id),
    type: 'event',
    content,
    created: new Date(),
  })
}

// an event line with a literal author ("You kicked bob")
export function line(name: string, content: string) {
  pushChat({ id: '', name, type: 'event', content, created: new Date() })
}

export function sendChat(text: string) {
  client.send('chat/message', { text })
}

// who toggled a member's chat permission, when it was us (the server does not say)
export const mutedByMe = new Set<string>()

export function mute(id: string, muted: boolean) {
  mutedByMe.add(id)
  return api('POST', `/members/${encodeURIComponent(id)}`, { plugins: { 'chat.can_send': !muted } })
}

export const openInApp = (url: string) => api('POST', '/openinapp/openlink', { text: url })
