<template>
  <div>
    <button @click="connect()">Connect</button>
    <button @click="disconnect()">Disonnect</button>

    <template v-if="loaded && neko.connected">
      <button v-if="!is_controlling" @click="neko.control.request()">request control</button>
      <button v-else @click="neko.control.release()">release control</button>

      <button @click="neko.video.pause()">pause stream</button>
      <button @click="neko.video.play()">play stream</button><br />
    </template>

      <table class="states" v-if="loaded">
        <tr><th>is connected</th><td>{{ neko.connected ? 'yes' : 'no' }}</td></tr>
        <tr><th>is contolling</th><td>{{ is_controlling ? 'yes' : 'no' }}</td></tr>
        <tr><th>websocket state</th><td>{{ neko.state.websocket }}</td></tr>
        <tr><th>webrtc state</th><td>{{ neko.state.webrtc }}</td></tr>
      </table>

    <div ref="container" style="width: 1280px; height: 720px; border: 2px solid red">
      <neko-canvas ref="neko" />
    </div>

    <template v-if="loaded">
      <button
        v-for="{ width, height, rate } in available_screen_sizes"
        :key="width + height + rate"
        @click="neko.screen.size(width, height, rate)"
      >
        {{ width }}x{{ height }}@{{ rate }}
      </button>
    </template>
  </div>
</template>

<style lang="scss" scoped>
  .states {
    td, th {
      border: 1px solid black;
      padding: 4px;
    }
    
    th {
      text-align: right;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Ref, Watch } from 'vue-property-decorator'

  import Neko from '~/components/canvas.vue'

  @Component({
    name: 'neko',
    components: {
      'neko-canvas': Neko,
    },
  })
  export default class extends Vue {
    @Ref('container') readonly container!: HTMLElement
    @Ref('neko') readonly neko!: Neko
    loaded: boolean = false

    get is_controlling() {
      return this.neko.state.is_controlling
    }

    get available_screen_sizes() {
      return this.neko.state.available_screen_sizes
    }

    connect() {
      this.neko.connect('ws://192.168.1.20:3000/', 'admin', 'test')
    }

    disconnect() {
      this.neko.disconnect()
    }

    mounted() {
      this.loaded = true

      this.neko.events.on('system.websocket', (status) => {
        console.log('system.websocket', status)
      })
      this.neko.events.on('system.webrtc', (status) => {
        console.log('system.webrtc', status)
      })
      this.neko.events.on('system.connect', () => {
        console.log('system.connect')
      })
      this.neko.events.on('system.disconnect', (message) => {
        console.log('system.disconnect', message)
      })
      this.neko.events.on('member.list', (members) => {
        console.log('member.list', members)
      })
      this.neko.events.on('member.connected', (id) => {
        console.log('member.connected', id)
      })
      this.neko.events.on('member.disconnected', (id) => {
        console.log('member.disconnected', id)
      })
      this.neko.events.on('control.host', (id) => {
        console.log('control.host', id)
      })
      this.neko.events.on('control.request', (id) => {
        console.log('control.request', id)
      })
      this.neko.events.on('control.requesting', (id) => {
        console.log('control.requesting', id)
      })
      this.neko.events.on('clipboard.update', (text) => {
        console.log('clipboard.update', text)
      })
      this.neko.events.on('screen.configuration', (configurations) => {
        console.log('screen.configuration', configurations)
      })
      this.neko.events.on('screen.size', (width, height, rate) => {
        console.log('screen.size', width, height, rate)
      })
      this.neko.events.on('broadcast.status', (payload) => {
        console.log('broadcast.status', payload)
      })
      this.neko.events.on('member.ban', (id, target) => {
        console.log('member.ban', id, target)
      })
      this.neko.events.on('member.kick', (id, target) => {
        console.log('member.kick', id, target)
      })
      this.neko.events.on('member.muted', (id, target) => {
        console.log('member.muted', id, target)
      })
      this.neko.events.on('member.unmuted', (id, target) => {
        console.log('member.unmuted', id, target)
      })
      this.neko.events.on('room.locked', (id) => {
        console.log('room.locked', id)
      })
      this.neko.events.on('room.unlocked', (id) => {
        console.log('room.unlocked', id)
      })
    }
  }
</script>
