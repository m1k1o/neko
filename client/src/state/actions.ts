// what the GUI can do to the room; `actions(neko)` is the set bound to one instance, `useActions()`
// the same in a component
import type { MemberData } from '@m1k1o/neko'
import { isLocked, name, scoped, type LockResource } from './app'
import { api } from './api'
import { toast } from './dialogs'
import { line } from './bus'
import { sendEmote } from './emotes'
import { useNeko } from './provider'
import { t } from '@/i18n'

export const actions = scoped((neko) => {
  const { client, app } = neko
  const banned = () => app.setState((s) => ({ bans: s.bans + 1 }))
  // accounts stored by the auth provider (empty for multiuser / noauth / OAuth)
  const members = () => client.api.req<MemberData[]>('GET', '/members').catch(() => [] as MemberData[])
  return {
    login: async (username: string, password: string) => {
      await client.login(username, password)
      client.connect()
    },
    logout: () => {
      neko.bus.emit('logout')
      app.setState({ ignored: {}, broadcast: { active: false, url: '' } })
      return client.logout().catch(() => {})
    },

    toggleControl() {
      if (client.controlling) return client.release()
      const host = client.state.control.host_id
      client.request()
      // server only notifies the host; tell the requester what happened
      if (host && !client.state.settings.implicit_hosting)
        toast(t('notifications.controls_has', { name: name(client, host) }), t('notifications.controls_has_alt'))
    },

    sendEmote: (emote: string) => sendEmote(neko, emote),

    toggleLock(r: LockResource) {
      const locked = !isLocked(r, client.state.settings)
      if (r === 'login') return api(neko, 'POST', '/room/settings', { locked_logins: locked })
      return api(neko, 'POST', '/room/settings', { locked_controls: locked })
    },

    // the member context menu, opened from the member list and from chat authors
    openMenu(e: React.MouseEvent, id: string) {
      e.preventDefault()
      e.stopPropagation()
      if (id !== client.state.session_id) app.setState({ menu: { x: e.clientX, y: e.clientY, id } })
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
      const who = name(client, id)
      if (await api(neko, 'DELETE', `/sessions/${encodeURIComponent(id)}`))
        line(neko, t('you'), t('notifications.kicked', { name: who }))
    },
    async ban(id: string) {
      const who = name(client, id)
      if (
        (await api(neko, 'POST', `/members/${encodeURIComponent(id)}`, { can_login: false })) &&
        (await api(neko, 'DELETE', `/sessions/${encodeURIComponent(id)}`))
      )
        (banned(), line(neko, t('you'), t('notifications.banned', { name: who })))
    },
    unban: (id: string) =>
      api(neko, 'POST', `/members/${encodeURIComponent(id)}`, { can_login: true }).then((ok) => (ok && banned(), ok)),
    members,
    async canBan(id: string) {
      return (await members()).some((m) => m.id === id)
    },
    give: (id: string) => api(neko, 'POST', `/room/control/give/${encodeURIComponent(id)}`),
    take: () => api(neko, 'POST', '/room/control/take'),
    reset: () => api(neko, 'POST', '/room/control/reset'),

    broadcastStart: (url: string) => api(neko, 'POST', '/room/broadcast/start', { url }),
    broadcastStop: () => api(neko, 'POST', '/room/broadcast/stop'),
  }
})
export type Actions = ReturnType<typeof actions>

export const useActions = () => actions(useNeko())
