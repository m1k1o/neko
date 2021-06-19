<template>
  <div ref="component" class="neko-component">
    <div ref="container" class="neko-container">
      <video ref="video" :autoplay="autoplay" :muted="autoplay" playsinline />
      <neko-overlay
        :webrtc="connection.webrtc"
        :scroll="state.control.scroll"
        :screenSize="state.screen.size"
        :canvasSize="canvasSize"
        :isControling="controlling"
        :cursorTag="
          state.control.implicit_hosting && state.control.host_id != null
            ? state.sessions[state.control.host_id].profile.name
            : ''
        "
        :implicitControl="state.control.implicit_hosting && state.sessions[state.session_id].profile.can_host"
        @implicitControlRequest="connection.websocket.send('control/request')"
        @implicitControlRelease="connection.websocket.send('control/release')"
        @updateKeyboardModifiers="updateKeyboardModifiers($event)"
        @uploadDrop="uploadDrop($event)"
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

  import { NekoApi, MembersApi, RoomApi } from './internal/api'
  import { NekoConnection } from './internal/connection'
  import { WebRTCStats } from './internal/webrtc'
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
    observer = new ResizeObserver(this.onResize.bind(this))
    canvasSize: { width: number; height: number } = {
      width: 0,
      height: 0,
    }

    @Prop({ type: String })
    private readonly server!: string

    @Prop({ type: Boolean })
    private readonly autologin!: boolean

    @Prop({ type: Boolean })
    private readonly autoconnect!: boolean

    @Prop({ type: Boolean })
    private readonly autoplay!: boolean

    /////////////////////////////
    // Public state
    /////////////////////////////
    public state = {
      authenticated: false,
      connection: {
        status: 'disconnected',
        webrtc: {
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
          sensitivity: 0,
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

    /////////////////////////////
    // Public connection manager
    /////////////////////////////
    public connection = new NekoConnection(this.state.connection)

    public get connected() {
      return this.state.connection.status == 'connected'
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
    public events = new NekoMessages(this.connection, this.state)

    /////////////////////////////
    // Public methods
    /////////////////////////////
    @Watch('server', { immediate: true })
    public async setUrl(url: string) {
      if (!url) {
        url = location.href
      }

      const httpURL = url.replace(/^ws/, 'http').replace(/\/$|\/ws\/?$/, '')
      this.api.setUrl(httpURL)
      this.connection.setUrl(httpURL)

      if (this.connected) {
        this.disconnect()
      }

      if (this.state.authenticated) {
        Vue.set(this.state, 'authenticated', false)
      }

      // check if is user logged in
      if (this.autologin) {
        const token = localStorage.getItem('neko_session')
        if (token) {
          this.api.setToken(token)
          this.connection.setToken(token)
        }

        await this.api.session.whoami()
        Vue.set(this.state, 'authenticated', true)

        if (this.autoconnect) {
          await this.connect()
        }
      }
    }

    public async login(username: string, password: string) {
      if (this.state.authenticated) {
        throw new Error('client already authenticated')
      }

      const res = await this.api.session.login({ username, password })
      if (res.data.token) {
        this.api.setToken(res.data.token)
        this.connection.setToken(res.data.token)

        if (this.autologin) {
          localStorage.setItem('neko_session', res.data.token)
        }
      }

      Vue.set(this.state, 'authenticated', true)

      if (this.autoconnect) {
        await this.connect()
      }
    }

    public async logout() {
      if (!this.state.authenticated) {
        throw new Error('client not authenticated')
      }

      if (this.connected) {
        this.disconnect()
      }

      try {
        await this.api.session.logout()
      } finally {
        this.api.setToken('')
        this.connection.setToken('')

        if (this.autologin) {
          localStorage.removeItem('neko_session')
        }

        Vue.set(this.state, 'authenticated', false)
      }
    }

    // TODO: Refactor.
    public async connect(video?: string) {
      if (!this.state.authenticated) {
        throw new Error('client not authenticated')
      }

      if (this.connected) {
        throw new Error('client is already connected')
      }

      await this.connection.connect()

      if (video && !this.state.connection.webrtc.videos.includes(video)) {
        throw new Error('video id not found')
      }

      this.connection.websocket.send(EVENT.SIGNAL_REQUEST, { video: video })
    }

    public disconnect() {
      if (!this.connected) {
        throw new Error('client is not connected')
      }

      // TODO: Refactor.
      this.connection.webrtc.disconnect()

      this.connection.disconnect()
      this.clear()
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
      this.connection.websocket.send(EVENT.SCREEN_SET, { width, height, rate })
    }

    public setWebRTCVideo(video: string) {
      if (!this.state.connection.webrtc.videos.includes(video)) {
        throw new Error('video id not found')
      }

      this.connection.websocket.send(EVENT.SIGNAL_VIDEO, { video: video })
    }

    public setWebRTCAuto(auto: boolean = true) {
      Vue.set(this.state.connection.webrtc, 'auto', auto)
    }

    public sendUnicast(receiver: string, subject: string, body: any) {
      this.connection.websocket.send(EVENT.SEND_UNICAST, { receiver, subject, body })
    }

    public sendBroadcast(subject: string, body: any) {
      this.connection.websocket.send(EVENT.SEND_BROADCAST, { subject, body })
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

      // webrtc
      this.connection.webrtc.on('track', (event: RTCTrackEvent) => {
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

      let webrtcCongestion: number = 0
      this.connection.webrtc.on('stats', (stats: WebRTCStats) => {
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

      // unmute on users first interaction
      if (this.autoplay) {
        document.addEventListener('click', this.unmute, { once: true })
      }
    }

    beforeDestroy() {
      this.observer.disconnect()
      this.disconnect()

      // remove users first interaction event
      document.removeEventListener('click', this.unmute)
    }

    @Watch('controlling')
    @Watch('state.control.keyboard')
    updateKeyboard() {
      if (this.controlling && this.state.control.keyboard.layout) {
        this.connection.websocket.send(EVENT.KEYBOARD_MAP, this.state.control.keyboard)
      }
    }

    updateKeyboardModifiers(modifiers: { capslock: boolean; numlock: boolean }) {
      this.connection.websocket.send(EVENT.KEYBOARD_MODIFIERS, modifiers)
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

    clear() {
      // destroy video
      if (this._video) {
        if ('srcObject' in this._video) {
          this._video.srcObject = null
        } else {
          // @ts-ignore
          this._video.removeAttribute('src')
        }
      }

      // websocket
      Vue.set(this.state.connection.webrtc, 'videos', [])
      Vue.set(this.state.control, 'clipboard', null)
      Vue.set(this.state.control, 'host_id', null)
      Vue.set(this.state.control, 'implicit_hosting', false)
      Vue.set(this.state.screen, 'size', { width: 1280, height: 720, rate: 30 })
      Vue.set(this.state.screen, 'configurations', [])
      Vue.set(this.state, 'session_id', null)
      Vue.set(this.state, 'sessions', {})

      // webrtc
      Vue.set(this.state.connection.webrtc, 'stats', null)
      Vue.set(this.state.connection.webrtc, 'video', null)
      Vue.set(this.state.connection, 'type', 'none')
    }
  }
</script>
