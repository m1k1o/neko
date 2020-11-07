<template>
  <div>
    <button @click="connect()">Connect</button>
    <button @click="disconnect()">Disonnect</button>
    <button @click="neko.control.request()">request control</button>
    <button @click="neko.control.release()">release control</button>
    <button @click="neko.video.pause()">stop</button><br />
    W: <input type="text" v-model="width" /><br />
    H: <input type="text" v-model="height" /><br />
    <button @click="resize()">Resize</button>
    <br />
    <!--websocket_state: {{ websocket_state }}<br />
    webrtc_state: {{ webrtc_state }}<br />-->

    <div ref="container" style="width: 1280px; height: 720px; border: 2px solid red">
      <neko-canvas ref="neko" />
    </div>
  </div>
</template>

<style lang="scss" scoped></style>

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

    width = '720px'
    height = '1280px'

    connect() {
      this.neko.connect('ws://192.168.1.20:3000/', 'neko')
    }

    disconnect() {
      this.neko.disconnect()
    }

    resize() {
      this.container.style.width = this.width
      this.container.style.height = this.height
    }

    mounted() {
      this.neko.events.on('connect', () => {
        console.log('connected...')
      })
      this.neko.events.on('host.change', (id) => {
        console.log('host.change', id)
      })
      this.neko.events.on('disconnect', (message) => {
        console.log('disconnect', message)
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
