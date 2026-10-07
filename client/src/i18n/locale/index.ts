import * as en from './en-us'
import * as es from './es-sp'
import * as sk from './sk-sk'
import * as sv from './sv-se'
import * as nb from './nb-no'
import * as fr from './fr-fr'
import * as de from './de-de'
import * as ko from './ko-kr'
import * as fi from './fi-fi'
import * as ru from './ru-ru'
import * as cn from './zh-cn'
import * as tw from './zh-tw'
import * as ja from './ja-jp'
import * as id from './id-id'
import * as pl from './pl-pl'

// plain copies: plugins add their strings into these at start-up (extendMessages)
export const messages = {
  en: { ...en },
  es: { ...es },
  sk: { ...sk },
  sv: { ...sv },
  nb: { ...nb },
  fr: { ...fr },
  de: { ...de },
  ko: { ...ko },
  fi: { ...fi },
  ru: { ...ru },
  cn: { ...cn },
  tw: { ...tw },
  ja: { ...ja },
  id: { ...id },
  pl: { ...pl },
}

export type Lang = keyof typeof messages
export const langs = Object.keys(messages) as Lang[]

export function detectLang(): Lang {
  const browser = navigator.language.toLowerCase()
  const base = browser.split('-')[0]
  return (langs.find((l) => l === browser) ?? langs.find((l) => l.startsWith(base)) ?? 'en') as Lang
}

// a plugin's strings per language, merged into the table so they can sit inside existing groups
// (`side.files`, `context.mute`, ...); unknown languages are ignored, English is the fallback anyway
export function extendMessages(table: Record<string, object>) {
  for (const [lang, strings] of Object.entries(table)) if (lang in messages) merge((messages as any)[lang], strings)
}
function merge(into: any, from: any) {
  for (const [k, v] of Object.entries(from)) {
    if (v && typeof v === 'object') merge(into[k] ?? (into[k] = {}), v)
    else into[k] = v
  }
}
