<template>
  <div ref="component" class="video">
    <button @click="connect()">Connect</button>
    <button @click="disconnect()">Disonnect</button><br />
    websocket_state: {{ websocket_state }}<br />
    webrtc_state: {{ webrtc_state }}<br />

    <div ref="container" class="player-container">
      <video ref="video" />
      <div ref="overlay" class="overlay" tabindex="0" @click.stop.prevent @contextmenu.stop.prevent />
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

    .overlay {
      position: absolute;
      top: 0;
      bottom: 0;
      width: 100%;
      height: 100%;
    }
  }
</style>

<script lang="ts">
  import ResizeObserver from 'resize-observer-polyfill'
  import { Vue, Component, Ref, Watch } from 'vue-property-decorator'
  import { NekoWebSocket } from './internal/websocket'
  import { NekoWebRTC } from './internal/webrtc'

  @Component({
    name: 'neko',
  })
  export default class extends Vue {
    @Ref('component') readonly _component!: HTMLElement
    @Ref('container') readonly _container!: HTMLElement
    @Ref('video') readonly _video!: HTMLVideoElement

    private observer = new ResizeObserver(this.onResize.bind(this))

    protected _websocket?: NekoWebSocket
    protected _webrtc?: NekoWebRTC

    private websocket_state = 'disconnected'
    private webrtc_state = 'disconnected'

    public connect() {
      try {
        this._websocket?.connect('ws://192.168.1.20:3000/', 'admin')
      } catch (e) {}
    }

    public disconnect() {
      this._websocket?.disconnect()
    }

    mounted() {
      // Update canvas on resize
      this._container.addEventListener('resize', this.onResize)
      this.observer.observe(this._component)

      // WebSocket
      this._websocket = new NekoWebSocket()
      this._websocket?.on('message', async (event: string, payload: any) => {
        switch (event) {
          case 'signal/provide':
            try {
              let sdp = await this._webrtc?.connect(payload.sdp, payload.lite, payload.ice)
              this._websocket?.send('signal/answer', { sdp, displayname: 'test' })
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

      // WebRTC
      this._webrtc = new NekoWebRTC()
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

    destroyed() {
      this._webrtc?.disconnect()
      this._websocket?.disconnect()
    }

    public onResize() {
      console.log('Resize event triggered.')
    }
  }
</script>
