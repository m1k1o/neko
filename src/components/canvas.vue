<template>
  <div ref="component" class="component">
    <div ref="container" class="player-container" v-show="state.websocket == 'connected' && state.webrtc == 'connected'">
      <video ref="video" />
      <neko-overlay
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
  import EventEmitter from 'eventemitter3'

  import { NekoWebSocket } from '~/internal/websocket'
  import { NekoWebRTC } from '~/internal/webrtc'
  import { NekoMessages } from '~/internal/messages'

  import NekoState from '~/types/state'
  import Overlay from './overlay.vue'

  export interface NekoEvents {
    connecting: () => void
    connected: () => void
    disconnected: (error?: Error) => void
  }

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
    private messages = new NekoMessages()
    public readonly events = new EventEmitter<NekoEvents>()

    private state = {
      id: null,
      display_name: null,
      screen_size: {
        width: 1280,
        height: 720,
        rate: 30,
      },
      is_controlling: false,
      websocket: 'disconnected',
      webrtc: 'disconnected',
    } as NekoState

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
            // @ts-ignore
            if (typeof this.messages[event] === 'function') {
              // @ts-ignore
              this.messages[event](payload)
            } else {
              console.log(`unhandled websocket event '${event}':`, payload)
            }
        }
      })
      this.websocket.on('connecting', () => {
        Vue.set(this.state, 'websocket', 'connecting')
      })
      this.websocket.on('connected', () => {
        Vue.set(this.state, 'websocket', 'connected')
      })
      this.websocket.on('disconnected', () => {
        Vue.set(this.state, 'websocket', 'disconnected')
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
        Vue.set(this.state, 'webrtc', 'connecting')
      })
      this.webrtc.on('connected', () => {
        Vue.set(this.state, 'webrtc', 'connected')
        this.events.emit('connected')
      })
      this.webrtc.on('disconnected', () => {
        Vue.set(this.state, 'webrtc', 'disconnected')
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
