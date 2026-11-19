// The browser, faked just far enough for an instance of the GUI (createNekoApp) to be made under
// vitest: location, localStorage, the elements the input overlay creates, and a stream transport
// that does nothing. Installed before the modules that read them are imported (`fakeBrowser()`,
// then `await import(...)`). The strings are not loaded: tests mock `@/i18n` with `t` as the key.
import type { StreamTransport, TransportEvents } from '@m1k1o/neko'

class FakeElement {
  style: Record<string, string> = {}
  spellcheck = true
  children: unknown[] = []
  onload: unknown = null
  setAttribute() {}
  append(...nodes: unknown[]) {
    this.children.push(...nodes)
  }
  addEventListener() {}
  removeEventListener() {}
  getContext() {
    return {}
  }
}

const fake = (name: string, value: unknown) =>
  Object.defineProperty(globalThis, name, { value, configurable: true, writable: true })

// returns what localStorage holds
export function fakeBrowser(href = 'http://neko.test/') {
  const stored = new Map<string, string>()
  fake('location', { href })
  fake('localStorage', {
    getItem: (k: string) => stored.get(k) ?? null,
    setItem: (k: string, v: string) => stored.set(k, String(v)),
    removeItem: (k: string) => stored.delete(k),
  })
  fake('document', { createElement: () => new FakeElement(), visibilityState: 'visible' })
  fake('Image', FakeElement)
  fake(
    'Audio',
    class {
      play = () => Promise.resolve()
    },
  )
  return stored
}

// a transport that records the volume it is given
export function fakeTransport(): StreamTransport & { volume: number } {
  const handlers = new Map<keyof TransportEvents, Set<(...args: never[]) => void>>()
  return {
    kind: 'webrtc',
    element: null,
    volume: 1,
    connect: async () => {},
    close() {},
    attach: () => () => {},
    setPlaying: async () => {},
    setVolume(v) {
      this.volume = v
    },
    setMuted() {},
    on(event, cb) {
      if (!handlers.has(event)) handlers.set(event, new Set())
      handlers.get(event)!.add(cb)
      return () => void handlers.get(event)!.delete(cb)
    },
  }
}
