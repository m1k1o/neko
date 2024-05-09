import type Vue from 'vue'
import Neko from './component/main.vue'

// TODO
export * as ApiModels from './component/api/models'
export * as StateModels from './component/types/state'
export * as webrtcTypes from './component/types/webrtc'

/**
 * Fügt eine "install" function hinzu
 *
 * Weitere Infos:
 *      https://vuejs.org/v2/cookbook/packaging-sfc-for-npm.html#Packaging-Components-for-npm
 */
const NekoElements = {
  install(vue: typeof Vue): void {
    // TODO
    // @ts-ignore
    vue.component('Neko', Neko)
  },
}

// TODO
// @ts-ignore
if (typeof window !== 'undefined' && window.Vue) {
  // @ts-ignore
  window.Vue.use(NekoElements, {})
}

export default Neko
