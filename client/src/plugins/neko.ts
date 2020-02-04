import { PluginObject } from 'vue'
import { NekoClient } from '~/neko'

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
      .on('error', window.$log.error)
      .on('warn', window.$log.warn)
      .on('info', window.$log.info)
      .on('debug', window.$log.debug)

    Vue.prototype.$client = window.$client
  },
}

export default plugin
