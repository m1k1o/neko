// emoji (legacy emoji.json + twitter sprite sheet), loaded once when chat first needs it
import { set } from '@/state/storage'
import { chat } from './store'

export interface EmojiGroup {
  id: string
  name: string
  list: string[]
}
export const emoji = { groups: [] as EmojiGroup[], keywords: {} as Record<string, string[]>, names: new Set<string>() }
let emojiLoading = false

export function loadEmoji() {
  if (emojiLoading) return
  emojiLoading = true
  try {
    chat.setState({ emojiRecent: JSON.parse(localStorage.getItem('emoji_recent') || '[]') })
  } catch {}
  fetch('emoji.json')
    .then((r) => r.json())
    .then((d: { groups: EmojiGroup[]; keywords: Record<string, string[]>; list: string[] }) => {
      Object.assign(emoji, { groups: d.groups, keywords: d.keywords, names: new Set(d.list) })
      chat.setState({ emojiReady: true })
    })
    .catch(() => (emojiLoading = false))
}

export function pickedEmoji(name: string) {
  const recent = chat.getState().emojiRecent
  if (recent.includes(name)) return
  const next = [...recent.slice(-30), name]
  chat.setState({ emojiRecent: next })
  set('emoji_recent', JSON.stringify(next))
}
