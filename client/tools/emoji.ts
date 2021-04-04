import * as fs from 'fs'
import { custom } from './emoji_custom'

const datasource = require('emoji-datasource/emoji.json') as EmojiDatasource[]
const emojis = require('emojilib') 

interface EmojiDatasource {
  name: string
  unified: string
  non_qualified: string | null
  docomo: string | null
  au: string | null
  softbank: string | null
  google: string | null
  image: string
  sheet_x: number
  sheet_y: number
  short_name: string
  short_names: string[]
  text: string | null
  texts: string | null
  category: string
  sort_order: number
  added_in: string
  has_img_apple: boolean
  has_img_google: boolean
  has_img_twitter: boolean
  has_img_facebook: boolean
  skin_variations: {
    [id: string]: {
      unified: string
      image: string
      sheet_x: number
      sheet_y: number
      added_in: string
      has_img_apple: boolean
      has_img_google: boolean
      has_img_twitter: boolean
      has_img_facebook: boolean
    }
  }
  obsoletes: string
  obsoleted_by: string
}

const SHEET_COLUMNS = 58
const MULTIPLY = 100 / (SHEET_COLUMNS - 1)

const css: string[] = []
const keywords: { [name: string]: string[] } = {}
const list: string[] = []
const groups: { [name: string]: string[] } = { neko: [] }

for (const emoji of custom) {
  groups['neko'].push(emoji.name)
  list.push(emoji.name)
  keywords[emoji.name] = emoji.keywords

  // prettier-ignore
  css.push(`&[data-emoji='${emoji.name}'] { background-size: contain; background-image: url('../images/emoji/${emoji.file}'); }`)
}

for (const source of datasource) {
  const unified = source.unified.split('-').map(v => v.toLowerCase())

  if (!source.has_img_twitter) {
    console.log(source.short_name, 'not avalible for set twitter')
    continue
  }

  // keywords
  let words: string[] = []
  for (const id of Object.keys(emojis)) {
    if (unified.includes(id.codePointAt(0)!.toString(16))) {
      words = [id, ...emojis[id]]
      break
    }
  }

  if (words.length == 0) {
    console.log(source.short_name, 'no keywords')
  }

  for (const name of source.short_names) {
    if (!words.includes(name)) {
      words.push(name)
    }
  }

  keywords[source.short_name] = words

  // keywords
  let group = ''
  switch (source.category) {
    case 'Symbols':
      group = 'symbols'
      break
    case 'Activities':
      group = 'activity'
      break
    case 'Flags':
      group = 'flags'
      break
    case 'Travel & Places':
      group = 'travel'
      break
    case 'Food & Drink':
      group = 'food'
      break
    case 'Animals & Nature':
      group = 'nature'
      break
    case 'People & Body':
      group = 'people'
      break
    case 'Smileys & Emotion':
      group = 'emotion'
      break
    case 'Objects':
      group = 'objects'
      break
    case 'Skin Tones':
      continue
    default:
      console.log(`unknown category ${source.category}`)
      continue
  }

  if (!groups[group]) {
    groups[group] = [source.short_name]
  } else {
    groups[group].push(source.short_name)
  }

  // list
  list.push(source.short_name)

  // css
  // prettier-ignore
  css.push(`&[data-emoji='${source.short_name}'] { background-position: ${MULTIPLY * source.sheet_x}% ${MULTIPLY * source.sheet_y}% }`)
}

fs.writeFile(
  'src/assets/styles/vendor/_emoji.scss',
  `
.emoji {
  display: inline-block;
  background-size: ${SHEET_COLUMNS * 100}%;
  background-image: url('~emoji-datasource/img/twitter/sheets/32.png');
  background-repeat: no-repeat;
  vertical-align: bottom;
  height: 22px;
  width: 22px;

${css.map(v => `  ${v}`).join('\n')}
}
`,
  () => {
    console.log('_emoji.scss done')
  },
)

const data = {
  groups: [
    {
      id: 'neko',
      name: 'Neko',
      list: groups['neko'] ? groups['neko'] : [],
    },
    {
      id: 'emotion',
      name: 'Emotion',
      list: groups['emotion'] ? groups['emotion'] : [],
    },
    {
      id: 'people',
      name: 'People',
      list: groups['people'] ? groups['people'] : [],
    },
    {
      id: 'nature',
      name: 'Nature',
      list: groups['nature'] ? groups['nature'] : [],
    },
    {
      id: 'food',
      name: 'Food',
      list: groups['food'] ? groups['food'] : [],
    },
    {
      id: 'activity',
      name: 'Activity',
      list: groups['activity'] ? groups['activity'] : [],
    },
    {
      id: 'travel',
      name: 'Travel',
      list: groups['travel'] ? groups['travel'] : [],
    },
    {
      id: 'objects',
      name: 'Objects',
      list: groups['objects'] ? groups['objects'] : [],
    },
    {
      id: 'symbols',
      name: 'Symbols',
      list: groups['symbols'] ? groups['symbols'] : [],
    },
    {
      id: 'flags',
      name: 'Flags',
      list: groups['flags'] ? groups['flags'] : [],
    },
  ],
  list,
  keywords,
}

fs.writeFile('public/emoji.json', JSON.stringify(data), () => {
  console.log('emoji.json done')
})
