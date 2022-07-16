import Vue from 'vue'
import app from './page/main.vue'

Vue.config.productionTip = false

new Vue({
  render: (h) => h(app),
}).$mount('#app')
