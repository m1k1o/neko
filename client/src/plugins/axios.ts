import { PluginObject } from 'vue'
import axios, { AxiosStatic } from 'axios'

declare module 'vue/types/vue' {
  interface Vue {
    $http: AxiosStatic
  }
}

const plugin: PluginObject<undefined> = {
  install(Vue) {
    Vue.prototype.$http = axios
  },
}

export default plugin
