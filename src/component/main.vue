<template>
  <div ref="component" class="neko-component">
    <div ref="container" class="neko-container">
      <video ref="video" :autoplay="autoplay" :muted="autoplay" playsinline />
      <neko-overlay
        :webrtc="webrtc"
        :scroll="state.control.scroll"
        :screenSize="state.screen.size"
        :canvasSize="canvasSize"
        :isControling="controlling && watching"
        :cursorTag="
          state.control.implicit_hosting && state.control.host_id != null
            ? state.sessions[state.control.host_id].profile.name
            : ''
        "
        :implicitControl="state.control.implicit_hosting && state.sessions[state.session_id].profile.can_host"
        @implicit-control-request="websocket.send('control/request')"
        @implicit-control-release="websocket.send('control/release')"
        @update-kbd-modifiers="updateKeyboardModifiers($event)"
        @drop-files="uploadDrop($event)"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .neko-component {
    width: 100%;
    height: 100%;
  }

  .neko-container {
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
  export * as ApiModels from './api/models'
  export * as StateModels from './types/state'
  import * as EVENT from './types/events'

  import { Vue, Component, Ref, Watch, Prop } from 'vue-property-decorator'
  import ResizeObserver from 'resize-observer-polyfill'
  import EventEmitter from 'eventemitter3'

  import { NekoApi, MembersApi, RoomApi } from './internal/api'
  import { NekoWebSocket } from './internal/websocket'
  import { NekoWebRTC, WebRTCStats } from './internal/webrtc'
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
    canvasSize: { width: number; height: number } = {
      width: 0,
      height: 0,
    }

    @Prop({ type: Boolean })
    private readonly autoplay!: boolean

    @Prop({ type: Boolean })
    private readonly autologin!: boolean

    @Prop({ type: String })
    private readonly server!: string

    /////////////////////////////
    // Public state
    /////////////////////////////
    public state = {
      connection: {
        authenticated: false,
        websocket: this.websocket.supported ? 'disconnected' : 'unavailable',
        webrtc: {
          status: this.webrtc.supported ? 'disconnected' : 'unavailable',
          stats: null,
          video: null,
          videos: [],
          auto: true,
        },
        type: 'none',
      },
      video: {
        playable: false,
        playing: false,
        volume: 0,
        muted: false,
      },
      control: {
        scroll: {
          inverse: true,
          sensitivity: 1,
        },
        clipboard: null,
        keyboard: {
          layout: 'us',
          variant: '',
        },
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
      session_id: null,
      sessions: {},
    } as NekoState

    public get authenticated() {
      return this.state.connection.authenticated
    }

    public get connected() {
      return this.state.connection.websocket == 'connected'
    }

    public get watching() {
      return this.state.connection.webrtc.status == 'connected'
    }

    public get controlling() {
      return this.state.control.host_id !== null && this.state.session_id === this.state.control.host_id
    }

    public get is_admin() {
      return this.state.session_id != null ? this.state.sessions[this.state.session_id].profile.is_admin : false
    }

    /////////////////////////////
    // Public events
    /////////////////////////////
    public events = new NekoMessages(this.websocket, this.state)

    /////////////////////////////
    // Public methods
    /////////////////////////////
    @Watch('server', { immediate: true })
    public setUrl(url: string) {
      if (!url) {
        url = location.href
      }

      const httpURL = url.replace(/^ws/, 'http').replace(/\/$|\/ws\/?$/, '')
      this.api.setUrl(httpURL)
      this.websocket.setUrl(httpURL)

      if (this.connected) {
        this.websocket.disconnect(new Error('url changed'))
      }

      if (this.watching) {
        this.webrtc.disconnect()
      }

      if (this.authenticated) {
        Vue.set(this.state.connection, 'authenticated', false)
      }

      // check if is user logged in
      if (this.autologin) {
        this.api.session.whoami().then(() => {
          Vue.set(this.state.connection, 'authenticated', true)
          this.websocket.connect()
        })
      }
    }

    public async login(username: string, password: string) {
      if (this.authenticated) {
        throw new Error('client already authenticated')
      }

      await this.api.session.login({ username, password })
      Vue.set(this.state.connection, 'authenticated', true)
      this.websocket.connect()
    }

    public async logout() {
      if (!this.authenticated) {
        throw new Error('client not authenticated')
      }

      if (this.connected) {
        this.websocket.disconnect(new Error('logged out'))
      }

      try {
        await this.api.session.logout()
      } finally {
        Vue.set(this.state.connection, 'authenticated', false)
      }
    }

    public websocketConnect() {
      if (!this.authenticated) {
        throw new Error('client not authenticated')
      }

      if (this.connected) {
        throw new Error('client already connected to websocket')
      }

      this.websocket.connect()
    }

    public websocketDisconnect() {
      if (!this.connected) {
        throw new Error('client not connected to websocket')
      }

      this.websocket.disconnect(new Error('manual action'))
    }

    public webrtcConnect() {
      if (!this.connected) {
        throw new Error('client not connected to websocket')
      }

      if (this.watching) {
        throw new Error('client already connected to webrtc')
      }

      this.websocket.send(EVENT.SIGNAL_REQUEST)
    }

    public webrtcDisconnect() {
      if (!this.watching) {
        throw new Error('client not connected to webrtc')
      }

      this.webrtc.disconnect()
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
        throw new Error('volume must be between 0 and 1')
      }

      this._video.volume = value
    }

    public setScrollInverse(value: boolean = true) {
      Vue.set(this.state.control.scroll, 'inverse', value)
    }

    public setScrollSensitivity(value: number) {
      Vue.set(this.state.control.scroll, 'sensitivity', value)
    }

    public setKeyboard(layout: string, variant: string = '') {
      Vue.set(this.state.control, 'keyboard', { layout, variant })
    }

    // TODO: Remove? Use REST API only?
    public setScreenSize(width: number, height: number, rate: number) {
      this.websocket.send(EVENT.SCREEN_SET, { width, height, rate })
    }

    public setWebRTCVideo(video: string) {
      if (!this.state.connection.webrtc.videos.includes(video)) {
        throw new Error('video id not found')
      }

      this.websocket.send(EVENT.SIGNAL_VIDEO, { video: video })
    }

    public setWebRTCAuto(auto: boolean = true) {
      Vue.set(this.state.connection.webrtc, 'auto', auto)
    }

    public sendUnicast(receiver: string, subject: string, body: any) {
      this.websocket.send(EVENT.SEND_UNICAST, { receiver, subject, body })
    }

    public sendBroadcast(subject: string, body: any) {
      this.websocket.send(EVENT.SEND_BROADCAST, { subject, body })
    }

    public get room(): RoomApi {
      return this.api.room
    }

    public get members(): MembersApi {
      return this.api.members
    }

    async uploadDrop({ x, y, files }: { x: number; y: number; files: Array<Blob> }) {
      try {
        this.events.emit('upload.drop.started')

        await this.api.room.uploadDrop(x, y, files, {
          onUploadProgress: (progressEvent: ProgressEvent) => {
            this.events.emit('upload.drop.progress', progressEvent)
          },
        })

        this.events.emit('upload.drop.finished', null)
      } catch (err) {
        this.events.emit('upload.drop.finished', err)
      }
    }

    /////////////////////////////
    // Component lifecycle
    /////////////////////////////
    mounted() {
      // component size change
      this.observer.observe(this._component)

      // video events
      VideoRegister(this._video, this.state.video)

      // websocket
      this.websocket.on('message', async (event: string, payload: any) => {
        switch (event) {
          case EVENT.SIGNAL_PROVIDE:
            try {
              let sdp = await this.webrtc.connect(payload.sdp, payload.iceservers)
              this.websocket.send(EVENT.SIGNAL_ANSWER, { sdp })
              this.events.emit('connection.webrtc.sdp', 'local', sdp)
            } catch (e) {}
            break
          case EVENT.SIGNAL_CANDIDATE:
            this.webrtc.setCandidate(payload)
            break
          case EVENT.SYSTEM_DISCONNECT:
            this.websocket.disconnect(new Error('disconnected by server'))
            break
        }
      })
      this.websocket.on('connecting', () => {
        Vue.set(this.state.connection, 'websocket', 'connecting')
        this.events.emit('connection.websocket', 'connecting')
      })
      this.websocket.on('connected', () => {
        Vue.set(this.state.connection, 'websocket', 'connected')
        this.events.emit('connection.websocket', 'connected')

        if (!this.watching) {
          this.webrtcConnect()
        }
      })
      this.websocket.on('disconnected', () => {
        Vue.set(this.state.connection, 'websocket', 'disconnected')
        this.events.emit('connection.websocket', 'disconnected')

        this.clearWebSocketState()
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

        if (this.autoplay) {
          this._video.play()
        }
      })
      this.webrtc.on('candidate', (candidate: RTCIceCandidateInit) => {
        this.websocket.send(EVENT.SIGNAL_CANDIDATE, candidate)
        this.events.emit('connection.webrtc.sdp.candidate', 'local', candidate)
      })

      let webrtcCongestion: number = 0
      this.webrtc.on('stats', (stats: WebRTCStats) => {
        Vue.set(this.state.connection.webrtc, 'stats', stats)

        // if automatic quality adjusting is turned off
        if (!this.state.connection.webrtc.auto) return

        // if there are no or just one quality, no switching can be done
        if (this.state.connection.webrtc.videos.length <= 1) return

        // current quality is not known
        if (this.state.connection.webrtc.video == null) return

        // check if video is not playing
        if (stats.fps) {
          webrtcCongestion = 0
          return
        }

        // try to downgrade quality if it happend many times
        if (++webrtcCongestion >= 3) {
          let index = this.state.connection.webrtc.videos.indexOf(this.state.connection.webrtc.video)

          // edge case: current quality is not in qualities list
          if (index === -1) return

          // current quality is the lowest one
          if (index + 1 == this.state.connection.webrtc.videos.length) return

          // downgrade video quality
          this.setWebRTCVideo(this.state.connection.webrtc.videos[index + 1])
          webrtcCongestion = 0
        }
      })
      this.webrtc.on('connecting', () => {
        Vue.set(this.state.connection.webrtc, 'status', 'connecting')
        this.events.emit('connection.webrtc', 'connecting')
      })
      this.webrtc.on('connected', () => {
        Vue.set(this.state.connection.webrtc, 'status', 'connected')
        Vue.set(this.state.connection, 'type', 'webrtc')
        this.events.emit('connection.webrtc', 'connected')
      })

      let webrtcReconnect: any = null
      this.webrtc.on('disconnected', () => {
        this.events.emit('connection.webrtc', 'disconnected')
        this.clearWebRTCState()

        // destroy video
        if (this._video) {
          if ('srcObject' in this._video) {
            this._video.srcObject = null
          } else {
            // @ts-ignore
            this._video.removeAttribute('src')
          }
        }

        // periodically reconnect WebRTC
        if (this.connected && !webrtcReconnect) {
          webrtcReconnect = setInterval(() => {
            // connect only if disconnected
            if (this.state.connection.webrtc.status == 'disconnected') {
              try {
                this.webrtcConnect()
              } catch (e) {}
            }

            // stop reconnecting if connected to webrtc or disconnected from WS
            if (this.state.connection.webrtc.status == 'connected' || !this.connected) {
              clearInterval(webrtcReconnect)
              webrtcReconnect = null
            }
          }, 1000)
        }
      })

      // unmute on users first interaction
      if (this.autoplay) {
        document.addEventListener('click', this.unmute, { once: true })
      }
    }

    beforeDestroy() {
      this.observer.disconnect()
      this.webrtc.disconnect()
      this.websocket.disconnect()

      // remove users first interaction event
      document.removeEventListener('click', this.unmute)
    }

    @Watch('controlling')
    @Watch('state.control.keyboard')
    updateKeyboard() {
      if (this.controlling && this.state.control.keyboard.layout) {
        this.websocket.send(EVENT.KEYBOARD_MAP, this.state.control.keyboard)
      }
    }

    updateKeyboardModifiers(modifiers: { capslock: boolean; numlock: boolean }) {
      this.websocket.send(EVENT.KEYBOARD_MODIFIERS, modifiers)
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

        Vue.set(this, 'canvasSize', {
          width: offsetWidth,
          height: vertical,
        })
      }
      // horizontal centering
      else if (screen_ratio < canvas_ratio) {
        const horizontal = screen_ratio * offsetHeight
        this._container.style.width = `${horizontal}px`
        this._container.style.height = `${offsetHeight}px`
        this._container.style.marginTop = `0px`
        this._container.style.marginLeft = `${(offsetWidth - horizontal) / 2}px`

        Vue.set(this, 'canvasSize', {
          width: horizontal,
          height: offsetHeight,
        })
      }
      // no centering
      else {
        this._container.style.width = `${offsetWidth}px`
        this._container.style.height = `${offsetHeight}px`
        this._container.style.marginTop = `0px`
        this._container.style.marginLeft = `0px`

        Vue.set(this, 'canvasSize', {
          width: offsetWidth,
          height: offsetHeight,
        })
      }
    }

    clearWebSocketState() {
      Vue.set(this.state.control, 'clipboard', null)
      Vue.set(this.state.control, 'host_id', null)
      Vue.set(this.state.control, 'implicit_hosting', false)
      Vue.set(this.state.screen, 'size', { width: 1280, height: 720, rate: 30 })
      Vue.set(this.state.screen, 'configurations', [])
      Vue.set(this.state, 'session_id', null)
      Vue.set(this.state, 'sessions', {})
    }

    clearWebRTCState() {
      Vue.set(this.state.connection.webrtc, 'status', 'disconnected')
      Vue.set(this.state.connection.webrtc, 'stats', null)
      Vue.set(this.state.connection.webrtc, 'video', null)
      Vue.set(this.state.connection.webrtc, 'videos', [])
      Vue.set(this.state.connection, 'type', 'none')
    }
  }
</script>
