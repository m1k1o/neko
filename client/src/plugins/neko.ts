import { PluginObject } from 'vue'
import { NekoClient } from '~/client'

declare global {
  const $client: NekoClient

  interface Window {
    $client: NekoClient
  }
}

declare module 'vue/types/vue' {
  interface Vue {
    $client: NekoClient
  }
}

const plugin: PluginObject<undefined> = {
  install(Vue) {
    window.$client = new NekoClient()
      .on('error', error => console.error('[%cNEKO%c] %cERR', 'color: #498ad8;', '', 'color: #d84949;', error))
      .on('warn', (...log) => console.warn('[%cNEKO%c] %cWRN', 'color: #498ad8;', '', 'color: #eae364;', ...log))
      .on('info', (...log) => console.info('[%cNEKO%c] %cINF', 'color: #498ad8;', '', 'color: #4ac94c;', ...log))
      .on('debug', (...log) => console.log('[%cNEKO%c] %cDBG', 'color: #498ad8;', '', 'color: #eae364;', ...log))

    Vue.prototype.$client = window.$client
  },
}

export default plugin
