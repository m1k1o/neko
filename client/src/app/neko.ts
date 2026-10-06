// App state + actions on top of the framework-free core. One module on purpose:
// the GUI is small, and every component reads the same two stores.
import { useSyncExternalStore } from 'react'
import { NekoClient } from '../core/client'
import { Store } from '../core/store'
import type { Settings, MemberData } from '../core/types'
import { messages } from './locale'

/////////////////////////////
// i18n: legacy locale files, missing keys fall back to English
/////////////////////////////

export const langs = Object.keys(messages) as Lang[]
type Lang = keyof typeof messages

function detectLang(): Lang {
  const browser = navigator.language.toLowerCase()
  const base = browser.split('-')[0]
  return (langs.find((l) => l === browser) ?? langs.find((l) => l.startsWith(base)) ?? 'en') as Lang
}

const lookup = (m: unknown, key: string) => key.split('.').reduce<any>((o, k) => o?.[k], m)

export function t(key: string, vars: Record<string, string | number> = {}): string {
  let msg = lookup(messages[s.lang], key)
  if (typeof msg !== 'string') msg = lookup(messages.en, key)
  if (typeof msg !== 'string') return key
  return msg.replace(/\{(\w+)\}/g, (_: string, k: string) => String(vars[k] ?? ''))
}

/////////////////////////////
// per-viewer settings, stored like the legacy client did (same keys, '1'/'0' booleans)
/////////////////////////////

function get<T extends string | number | boolean>(key: string, def: T): T {
  try {
    const v = localStorage.getItem(key)
    if (!v) return def
    if (typeof def === 'boolean') return (v === '1') as T
    if (typeof def === 'number') return (isNaN(parseInt(v)) ? def : parseInt(v)) as T
    return v as T
  } catch {
    return def
  }
}

export const remember = (key: string, val: string | number | boolean) => set(key, val)
function set(key: string, val: string | number | boolean) {
  try {
    localStorage.setItem(key, typeof val === 'boolean' ? (val ? '1' : '0') : String(val))
  } catch {}
}

const defaults = {
  // v3 scroll steps -5..5; new key because the legacy `scroll` (1..100 px clamp) meant something else
  scroll_sensitivity: 0,
  scroll_invert: true,
  autoplay: true,
  ignore_emotes: false,
  chat_sound: true,
  // open chat links on the remote desktop instead of locally (needs the openinapp plugin)
  links_in_app: false,
  keyboard_layout: 'us',
}
type ViewerSettings = typeof defaults

function load(): ViewerSettings {
  const out = { ...defaults }
  for (const k of Object.keys(defaults) as (keyof ViewerSettings)[]) (out as any)[k] = get(k, defaults[k])
  return out
}

/////////////////////////////
// stores
/////////////////////////////

export interface ChatLine {
  seq: number // unique and increasing: the React key, and what scrolling follows
  id: string
  name: string
  type: 'text' | 'event'
  content: string
  created: Date
}

export interface FileItem {
  name: string
  type: 'file' | 'dir'
  size?: number
}

export interface FileTransfer {
  enabled: boolean
  root_dir: string
  user_download: boolean
  user_upload: boolean
  user_delete: boolean
  files: FileItem[]
}

export interface Upload {
  id: number
  name: string
  size: number
  progress: number
  status: 'inprogress' | 'completed' | 'failed'
  error?: string
}

export interface Dialog {
  title: string
  text?: string
  icon: 'warning' | 'error' | 'info'
  cancel: boolean
  resolve: (ok: boolean) => void
}

export interface Toast {
  id: number
  kind: 'info' | 'success' | 'warning' | 'error'
  title: string
  text?: string
}

const params = new URL(location.href).searchParams
const settings = load()
for (const k of ['displayname', 'password']) localStorage.removeItem(k) // the Vue client's stored login

export const client = new NekoClient({
  autologin: true,
  autoconnect: true,
  // read when a track arrives, so the setting applies without a reload
  get autoplay() {
    return s.settings.autoplay
  },
})

