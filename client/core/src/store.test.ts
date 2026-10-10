// Unit tests of the client's store (zustand, immutable updates, selector subscriptions) and the
// emitter: `npm test` (vitest)
import { test } from 'vitest'
import assert from 'node:assert/strict'
import { reset, patch, session } from './test/browser.ts'
import { Emitter } from './emitter.ts'
const { NekoClient, selectControlling, selectSession, selectIsAdmin } = await import('./index.ts')

const fresh = () => {
  reset()
  return new NekoClient({ autologin: false, autoconnect: false })
}

test('a selector subscription fires only when the selected value changes, with the value before', () => {
  const client = fresh()
  const seen: unknown[] = []
  const off = client.store.subscribe(
    (s) => s.video.playing,
    (v, old) => seen.push([v, old]),
  )
  // writes elsewhere in the state: not seen (the cursor never enters the store at all, it is a
  // transport event the overlay draws)
  patch(client, 'control', { clipboard: { text: 'x' } })
  client.store.setState({ mobile_keyboard_open: true })
  patch(client, 'video', { volume: 0.5 }) // the same slice, another field: a new object, the same value
  assert.deepEqual(seen, [])
  patch(client, 'video', { playing: true })
  assert.deepEqual(seen, [[true, false]])
  patch(client, 'video', { playing: true }) // no change
  assert.deepEqual(seen, [[true, false]])
  off()
  patch(client, 'video', { playing: false })
  assert.equal(seen.length, 1, 'unsubscribed')
})

test('updates are immutable: the changed slice is a new object, untouched slices keep their identity', () => {
  const client = fresh()
  const before = client.state
  client.lock()
  const after = client.state
  assert.notEqual(after, before)
  assert.notEqual(after.control, before.control)
  assert.equal(after.control.locked, true)
  assert.equal(before.control.locked, false, 'the state before is not touched')
  for (const k of ['connection', 'video', 'screen', 'sessions', 'settings'] as const)
    assert.equal(after[k], before[k], `${k} keeps its identity`)
  client.setScrollInverse(false)
  assert.equal(
    client.state.control.keyboard,
    after.control.keyboard,
    'the sibling of a nested write keeps its identity',
  )
  assert.notEqual(client.state.control.scroll, after.control.scroll)
})

test('client.state is the store’s current state; a plain listener sees every update', () => {
  const client = fresh()
  assert.equal(client.state, client.store.getState())
  let runs = 0
  const off = client.store.subscribe(() => runs++)
  client.store.setState({ authenticated: true })
  assert.equal(client.state, client.store.getState())
  assert.deepEqual([client.state.authenticated, runs], [true, 1])
  patch(client, 'video', { volume: 0.5 })
  assert.equal(runs, 2)
  off()
  assert.ok(!('watch' in client.store) && !('version' in client.store), 'the proxy store’s API is gone')
})

test('selectors: session, controlling and isAdmin from the state, as the getters report them', () => {
  const client = fresh()
  const s1 = session('s1', { is_admin: true })
  assert.deepEqual(
    [selectSession(client.state), selectControlling(client.state), selectIsAdmin(client.state)],
    [null, false, false],
  )
  client.store.setState({ session_id: 's1', sessions: { s1 } })
  assert.deepEqual([selectSession(client.state), client.session], [s1, s1])
  assert.deepEqual([selectIsAdmin(client.state), client.isAdmin], [true, true])
  patch(client, 'control', { host_id: 's2' })
  assert.deepEqual([selectControlling(client.state), client.controlling], [false, false])
  patch(client, 'control', { host_id: 's1' })
  assert.deepEqual([selectControlling(client.state), client.controlling], [true, true])
})

test('emitter: on returns the unsubscribe, once fires one time', () => {
  const e = new Emitter<{ x: (v: number) => void }>()
  const got: number[] = []
  const off = e.on('x', (v) => got.push(v))
  e.once('x', (v) => got.push(v * 10))
  e.emit('x', 1)
  off()
  e.emit('x', 2)
  assert.deepEqual(got, [1, 10])
})
