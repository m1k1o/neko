// event wiring: turn v3 events into event lines (shown by the chat) and toasts
import type { Settings } from '@m1k1o/neko'
import { app } from './app'
import { client, name, hostId, isLocked, type LockResource } from './client'
import { actions } from './actions'
import { toast, tell } from './dialogs'
import { bus, event } from './bus'
import { EMOTES, showEmote } from './emotes'
import { t } from '@/i18n'

const ev = client.events
let initialized = false
let lastHost: string | null = null
let lastSettings: Settings | null = null
// last seen state per session; names outlive session/deleted (core drops the session first)
const known: Record<string, { connected: boolean; name: string }> = {}

// session/created events during system/init are the existing member list, not joins
client.store.subscribe(
  (s) => s.session_id,
  (id) => {
    initialized = !!id
    if (!id) return
    lastHost = hostId()
    lastSettings = { ...client.state.settings, plugins: { ...client.state.settings.plugins } }
    for (const [sid, sess] of Object.entries(client.state.sessions))
      known[sid] = { connected: sess.state.is_connected, name: sess.profile.name }
    event(id, t('notifications.connected', { name: '' }))
  },
)

function onSession(id: string) {
  if (!initialized) return
  const sess = client.state.sessions[id]
  const now = !!sess?.state.is_connected
  if (now !== !!known[id]?.connected && id !== client.state.session_id) {
    event(id, t(now ? 'notifications.connected' : 'notifications.disconnected', { name: '' }))
  }
  known[id] = { connected: now, name: sess?.profile.name ?? known[id]?.name ?? '' }
}
ev.on('session.created', onSession)
ev.on('session.updated', onSession)
ev.on('session.deleted', (id) => {
  const k = known[id]
  if (k?.connected) bus.emit('log', id, k.name || t('somebody'), t('notifications.disconnected', { name: '' }))
  delete known[id]
})

ev.on('room.control.host', (hasHost, hostID, by) => {
  const me = client.state.session_id
  if (hasHost && hostID) {
    if (by && by !== hostID) event(by, t('notifications.controls_given', { name: name(hostID) }))
    else if (lastHost && lastHost !== hostID)
      event(hostID, t('notifications.controls_taken_steal', { name: name(lastHost) }))
    else event(hostID, t('notifications.controls_taken', { name: '' }))
    if (hostID === me) toast(t('notifications.controls_taken', { name: t('you') }))
    else if (lastHost === me) toast(t('notifications.controls_released', { name: t('you') }))
  } else if (lastHost) {
    if (by && by !== lastHost) event(by, t('notifications.controls_released_steal', { name: name(lastHost) }))
    else event(lastHost, t('notifications.controls_released', { name: '' }))
    if (lastHost === me) toast(t('notifications.controls_released', { name: t('you') }))
  }
  lastHost = hasHost ? (hostID ?? null) : null
})

// connection toasts, as the legacy client showed them
client.store.subscribe(
  (s) => s.connection.status,
  (status, old) => {
    if (status === 'connecting' && old === 'connected') toast(t('connection.reconnecting'), undefined, 'warning')
    if (status === 'connected') {
      app.setState({ toasts: [] })
      toast(t('connection.connected'), undefined, 'success')
    }
  },
)

ev.on('room.control.request', (id) => {
  if (!app.getState().ignored[id]) toast(t('notifications.controls_requesting', { name: name(id) }))
})

ev.on('room.screen.updated', (width, height, rate, id) => {
  if (id) event(id, t('notifications.resolution', { width, height, rate }))
})

ev.on('room.settings.updated', (next, id) => {
  if (lastSettings && id) {
    for (const r of ['control', 'login'] as LockResource[]) {
      const locked = isLocked(r, next)
      if (locked !== isLocked(r, lastSettings)) event(id, t(`locks.${r}.notif_${locked ? 'locked' : 'unlocked'}`))
    }
  }
  lastSettings = { ...next, plugins: { ...next.plugins } }
})

// files dropped on the video go to the remote desktop through the core (room/upload/drop)
ev.on('upload.drop.finished', (error) => error && toast(error.message, undefined, 'error'))

ev.on('room.broadcast.status', (active, url) => app.setState({ broadcast: { active, url: url || '' } }))

ev.on('receive.broadcast', (sender, subject, body) => {
  // other clients may send anything; only known names become class names on screen
  if (subject === 'emote' && EMOTES.includes(body) && !app.getState().ignored[sender]) showEmote(body)
})

// plugin events (`chat/*`, ...) are dispatched by the plugin registry
ev.on('message', (event, payload) => {
  if (event === 'openinapp/init') app.setState({ openInApp: !!payload.enabled })
})

ev.on('connection.closed', (error) => {
  initialized = false
  if (!error) return
  const reason =
    error.message === 'session deleted'
      ? t('connection.kicked')
      : error.message === 'connection replaced'
        ? t('connection.replaced')
        : error.message
  // session gone server-side (kicked, logged out elsewhere, server restarted): back to the login
  // screen. Any other failure, such as no network, keeps the session for the Connect button.
  client.api
    .req('GET', '/whoami')
    .catch((err) => (err.status === 401 ? actions.logout() : undefined))
    .finally(() => tell(t('connection.disconnected'), reason))
})
