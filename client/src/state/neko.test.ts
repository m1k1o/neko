// Unit tests of an instance of the GUI (createNekoApp): `npm test` (vitest). Two instances in one
// process share nothing: stores, bus, event wiring, settings, actions. The browser is faked
// (src/test/browser.ts), the strings are their keys, the toasts are sonner's mocked.
import { test, vi } from 'vitest'
import assert from 'node:assert/strict'
import type { Session } from '@m1k1o/neko'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({
  toast: { info: vi.fn(), success: vi.fn(), warning: vi.fn(), error: vi.fn(), dismiss: vi.fn() },
}))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
let stored = fakeBrowser()
const { toast: sonner } = await import('sonner')
const { createNekoApp } = await import('./neko.ts')
const { actions } = await import('./actions.ts')
const { setSetting } = await import('./settings.ts')
const { event } = await import('./bus.ts')

const make = () => createNekoApp({ transport: fakeTransport() })
const session = (id: string, name: string): Session => ({
  id,
  profile: {
    name,
    is_admin: false,
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
})
// what reaches an instance's bus
const logs = (neko: ReturnType<typeof make>) => {
  const out: string[][] = []
  neko.bus.on('log', (id, name, content) => out.push([id, name, content]))
  return out
}
// system/init as the core reports it: the session and the member list
const introduce = (neko: ReturnType<typeof make>, me: string, others: string[] = []) =>
  neko.client.store.setState({
    session_id: me,
    sessions: Object.fromEntries([me, ...others].map((id) => [id, session(id, id.toUpperCase())])),
  })

test('two instances: the client, the overlay, the store and the bus are each their own', () => {
  const a = make()
  const b = make()
  assert.notEqual(a.client, b.client)
  assert.notEqual(a.client.store, b.client.store)
  assert.notEqual(a.overlay, b.overlay)
  assert.notEqual(a.app, b.app)
  assert.notEqual(a.bus, b.bus)
  a.app.setState({ about: true })
  assert.deepEqual([a.app.getState().about, b.app.getState().about], [true, false])
  a.client.store.setState({ session_id: 'x' })
  assert.deepEqual([a.client.state.session_id, b.client.state.session_id], ['x', null])
})

test('the event wiring is per instance and once: a server event of one becomes one event line on its bus only', () => {
  const a = make()
  const b = make()
  const la = logs(a)
  const lb = logs(b)
  introduce(a, 'me', ['bob'])
  assert.deepEqual(la, [['me', 'you', 'notifications.connected']])
  assert.deepEqual(lb, [])
  introduce(b, 'other')
  assert.deepEqual(lb, [['other', 'you', 'notifications.connected']])
  assert.equal(la.length, 1)

  // a's bob takes the controls: one line on a (not two: wired once), none on b
  a.client.events.emit('room.control.host', true, 'bob', 'bob')
  assert.deepEqual(la.at(-1), ['bob', 'BOB', 'notifications.controls_taken'])
  assert.deepEqual([la.length, lb.length], [2, 1])
  // b's own controls: a toast for b's viewer, nothing on a
  b.client.events.emit('room.control.host', true, 'other', 'other')
  assert.deepEqual(lb.at(-1), ['other', 'you', 'notifications.controls_taken'])
  assert.equal(sonner.info.mock.calls.at(-1)?.[0], 'notifications.controls_taken')
  assert.deepEqual([la.length, lb.length], [2, 2])
  // the member list is per instance too: a's session events do not reach b's wiring
  a.client.store.setState({ sessions: { ...a.client.state.sessions, carl: session('carl', 'Carl') } })
  a.client.events.emit('session.created', 'carl')
  b.client.events.emit('session.created', 'carl') // unknown to b: nothing
  assert.deepEqual(la.at(-1), ['carl', 'Carl', 'notifications.connected'])
  assert.deepEqual([la.length, lb.length], [3, 2])
})

test('an event line names the session of its instance', () => {
  const a = make()
  const b = make()
  const la = logs(a)
  const lb = logs(b)
  introduce(a, 'me', ['bob'])
  event(a, 'bob', 'x')
  event(b, 'bob', 'y') // b does not know bob
  assert.deepEqual(la.at(-1), ['bob', 'BOB', 'x'])
  assert.deepEqual(lb.at(-1), ['bob', 'somebody', 'y'])
})

test('a setting written on one instance reaches its client and localStorage, not the other instance', () => {
  const a = make()
  const b = make()
  setSetting(a, 'scroll_invert', false)
  assert.deepEqual([a.app.getState().settings.scroll_invert, b.app.getState().settings.scroll_invert], [false, true])
  assert.deepEqual([a.client.state.control.scroll.inverse, b.client.state.control.scroll.inverse], [false, true])
  assert.equal(stored.get('scroll_invert'), '0')
})

test('the actions are bound to their instance, one set per instance', () => {
  const a = make()
  const b = make()
  assert.equal(actions(a), actions(a))
  assert.notEqual(actions(a), actions(b))
  const click = { preventDefault() {}, stopPropagation() {}, clientX: 1, clientY: 2 } as React.MouseEvent
  actions(a).openMenu(click, 'bob')
  assert.deepEqual(a.app.getState().menu, { x: 1, y: 2, id: 'bob' })
  assert.equal(b.app.getState().menu, null)
})

test('start-up: the URL parameters and the stored settings of the page, read when the instance is made', () => {
  stored = fakeBrowser('http://neko.test/?show_side=1&mute_chat=1&scroll=20&volume=0.5')
  stored.set('tab', 'files')
  const transport = fakeTransport()
  const neko = createNekoApp({ transport })
  const s = neko.app.getState()
  assert.deepEqual([s.side, s.tab, s.settings.chat_sound, s.settings.scroll_sensitivity], [true, 'files', false, 2])
  assert.equal(neko.client.state.control.scroll.sensitivity, 2)
  assert.equal(stored.get('scroll_sensitivity'), '2')
  // the volume is applied once the stream is playable, then kept
  assert.equal(transport.volume, 1)
  neko.client.store.setState((st) => ({ video: { ...st.video, playable: true } }))
  assert.equal(transport.volume, 0.5)
  neko.client.store.setState((st) => ({ video: { ...st.video, volume: 0.3 } }))
  assert.equal(stored.get('volume'), '30')

  stored = fakeBrowser('http://neko.test/?cast=1')
  const cast = createNekoApp({ transport: fakeTransport() })
  assert.equal(cast.app.getState().settings.chat_sound, false)
  assert.equal(stored.get('chat_sound'), '0')
  stored = fakeBrowser()
})