export const app = new Store({
  lang: get<string>('lang', detectLang()) as Lang,
  settings,
  side: params.has('show_side') ? params.get('show_side') === '1' : get('side', false),
  tab: get<'chat' | 'files' | 'settings'>('tab', 'chat'),
  chat: [] as ChatLine[],
  texts: 0,
  chatEnabled: true,
  openInApp: false,
  emojiReady: false,
  emojiRecent: [] as string[],
  emotes: {} as Record<string, string>,
  files: null as FileTransfer | null,
  uploads: [] as Upload[],
  toasts: [] as Toast[],
  ignored: {} as Record<string, boolean>,
  broadcast: { active: false, url: '' },
  keyboardLayouts: {} as Record<string, string>,
  menu: null as null | { x: number; y: number; id: string },
  dialog: null as Dialog | null,
  about: false,
  bans: 0, // bumped after a ban or unban, so lists that show them reload
})
const s = app.state

if (params.get('mute_chat') !== null) s.settings.chat_sound = params.get('mute_chat') !== '1'

export function setSetting<K extends keyof ViewerSettings>(key: K, value: ViewerSettings[K]) {
  s.settings[key] = value
  set(key, value)
  applySettings()
}

const htmlLang = (lang: Lang) => ({ cn: 'zh-CN', tw: 'zh-TW' })[lang as string] ?? lang
export function setLang(lang: Lang) {
  s.lang = lang
  set('lang', lang)
  document.documentElement.lang = htmlLang(lang)
}
document.documentElement.lang = htmlLang(s.lang)
const urlLang = params.get('lang') as Lang | null
if (urlLang && langs.includes(urlLang)) setLang(urlLang)

// legacy ?scroll= was a 1..100 px clamp (default 10); map it onto v3 steps around the same default
if (params.has('scroll')) {
  const px = parseInt(params.get('scroll') || '', 10)
  if (!isNaN(px))
    setSetting('scroll_sensitivity', Math.max(-5, Math.min(5, Math.round(2 * Math.log2(Math.max(1, px) / 10)))))
}

function applySettings() {
  client.setScrollSensitivity(s.settings.scroll_sensitivity)
  client.setScrollInverse(s.settings.scroll_invert)
  client.setKeyboard(s.settings.keyboard_layout)
}
applySettings()

// volume survives a reload, as in the Vue client (same key and 0..100 scale); a ?volume= url
// parameter overrides it for this visit
const urlVolume = params.has('volume') ? parseFloat(params.get('volume') || '1') : NaN
const startVolume = isNaN(urlVolume) ? get('volume', 100) / 100 : Math.max(0, Math.min(urlVolume, 1))
const unwatchVolume = client.store.watch(
  () => client.state.video.playable,
  (playable) => {
    if (!playable) return
    client.setVolume(startVolume)
    unwatchVolume()
    client.store.watch(
      () => client.state.video.volume,
      (v) => set('volume', Math.round(v * 100)),
    )
  },
)

fetch('keyboard_layouts.json')
  .then((r) => r.json())
  .then((l) => (s.keyboardLayouts = l))
  .catch(() => {})

/////////////////////////////
// accessibility: icon controls are <i>/<span>; this makes them focusable, labelled buttons.
// Enter/Space activation is one delegated listener (installed below) instead of per element.
/////////////////////////////

// menus and pickers: a click anywhere or Escape closes them; returns the cleanup for an effect
export function closeOn(close: () => void) {
  const key = (e: KeyboardEvent) => e.key === 'Escape' && close()
  // after the click that opened it has finished propagating, or that click would close it
  const timer = setTimeout(() => {
    document.addEventListener('click', close)
    document.addEventListener('keydown', key)
  })
  return () => {
    clearTimeout(timer)
    document.removeEventListener('click', close)
    document.removeEventListener('keydown', key)
  }
}

export const a11y = (label: string, role: 'button' | 'menuitem' | 'tab' = 'button') => ({
  role,
  tabIndex: 0,
  'aria-label': label,
  title: label,
})

document.addEventListener('keydown', (e) => {
  const el = e.target as HTMLElement
  if ((e.key === 'Enter' || e.key === ' ') && /^(button|menuitem|tab)$/.test(el.getAttribute?.('role') ?? '')) {
    e.preventDefault()
    el.click()
  }
})

