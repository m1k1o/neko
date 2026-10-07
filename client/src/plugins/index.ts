import { client } from '@/state/client'
import { extendMessages } from '@/i18n'
import type { Plugin } from './types'
import { chat } from './chat'
import { filetransfer } from './filetransfer'

// the plugins of this build, in side-panel order; adding one is its folder plus a line here
export const plugins: Plugin[] = [chat, filetransfer]

export const tabs = () => plugins.flatMap((p) => (p.tab ? [p.tab] : []))
export const topBar = () => plugins.flatMap((p) => p.topBar ?? [])
export const memberMenu = () => plugins.flatMap((p) => p.memberMenu ?? [])

// once at start-up (app/boot.ts): strings, server event dispatch, then each plugin's init
export function initPlugins() {
  for (const p of plugins) if (p.locale) extendMessages(p.locale)
  client.events.on('message', (event, payload) => {
    plugins.find((p) => event.startsWith(p.id + '/'))?.onEvent?.(event, payload)
  })
  for (const p of plugins) p.init?.()
}
