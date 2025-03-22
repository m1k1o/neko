import Vue from 'vue'
import VueI18n from 'vue-i18n'
import { messages } from '~/locale'

Vue.use(VueI18n)

function detectBrowserLanguage(): string {
  const browserLang = navigator.language.toLowerCase()

  const supportedLangs = Object.keys(messages)
  console.log(supportedLangs)
  if (supportedLangs.includes(browserLang)) {
    return browserLang
  }

  const baseLang = browserLang.split('-')[0]
  const matchingLang = supportedLangs.find((lang) => lang.startsWith(baseLang))
  if (matchingLang) {
    return matchingLang
  }

  return 'en'
}

const storedLang = localStorage.getItem('neko_language')
const defaultLang = storedLang ?? detectBrowserLanguage()

export const i18n = new VueI18n({
  locale: defaultLang,
  fallbackLocale: 'en',
  messages,
})
