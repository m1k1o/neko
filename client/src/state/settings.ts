// viewer settings: writing one, pushing them into the core, and the URL parameters that set them
import { app, type ViewerSettings } from './app'
import { client } from './client'
import { get, set } from './storage'
import { langs, setLang, htmlLang, type Lang } from '@/i18n'

const s = app.state
const params = new URL(location.href).searchParams

export function setSetting<K extends keyof ViewerSettings>(key: K, value: ViewerSettings[K]) {
  s.settings[key] = value
  set(key, value)
  applySettings()
}

export function applySettings() {
  client.setScrollSensitivity(s.settings.scroll_sensitivity)
  client.setScrollInverse(s.settings.scroll_invert)
  client.setKeyboard(s.settings.keyboard_layout)
}

// start-up: what the URL and the stored settings say, applied once before the connection
export function initSettings() {
  for (const k of ['displayname', 'password']) localStorage.removeItem(k) // the Vue client's stored login

  if (params.get('mute_chat') !== null) s.settings.chat_sound = params.get('mute_chat') !== '1'

  document.documentElement.lang = htmlLang(s.lang)
  const urlLang = params.get('lang') as Lang | null
  if (urlLang && langs.includes(urlLang)) setLang(urlLang)

  // legacy ?scroll= was a 1..100 px clamp (default 10); map it onto v3 steps around the same default
  if (params.has('scroll')) {
    const px = parseInt(params.get('scroll') || '', 10)
    if (!isNaN(px))
      setSetting('scroll_sensitivity', Math.max(-5, Math.min(5, Math.round(2 * Math.log2(Math.max(1, px) / 10)))))
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
}
