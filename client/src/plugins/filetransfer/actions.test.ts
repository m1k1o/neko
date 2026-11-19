// Unit tests of the upload list: `npm test` (vitest). The browser is faked as far as an instance
// needs it; the core's upload call is stubbed to hang, so the list stays in progress.
import { expect, test, vi } from 'vitest'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: {} }))
vi.mock('@/i18n', () => ({ t: (key: string) => key }))
fakeBrowser()
const { createNekoApp } = await import('@/state/neko')
const { upload } = await import('./actions.ts')
const { store } = await import('./store.ts')

const UID = /^[0-9a-f]{16}$/

test('each upload gets an id of its own, a 64-bit hex key (crypto.getRandomValues), and its progress is tracked by it', () => {
  const neko = createNekoApp({ transport: fakeTransport() })
  const progress: ((p: { loaded: number; total: number }) => void)[] = []
  vi.spyOn(neko.client.api, 'upload').mockImplementation((_path, _form, onProgress) => {
    progress.push(onProgress!)
    return new Promise(() => {})
  })
  upload(neko, [new File(['a'], 'a.txt'), new File(['bb'], 'b.txt')])
  const { uploads } = store(neko).getState()
  expect(uploads.map((u) => [u.name, u.size, u.status])).toEqual([
    ['a.txt', 1, 'inprogress'],
    ['b.txt', 2, 'inprogress'],
  ])
  for (const u of uploads) expect(u.id).toMatch(UID)
  expect(uploads[0].id).not.toBe(uploads[1].id)
  progress[1]({ loaded: 1, total: 2 })
  expect(
    store(neko)
      .getState()
      .uploads.map((u) => u.progress),
  ).toEqual([0, 1])
})