/////////////////////////////
// React binding
/////////////////////////////

const subscribe = (fn: () => void) => {
  const a = client.store.subscribe(fn)
  const b = app.subscribe(fn)
  return () => {
    a()
    b()
  }
}
const snapshot = () => `${client.store.version}.${app.version}`

// every store change re-renders the subscribed tree; fine at this size,
// split per-slice selectors if profiling ever shows render cost
export function useNeko() {
  useSyncExternalStore(subscribe, snapshot)
  return { client, state: client.state, app: s }
}

/////////////////////////////
// derived
/////////////////////////////

export const name = (id?: string | null) => (id && client.state.sessions[id]?.profile.name) || t('somebody')
export const isMuted = (id = client.state.session_id) =>
  !!id && client.state.sessions[id]?.profile.plugins?.['chat.can_send'] === false
export const hostId = () => client.state.control.host_id

export type LockResource = 'login' | 'control' | 'file_transfer'
export function isLocked(r: LockResource, settings: Settings = client.state.settings) {
  if (r === 'login') return settings.locked_logins
  if (r === 'control') return settings.locked_controls
  return settings.plugins?.['filetransfer.enabled'] === false
}

/////////////////////////////
// actions
/////////////////////////////

// modal dialogs (legacy SweetAlert look): ask() resolves true on confirm, tell() when dismissed
function dialog(d: Omit<Dialog, 'resolve'>) {
  s.dialog?.resolve(false)
  return new Promise<boolean>((resolve) => (s.dialog = { ...d, resolve }))
}
export const ask = (title: string, text?: string) => dialog({ title, text, icon: 'warning', cancel: true })
export const tell = (title: string, text?: string, icon: Dialog['icon'] = 'error') =>
  dialog({ title, text, icon, cancel: false })

let toastId = 0
export function toast(title: string, text?: string, kind: Toast['kind'] = 'info') {
  const id = ++toastId
  s.toasts.push({ id, kind, title, text })
  setTimeout(() => (s.toasts = s.toasts.filter((x) => x.id !== id)), 5000)
}

// chat keeps the last CHAT_LIMIT lines in memory; nothing is persisted. Trimming in place keeps
// the proxies the store handed out (see store.ts wrap) and the row keys stable.
const CHAT_LIMIT = 1000
let chatSeq = 0
function pushChat(line: Omit<ChatLine, 'seq'>) {
  s.chat.push({ ...line, seq: ++chatSeq })
  if (s.chat.length > CHAT_LIMIT) s.chat.splice(0, s.chat.length - CHAT_LIMIT)
}

function event(id: string, content: string) {
  pushChat({
    id,
    name: id === client.state.session_id ? t('you') : name(id),
    type: 'event',
    content,
    created: new Date(),
  })
}

// REST call for user actions: failures become a toast, the result says whether it worked.
// Session ids are built from the login name on some providers, so they go through encodeURIComponent
// wherever they are part of a path: otherwise a name like `x/../../logout?` redirects the request.
const api = (method: string, path: string, body?: unknown) =>
  client.api.req(method, path, body).then(
    () => true,
    (err) => (toast(err.message), false),
  )

function line(name: string, content: string) {
  pushChat({ id: '', name, type: 'event', content, created: new Date() })
}

// who toggled a member's chat permission, when it was us (the server does not say)
const mutedByMe = new Set<string>()

