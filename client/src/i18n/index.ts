// i18n: i18next over one JSON file per language and namespace, src/locales/<lang>/<ns>.json (README,
// "i18n"). The files are chunks fetched on demand through the loader below (Vite's lazy
// import.meta.glob): a language is downloaded when it becomes the active one, and en, the fallback
// for the keys a language lacks, with it; the other languages never are.
import i18next, { type BackendModule, type ResourceKey } from 'i18next'
import { initReactI18next } from 'react-i18next'
import { get, set } from '@/state/storage'

// the languages, in picker order: a folder of three files under src/locales/ and its code here
export const langs = ['en', 'es', 'sk', 'sv', 'nb', 'fr', 'de', 'ko', 'fi', 'ru', 'cn', 'tw', 'ja', 'id', 'pl'] as const
export type Lang = (typeof langs)[number]
const isLang = (s: string): s is Lang => (langs as readonly string[]).includes(s)

// the browser's language if there is a locale for it, by exact code then by its base (`de-at` -> de)
export function detectLang(): Lang {
  const browser = navigator.language.toLowerCase()
  const base = browser.split('-')[0]
  return langs.find((l) => l === browser) ?? langs.find((l) => l.startsWith(base)) ?? 'en'
}

// the <html lang> tag: the two Chinese codes are the legacy short codes, not BCP 47 tags
const HTML_LANG: Record<string, string> = { cn: 'zh-CN', tw: 'zh-TW' }
export const htmlLang = (lang: string) => HTML_LANG[lang] ?? lang

export type LocaleFiles = Record<string, () => Promise<{ default: ResourceKey }>>

// the loader: i18next asks it for each <lang>/<ns> it needs (init, changeLanguage) and adds what it
// returns as that bundle; a file that does not exist is an error, which i18next takes as "nothing
// for this language", so the key falls back to en
export const loader = (files: LocaleFiles): BackendModule => ({
  type: 'backend',
  init() {},
  read(lang, ns, done) {
    const file = files[`../locales/${lang}/${ns}.json`]
    if (!file) return done(new Error(`no locale ${lang}/${ns}`), false)
    file().then(
      (m) => done(null, m.default),
      (err: Error) => done(err, false),
    )
  },
})

export const i18n = i18next
  .use(initReactI18next)
  .use(loader(import.meta.glob<{ default: ResourceKey }>('../locales/*/*.json')))
// for code outside components (components use react-i18next's useTranslation, which re-renders them
// on a language change)
export const { t } = i18n

// start-up: the language from ?lang= (kept, like a pick), else the stored one, else the browser's,
// with the namespaces of the plugins next to `common`; resolves once they are loaded
export function initI18n(ns: string[]) {
  const url = new URL(location.href).searchParams.get('lang')
  let lng = get<string>('lang', detectLang())
  if (url && isLang(url)) set('lang', (lng = url))
  document.documentElement.lang = htmlLang(lng)
  return i18n.init({
    lng,
    fallbackLng: 'en',
    ns: ['common', ...ns],
    defaultNS: 'common',
    interpolation: { escapeValue: false }, // React escapes
    // a key no language has (en included) is a bug: reported in development
    saveMissing: import.meta.env.DEV,
    missingKeyHandler: (lngs, ns, key) => console.warn(`i18n: missing ${ns}:${key} (${lngs.join(', ')})`),
  })
}

// a pick in the room menu: load the language if it is not yet, persist it, <html lang>
export function setLang(lang: Lang) {
  set('lang', lang)
  document.documentElement.lang = htmlLang(lang)
  return i18n.changeLanguage(lang)
}
