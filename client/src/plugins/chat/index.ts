import { app } from '@/state/app'
import { client, name, isMuted } from '@/state/client'
import { ask } from '@/state/dialogs'
import { bus } from '@/state/bus'
import { t } from '@/i18n'
import type { Plugin } from '@/plugins/types'
import { Chat } from './Chat'
import { chat as store, push } from './store'
import { mute, mutedByMe } from './actions'
import { locale } from './locale'

export const chat: Plugin = {
  id: 'chat',
  locale,
  tab: { id: 'chat', icon: 'fa-comment-alt', component: Chat, badge: () => store.state.texts },
  memberMenu: [
    {
      id: 'chat-mute',
      visible: () => client.isAdmin,
      label: (m) => t(isMuted(m.id) ? 'context.unmute' : 'context.mute'),
      onClick(m) {
        const muted = isMuted(m.id)
        const which = muted ? 'unmute' : 'mute'
        ask(
          t(`context.confirm.${which}_title`, { name: m.profile.name }),
          t(`context.confirm.${which}_text`, { name: m.profile.name }),
        ).then((ok) => ok && mute(m.id, !muted))
      },
    },
  ],
  onEvent(event, payload) {
    if (event === 'chat/init') store.state.enabled = payload.enabled
    if (event === 'chat/message') {
      if (app.state.ignored[payload.id]) return
      push({
        id: payload.id,
        name: name(payload.id),
        type: 'text',
        content: payload.content.text,
        created: new Date(payload.created),
      })
      store.state.texts++
      if (app.state.settings.chat_sound && payload.id !== client.state.session_id)
        new Audio('chat.mp3').play().catch(() => {})
    }
  },
  init() {
    // the room's event lines ("bob took the controls") are shown in the chat
    bus.on('log', (id, name, content) => push({ id, name, type: 'event', content, created: new Date() }))
    bus.on('logout', () => Object.assign(store.state, { lines: [], texts: 0 }))

    // muted / unmuted lines: who may send, per session; a change after the member list is known
    // is someone's mute (ours when mutedByMe says so, the server does not say)
    let initialized = false
    const canSend: Record<string, boolean> = {}
    const can = (id: string) => client.state.sessions[id]?.profile.plugins?.['chat.can_send'] !== false
    client.store.watch(
      () => client.state.session_id,
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
        const by = mutedByMe.delete(id) ? t('you') : t('somebody')
        push({
          id: '',
          name: by,
          type: 'event',
          content: t(now ? 'notifications.unmuted' : 'notifications.muted', { name: name(id) }),
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
