import Vue from 'vue'
import VueI18n from 'vue-i18n'
import { messages } from '~/locale'
import { get } from '~/utils/localstorage'

Vue.use(VueI18n)

const fallbackLocale = 'en'

function detectBrowserLanguage(): string {
  const browserLang = navigator.language.toLowerCase()

  const supportedLangs = Object.keys(messages)
  if (supportedLangs.includes(browserLang)) {
    return browserLang
  }

  const baseLang = browserLang.split('-')[0]
  const matchingLang = supportedLangs.find((lang) => lang.startsWith(baseLang))
  if (matchingLang) {
    return matchingLang
  }

  return fallbackLocale
}

export const i18n = new VueI18n({
  locale: get<string>('lang', detectBrowserLanguage()),
  fallbackLocale,
  messages,
})
