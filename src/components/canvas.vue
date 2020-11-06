<template>
  <div ref="component" class="component">
    <div ref="container" class="player-container">
      <video ref="video" />
      <neko-overlay
        v-if="websocket_state == 'connected' && webrtc_state == 'connected'"
        :webrtc="webrtc"
        :screenWidth="state.screen_size.width"
        :screenHeight="state.screen_size.height"
        :scrollSensitivity="5"
        :scrollInvert="true"
        :isControling="state.is_controlling"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .component {
    width: 100%;
    height: 100%;
  }

  .player-container {
    position: relative;

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
  import { Vue, Component, Ref, Watch, Prop } from 'vue-property-decorator'
  import ResizeObserver from 'resize-observer-polyfill'

  import { NekoWebSocket } from '~/internal/websocket'
  import { NekoWebRTC } from '~/internal/webrtc'

  import NekoState from '~/types/state'
  import Overlay from './overlay.vue'

  @Component({
    name: 'neko-canvas',
    components: {
      'neko-overlay': Overlay,
    },
  })
  export default class extends Vue {
    @Ref('component') readonly _component!: HTMLElement
    @Ref('container') readonly _container!: HTMLElement
    @Ref('video') public readonly video!: HTMLVideoElement

    private observer = new ResizeObserver(this.onResize.bind(this))
    private websocket = new NekoWebSocket()
    private webrtc = new NekoWebRTC()

    private state = {
      id: null,
      display_name: null,
      screen_size: {
        width: 1280,
        height: 720,
        rate: 30,
      },
      is_controlling: false,
    } as NekoState

    private websocket_state = 'disconnected'
    private webrtc_state = 'disconnected'

    public control = {
      request: () => {
        this.websocket.send('control/request')
      },
      release: () => {
        this.websocket.send('control/release')
      },
    }

    public connect(url: string, password: string) {
      if (this.websocket.connected) {
        throw new Error('client already connected')
      }

      this.websocket.connect(url, password)
    }

    public disconnect() {
      if (!this.websocket.connected) {
        throw new Error('client not connected')
      }

      this.websocket.disconnect()
    }

    private mounted() {
      // Update canvas on resize
      this.observer.observe(this._component)

      // WebSocket
      this.websocket.on('message', async (event: string, payload: any) => {
        switch (event) {
          case 'signal/provide':
            Vue.set(this.state, 'id', payload.id)

            try {
              let sdp = await this.webrtc.connect(payload.sdp, payload.lite, payload.ice)
              this.websocket.send('signal/answer', { sdp, displayname: this.state.display_name })
            } catch (e) {}
            break
          case 'screen/resolution':
            Vue.set(this.state, 'screen_size', payload)
            this.onResize()
            break
          case 'control/release':
            if (payload.id === this.state.id) {
              Vue.set(this.state, 'is_controlling', false)
            }
            break
          case 'control/locked':
            if (payload.id === this.state.id) {
              Vue.set(this.state, 'is_controlling', true)
            }
            break
          default:
            console.log(event, payload)
        }
      })
      this.websocket.on('connecting', () => {
        this.websocket_state = 'connecting'
      })
      this.websocket.on('connected', () => {
        this.websocket_state = 'connected'
      })
      this.websocket.on('disconnected', () => {
        this.websocket_state = 'disconnected'
        this.webrtc.disconnect()
      })

      // WebRTC
      this.webrtc.on('track', (event: RTCTrackEvent) => {
        const { track, streams } = event
        if (track.kind === 'audio') {
          return
        }

        // Create stream
        if ('srcObject' in this.video) {
          this.video.srcObject = streams[0]
        } else {
          // @ts-ignore
          this.video.src = window.URL.createObjectURL(streams[0]) // for older browsers
        }

        this.video.play()
      })
      this.webrtc.on('connecting', () => {
        this.webrtc_state = 'connecting'
      })
      this.webrtc.on('connected', () => {
        this.webrtc_state = 'connected'
      })
      this.webrtc.on('disconnected', () => {
        this.webrtc_state = 'disconnected'
      })
    }

    private beforeDestroy() {
      this.webrtc.disconnect()
      this.websocket.disconnect()
    }

    private onResize() {
      console.log('Resize event triggered.')

      const { width, height } = this.state.screen_size
      const screen_ratio = width / height

      const { offsetWidth, offsetHeight } = this._component
      const canvas_ratio = offsetWidth / offsetHeight

      // Vertical centering
      if (screen_ratio > canvas_ratio) {
        const vertical = offsetWidth / screen_ratio
        this._container.style.width = `${offsetWidth}px`
        this._container.style.height = `${vertical}px`
        this._container.style.marginTop = `${(offsetHeight - vertical) / 2}px`
        this._container.style.marginLeft = `0px`
      }
      // Horizontal centering
      else if (screen_ratio < canvas_ratio) {
        const horizontal = screen_ratio * offsetHeight
        this._container.style.width = `${horizontal}px`
        this._container.style.height = `${offsetHeight}px`
        this._container.style.marginTop = `0px`
        this._container.style.marginLeft = `${(offsetWidth - horizontal) / 2}px`
      }
      // No centering
      else {
        this._container.style.width = `${offsetWidth}px`
        this._container.style.height = `${offsetHeight}px`
        this._container.style.marginTop = `0px`
        this._container.style.marginLeft = `0px`
      }
    }
  }
</script>
