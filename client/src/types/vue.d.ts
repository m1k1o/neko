import { NekoClient } from '~/client'
import { accessor } from '~/store'

declare module 'vue/types/vue' {
  interface Vue {
    $accessor: typeof accessor
    $client: NekoClient
  }
}
