import type { ComponentType } from 'react'
import type { Session } from '../core/types'

// What a plugin can contribute to the GUI. The registry (index.ts) wires it in: the side panel
// shows its tab, the header its top-bar items, the member menu its items, and server events
// starting with `${id}/` reach onEvent. Adding a plugin is a folder and one line in the registry.
export interface Plugin {
  // also the prefix of its server events (`chat/*`, `filetransfer/*`)
  id: string
  // strings per language (`{ en: { side: { chat: 'Chat' } }, de: ... }`), merged into the i18n table
  locale?: Record<string, object>
  tab?: PluginTab
  topBar?: PluginTopBarItem[]
  memberMenu?: PluginMemberMenuItem[]
  onEvent?: (event: string, payload: any) => void
  // called once at start-up, before the connection: subscribe to client events here
  init?: () => void
}

export interface PluginTab {
  // the stored `tab` value and the `side.<id>` label
  id: string
  // Font Awesome class of the tab icon
  icon: string
  component: ComponentType
  // default: always
  visible?: () => boolean
  // unread count; the side-panel toggle shows a badge when it changes while the panel is closed
  badge?: () => number
}

// renders its own `<li>` in the header menu, or null
export interface PluginTopBarItem {
  id: string
  component: ComponentType
}

// an entry of the context menu of another member
export interface PluginMemberMenuItem {
  id: string
  visible?: (member: Session) => boolean
  label: (member: Session) => string
  onClick: (member: Session) => void
}
