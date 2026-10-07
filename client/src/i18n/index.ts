// i18n: legacy locale files, missing keys fall back to English
import { app } from '@/state/app'
import { set } from '@/state/storage'
import { messages, langs, type Lang } from './locale'

export { langs, type Lang }

const lookup = (m: unknown, key: string) => key.split('.').reduce<any>((o, k) => o?.[k], m)

export function t(key: string, vars: Record<string, string | number> = {}): string {
  let msg = lookup(messages[app.state.lang], key)
  if (typeof msg !== 'string') msg = lookup(messages.en, key)
  if (typeof msg !== 'string') return key
  return msg.replace(/\{(\w+)\}/g, (_: string, k: string) => String(vars[k] ?? ''))
}

export const htmlLang = (lang: Lang) => ({ cn: 'zh-CN', tw: 'zh-TW' })[lang as string] ?? lang

export function setLang(lang: Lang) {
  app.state.lang = lang
  set('lang', lang)
  document.documentElement.lang = htmlLang(lang)
}
