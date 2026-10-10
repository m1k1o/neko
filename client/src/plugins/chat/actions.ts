import { scoped, type NekoApp } from '@/state/app'
import { api } from '@/state/api'

export function sendChat({ client }: NekoApp, text: string) {
  client.send('chat/message', { text })
}

// who toggled a member's chat permission, when it was us (the server does not say)
export const mutedByMe = scoped(() => new Set<string>())

export function mute(neko: NekoApp, id: string, muted: boolean) {
  mutedByMe(neko).add(id)
  return api(neko, 'POST', `/members/${encodeURIComponent(id)}`, { plugins: { 'chat.can_send': !muted } })
}

export const openInApp = (neko: NekoApp, url: string) => api(neko, 'POST', '/openinapp/openlink', { text: url })
