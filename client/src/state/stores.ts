import { Store } from '@m1k1o/neko'

// Every GUI store is created through here, so one hook (useNeko) can subscribe to all of them;
// plugins add their own stores the same way without the hook knowing them.
const stores: Store<any>[] = []

export function createStore<T extends object>(initial: T): Store<T> {
  const store = new Store(initial)
  stores.push(store)
  return store
}

export function subscribeAll(fn: () => void) {
  const offs = stores.map((s) => s.subscribe(fn))
  return () => offs.forEach((off) => off())
}

export const versionAll = () => stores.map((s) => s.version).join('.')
