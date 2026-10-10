// Unit tests of the emote keys: `npm test` (vitest). The browser is faked as far as the app store
// and showEmote need it (location, localStorage, document.visibilityState).
import { expect, test } from 'vitest'
import { fakeBrowser } from '@/test/browser'

fakeBrowser()
const { createAppStore } = await import('./app.ts')
const { showEmote, hideEmote } = await import('./emotes.ts')

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/

test('each shown emote gets a key of its own, a UUID (crypto.randomUUID), and hideEmote drops it', () => {
  const app = createAppStore()
  showEmote(app, 'wave')
  showEmote(app, 'wave')
  const keys = Object.keys(app.getState().emotes)
  expect(keys).toHaveLength(2)
  for (const k of keys) expect(k).toMatch(UUID)
  expect(keys[0]).not.toBe(keys[1])
  hideEmote(app, keys[0])
  expect(Object.keys(app.getState().emotes)).toEqual([keys[1]])
})
