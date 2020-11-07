<template>
  <div ref="component" class="component">
    <div
      ref="container"
      class="player-container"
      v-show="state.websocket == 'connected' && state.webrtc == 'connected'"
    >
      <video ref="video" />
      <neko-overlay
        :webrtc="webrtc"
        :screenWidth="state.screen_size.width"
        :screenHeight="state.screen_size.height"
        :isControling="state.is_controlling"
        :scrollSensitivity="state.scroll.sensitivity"
        :scrollInvert="state.scroll.invert"
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

    private websocket = new NekoWebSocket()
    private webrtc = new NekoWebRTC()
    private observer = new ResizeObserver(this.onResize.bind(this))

    public events = new NekoMessages(this.websocket)
    public state = {
      id: null,
      display_name: null,
      screen_size: {
        width: 1280,
        height: 720,
        rate: 30,
      },
      available_screen_sizes: [],
      scroll: {
        sensitivity: 10,
        invert: true,
      },
      is_controlling: false,
      websocket: 'disconnected',
      webrtc: 'disconnected',
    } as NekoState

    public get connected() {
      return this.state.websocket == 'connected' && this.state.webrtc == 'connected'
    }

    public control = {
      request: () => {
        this.websocket.send('control/request')
      },
      release: () => {
        this.websocket.send('control/release')
      },
    }

    public screen = {
      size: (width: number, height: number, rate: number) => {
        this.websocket.send('screen/set', { width, height, rate })
      },
    }

    public scroll = {
      sensitivity: (sensitivity: number) => {
        Vue.set(this.state.scroll, 'sensitivity', sensitivity)
      },
      inverse: (inverse: boolean) => {
        Vue.set(this.state.scroll, 'inverse', inverse)
      },
    }

    public connect(url: string, password: string, name: string) {
      if (this.websocket.connected) {
        throw new Error('client already connected')
      }

      Vue.set(this.state, 'display_name', name)
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

      this.events.on('control.host', (id: string | null) => {
        Vue.set(this.state, 'is_controlling', id != null && id === this.state.id)
      })

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
          case 'screen/configurations':
            let data = []
            for (const i of Object.keys(payload.configurations)) {
              const { width, height, rates } = payload.configurations[i]
              if (width >= 600 && height >= 300) {
                for (const j of Object.keys(rates)) {
                  const rate = rates[j]
                  if (rate === 30 || rate === 60) {
                    data.push({
                      width,
                      height,
                      rate,
                    })
                  }
                }
              }
            }

            let conf = data.sort((a, b) => {
              if (b.width === a.width && b.height == a.height) {
                return b.rate - a.rate
              } else if (b.width === a.width) {
                return b.height - a.height
              }
              return b.width - a.width
            })

            Vue.set(this.state, 'available_screen_sizes', conf)
            this.onResize()
            break
        }
      })
      this.websocket.on('connecting', () => {
        Vue.set(this.state, 'websocket', 'connecting')
        this.events.emit('system.websocket', 'connecting')
      })
      this.websocket.on('connected', () => {
        Vue.set(this.state, 'websocket', 'connected')
        this.events.emit('system.websocket', 'connected')
      })
      this.websocket.on('disconnected', () => {
        Vue.set(this.state, 'websocket', 'disconnected')
        this.events.emit('system.websocket', 'disconnected')
        this.webrtc.disconnect()
      })

      // WebRTC
      this.webrtc.on('track', (event: RTCTrackEvent) => {
        const { track, streams } = event
        if (track.kind === 'audio') return

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
        this.events.emit('system.webrtc', 'connecting')
      })
      this.webrtc.on('connected', () => {
        Vue.set(this.state, 'webrtc', 'connected')
        this.events.emit('system.webrtc', 'connected')
      })
      this.webrtc.on('disconnected', () => {
        Vue.set(this.state, 'webrtc', 'disconnected')
        this.events.emit('system.webrtc', 'disconnected')
      })
    }

    private beforeDestroy() {
      this.observer.disconnect()
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
