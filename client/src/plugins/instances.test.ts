// Unit tests of the plugins on two instances: `npm test` (vitest). The chat and file-transfer
// stores, the event dispatch and each plugin's wiring are per instance; the slots are the build's.
import { test, vi } from 'vitest'
import assert from 'node:assert/strict'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({
  toast: { info: vi.fn(), success: vi.fn(), warning: vi.fn(), error: vi.fn(), dismiss: vi.fn() },
}))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
fakeBrowser()
const { createNekoApp } = await import('@/state/neko')
const { initPlugins, useSlot } = await import('./index.ts')
const { chat } = await import('./chat/store.ts')
const { store: files } = await import('./filetransfer/store.ts')
const { mutedByMe } = await import('./chat/actions.ts')

const a = createNekoApp({ transport: fakeTransport() })
const b = createNekoApp({ transport: fakeTransport() })
initPlugins(a)
initPlugins(b)
const message = (id: string, text: string) => ({ id, created: '2025-01-01T00:00:00Z', content: { text } })
const update = { enabled: true, root_dir: '/', user_download: true, user_upload: true, user_delete: true, files: [] }

test('the slots are filled once, whatever the number of instances', () => {
  assert.deepEqual(
    useSlot('side.tab').map((x) => x.id),
    ['chat', 'files'],
  )
  assert.deepEqual(
    useSlot('header.item').map((x) => x.id),
    ['filetransfer-lock'],
  )
  assert.deepEqual(
    useSlot('member.menu').map((x) => x.id),
    ['chat-mute'],
  )
})

test('a plugin store is per instance: a server event of one fills its store only', () => {
  assert.notEqual(chat(a), chat(b))
  assert.equal(chat(a), chat(a))
  assert.notEqual(files(a), files(b))
  assert.notEqual(mutedByMe(a), mutedByMe(b))

  a.client.events.emit('message', 'chat/message', message('s1', 'hi'))
  assert.deepEqual(
    chat(a)
      .getState()
      .lines.map((l) => [l.type, l.content]),
    [['text', 'hi']],
  )
  assert.deepEqual([chat(a).getState().texts, chat(b).getState().texts], [1, 0])
  assert.deepEqual(chat(b).getState().lines, [])

  b.client.events.emit('message', 'filetransfer/update', update)
  assert.deepEqual([files(a).getState().files, files(b).getState().files], [null, update])
})

test("a plugin's wiring is per instance: the bus lines and the logout of one", () => {
  a.bus.emit('log', 'bob', 'Bob', 'took the controls')
  assert.deepEqual(
    chat(a)
      .getState()
      .lines.map((l) => [l.type, l.name]),
    [
      ['text', 'somebody'],
      ['event', 'Bob'],
    ],
  )
  assert.deepEqual(chat(b).getState().lines, [])

  files(a).setState({ uploads: [{ id: 1, name: 'x', size: 1, progress: 1, status: 'completed' }] })
  files(b).setState({ uploads: [{ id: 2, name: 'y', size: 1, progress: 1, status: 'completed' }] })
  a.bus.emit('logout')
  assert.deepEqual([chat(a).getState().lines, files(a).getState().uploads], [[], []])
  assert.equal(files(b).getState().uploads.length, 1)
})
