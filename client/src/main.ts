import './assets/styles/main.scss'

import Vue from 'vue'
import Notifications from 'vue-notification'
import App from './App.vue'

Vue.config.productionTip = false
Vue.use(Notifications)

new Vue({
  render: h => h(App),
}).$mount('#neko')
