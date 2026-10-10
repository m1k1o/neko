import { client } from '@/state/client'
import type { Plugin, PluginEventArgs, Slots } from './types'
import { chat } from './chat'
import { filetransfer } from './filetransfer'

// the plugins of this build, in side-panel order; adding one is its folder plus a line here
export const plugins: Plugin[] = [chat, filetransfer]

// what the plugins put into each place of the GUI, in registry order; filled once by initPlugins
const slots: { [K in keyof Slots]: Slots[K][] } = { 'side.tab': [], 'header.item': [], 'member.menu': [] }
export function registerSlot<K extends keyof Slots>(name: K, item: Slots[K]) {
  slots[name].push(item)
}
// read at render (Side, Header, App); the slots do not change after start-up, so there is no state
export function useSlot<K extends keyof Slots>(name: K): readonly Slots[K][] {
  return slots[name]
}

// once at start-up (app/boot.ts): the slots, the server event dispatch, then each plugin's init
export function initPlugins() {
  for (const p of plugins) {
    if (p.tab) registerSlot('side.tab', p.tab)
    for (const item of p.topBar ?? []) registerSlot('header.item', item)
    for (const item of p.memberMenu ?? []) registerSlot('member.menu', item)
  }
  // `chat/message` goes to the plugin with id `chat`: one lookup per message, the map built here once
  const byPrefix = new Map(plugins.map((p) => [p.id, p]))
  client.events.on('message', (event, payload) => {
    // the wire is untyped; the names and payloads are the server's (PluginEvents)
    const slash = event.indexOf('/')
    if (slash !== -1) byPrefix.get(event.slice(0, slash))?.onEvent?.(...([event, payload] as PluginEventArgs))
  })
  for (const p of plugins) p.init?.()
}
