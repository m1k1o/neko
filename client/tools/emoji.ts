// Regenerates public/emoji.json: `npm run build:emoji`. The chat renders an emoji as text, the
// Unicode character of its name (the browser's emoji font), so the data is the names, their
// characters (emoji-datasource's `unified` code points), the groups and the keywords (emojilib);
// the custom ones (emoji_custom.ts) are images and have no character.
import * as fs from 'node:fs'
import { createRequire } from 'node:module'
import { custom } from './emoji_custom.ts'

const require = createRequire(import.meta.url)
const datasource = require('emoji-datasource/emoji.json') as EmojiDatasource[]
const emojis = require('emojilib') as Record<string, string[]>

interface EmojiDatasource {
  unified: string
  short_name: string
  short_names: string[]
  category: string
}

const GROUP: Record<string, string> = {
  Symbols: 'symbols',
  Activities: 'activity',
  Flags: 'flags',
  'Travel & Places': 'travel',
  'Food & Drink': 'food',
  'Animals & Nature': 'nature',
  'People & Body': 'people',
  'Smileys & Emotion': 'emotion',
  Objects: 'objects',
}
const GROUP_NAME: Record<string, string> = {
  neko: 'Neko',
  emotion: 'Emotion',
  people: 'People',
  nature: 'Nature',
  food: 'Food',
  activity: 'Activity',
  travel: 'Travel',
  objects: 'Objects',
  symbols: 'Symbols',
  flags: 'Flags',
}

const keywords: Record<string, string[]> = {}
const chars: Record<string, string> = {}
const list: string[] = []
const groups: Record<string, string[]> = Object.fromEntries(Object.keys(GROUP_NAME).map((g) => [g, []]))

for (const emoji of custom) {
  groups.neko.push(emoji.name)
  list.push(emoji.name)
  keywords[emoji.name] = emoji.keywords
}

for (const source of datasource) {
  const codes = source.unified.split('-').map((h) => parseInt(h, 16))
  const group = GROUP[source.category]
  if (!group) {
    // Component: the skin tone and hair modifiers, not emoji of their own
    if (source.category !== 'Component') console.log(`unknown category ${source.category}`)
    continue
  }
  // emojilib keys by the character; match on the first code point
  const hex = codes[0].toString(16)
  let words =
    Object.entries(emojis)
      .find(([ch]) => ch.codePointAt(0)!.toString(16) === hex)
      ?.flat() ?? []
  if (words.length === 0) console.log(source.short_name, 'no keywords')
  for (const name of source.short_names) if (!words.includes(name)) words = [...words, name]
  keywords[source.short_name] = words
  groups[group].push(source.short_name)
  list.push(source.short_name)
  chars[source.short_name] = String.fromCodePoint(...codes)
}

const data = {
  groups: Object.entries(GROUP_NAME).map(([id, name]) => ({ id, name, list: groups[id] })),
  list,
  keywords,
  chars,
}
fs.writeFileSync('public/emoji.json', JSON.stringify(data))
console.log(`emoji.json done: ${list.length} names, ${Object.keys(chars).length} characters`)
