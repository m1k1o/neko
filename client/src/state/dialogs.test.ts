// Unit tests of the dialog promise contract and the toast mapping: `npm test` (vitest). The browser
// is faked as far as the app store needs it (location, localStorage).
import { beforeEach, describe, expect, it, vi } from 'vitest'

const fake = (name: string, value: unknown) =>
  Object.defineProperty(globalThis, name, { value, configurable: true, writable: true })
fake('location', { href: 'http://neko.test/' })
fake('localStorage', { getItem: () => null, setItem: () => {}, removeItem: () => {} })
vi.mock('sonner', () => ({
  toast: { info: vi.fn(), success: vi.fn(), warning: vi.fn(), error: vi.fn(), dismiss: vi.fn() },
}))
const { toast: sonner } = await import('sonner')
const { app } = await import('./app')
const { ask, tell, toast, dismissToasts } = await import('./dialogs')

describe('ask() and tell()', () => {
  beforeEach(() => app.setState({ dialog: null }))

  it('ask() puts a cancellable warning in the store and resolves with the answer', async () => {
    const p = ask('Kick bob?', 'Really?')
    const d = app.getState().dialog!
    expect(d).toMatchObject({ title: 'Kick bob?', text: 'Really?', icon: 'warning', cancel: true })
    d.resolve(true)
    await expect(p).resolves.toBe(true)
  })

  it('tell() is an error without cancel by default, with the icon given', async () => {
    const p = tell('Oops', undefined, 'info')
    const d = app.getState().dialog!
    expect(d).toMatchObject({ title: 'Oops', icon: 'info', cancel: false })
    expect(d.text).toBeUndefined()
    d.resolve(false)
    await expect(p).resolves.toBe(false)
  })

  it('a new dialog resolves the one still open with false', async () => {
    const first = ask('one')
    const second = tell('two')
    await expect(first).resolves.toBe(false)
    expect(app.getState().dialog!.title).toBe('two')
    app.getState().dialog!.resolve(true)
    await expect(second).resolves.toBe(true)
  })
})

describe('toast()', () => {
  it('maps the kind to the sonner call, 5 s, the text as description', () => {
    toast('Connected', undefined, 'success')
    expect(sonner.success).toHaveBeenCalledWith('Connected', { description: undefined, duration: 5000 })
    toast('Reconnecting', 'hold on', 'warning')
    expect(sonner.warning).toHaveBeenCalledWith('Reconnecting', { description: 'hold on', duration: 5000 })
    toast('bob wants control')
    expect(sonner.info).toHaveBeenCalledWith('bob wants control', { description: undefined, duration: 5000 })
    toast('Failed', 'x', 'error')
    expect(sonner.error).toHaveBeenCalledWith('Failed', { description: 'x', duration: 5000 })
  })
  it('dismissToasts() clears them all', () => {
    dismissToasts()
    expect(sonner.dismiss).toHaveBeenCalledWith()
  })
})
