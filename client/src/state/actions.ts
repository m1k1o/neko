// what the GUI can do to the room
import type { MemberData } from '../core/types'
import { app } from './app'
import { client, name, hostId, isLocked, type LockResource } from './client'
import { api } from './api'
import { toast } from './dialogs'
import { bus, line } from './bus'
import { sendEmote } from './emotes'
import { t } from '@/i18n'

const s = app.state

// the member context menu, opened from the member list and from chat authors
export const openMenu = (e: React.MouseEvent, id: string) => {
  e.preventDefault()
  e.stopPropagation()
  if (id !== client.state.session_id) s.menu = { x: e.clientX, y: e.clientY, id }
}

export const actions = {
  login: async (username: string, password: string) => {
    await client.login(username, password)
    client.connect()
  },
  logout: () => {
    bus.emit('logout')
    Object.assign(s, { ignored: {}, broadcast: { active: false, url: '' } })
    return client.logout().catch(() => {})
  },

  toggleControl() {
    if (client.controlling) return client.release()
    const host = hostId()
    client.request()
    // server only notifies the host; tell the requester what happened
    if (host && !client.state.settings.implicit_hosting)
      toast(t('notifications.controls_has', { name: name(host) }), t('notifications.controls_has_alt'))
  },

  sendEmote,

  toggleLock(r: LockResource) {
    const locked = !isLocked(r)
    if (r === 'login') return api('POST', '/room/settings', { locked_logins: locked })
    return api('POST', '/room/settings', { locked_controls: locked })
  },

  // Kick & ban: assumed intent. The legacy client banned IP addresses inside the v2
  // compatibility layer; v3 has no IP bans, so these follow what chat/meeting apps do.
  //
  //   kick = remove the user now; they may log straight back in (Discord, Jitsi, Mumble).
  //          DELETE /api/sessions/{id} drops their websocket + WebRTC and invalidates the token.
  //   ban  = keep the user out until an admin unbans them (Settings > Banned).
  //          POST /api/members/{id} {can_login:false}, which persists on providers that store
  //          accounts (file, object), then the kick above.
  //
  // Ban is only offered for members listed by GET /api/members (see canBan): with shared-password
  // (multiuser), noauth or OAuth logins there is no stored account, every login is a new member,
  // so a ban could not stick.
  //
  // To change: edit kick/ban here. A softer kick is POST /api/sessions/{id}/disconnect (same session
  // can reconnect). IP bans would need a server-side ban list checked on /api/login and /api/ws.
  async kick(id: string) {
    const who = name(id)
    if (await api('DELETE', `/sessions/${encodeURIComponent(id)}`))
      line(t('you'), t('notifications.kicked', { name: who }))
  },
  async ban(id: string) {
    const who = name(id)
    if (
      (await api('POST', `/members/${encodeURIComponent(id)}`, { can_login: false })) &&
      (await api('DELETE', `/sessions/${encodeURIComponent(id)}`))
    )
      (s.bans++, line(t('you'), t('notifications.banned', { name: who })))
  },
  unban: (id: string) =>
    api('POST', `/members/${encodeURIComponent(id)}`, { can_login: true }).then((ok) => (ok && s.bans++, ok)),
  // accounts stored by the auth provider (empty for multiuser / noauth / OAuth)
  members: () => client.api.req<MemberData[]>('GET', '/members').catch(() => [] as MemberData[]),
  async canBan(id: string) {
    return (await actions.members()).some((m) => m.id === id)
  },
  give: (id: string) => api('POST', `/room/control/give/${encodeURIComponent(id)}`),
  take: () => api('POST', '/room/control/take'),
  reset: () => api('POST', '/room/control/reset'),

  broadcastStart: (url: string) => api('POST', '/room/broadcast/start', { url }),
  broadcastStop: () => api('POST', '/room/broadcast/stop'),
}
