import type { ComponentType } from 'react'
import type { LucideIcon } from 'lucide-react'
import type { Session } from '@m1k1o/neko'
import type { NekoApp } from '@/state/app'

// What a plugin can contribute to the GUI. The registry (index.ts) wires it in: its tab, header
// items and member-menu entries go into the slots the layout renders (`side.tab`, `header.item`,
// `member.menu`), server events starting with `${id}/` reach onEvent, and its strings are the
// namespace `ns`. Adding a plugin is a folder and one line in the registry.
//
// The instance (`NekoApp`: the client, the GUI store, the bus) comes first to every callback;
// hooks and components read it from the context (`useNeko()`, `useClient()`). A plugin keeps its
// own store per instance with `scoped()` (chat/store.ts).
export interface Plugin {
  // also the prefix of its server events (`chat/*`, `filetransfer/*`)
  id: string
  // the namespace of its strings, src/locales/<lang>/<ns>.json, used as `<ns>:key`
  ns: string
  tab?: PluginTab
  topBar?: PluginTopBarItem[]
  memberMenu?: PluginMemberMenuItem[]
  // its server events, typed by PluginEvents; the rest parameter narrows the payload on the event:
  // `onEvent(neko, ...[event, payload]) { if (event === 'chat/init') store.setState({ enabled: payload.enabled }) }`
  onEvent?: (neko: NekoApp, ...args: PluginEventArgs) => void
  // called once per instance at start-up, before the connection: subscribe to client events here
  init?: (neko: NekoApp) => void
}

// the server's plugin events (server/internal/plugins/*/types.go) and their payloads
export interface PluginEvents {
  'chat/init': { enabled: boolean }
  'chat/message': { id: string; created: string; content: { text: string } }
  'filetransfer/update': {
    enabled: boolean
    root_dir: string
    user_download: boolean
    user_upload: boolean
    user_delete: boolean
    files: { name: string; type: 'file' | 'dir'; size?: number }[]
  }
  // no GUI plugin of its own: state/events.ts turns it into app.openInApp
  'openinapp/init': { enabled: boolean }
}
// `[event, payload]` of one PluginEvents entry
export type PluginEventArgs = { [K in keyof PluginEvents]: [event: K, payload: PluginEvents[K]] }[keyof PluginEvents]

// the places of the GUI a plugin contributes to, and what goes there (registry: registerSlot, useSlot)
export interface Slots {
  'side.tab': PluginTab
  'header.item': PluginTopBarItem
  'member.menu': PluginMemberMenuItem
}

// `useVisible` and `useBadge` are hooks (they select from the plugin's store with useStore), so
// what they return re-renders only what reads it; the GUI runs each in a component of its own
export interface PluginTab {
  // the stored `tab` value
  id: string
  // the tab's icon, a lucide-react component
  icon: LucideIcon
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
  label: (neko: NekoApp, member: Session) => string
  onClick: (neko: NekoApp, member: Session) => void
}
