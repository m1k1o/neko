// viewer settings: writing one, pushing them into the core, and the URL parameters that set them
import type { NekoApp, ViewerSettings } from './app'
import { get, set } from './storage'

export function setSetting<K extends keyof ViewerSettings>(neko: NekoApp, key: K, value: ViewerSettings[K]) {
  neko.app.setState((s) => ({ settings: { ...s.settings, [key]: value } }))
  set(key, value)
  applySettings(neko)
}

export function applySettings({ client, app }: NekoApp) {
  const { settings } = app.getState()
  client.setScrollSensitivity(settings.scroll_sensitivity)
  client.setScrollInverse(settings.scroll_invert)
  client.setKeyboard(settings.keyboard_layout)
}

// start-up (createNekoApp): what the URL and the stored settings say, applied once before the connection
export function initSettings(neko: NekoApp) {
  const { client, app } = neko
  const params = new URL(location.href).searchParams
  for (const k of ['displayname', 'password']) localStorage.removeItem(k) // the Vue client's stored login

  if (params.get('mute_chat') !== null)
    app.setState((s) => ({ settings: { ...s.settings, chat_sound: params.get('mute_chat') !== '1' } }))

  // legacy ?scroll= was a 1..100 px clamp (default 10); map it onto v3 steps around the same default
  if (params.has('scroll')) {
    const px = parseInt(params.get('scroll') || '', 10)
    if (!isNaN(px))
      setSetting(neko, 'scroll_sensitivity', Math.max(-5, Math.min(5, Math.round(2 * Math.log2(Math.max(1, px) / 10)))))
  }

  applySettings(neko)

  // volume survives a reload, as in the Vue client (same key and 0..100 scale); a ?volume= url
  // parameter overrides it for this visit
  const urlVolume = params.has('volume') ? parseFloat(params.get('volume') || '1') : NaN
  const startVolume = isNaN(urlVolume) ? get('volume', 100) / 100 : Math.max(0, Math.min(urlVolume, 1))
  const unwatchVolume = client.store.subscribe(
    (s) => s.video.playable,
    (playable) => {
      if (!playable) return
      client.setVolume(startVolume)
      unwatchVolume()
      client.store.subscribe(
        (s) => s.video.volume,
        (v) => set('volume', Math.round(v * 100)),
      )
    },
  )

  fetch('keyboard_layouts.json')
    .then((r) => r.json())
    .then((l) => app.setState({ keyboardLayouts: l }))
    .catch(() => {})

  // ?cast= (video only, for a stream) is silent
  if (params.get('cast')) setSetting(neko, 'chat_sound', false)
}
