<template>
  <div>
    <button @click="connect()">Connect</button>
    <button @click="disconnect()">Disonnect</button>
    <button @click="send()">Send</button><br />
    websocket_state: {{ websocket_state }}<br />
    webrtc_state: {{ webrtc_state }}<br />

    <video ref="video" />
  </div>
</template>

<script lang="ts">
  import { Vue, Component, Ref, Watch } from 'vue-property-decorator'
  import { NekoWebSocket } from './internal/websocket'
  import { NekoWebRTC } from './internal/webrtc'

  @Component({
    name: 'neko',
  })
  export default class extends Vue {
    @Ref('video') readonly _video!: HTMLVideoElement

    protected _websocket?: NekoWebSocket
    protected _webrtc?: NekoWebRTC

    websocket_state = 'disconnected'
    webrtc_state = 'disconnected'

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

      this._websocket?.on('message', async (event: string, payload: any) => {
        if (event == 'signal/provide') {
          try {
            let sdp = await this._webrtc?.connect(payload.sdp, payload.lite, payload.ice)
            this._websocket?.send('signal/answer', { sdp, displayname: 'test' })
          } catch (e) {}
        } else {
          console.log(event, payload)
        }
      })
      this._websocket?.on('connecting', () => {
        this.websocket_state = 'connecting'
      })
      this._websocket?.on('connected', () => {
        this.websocket_state = 'connected'
      })
      this._websocket?.on('disconnected', () => {
        this.websocket_state = 'disconnected'
        this._webrtc?.disconnect()
      })

      this._webrtc?.on('track', (event: RTCTrackEvent) => {
        const { track, streams } = event
        if (track.kind === 'audio') {
          return
        }

        // Create stream
        if ('srcObject' in this._video) {
          this._video.srcObject = streams[0]
        } else {
          // @ts-ignore
          this._video.src = window.URL.createObjectURL(streams[0]) // for older browsers
        }

        this._video.play()
      })
      this._webrtc?.on('connecting', () => {
        this.webrtc_state = 'connecting'
      })
      this._webrtc?.on('connected', () => {
        this.webrtc_state = 'connected'
      })
      this._webrtc?.on('disconnected', () => {
        this.webrtc_state = 'disconnected'
      })
    }
  }
</script>
