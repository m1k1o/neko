import { client } from '@/state/client'
import { api } from '@/state/api'

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
