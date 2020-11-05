<template>
  <div>
    <h1>Hello world, Neko.</h1>
    <button @click="connect()">Connect</button>
    <button @click="disconnect()">Disonnect</button>
    <button @click="send()">Send</button>
  </div>
</template>

<script lang="ts">
  import { Vue, Component, Ref } from 'vue-property-decorator'
  import { NekoWebSocket } from './internal/websocket'
  import { NekoWebRTC } from './internal/webrtc'

  @Component({
    name: 'neko',
  })
  export default class extends Vue {
    protected _websocket?: NekoWebSocket
    protected _webrtc?: NekoWebRTC

    public connect() {
      try {
        this._websocket?.connect('ws://192.168.1.20:3000/', 'admin')
      } catch (e) {}
    }

    public disconnect() {
      this._websocket?.disconnect()
    }

    public send() {
      this._websocket?.send('test', 'abc')
    }

    mounted() {
      this._websocket = new NekoWebSocket()
      this._webrtc = new NekoWebRTC()

      this._websocket?.on('message', async (event, payload) => {
        if (event == 'signal/provide') {
          try {
            let sdp = await this._webrtc?.connect(payload.sdp, payload.lite, payload.ice)
            this._websocket?.send('signal/answer', { sdp, displayname: 'test' })
          } catch (e) {}
        } else {
          console.log(event, payload)
        }
      })
      this._websocket?.on('disconnected', () => {
        this._webrtc?.disconnect()
      })
    }
  }
</script>