export const actions = {
  login: async (username: string, password: string) => {
    await client.login(username, password)
    client.connect()
  },
  logout: () => {
    Object.assign(s, { chat: [], texts: 0, uploads: [], ignored: {}, broadcast: { active: false, url: '' } })
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

  sendChat(text: string) {
    client.send('chat/message', { text })
  },
  sendEmote(emote: string) {
    if (isMuted()) return
    client.sendBroadcast('emote', emote)
    showEmote(emote) // server does not echo broadcasts to the sender
  },

  toggleLock(r: LockResource) {
    const locked = !isLocked(r)
    if (r === 'login') return api('POST', '/room/settings', { locked_logins: locked })
    if (r === 'control') return api('POST', '/room/settings', { locked_controls: locked })
    return api('POST', '/room/settings', { plugins: { 'filetransfer.enabled': !locked } })
  },

  mute(id: string, muted: boolean) {
    mutedByMe.add(id)
    return api('POST', `/members/${encodeURIComponent(id)}`, { plugins: { 'chat.can_send': !muted } })
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

  openInApp: (url: string) => api('POST', '/openinapp/openlink', { text: url }),
  filesRefresh: () => client.send('filetransfer/update'),
  // plain link so the browser streams the download; the token only goes in the URL when the
  // server runs without cookies (otherwise the session cookie authenticates it)
  fileUrl: (name: string) =>
    `${client.api.url}/api/filetransfer?filename=${encodeURIComponent(name)}` +
    (client.api.token ? `&token=${encodeURIComponent(client.api.token)}` : ''),
  fileDelete: (name: string) => api('DELETE', `/filetransfer?filename=${encodeURIComponent(name)}`),
  upload(files: FileList | File[]) {
    for (const file of files) {
      const u: Upload = { id: ++toastId, name: file.name, size: file.size, progress: 0, status: 'inprogress' }
      s.uploads.push(u)
      const live = () => s.uploads.find((x) => x.id === u.id)!
      const form = new FormData()
      form.append('files', file)
      client.api
        .upload('/filetransfer', form, (p) => (live().progress = p.loaded))
        .then(() => Object.assign(live(), { progress: file.size, status: 'completed' }))
        .catch((err) => Object.assign(live(), { status: 'failed', error: err.message }))
    }
  },
}

/////////////////////////////
// emoji (legacy emoji.json + twitter sprite sheet), loaded once when chat first needs it
/////////////////////////////

export interface EmojiGroup {
  id: string
  name: string
  list: string[]
}
export const emoji = { groups: [] as EmojiGroup[], keywords: {} as Record<string, string[]>, names: new Set<string>() }
let emojiLoading = false

export function loadEmoji() {
  if (emojiLoading) return
  emojiLoading = true
  try {
    s.emojiRecent = JSON.parse(localStorage.getItem('emoji_recent') || '[]')
  } catch {}
  fetch('emoji.json')
    .then((r) => r.json())
    .then((d: { groups: EmojiGroup[]; keywords: Record<string, string[]>; list: string[] }) => {
      Object.assign(emoji, { groups: d.groups, keywords: d.keywords, names: new Set(d.list) })
      s.emojiReady = true
    })
    .catch(() => (emojiLoading = false))
}

export function pickedEmoji(name: string) {
  if (s.emojiRecent.includes(name)) return
  s.emojiRecent = [...s.emojiRecent.slice(-30), name]
  set('emoji_recent', JSON.stringify(s.emojiRecent))
}

export const EMOTES = [
  'anger',
  'bomb',
  'sleep',
  'explode',
  'sweat',
  'poo',
  'hundred',
  'alert',
  'punch',
  'wave',
  'okay',
  'thumbs-up',
  'clap',
  'prey',
  'celebrate',
  'flame',
  'goof',
  'love',
  'cool',
  'smerk',
  'worry',
  'ouch',
  'cry',
  'surprised',
  'quiet',
  'rage',
  'annoy',
  'steamed',
  'scared',
  'terrified',
  'sleepy',
  'dead',
  'happy',
  'roll-eyes',
  'thinking',
  'clown',
  'sick',
  'rofl',
  'drule',
  'sniff',
  'sus',
  'party',
  'odd',
  'hot',
  'cold',
  'blush',
  'sad',
]

export function showEmote(emote: string) {
  if (s.settings.ignore_emotes || document.visibilityState === 'hidden') return
  s.emotes[Math.random().toString(36).slice(2)] = emote
}

/////////////////////////////
// event wiring: turn v3 events into the legacy chat log + toasts
/////////////////////////////

const ev = client.events
let initialized = false
let lastHost: string | null = null
let lastSettings: Settings | null = null
// last seen state per session; names outlive session/deleted (core drops the session first)
const known: Record<string, { connected: boolean; canSend: boolean; name: string }> = {}

// session/created events during system/init are the existing member list, not joins
client.store.watch(
  () => client.state.session_id,
  (id) => {
    initialized = !!id
    if (!id) return
    lastHost = hostId()
    lastSettings = { ...client.state.settings, plugins: { ...client.state.settings.plugins } }
    for (const [sid, sess] of Object.entries(client.state.sessions))
      known[sid] = {
        connected: sess.state.is_connected,
        canSend: sess.profile.plugins?.['chat.can_send'] !== false,
        name: sess.profile.name,
      }
    event(id, t('notifications.connected', { name: '' }))
    actions.filesRefresh()
  },
)

function onSession(id: string) {
  if (!initialized) return
  const sess = client.state.sessions[id]
  const now = !!sess?.state.is_connected
  if (now !== !!known[id]?.connected && id !== client.state.session_id) {
    event(id, t(now ? 'notifications.connected' : 'notifications.disconnected', { name: '' }))
  }
  const canSend = sess?.profile.plugins?.['chat.can_send'] !== false
  if (known[id] && known[id].canSend !== canSend) {
    const by = mutedByMe.delete(id) ? t('you') : t('somebody')
    line(by, t(canSend ? 'notifications.unmuted' : 'notifications.muted', { name: name(id) }))
  }
  known[id] = { connected: now, canSend, name: sess?.profile.name ?? known[id]?.name ?? '' }
}
ev.on('session.created', onSession)
ev.on('session.updated', onSession)
ev.on('session.deleted', (id) => {
  const k = known[id]
  if (k?.connected)
    pushChat({
      id,
      name: k.name || t('somebody'),
      type: 'event',
      content: t('notifications.disconnected', { name: '' }),
      created: new Date(),
    })
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
client.store.watch(
  () => client.state.connection.status,
  (status, old) => {
    if (status === 'connecting' && old === 'connected') toast(t('connection.reconnecting'), undefined, 'warning')
    if (status === 'connected') {
      s.toasts = []
      toast(t('connection.connected'), undefined, 'success')
    }
  },
)

ev.on('room.control.request', (id) => {
  if (!s.ignored[id]) toast(t('notifications.controls_requesting', { name: name(id) }))
})

ev.on('room.screen.updated', (width, height, rate, id) => {
  if (id) event(id, t('notifications.resolution', { width, height, rate }))
})

ev.on('room.settings.updated', (next, id) => {
  if (lastSettings && id) {
    for (const r of ['control', 'login', 'file_transfer'] as LockResource[]) {
      const locked = isLocked(r, next)
      if (locked !== isLocked(r, lastSettings)) event(id, t(`locks.${r}.notif_${locked ? 'locked' : 'unlocked'}`))
    }
  }
  lastSettings = { ...next, plugins: { ...next.plugins } }
})

ev.on('upload.drop.finished', (error) => error && toast(error.message, undefined, 'error'))

ev.on('room.broadcast.status', (active, url) => (s.broadcast = { active, url: url || '' }))

ev.on('receive.broadcast', (sender, subject, body) => {
  // other clients may send anything; only known names become class names on screen
  if (subject === 'emote' && EMOTES.includes(body) && !s.ignored[sender]) showEmote(body)
})

ev.on('message', (event, payload) => {
  switch (event) {
    case 'chat/init':
      s.chatEnabled = payload.enabled
      break
    case 'chat/message':
      if (s.ignored[payload.id]) return
      pushChat({
        id: payload.id,
        name: name(payload.id),
        type: 'text',
        content: payload.content.text,
        created: new Date(payload.created),
      })
      s.texts++
      if (s.settings.chat_sound && payload.id !== client.state.session_id) new Audio('chat.mp3').play().catch(() => {})
      break
    case 'filetransfer/update':
      s.files = payload
      break
    case 'openinapp/init':
      s.openInApp = !!payload.enabled
      break
  }
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

// start: resume a saved session (autologin) and connect
client.setUrl(location.href)
