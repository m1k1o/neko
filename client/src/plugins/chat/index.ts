import { useStore } from 'zustand'
import { selectIsAdmin } from '@m1k1o/neko'
import { MessageSquare } from 'lucide-react'
import { isMuted, name } from '@/state/app'
import { ask } from '@/state/dialogs'
import { useClient, useNeko } from '@/state/provider'
import { t } from '@/i18n'
import type { Plugin } from '@/plugins/types'
import { Chat } from './Chat'
import { chat as store, push } from './store'
import { mute, mutedByMe } from './actions'

export const chat: Plugin = {
  id: 'chat',
  ns: 'chat',
  tab: {
    id: 'chat',
    icon: MessageSquare,
    label: 'chat:tab',
    component: Chat,
    useBadge: () => useStore(store(useNeko()), (s) => s.texts),
  },
  memberMenu: [
    {
      id: 'chat-mute',
      useVisible: () => useStore(useClient().store, selectIsAdmin),
      label: ({ client }, m) => t(isMuted(client, m.id) ? 'chat:unmute' : 'chat:mute'),
      onClick(neko, m) {
        const muted = isMuted(neko.client, m.id)
        const which = muted ? 'unmute' : 'mute'
        ask(
          neko.app,
          t(`chat:confirm.${which}_title`, { name: m.profile.name }),
          t(`chat:confirm.${which}_text`, { name: m.profile.name }),
        ).then((ok) => ok && mute(neko, m.id, !muted))
      },
    },
  ],
  onEvent(neko, ...[event, payload]) {
    const { client, app } = neko
    if (event === 'chat/init') store(neko).setState({ enabled: payload.enabled })
    if (event === 'chat/message') {
      if (app.getState().ignored[payload.id]) return
      push(store(neko), {
        id: payload.id,
        name: name(client, payload.id),
        type: 'text',
        content: payload.content.text,
        created: new Date(payload.created),
      })
      store(neko).setState((s) => ({ texts: s.texts + 1 }))
      if (app.getState().settings.chat_sound && payload.id !== client.state.session_id)
        new Audio('chat.mp3').play().catch(() => {})
    }
  },
  init(neko) {
    const { client, bus } = neko
    const lines = store(neko)
    // the room's event lines ("bob took the controls") are shown in the chat
    bus.on('log', (id, name, content) => push(lines, { id, name, type: 'event', content, created: new Date() }))
    bus.on('logout', () => lines.setState({ lines: [], texts: 0 }))

    // muted / unmuted lines: who may send, per session; a change after the member list is known
    // is someone's mute (ours when mutedByMe says so, the server does not say)
    let initialized = false
    const canSend: Record<string, boolean> = {}
    const can = (id: string) => client.state.sessions[id]?.profile.plugins?.['chat.can_send'] !== false
    client.store.subscribe(
      (s) => s.session_id,
      (id) => {
        initialized = !!id
        if (!id) return
        for (const sid of Object.keys(client.state.sessions)) canSend[sid] = can(sid)
      },
    )
    const onSession = (id: string) => {
      if (!initialized) return
      const now = can(id)
      if (id in canSend && canSend[id] !== now) {
        const by = mutedByMe(neko).delete(id) ? t('you') : t('somebody')
        push(lines, {
          id: '',
          name: by,
          type: 'event',
          content: t(now ? 'chat:unmuted' : 'chat:muted', { name: name(client, id) }),
          created: new Date(),
        })
      }
      canSend[id] = now
    }
    client.events.on('session.created', onSession)
    client.events.on('session.updated', onSession)
    client.events.on('session.deleted', (id) => delete canSend[id])
    client.events.on('connection.closed', () => (initialized = false))
  },
}
