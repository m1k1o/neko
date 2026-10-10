import { client } from '@/state/client'
import type { Plugin } from './types'
import { chat } from './chat'
import { filetransfer } from './filetransfer'

// the plugins of this build, in side-panel order; adding one is its folder plus a line here
export const plugins: Plugin[] = [chat, filetransfer]

export const tabs = () => plugins.flatMap((p) => (p.tab ? [p.tab] : []))
export const topBar = () => plugins.flatMap((p) => p.topBar ?? [])
export const memberMenu = () => plugins.flatMap((p) => p.memberMenu ?? [])

// once at start-up (app/boot.ts): server event dispatch, then each plugin's init (the strings are
// the plugins' namespaces, loaded with the language by initI18n)
export function initPlugins() {
  client.events.on('message', (event, payload) => {
    plugins.find((p) => event.startsWith(p.id + '/'))?.onEvent?.(event, payload)
  })
  for (const p of plugins) p.init?.()
}
