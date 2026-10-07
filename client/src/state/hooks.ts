import { useSyncExternalStore } from 'react'
import { client } from './client'
import { app } from './app'
import { subscribeAll, versionAll } from './stores'

const subscribe = (fn: () => void) => {
  const a = client.store.subscribe(fn)
  const b = subscribeAll(fn)
  return () => {
    a()
    b()
  }
}
const snapshot = () => `${client.store.version}.${versionAll()}`

// every store change re-renders the subscribed tree; fine at this size,
// split per-slice selectors if profiling ever shows render cost
export function useNeko() {
  useSyncExternalStore(subscribe, snapshot)
  return { client, state: client.state, app: app.state }
}
