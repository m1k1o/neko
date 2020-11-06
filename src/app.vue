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
  }
</script>
