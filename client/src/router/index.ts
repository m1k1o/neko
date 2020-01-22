import Vue from 'vue'
import VueRouter from 'vue-router'
import chat from '~/pages/chat.vue'
import about from '~/pages/about.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'chat',
    component: chat,
  },
  {
    path: '/about',
    name: 'about',
    component: about,
  },
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes,
})

export default router
