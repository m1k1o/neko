<template>
  <div ref="component" class="video">
    <button @click="connect()">Connect</button>
    <button @click="disconnect()">Disonnect</button><br />
    websocket_state: {{ websocket_state }}<br />
    webrtc_state: {{ webrtc_state }}<br />

    <div ref="container" class="player-container">
      <video ref="video" />
      <neko-overlay
        v-if="websocket_state == 'connected' && webrtc_state == 'connected'"
        :webrtc="webrtc"
        :screenWidth="1280"
        :screenHeight="720"
        :scrollSensitivity="5"
        :scrollInvert="true"
        :isControling="true"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .player-container {
    position: relative;
    width: 1280px;
    height: 720px;

    video {
      position: absolute;
      top: 0;
      bottom: 0;
      width: 100%;
      height: 100%;
      display: flex;
      background: #000;

      &::-webkit-media-controls {
        display: none !important;
      }
    }
  }
</style>

<script lang="ts">
  import ResizeObserver from 'resize-observer-polyfill'
  import { Vue, Component, Ref, Watch } from 'vue-property-decorator'
  import { NekoWebSocket } from './internal/websocket'
  import { NekoWebRTC } from './internal/webrtc'

  import Overlay from '~/components/overlay.vue'

  @Component({
    name: 'neko',
    components: {
      'neko-overlay': Overlay,
    },
  })
  export default class extends Vue {
    @Ref('component') readonly _component!: HTMLElement
    @Ref('container') readonly _container!: HTMLElement
    @Ref('video') readonly _video!: HTMLVideoElement

    private observer = new ResizeObserver(this.onResize.bind(this))

    websocket: NekoWebSocket | null = null
    webrtc: NekoWebRTC | null = null

    private websocket_state = 'disconnected'
    private webrtc_state = 'disconnected'

    public connect() {
      try {
        this.websocket?.connect('ws://192.168.1.20:3000/', 'admin')
      } catch (e) {}
    }

    public disconnect() {
      this.websocket?.disconnect()
    }

    mounted() {
      // Update canvas on resize
      this._container.addEventListener('resize', this.onResize)
      this.observer.observe(this._component)

      // WebSocket
      this.websocket = new NekoWebSocket()
      this.websocket?.on('message', async (event: string, payload: any) => {
        switch (event) {
          case 'signal/provide':
            try {
              let sdp = await this.webrtc?.connect(payload.sdp, payload.lite, payload.ice)
              this.websocket?.send('signal/answer', { sdp, displayname: 'test' })
            } catch (e) {}
            break
          case 'screen/resolution':
            payload.width
            payload.height
            payload.rate
            break
          default:
            console.log(event, payload)
        }
      })
      this.websocket?.on('connecting', () => {
        this.websocket_state = 'connecting'
      })
      this.websocket?.on('connected', () => {
        this.websocket_state = 'connected'
      })
      this.websocket?.on('disconnected', () => {
        this.websocket_state = 'disconnected'
        this.webrtc?.disconnect()
      })

      // WebRTC
      this.webrtc = new NekoWebRTC()
      this.webrtc?.on('track', (event: RTCTrackEvent) => {
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
      this.webrtc?.on('connecting', () => {
        this.webrtc_state = 'connecting'
      })
      this.webrtc?.on('connected', () => {
        this.webrtc_state = 'connected'
      })
      this.webrtc?.on('disconnected', () => {
        this.webrtc_state = 'disconnected'
      })
    }

    destroyed() {
      this.webrtc?.disconnect()
      this.websocket?.disconnect()
    }

    public onResize() {
      console.log('Resize event triggered.')
    }
  }
</script>
