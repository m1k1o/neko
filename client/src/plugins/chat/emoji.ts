// emoji (public/emoji.json, see tools/emoji.ts), loaded once when chat first needs it: the names,
// their Unicode characters, the groups and keywords of the picker; the custom ones are images
import { set } from '@/state/storage'
import neko from '@/assets/images/emoji/neko.png'
import { chat } from './store'

export interface EmojiGroup {
  id: string
  name: string
  list: string[]
}
export const emoji = {
  groups: [] as EmojiGroup[],
  keywords: {} as Record<string, string[]>,
  names: new Set<string>(),
  chars: {} as Record<string, string>,
}
// the emoji that are images, by name (tools/emoji_custom.ts lists them for the generator)
export const custom: Record<string, string> = { neko }
let emojiLoading = false

export function loadEmoji() {
  if (emojiLoading) return
  emojiLoading = true
  try {
    chat.setState({ emojiRecent: JSON.parse(localStorage.getItem('emoji_recent') || '[]') })
  } catch {}
  fetch('emoji.json')
    .then((r) => r.json())
    .then(
      (d: {
        groups: EmojiGroup[]
        keywords: Record<string, string[]>
        list: string[]
        chars: Record<string, string>
      }) => {
        Object.assign(emoji, { groups: d.groups, keywords: d.keywords, names: new Set(d.list), chars: d.chars })
        chat.setState({ emojiReady: true })
      },
    )
    .catch(() => (emojiLoading = false))
}

export function pickedEmoji(name: string) {
  const recent = chat.getState().emojiRecent
  if (recent.includes(name)) return
  const next = [...recent.slice(-30), name]
  chat.setState({ emojiRecent: next })
  set('emoji_recent', JSON.stringify(next))
}
