import Vue from 'vue'

import { SweetAlertOptions } from 'sweetalert2'
import Swal from 'sweetalert2/dist/sweetalert2.js'

type VueSwalInstance = typeof Swal.fire

declare module 'vue/types/vue' {
  interface Vue {
    $swal: VueSwalInstance
  }

  interface VueConstructor {
    swal: VueSwalInstance
  }
}

interface VueSweetalert2Options extends SweetAlertOptions {
  // includeCss?: boolean;
}

class VueSweetalert2 {
  static install(vue: Vue | any, options?: VueSweetalert2Options): void {
    const swalFunction = (...args: [SweetAlertOptions]) => {
      if (options) {
        const mixed = Swal.mixin(options)

        return mixed.fire(...args)
      }

      return Swal.fire(...args)
    }

    let methodName: string | number | symbol

    for (methodName in Swal) {
      // @ts-ignore
      if (Object.prototype.hasOwnProperty.call(Swal, methodName) && typeof Swal[methodName] === 'function') {
        // @ts-ignore
        swalFunction[methodName] = ((method) => {
          return (...args: any[]) => {
            // @ts-ignore
            return Swal[method](...args)
          }
        })(methodName)
      }
    }

    vue['swal'] = swalFunction

    // add the instance method
    if (!vue.prototype.hasOwnProperty('$swal')) {
      vue.prototype.$swal = swalFunction
    }
  }
}

export default VueSweetalert2
