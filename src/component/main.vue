<template>
  <div ref="component" class="component">
    <div ref="container" class="player-container">
      <video ref="video" />
      <neko-overlay
        :webrtc="webrtc"
        :screenWidth="state.screen.size.width"
        :screenHeight="state.screen.size.height"
        :isControling="controlling"
        :scrollSensitivity="state.control.scroll.sensitivity"
        :scrollInvert="state.control.scroll.inverse"
        :implicitControl="state.control.implicit_hosting && state.members[state.member_id].profile.can_host"
        @implicit-control-request="websocket.send('control/request')"
        @implicit-control-release="websocket.send('control/release')"
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

  import { NekoApi } from './internal/api'
  import { NekoWebSocket } from './internal/websocket'
  import { NekoWebRTC } from './internal/webrtc'
  import { NekoMessages } from './internal/messages'
  import { register as VideoRegister } from './internal/video'

  import NekoState from './types/state'
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

    api = new NekoApi()
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
        muted: false,
        fullscreen: false,
      },
      control: {
        scroll: {
          inverse: true,
          sensitivity: 1,
        },
        clipboard: null,
        host_id: null,
        implicit_hosting: false,
      },
      screen: {
        size: {
          width: 1280,
          height: 720,
          rate: 30,
        },
        configurations: [],
      },
      member_id: null,
      members: {},
    } as NekoState

    public get connected() {
      return this.state.connection.websocket == 'connected' && this.state.connection.webrtc == 'connected'
    }

    public get controlling() {
      return this.state.control.host_id !== null && this.state.member_id === this.state.control.host_id
    }

    public get is_admin() {
      return this.state.member_id != null ? this.state.members[this.state.member_id].profile.is_admin : false
    }

    /////////////////////////////
    // Public events
    /////////////////////////////
    public events = new NekoMessages(this.websocket, this.state)

    /////////////////////////////
    // Public methods
    /////////////////////////////
    public connect(url: string, id: string, secret: string) {
      if (this.connected) {
        throw new Error('client already connected')
      }

      const wsURL = url.replace(/^http/, 'ws').replace(/\/$|\/ws\/?$/, '')
      this.websocket.connect(wsURL, id, secret)

      const httpURL = url.replace(/^ws/, 'http').replace(/\/$|\/ws\/?$/, '')
      this.api.connect(httpURL, id, secret)
    }

    public disconnect() {
      if (!this.connected) {
        throw new Error('client not connected')
      }

      this.websocket.disconnect()
      this.api.disconnect()
    }

    public play() {
      this._video.play()
    }

    public pause() {
      this._video.pause()
    }

    public mute() {
      this._video.muted = true
    }

    public unmute() {
      this._video.muted = false
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

    public setClipboardData(text: string) {
      const clipboardPayload = { text }
      this.api.host.clipboardWrite({ clipboardPayload })
    }

    public requestControl() {
      this.api.user.controlRequest()
    }

    public releaseControl() {
      this.api.user.controlRelease()
    }

    public takeControl() {
      this.api.admin.controlTake()
    }

    public giveControl(id: string) {
      const controlTargetPayload = { id }
      this.api.admin.controlGive({ controlTargetPayload })
    }

    public resetControl() {
      this.api.admin.controlReset()
    }

    public setScreenSize(width: number, height: number, rate: number) {
      //this.api.admin.screenConfigurationChange({ screenConfigurationPayload: { width, height, rate } })
      this.websocket.send('screen/set', { width, height, rate })
    }

    public memberCreate(memberDataPayload: {
      id: string
      secret: string
      name: string
      isAdmin: boolean
      canLogin: boolean
      canConnect: boolean
      canWatch: boolean
      canHost: boolean
      canAccessClipboard: boolean
    }) {
      this.api.admin.membersCreate({ memberDataPayload })
    }

    public memberUpdate(
      memberId: string,
      memberDataPayload: {
        secret: string
        name: string
        isAdmin: boolean
        canLogin: boolean
        canConnect: boolean
        canWatch: boolean
        canHost: boolean
        canAccessClipboard: boolean
      },
    ) {
      this.api.admin.membersUpdate({ memberId, memberDataPayload })
    }

    public memberDelete(memberId: string) {
      this.api.admin.membersDelete({ memberId })
    }

    /////////////////////////////
    // Component lifecycle
    /////////////////////////////
    mounted() {
      // component size change
      this.observer.observe(this._component)

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
            try {
              let sdp = await this.webrtc.connect(payload.sdp, payload.lite, payload.ice)
              this.websocket.send('signal/answer', { sdp })
            } catch (e) {}
            break
        }
      })
      this.websocket.on('connecting', () => {
        Vue.set(this.state.connection, 'websocket', 'connecting')
        this.events.emit('internal.websocket', 'connecting')
      })
      this.websocket.on('connected', () => {
        this.websocket.send('signal/request')
        Vue.set(this.state.connection, 'websocket', 'connected')
        this.events.emit('internal.websocket', 'connected')
      })
      this.websocket.on('disconnected', () => {
        Vue.set(this.state.connection, 'websocket', 'disconnected')
        this.events.emit('internal.websocket', 'disconnected')

        this.webrtc.disconnect()
        this.clearState()
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
        this.events.emit('internal.webrtc', 'connecting')
      })
      this.webrtc.on('connected', () => {
        Vue.set(this.state.connection, 'webrtc', 'connected')
        this.events.emit('internal.webrtc', 'connected')
      })
      this.webrtc.on('disconnected', () => {
        Vue.set(this.state.connection, 'webrtc', 'disconnected')
        this.events.emit('internal.webrtc', 'disconnected')

        // destroy stream
        if ('srcObject' in this._video) {
          this._video.srcObject = null
        } else {
          // @ts-ignore
          this._video.removeAttribute('src')
        }
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

    clearState() {
      Vue.set(this.state.control, 'clipboard', null)
      Vue.set(this.state.control, 'host_id', null)
      Vue.set(this.state.control, 'implicit_hosting', false)
      Vue.set(this.state.screen, 'size', { width: 1280, height: 720, rate: 30 })
      Vue.set(this.state.screen, 'configurations', [])
      Vue.set(this.state, 'member_id', null)
      Vue.set(this.state, 'members', {})
    }
  }
</script>
