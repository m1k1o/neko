// Minimal deep-observable store, replacing Vue reactivity for the core.
// Internals keep mutating state in place (as they did under Vue); every write
// marks the store dirty, and one microtask later watchers run and subscribers
// are notified. Subscribers re-read state; `version` changes on every flush,
// which is what React's useSyncExternalStore compares.

const isPlain = (v: unknown): v is object => {
  if (v === null || typeof v !== 'object') return false
  const proto = Object.getPrototypeOf(v)
  return proto === Object.prototype || proto === Array.prototype || proto === null
}

interface Watcher {
  get: () => unknown
  cb: (value: any, old: any) => void
  last: unknown
}

export class Store<T extends object> {
  readonly state: T
  version = 0

  private proxies = new WeakMap<object, object>()
  private raw = new WeakMap<object, object>()
  private listeners = new Set<() => void>()
  private watchers: Watcher[] = []
  private queued = false

  constructor(initial: T) {
    this.state = this.wrap(initial)
  }

  // subscribe to any change, returns unsubscribe
  subscribe = (fn: () => void) => {
    this.listeners.add(fn)
    return () => {
      this.listeners.delete(fn)
    }
  }

  // run cb when getter's value changes (=== comparison), returns unwatch
  watch<V>(get: () => V, cb: (value: V, old: V) => void): () => void {
    const w: Watcher = { get, cb, last: get() }
    this.watchers.push(w)
    return () => {
      this.watchers = this.watchers.filter((x) => x !== w)
    }
  }

  private wrap<O extends object>(obj: O): O {
    let p = this.proxies.get(obj)
    if (!p) {
      p = new Proxy(obj, {
        get: (t, k, r) => {
          const v = Reflect.get(t, k, r)
          return isPlain(v) ? this.wrap(v) : v
        },
        set: (t, k, v, r) => {
          v = (typeof v === 'object' && v && this.raw.get(v)) || v
          if (Reflect.get(t, k, r) !== v) {
            Reflect.set(t, k, v, r)
            this.changed()
          }
          return true
        },
        deleteProperty: (t, k) => {
          if (k in t) {
            Reflect.deleteProperty(t, k)
            this.changed()
          }
          return true
        },
      })
      this.proxies.set(obj, p)
      this.raw.set(p, obj)
    }
    return p as O
  }

  private changed() {
    if (this.queued) return
    this.queued = true
    queueMicrotask(() => {
      this.queued = false
      this.version++
      for (const w of [...this.watchers]) {
        const v = w.get()
        if (v !== w.last) {
          const old = w.last
          w.last = v
          w.cb(v, old)
        }
      }
      this.listeners.forEach((fn) => fn())
    })
  }
}

// Minimal typed event emitter (on/off/once/emit).
export class Emitter<E extends { [K in keyof E]: (...args: any[]) => void }> {
  private handlers = new Map<keyof E, Set<Function>>()

  on<K extends keyof E>(event: K, fn: E[K]) {
    if (!this.handlers.has(event)) this.handlers.set(event, new Set())
    this.handlers.get(event)!.add(fn)
    return () => this.off(event, fn)
  }

  off<K extends keyof E>(event: K, fn: E[K]) {
    this.handlers.get(event)?.delete(fn)
  }

  once<K extends keyof E>(event: K, fn: E[K]) {
    const wrap = ((...args: any[]) => (this.off(event, wrap), fn(...args))) as E[K]
    return this.on(event, wrap)
  }

  emit<K extends keyof E>(event: K, ...args: Parameters<E[K]>) {
    for (const fn of [...(this.handlers.get(event) ?? [])]) fn(...args)
  }
}
