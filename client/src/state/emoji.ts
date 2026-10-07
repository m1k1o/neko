// emoji (legacy emoji.json + twitter sprite sheet), loaded once when chat first needs it
import { app } from './app'
import { set } from './storage'

export interface EmojiGroup {
  id: string
  name: string
  list: string[]
}
export const emoji = { groups: [] as EmojiGroup[], keywords: {} as Record<string, string[]>, names: new Set<string>() }
let emojiLoading = false

const s = app.state

export function loadEmoji() {
  if (emojiLoading) return
  emojiLoading = true
  try {
    s.emojiRecent = JSON.parse(localStorage.getItem('emoji_recent') || '[]')
  } catch {}
  fetch('emoji.json')
    .then((r) => r.json())
    .then((d: { groups: EmojiGroup[]; keywords: Record<string, string[]>; list: string[] }) => {
      Object.assign(emoji, { groups: d.groups, keywords: d.keywords, names: new Set(d.list) })
      s.emojiReady = true
    })
    .catch(() => (emojiLoading = false))
}

export function pickedEmoji(name: string) {
  if (s.emojiRecent.includes(name)) return
  s.emojiRecent = [...s.emojiRecent.slice(-30), name]
  set('emoji_recent', JSON.stringify(s.emojiRecent))
}
