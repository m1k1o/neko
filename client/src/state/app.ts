// The GUI's own state next to the core's `client.state`: viewer settings, the side panel, the
// dialog. A zustand store: components select from it (`useStore(app, (s) => s.side)`), everything
// else reads `app.getState()` and writes with `app.setState()`. Plugins keep their own stores.
// Also the bundle one instance of the GUI runs on (`NekoApp`) and the small readers of the
// client's state that components share.
import { createStore } from 'zustand/vanilla'
import type { NekoClient, Overlay, Settings, State } from '@m1k1o/neko'
import { get } from './storage'
import type { Dialog } from './dialogs'
import type { Bus } from './bus'
import { t } from '@/i18n'

export const defaults = {
  // v3 scroll steps -5..5; new key because the legacy `scroll` (1..100 px clamp) meant something else
  scroll_sensitivity: 0,
  scroll_invert: true,
  autoplay: true,
  ignore_emotes: false,
  chat_sound: true,
  // open chat links on the remote desktop instead of locally (needs the openinapp plugin)
  links_in_app: false,
  keyboard_layout: 'us',
  // how the stream arrives: 'auto' tries WebRTC and falls back to the websocket when it fails,
  // the others force one. Read once when the client is made, so a change applies on reload.
  transport: 'auto' as TransportMode,
}
export type TransportMode = 'auto' | 'webrtc' | 'websocket'
export const TRANSPORT_MODES: TransportMode[] = ['auto', 'webrtc', 'websocket']
export const isTransportMode = (v: unknown): v is TransportMode => TRANSPORT_MODES.includes(v as TransportMode)
export type ViewerSettings = typeof defaults

function load(): ViewerSettings {
  const out = { ...defaults }
  for (const k of Object.keys(defaults) as (keyof ViewerSettings)[]) (out as any)[k] = get(k, defaults[k])
  // a stored value from an older build, or typed by hand, must not leave the stream unreachable
  if (!isTransportMode(out.transport)) out.transport = defaults.transport
  return out
}

// one per instance (createNekoApp); what the URL and the stored settings say at start-up
export function createAppStore() {
  const params = new URL(location.href).searchParams
  return createStore(() => ({
    settings: load(),
    side: params.has('show_side') ? params.get('show_side') === '1' : get('side', false),
    // id of the side panel's tab (a plugin's or 'settings')
    tab: get<string>('tab', 'chat'),
    // openinapp/init: links can be opened on the remote desktop
    openInApp: false,
    emotes: {} as Record<string, string>,
    ignored: {} as Record<string, boolean>,
    broadcast: { active: false, url: '' },
    keyboardLayouts: {} as Record<string, string>,
    menu: null as null | { x: number; y: number; id: string },
    dialog: null as Dialog | null,
    about: false,
  }))
}
export type AppStore = ReturnType<typeof createAppStore>

// What one instance of the GUI runs on: the core client with its keyboard/mouse/touch overlay, the
// store above and the bus. Made by createNekoApp (neko.ts), handed to <NekoProvider>, read by the
// hooks (provider.tsx) in components and passed to everything else (actions, plugins, ...).
export interface NekoApp {
  client: NekoClient
  overlay: Overlay
  app: AppStore
  bus: Bus
}

// one value per instance, made on first use: a plugin's store, the bound actions
export function scoped<T>(create: (neko: NekoApp) => T): (neko: NekoApp) => T {
  const cache = new WeakMap<NekoApp, T>()
  return (neko) => {
    if (!cache.has(neko)) cache.set(neko, create(neko))
    return cache.get(neko)!
  }
}

export const name = (client: NekoClient, id?: string | null) =>
  (id && client.state.sessions[id]?.profile.name) || t('somebody')
// the server's chat plugin took this member's right to send (chat.can_send); emotes follow it too
export const selectMuted = (s: State, id = s.session_id) =>
  !!id && s.sessions[id]?.profile.plugins?.['chat.can_send'] === false
export const isMuted = (client: NekoClient, id?: string | null) => selectMuted(client.state, id)

export type LockResource = 'login' | 'control'
export const isLocked = (r: LockResource, settings: Settings) =>
  r === 'login' ? settings.locked_logins : settings.locked_controls
