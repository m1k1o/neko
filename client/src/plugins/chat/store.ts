import { createStore } from 'zustand/vanilla'
import { scoped } from '@/state/app'

export interface ChatLine {
  seq: number // unique and increasing: the React key, and what scrolling follows
  id: string
  name: string
  type: 'text' | 'event'
  content: string
  created: Date
}

export const createChatStore = () =>
  createStore(() => ({
    lines: [] as ChatLine[],
    // messages received, for the unread badge
    texts: 0,
    // chat/init: the server's chat plugin is on
    enabled: true,
    emojiReady: false,
    emojiRecent: [] as string[],
  }))
export type ChatStore = ReturnType<typeof createChatStore>
// the store of an instance (`chat(neko)`, `chat(useNeko())`)
export const chat = scoped(createChatStore)

// chat keeps the last CHAT_LIMIT lines in memory; nothing is persisted
const CHAT_LIMIT = 1000
let seq = 0
export function push(store: ChatStore, line: Omit<ChatLine, 'seq'>) {
  store.setState((s) => ({ lines: [...s.lines, { ...line, seq: ++seq }].slice(-CHAT_LIMIT) }))
}
