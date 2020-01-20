import './assets/styles/main.scss'

import { EVENT } from '~/client/events'

import Vue from 'vue'
import Notifications from 'vue-notification'
import Client from './plugins/neko'
import App from './App.vue'
import store from './store'

Vue.config.productionTip = false

Vue.use(Notifications)
Vue.use(Client)

new Vue({
  store,
  render: h => h(App),
  created() {
    this.$client.init(this)
  },
}).$mount('#neko')
