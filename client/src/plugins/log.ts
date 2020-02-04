import { PluginObject } from 'vue'

interface Logger {
  error(error: Error): void
  warn(...log: any[]): void
  info(...log: any[]): void
  debug(...log: any[]): void
}

declare global {
  const $log: Logger

  interface Window {
    $log: Logger
  }
}

declare module 'vue/types/vue' {
  interface Vue {
    $log: Logger
  }
}

const plugin: PluginObject<undefined> = {
  install(Vue) {
    window.$log = {
      error: (error: Error) => console.error('[%cNEKO%c] %cERR', 'color: #498ad8;', '', 'color: #d84949;', error),
      warn: (...log: any[]) => console.warn('[%cNEKO%c] %cWRN', 'color: #498ad8;', '', 'color: #eae364;', ...log),
      info: (...log: any[]) => console.info('[%cNEKO%c] %cINF', 'color: #498ad8;', '', 'color: #4ac94c;', ...log),
      debug: (...log: any[]) => console.log('[%cNEKO%c] %cDBG', 'color: #498ad8;', '', 'color: #eae364;', ...log),
    }

    Vue.prototype.$log = window.$log
  },
}

export default plugin
