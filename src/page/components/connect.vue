<template>
  <div>
    <table v-if="!neko.state.authenticated">
      <tr>
        <th style="padding: 5px; text-align: left">Username</th>
        <td><input type="text" placeholder="Username" v-model="username" /></td>
      </tr>
      <tr>
        <th style="padding: 5px; text-align: left">Password</th>
        <td><input type="password" placeholder="Password" v-model="password" /></td>
      </tr>
      <tr>
        <th></th>
        <td><button @click="login()">Login</button></td>
      </tr>
    </table>
    <div v-else style="text-align: center">
      <p style="padding-bottom: 10px">You are not connected to the server.</p>
      <button @click="connect()">Connect</button> or
      <button @click="logout()">Logout</button>
    </div>
  </div>
</template>

<style lang="scss"></style>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import type Neko from '@/component/main.vue'

const props = defineProps<{
  neko: typeof Neko
}>()

const username = ref('admin')
const password = ref('admin')

async function login() {
  localStorage.setItem('username', username.value)
  localStorage.setItem('password', password.value)

  try {
    await props.neko.login(username.value, password.value)
  } catch (e: any) {
    alert(e.response ? e.response.data.message : e)
  }
}

async function connect() {
  try {
    await props.neko.connect()
  } catch (e: any) {
    alert(e)
  }
}

async function logout() {
  try {
    await props.neko.logout()
  } catch (e: any) {
    alert(e.response ? e.response.data.message : e)
  }
}

onMounted(() => {
  const u = localStorage.getItem('username')
  if (u) {
    username.value = u
  }

  const p = localStorage.getItem('password')
  if (p) {
    password.value = p
  }
})
</script>
