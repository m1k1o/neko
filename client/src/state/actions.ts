// what the GUI can do to the room; `actions(neko)` is the set bound to one instance, `useActions()`
// the same in a component
import { isLocked, name, scoped, type LockResource } from './app'
import { api } from './api'
import { toast } from './dialogs'
import { line } from './bus'
import { sendEmote } from './emotes'
import { useNeko } from './provider'
import { t } from '@/i18n'

export const actions = scoped((neko) => {
  const { client, app } = neko
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

    // Kick: DELETE /api/sessions/{id} drops the user's websocket + WebRTC and invalidates the token;
    // they may log straight back in. A softer kick is POST /api/sessions/{id}/disconnect (the same
    // session can reconnect). The v2 IP ban lived in the legacy layer and has no v3 counterpart.
    async kick(id: string) {
      const who = name(client, id)
      if (await api(neko, 'DELETE', `/sessions/${encodeURIComponent(id)}`))
        line(neko, t('you'), t('notifications.kicked', { name: who }))
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
