import { createStore } from '@/state/stores'

export interface ChatLine {
  seq: number // unique and increasing: the React key, and what scrolling follows
  id: string
  name: string
  type: 'text' | 'event'
  content: string
  created: Date
}

export const chat = createStore({
  lines: [] as ChatLine[],
  // messages received, for the unread badge
  texts: 0,
  // chat/init: the server's chat plugin is on
  enabled: true,
  emojiReady: false,
  emojiRecent: [] as string[],
})

// chat keeps the last CHAT_LIMIT lines in memory; nothing is persisted. Trimming in place keeps
// the proxies the store handed out (see core store.ts wrap) and the row keys stable.
const CHAT_LIMIT = 1000
let seq = 0
export function push(line: Omit<ChatLine, 'seq'>) {
  const s = chat.state
  s.lines.push({ ...line, seq: ++seq })
  if (s.lines.length > CHAT_LIMIT) s.lines.splice(0, s.lines.length - CHAT_LIMIT)
}
