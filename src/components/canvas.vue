<template>
  <div ref="component" class="component">
    <div ref="container" class="player-container">
      <video ref="video" />
      <neko-overlay
        :webrtc="webrtc"
        :screenWidth="state.screen.size.width"
        :screenHeight="state.screen.size.height"
        :isControling="state.member.is_controlling"
        :scrollSensitivity="state.control.scroll.sensitivity"
        :scrollInvert="state.control.scroll.inverse"
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
  import { register as VideoRegister } from '~/internal/video'

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
    @Ref('video') readonly _video!: HTMLVideoElement

    websocket = new NekoWebSocket()
    webrtc = new NekoWebRTC()
    observer = new ResizeObserver(this.onResize.bind(this))

    /////////////////////////////
    // Public state
    /////////////////////////////
    public state = {
      connection: {
        websocket: 'disconnected',
        webrtc: 'disconnected',
        type: 'none',
        can_watch: false,
        can_control: false,
        clipboard_access: false,
      },
      video: {
        playable: false,
        playing: false,
        volume: 0,
        fullscreen: false,
      },
      control: {
        scroll: {
          inverse: true,
          sensitivity: 1,
        },
        clipboard: {
          data: null,
        },
        host: null,
      },
      screen: {
        size: {
          width: 1280,
          height: 720,
          rate: 30,
        },
        configurations: [],
      },
      member: {
        id: null,
        name: null,
        is_admin: false,
        is_watching: false,
        is_controlling: false,
        can_watch: false,
        can_control: false,
        clipboard_access: false,
      },
      members: [],
    } as NekoState

    public get connected() {
      return this.state.connection.websocket == 'connected' && this.state.connection.webrtc == 'connected'
    }

    /////////////////////////////
    // Public events
    /////////////////////////////
    public events = new NekoMessages(this.websocket, this.state)

    /////////////////////////////
    // Public methods
    /////////////////////////////
    public connect(url: string, password: string, name: string) {
      if (this.connected) {
        throw new Error('client already connected')
      }

      Vue.set(this.state.member, 'name', name)
      this.websocket.connect(url, password)
    }

    public disconnect() {
      if (!this.connected) {
        throw new Error('client not connected')
      }

      this.websocket.disconnect()
    }

    public play() {
      this._video.play()
    }

    public pause() {
      this._video.pause()
    }

    public setVolume(value: number) {
      if (value < 0 || value > 1) {
        throw new Error('Out of range. Value must be between 0 and 1.')
      }

      this._video.volume = value
    }

    public requestFullscreen() {
      this._component.requestFullscreen()
    }

    public exitFullscreen() {
      document.exitFullscreen()
    }

    public setScrollInverse(value: boolean = true) {
      Vue.set(this.state.control.scroll, 'inverse', value)
    }

    public setScrollSensitivity(value: number) {
      Vue.set(this.state.control.scroll, 'sensitivity', value)
    }

    public setClipboardData(value: number) {
      // TODO: Via REST API.
    }

    public requestControl() {
      // TODO: Via REST API.
      this.websocket.send('control/request')
    }

    public releaseControl() {
      // TODO: Via REST API.
      this.websocket.send('control/release')
    }

    public takeControl() {
      // TODO: Via REST API.
    }

    public giveControl(id: string) {
      // TODO: Via REST API.
    }

    public resetControl() {
      // TODO: Via REST API.
    }

    public setScreenSize(width: number, height: number, rate: number) {
      // TODO: Via REST API.
      this.websocket.send('screen/set', { width, height, rate })
    }

    /////////////////////////////
    // Component lifecycle
    /////////////////////////////
    mounted() {
      // component size change
      this.observer.observe(this._component)

      // host change
      this.events.on('control.host', (id: string | null) => {
        Vue.set(this.state.member, 'is_controlling', id != null && id === this.state.member.id)
      })

      // fullscreen change
      this._component.addEventListener('fullscreenchange', () => {
        Vue.set(this.state.video, 'fullscreen', document.fullscreenElement !== null)
        this.onResize()
      })

      // video events
      VideoRegister(this._video, this.state.video)

      // websocket
      this.websocket.on('message', async (event: string, payload: any) => {
        switch (event) {
          case 'signal/provide':
            Vue.set(this.state.member, 'id', payload.id)

            try {
              let sdp = await this.webrtc.connect(payload.sdp, payload.lite, payload.ice)
              this.websocket.send('signal/answer', { sdp, displayname: this.state.member.name })
            } catch (e) {}
            break
        }
      })
      this.websocket.on('connecting', () => {
        Vue.set(this.state.connection, 'websocket', 'connecting')
        this.events.emit('system.websocket', 'connecting')
      })
      this.websocket.on('connected', () => {
        Vue.set(this.state.connection, 'websocket', 'connected')
        this.events.emit('system.websocket', 'connected')
      })
      this.websocket.on('disconnected', () => {
        Vue.set(this.state.connection, 'websocket', 'disconnected')
        this.events.emit('system.websocket', 'disconnected')
        this.webrtc.disconnect()

        // TODO: reset state
        Vue.set(this.state, 'member', {
          id: null,
          name: null,
          is_admin: false,
          is_watching: false,
          is_controlling: false,
          can_watch: false,
          can_control: false,
          clipboard_access: false,
        })
      })

      // webrtc
      this.webrtc.on('track', (event: RTCTrackEvent) => {
        const { track, streams } = event
        if (track.kind === 'audio') return

        // create stream
        if ('srcObject' in this._video) {
          this._video.srcObject = streams[0]
        } else {
          // @ts-ignore
          this._video.src = window.URL.createObjectURL(streams[0]) // for older browsers
        }

        this._video.play()
      })
      this.webrtc.on('connecting', () => {
        Vue.set(this.state.connection, 'webrtc', 'connecting')
        this.events.emit('system.webrtc', 'connecting')
      })
      this.webrtc.on('connected', () => {
        Vue.set(this.state.connection, 'webrtc', 'connected')
        this.events.emit('system.webrtc', 'connected')
      })
      this.webrtc.on('disconnected', () => {
        Vue.set(this.state.connection, 'webrtc', 'disconnected')
        this.events.emit('system.webrtc', 'disconnected')
        // @ts-ignore
        this._video.src = null
      })

      // hardcoded webrtc for now
      Vue.set(this.state.connection, 'type', 'webrtc')
      Vue.set(this.state.connection, 'can_watch', this.webrtc.supported)
      Vue.set(this.state.connection, 'can_control', this.webrtc.supported)
    }

    beforeDestroy() {
      this.observer.disconnect()
      this.webrtc.disconnect()
      this.websocket.disconnect()
    }

    @Watch('state.video.playing')
    onVideoPlayingChanged(play: boolean) {
      // TODO: check if user has tab focused and send via websocket
      Vue.set(this.state.member, 'is_watching', play)
    }

    @Watch('state.screen.size')
    onResize() {
      const { width, height } = this.state.screen.size
      const screen_ratio = width / height

      const { offsetWidth, offsetHeight } = this._component
      const canvas_ratio = offsetWidth / offsetHeight

      // vertical centering
      if (screen_ratio > canvas_ratio) {
        const vertical = offsetWidth / screen_ratio
        this._container.style.width = `${offsetWidth}px`
        this._container.style.height = `${vertical}px`
        this._container.style.marginTop = `${(offsetHeight - vertical) / 2}px`
        this._container.style.marginLeft = `0px`
      }
      // horizontal centering
      else if (screen_ratio < canvas_ratio) {
        const horizontal = screen_ratio * offsetHeight
        this._container.style.width = `${horizontal}px`
        this._container.style.height = `${offsetHeight}px`
        this._container.style.marginTop = `0px`
        this._container.style.marginLeft = `${(offsetWidth - horizontal) / 2}px`
      }
      // no centering
      else {
        this._container.style.width = `${offsetWidth}px`
        this._container.style.height = `${offsetHeight}px`
        this._container.style.marginTop = `0px`
        this._container.style.marginLeft = `0px`
      }
    }
  }
</script>
