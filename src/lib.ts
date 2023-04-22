import Vue from 'vue'
import Neko from './component/main.vue'

/**
 * Fügt eine "install" function hinzu
 *
 * Weitere Infos:
 *      https://vuejs.org/v2/cookbook/packaging-sfc-for-npm.html#Packaging-Components-for-npm
 */
const NekoElements = {
  install(vue: typeof Vue): void {
    vue.component('Neko', Neko)
  },
}

if (typeof window !== 'undefined' && window.Vue) {
  // @ts-ignore
  window.Vue.use(NekoElements, {})
}

export default Neko
