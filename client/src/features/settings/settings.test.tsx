// Unit tests of the settings tab: `npm test` (vitest, react-dom/server). The Settings component is
// rendered under a <NekoProvider> with the browser faked (src/test/browser.ts), the strings are
// their keys; the room settings are an admin's only and follow the client's state.
import { expect, test, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import type { Session } from '@m1k1o/neko'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: {} }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
vi.mock('react-i18next', () => ({ useTranslation: () => ({ t: (key: string) => key }) }))
fakeBrowser()
const { createNekoApp } = await import('@/state/neko')
const { NekoProvider } = await import('@/state/provider')
const { actions } = await import('@/state/actions')
const { Settings } = await import('./Settings.tsx')

const make = (is_admin: boolean) => {
  const neko = createNekoApp({ transport: fakeTransport() })
  // react-dom/server reads a zustand store's initial state (useSyncExternalStore): make it the current one
  neko.app.getInitialState = neko.app.getState
  neko.client.store.getInitialState = neko.client.store.getState
  const me: Session = {
    id: 'me',
    profile: {
      name: 'Me',
      is_admin,
      can_login: true,
      can_connect: true,
      can_watch: true,
      can_host: true,
      can_share_media: false,
      can_access_clipboard: false,
      sends_inactive_cursor: false,
      can_see_inactive_cursors: false,
    },
    state: { is_connected: true, is_watching: true },
  }
  neko.client.store.setState({ session_id: 'me', sessions: { me } })
  return neko
}
const html = (neko: ReturnType<typeof make>) =>
  renderToStaticMarkup(
    <NekoProvider neko={neko}>
      <Settings />
    </NekoProvider>,
  )
// the switches of the tab: name -> checked
const switches = (out: string) =>
  Object.fromEntries(
    [...out.matchAll(/<input type="checkbox"[^>]*aria-label="setting\.(\w+)"[^>]*\/>/g)].map((m) => [
      m[1],
      m[0].includes('checked=""'),
    ]),
  )

const ROOM = ['private_mode', 'control_protection', 'implicit_hosting', 'merciful_reconnect']

test('a user sees the viewer settings only', () => {
  const out = html(make(false))
  expect(Object.keys(switches(out))).toEqual(['scroll_invert', 'autoplay', 'ignore_emotes', 'chat_sound'])
  expect(out).not.toContain('setting.room_title')
})

test('an admin sees the room settings, in order, as the server last sent them', () => {
  const neko = make(true)
  let out = html(neko)
  expect(out).toContain('setting.room_title')
  const all = switches(out)
  expect(Object.keys(all).slice(4)).toEqual(ROOM) // inactive_cursors: this client draws no other cursors
  expect(ROOM.map((k) => all[k])).toEqual([false, false, false, false])
  for (const k of ROOM) expect(out).toContain(`title="setting.${k}_desc"`)

  const s = neko.client.state.settings
  neko.client.store.setState({ settings: { ...s, private_mode: true, implicit_hosting: true } })
  out = html(neko)
  expect(ROOM.map((k) => switches(out)[k])).toEqual([true, false, true, false])
})

test('setRoomSetting posts the one key it changes', async () => {
  const neko = make(true)
  const req = vi.spyOn(neko.client.api, 'req').mockResolvedValue(undefined)
  await expect(actions(neko).setRoomSetting('control_protection', true)).resolves.toBe(true)
  await actions(neko).setRoomSetting('merciful_reconnect', false)
  expect(req.mock.calls).toEqual([
    ['POST', '/room/settings', { control_protection: true }],
    ['POST', '/room/settings', { merciful_reconnect: false }],
  ])
})
