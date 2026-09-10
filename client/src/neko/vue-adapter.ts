import Vue from 'vue'

import { NekoClientRuntime } from './runtime'

/** Creates the legacy Vue/Vuex integration without leaking it into the SDK. */
export function createVueNekoRuntime(vue: Vue): NekoClientRuntime {
  return {
    http: vue.$http,
    state: vue.$accessor,
    ui: {
      translate(key, params) {
        return vue.$t(key, params) as string
      },
      notify(options) {
        vue.$notify(options)
      },
      alert(options) {
        vue.$swal(options)
      },
      log: vue.$log,
    },
  }
}
