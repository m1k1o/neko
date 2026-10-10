// Unit tests of the plugin registry: `npm test` (vitest). Slots and server event dispatch follow
// the plugins array; the registry needs only `client.events.on('message')` and the manifests.
import { test, vi } from 'vitest'
import assert from 'node:assert/strict'
import type { Plugin, PluginEventArgs } from './types.ts'

type Listener = (event: string, payload: unknown) => void
const listeners: Listener[] = []
vi.mock('@/state/client', () => ({ client: { events: { on: (_: string, fn: Listener) => listeners.push(fn) } } }))

const received: Record<string, PluginEventArgs[]> = {}
const inits: string[] = []
const Nothing = () => null
const manifest = (id: string, ns: string): Plugin => ({
  id,
  ns,
  tab: { id, icon: 'fa-x', label: `${ns}:tab`, component: Nothing },
  topBar: [{ id: `${id}-top`, component: Nothing }],
  memberMenu: [{ id: `${id}-menu`, label: () => id, onClick() {} }],
  onEvent: (...args) => (received[id] ??= []).push(args),
  init: () => inits.push(id),
})
vi.mock('./chat', () => ({ chat: manifest('chat', 'chat') }))
vi.mock('./filetransfer', () => ({ filetransfer: manifest('filetransfer', 'files') }))

// a registry as at start-up
const fresh = async () => {
  vi.resetModules()
  listeners.length = inits.length = 0
  for (const k of Object.keys(received)) delete received[k]
  return import('./index.ts')
}
const emit = (event: string, payload: unknown) => listeners.forEach((fn) => fn(event, payload))
const message = { id: 's1', created: '2025-01-01T00:00:00Z', content: { text: 'hi' } }
const update = { enabled: true, root_dir: '/', user_download: true, user_upload: true, user_delete: true, files: [] }

test('initPlugins fills the slots in registry order and runs each init', async () => {
  const { plugins, initPlugins, useSlot } = await fresh()
  assert.deepEqual(
    plugins.map((p) => [p.id, p.ns]),
    [
      ['chat', 'chat'],
      ['filetransfer', 'files'],
    ],
  )
  assert.deepEqual(useSlot('side.tab'), [])
  initPlugins()
  assert.deepEqual(
    useSlot('side.tab').map((x) => x.label),
    ['chat:tab', 'files:tab'],
  )
  assert.deepEqual(
    useSlot('header.item').map((x) => x.id),
    ['chat-top', 'filetransfer-top'],
  )
  assert.deepEqual(
    useSlot('member.menu').map((x) => x.id),
    ['chat-menu', 'filetransfer-menu'],
  )
  assert.deepEqual(inits, ['chat', 'filetransfer'])
})

test('a server event reaches the plugin of its prefix only; an unknown prefix reaches nobody', async () => {
  const { initPlugins } = await fresh()
  initPlugins()
  assert.equal(listeners.length, 1)
  emit('chat/message', message)
  emit('chat/init', { enabled: false })
  emit('filetransfer/update', update)
  emit('openinapp/init', { enabled: true })
  emit('chat', {}) // no slash: not a plugin event
  assert.deepEqual(received, {
    chat: [
      ['chat/message', message],
      ['chat/init', { enabled: false }],
    ],
    filetransfer: [['filetransfer/update', update]],
  })
})

test('the dispatch map is built once, at initPlugins: a plugin added afterwards gets nothing', async () => {
  const { plugins, initPlugins } = await fresh()
  initPlugins()
  plugins.push(manifest('late', 'late'))
  emit('late/x', {})
  assert.deepEqual(received, {})
})

test('removing a plugin from the array removes its slot entries and its event handling', async () => {
  const { plugins, initPlugins, useSlot } = await fresh()
  plugins.splice(
    plugins.findIndex((p) => p.id === 'chat'),
    1,
  )
  initPlugins()
  assert.deepEqual(
    useSlot('side.tab').map((x) => x.label),
    ['files:tab'],
  )
  assert.deepEqual(
    useSlot('header.item').map((x) => x.id),
    ['filetransfer-top'],
  )
  assert.deepEqual(
    useSlot('member.menu').map((x) => x.id),
    ['filetransfer-menu'],
  )
  assert.deepEqual(inits, ['filetransfer'])
  emit('chat/message', message)
  emit('filetransfer/update', update)
  assert.deepEqual(received, { filetransfer: [['filetransfer/update', update]] })
})

test('registerSlot: a contribution registered by hand goes after the manifests', async () => {
  const { initPlugins, registerSlot, useSlot } = await fresh()
  initPlugins()
  registerSlot('header.item', { id: 'extra', component: Nothing })
  assert.deepEqual(
    useSlot('header.item').map((x) => x.id),
    ['chat-top', 'filetransfer-top', 'extra'],
  )
})
