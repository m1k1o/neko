import type { ComponentType } from 'react'
import type { Session } from '@m1k1o/neko'

// What a plugin can contribute to the GUI. The registry (index.ts) wires it in: the side panel
// shows its tab, the header its top-bar items, the member menu its items, server events starting
// with `${id}/` reach onEvent, and its strings are the namespace `ns`. Adding a plugin is a folder
// and one line in the registry.
export interface Plugin {
  // also the prefix of its server events (`chat/*`, `filetransfer/*`)
  id: string
  // the namespace of its strings, src/locales/<lang>/<ns>.json, used as `<ns>:key`
  ns: string
  tab?: PluginTab
  topBar?: PluginTopBarItem[]
  memberMenu?: PluginMemberMenuItem[]
  onEvent?: (event: string, payload: any) => void
  // called once at start-up, before the connection: subscribe to client events here
  init?: () => void
}

// `useVisible` and `useBadge` are hooks (they select from the plugin's store with useStore), so
// what they return re-renders only what reads it; the GUI runs each in a component of its own
export interface PluginTab {
  // the stored `tab` value
  id: string
  // Font Awesome class of the tab icon
  icon: string
  // the i18n key of its label (`chat:tab`)
  label: string
  component: ComponentType
  // default: always
  useVisible?: () => boolean
  // unread count; the side-panel toggle shows a badge when it changes while the panel is closed
  useBadge?: () => number
}

// renders its own `<li>` in the header menu, or null
export interface PluginTopBarItem {
  id: string
  component: ComponentType
}

// an entry of the context menu of another member
export interface PluginMemberMenuItem {
  id: string
  // a hook; default: always
  useVisible?: (member: Session) => boolean
  label: (member: Session) => string
  onClick: (member: Session) => void
}
