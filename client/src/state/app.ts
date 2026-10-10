// The GUI's own state next to the core's `client.state`: viewer settings, the side panel, the
// dialog. A zustand store: components select from it (`useStore(app, (s) => s.side)`), everything
// else reads `app.getState()` and writes with `app.setState()`. Plugins keep their own stores.
import { createStore } from 'zustand/vanilla'
import { get } from './storage'
import type { Dialog } from './dialogs'

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
}
export type ViewerSettings = typeof defaults

function load(): ViewerSettings {
  const out = { ...defaults }
  for (const k of Object.keys(defaults) as (keyof ViewerSettings)[]) (out as any)[k] = get(k, defaults[k])
  return out
}

const params = new URL(location.href).searchParams

export const app = createStore(() => ({
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
  bans: 0, // bumped after a ban or unban, so lists that show them reload
}))
