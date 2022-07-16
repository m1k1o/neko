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

<script lang="ts">
  import { Vue, Component, Prop } from 'vue-property-decorator'
  import Neko from '~/component/main.vue'

  @Component({
    name: 'neko-controls',
  })
  export default class extends Vue {
    @Prop() readonly neko!: Neko

    username: string = 'admin'
    password: string = 'admin'

    async login() {
      localStorage.setItem('username', this.username)
      localStorage.setItem('password', this.password)

      try {
        await this.neko.login(this.username, this.password)
      } catch (e: any) {
        alert(e.response ? e.response.data.message : e)
      }
    }

    async connect() {
      try {
        await this.neko.connect()
      } catch (e: any) {
        alert(e)
      }
    }

    async logout() {
      try {
        await this.neko.logout()
      } catch (e: any) {
        alert(e.response ? e.response.data.message : e)
      }
    }

    mounted() {
      const username = localStorage.getItem('username')
      if (username) {
        this.username = username
      }

      const password = localStorage.getItem('password')
      if (password) {
        this.password = password
      }
    }
  }
</script>
