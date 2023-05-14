<template>
  <div ref="component" class="neko-component">
    <div ref="container" class="neko-container">
      <video ref="video" playsinline />
      <neko-screencast
        v-show="screencast && screencastReady"
        :image="fallbackImage"
        :enabled="screencast || (!state.connection.webrtc.stable && state.connection.webrtc.connected)"
        :api="api.room"
        @imageReady="screencastReady = $event"
      />
      <neko-cursors
        v-if="state.settings.inactive_cursors && session.profile.can_see_inactive_cursors"
        :sessions="state.sessions"
        :sessionId="state.session_id"
        :hostId="state.control.host_id"
        :screenSize="state.screen.size"
        :canvasSize="canvasSize"
        :cursors="state.cursors"
        :cursorDraw="inactiveCursorDrawFunction"
        :fps="fps"
      />
      <neko-overlay
        ref="overlay"
        v-show="!private_mode_enabled && state.connection.status != 'disconnected'"
        :style="{ pointerEvents: state.control.locked ? 'none' : 'auto' }"
        :wsControl="control"
        :sessions="state.sessions"
        :hostId="state.control.host_id"
        :webrtc="connection.webrtc"
        :scroll="state.control.scroll"
        :screenSize="state.screen.size"
        :canvasSize="canvasSize"
        :isControling="controlling"
        :cursorDraw="cursorDrawFunction"
        :implicitControl="state.settings.implicit_hosting && session.profile.can_host"
        :inactiveCursors="state.settings.inactive_cursors && session.profile.sends_inactive_cursor"
        :fps="fps"
        :hasMobileKeyboard="is_touch_device"
        @updateKeyboardModifiers="updateKeyboardModifiers($event)"
        @uploadDrop="uploadDrop($event)"
        @mobileKeyboardOpen="state.mobile_keyboard_open = $event"
      />
    </div>
  </div>
</template>

<style lang="scss">
  .neko-component {
    width: 100%;
    height: 100%;
  }

  .neko-container {
    position: relative;

    video,
    img {
      position: absolute;
      top: 0;
      bottom: 0;
      width: 100%;
      height: 100%;
      display: flex;
      background: transparent !important;

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

  import { Configuration } from './api/configuration'
  import { AxiosInstance } from 'axios'

  import { Vue, Component, Ref, Watch, Prop } from 'vue-property-decorator'
  import ResizeObserver from 'resize-observer-polyfill'

  import { NekoApi, MembersApi, RoomApi } from './internal/api'
  import { NekoConnection } from './internal/connection'
  import { NekoMessages } from './internal/messages'
  import { NekoControl } from './internal/control'
  import { register as VideoRegister } from './internal/video'

  import { ReconnectorConfig } from './types/reconnector'
  import NekoState from './types/state'
  import { CursorDrawFunction, InactiveCursorDrawFunction, Dimension } from './types/cursors'
  import Overlay from './overlay.vue'
  import Screencast from './screencast.vue'
  import Cursors from './cursors.vue'

  const SCREEN_SYNC_THROTTLE = 500 // wait 500ms before reacting to automatic screen size change

  @Component({
    name: 'neko-canvas',
    components: {
      'neko-overlay': Overlay,
      'neko-screencast': Screencast,
      'neko-cursors': Cursors,
    },
  })
  export default class extends Vue {
    @Ref('component') readonly _component!: HTMLElement
    @Ref('container') readonly _container!: HTMLElement
    @Ref('video') readonly _video!: HTMLVideoElement
    @Ref('overlay') readonly _overlay!: Overlay

    // fallback image for webrtc reconnections:
    // chrome shows black screen when closing webrtc connection, that's why
    // we need to grab video image before closing connection ans show that
    // while reconnecting, to not see black screen
    fallbackImage = ''

    api = new NekoApi()
    observer = new ResizeObserver(this.onResize.bind(this))
    canvasSize: Dimension = { width: 0, height: 0 }
    cursorDrawFunction: CursorDrawFunction | null = null
    inactiveCursorDrawFunction: InactiveCursorDrawFunction | null = null

    @Prop({ type: String })
    private readonly server!: string

    @Prop({ type: Boolean })
    private readonly autologin!: boolean

    @Prop({ type: Boolean })
    private readonly autoconnect!: boolean

    @Prop({ type: Boolean })
    private readonly autoplay!: boolean

    // fps for cursor rendering, 0 for no cap
    @Prop({ type: Number, default: 0 })
    private readonly fps!: number

    // auto / touch / mouse
    @Prop({ type: String, default: 'auto' })
    private readonly inputMode!: String

    /////////////////////////////
    // Public state
    /////////////////////////////
    public state = {
      authenticated: false,
      connection: {
        url: location.href,
        token: undefined,
        status: 'disconnected',
        websocket: {
          connected: false,
          config: {
            max_reconnects: 15,
            timeout_ms: 5000,
            backoff_ms: 1500,
          },
        },
        webrtc: {
          connected: false,
          stable: false,
          config: {
            max_reconnects: 15,
            timeout_ms: 10000,
            backoff_ms: 1500,
          },
          stats: null,
          video: null,
          bitrate: null,
          videos: [],
        },
        screencast: true, // TODO: Should get by API call.
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
        locked: false,
      },
      screen: {
        size: {
          width: 1280,
          height: 720,
          rate: 30,
        },
        configurations: [],
        sync: {
          enabled: false,
          multiplier: 0,
          rate: 30,
        },
      },
      session_id: null,
      sessions: {},
      settings: {
        private_mode: false,
        locked_controls: false,
        implicit_hosting: false,
        inactive_cursors: false,
        merciful_reconnect: false,
      },
      cursors: [],
      mobile_keyboard_open: false,
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

    public get session() {
      return this.state.session_id != null ? this.state.sessions[this.state.session_id] : null
    }

    public get is_admin() {
      return this.session?.profile.is_admin || false
    }

    public get private_mode_enabled() {
      return this.state.settings.private_mode && !this.is_admin
    }

    public get is_touch_device() {
      if (this.inputMode == 'mouse') return false
      if (this.inputMode == 'touch') return true

      return (
        // check if the device has a touch screen
        ('ontouchstart' in window || navigator.maxTouchPoints > 0) &&
        // we also check if the device has a pointer
        !window.matchMedia('(pointer:fine)').matches &&
        // and is capable of hover, then it probably has a mouse
        !window.matchMedia('(hover:hover)').matches
      )
    }

    @Watch('private_mode_enabled')
    private setWebRTCPaused(paused: boolean) {
      this.connection.webrtc.paused = paused
    }

    screencastReady = false
    public get screencast() {
      return (
        this.state.authenticated &&
        this.state.connection.status != 'disconnected' &&
        this.state.connection.screencast &&
        (!this.state.connection.webrtc.connected ||
          (this.state.connection.webrtc.connected && !this.state.video.playing))
      )
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

      // url can contain ?token=<string>
      let token = new URL(url).searchParams.get('token') || undefined

      // get URL without query params
      url = url.split('?')[0]

      const httpURL = url.replace(/^ws/, 'http').replace(/\/$|\/ws\/?$/, '')
      this.api.setUrl(httpURL)
      Vue.set(this.state.connection, 'url', httpURL)

      try {
        this.disconnect()
      } catch {}

      if (this.state.authenticated) {
        Vue.set(this.state, 'authenticated', false)
      }

      // save token to state
      Vue.set(this.state.connection, 'token', token)

      // try to authenticate and connect
      if (this.autoconnect) {
        try {
          await this.authenticate()
          this.connect()
        } catch {}
      }
    }

    public async authenticate(token?: string) {
      if (!token) {
        token = this.state.connection.token
      }

      if (!token && this.autologin) {
        token = localStorage.getItem('neko_session') ?? undefined
      }

      if (token) {
        this.api.setToken(token)
        Vue.set(this.state.connection, 'token', token)
      }

      await this.api.session.whoami()
      Vue.set(this.state, 'authenticated', true)

      if (token && this.autologin) {
        localStorage.setItem('neko_session', token)
      }
    }

    public async login(username: string, password: string) {
      if (this.state.authenticated) {
        throw new Error('client already authenticated')
      }

      const res = await this.api.session.login({ username, password })
      if (res.data.token) {
        this.api.setToken(res.data.token)
        Vue.set(this.state.connection, 'token', res.data.token)

        if (this.autologin) {
          localStorage.setItem('neko_session', res.data.token)
        }
      }

      Vue.set(this.state, 'authenticated', true)
    }

    public async logout() {
      if (!this.state.authenticated) {
        throw new Error('client not authenticated')
      }

      try {
        this.disconnect()
      } catch {}

      try {
        await this.api.session.logout()
      } finally {
        this.api.setToken('')
        Vue.delete(this.state.connection, 'token')

        if (this.autologin) {
          localStorage.removeItem('neko_session')
        }

        Vue.set(this.state, 'authenticated', false)
      }
    }

    public connect(video?: string) {
      if (!this.state.authenticated) {
        throw new Error('client not authenticated')
      }

      if (this.connected) {
        throw new Error('client is already connected')
      }

      this.connection.open(video)
    }

    public disconnect() {
      this.connection.close()
    }

    public setReconnectorConfig(type: 'websocket' | 'webrtc', config: ReconnectorConfig) {
      if (type != 'websocket' && type != 'webrtc') {
        throw new Error('unknown reconnector type')
      }

      Vue.set(this.state.connection[type], 'config', config)
      this.connection.reloadConfigs()
    }

    public async play() {
      // if autoplay is disabled, play() will throw an error
      // and we need to properly save the state otherwise we
      // would be thinking we're playing when we're not
      try {
        await this._video.play()
      } catch (e: any) {
        if (this._video.muted) {
          throw e
        }

        // video.play() can fail if audio is set due restrictive
        // browsers autoplay policy -> retry with muted audio
        try {
          this._video.muted = true
          await this._video.play()
          // unmute on users first interaction
          document.addEventListener('click', this.autoUnmute, { once: true })
          this.control.once('overlay.click', this.autoUnmute)
        } catch (e: any) {
          // if it still fails, we're not playing anything
          this._video.muted = false
          throw e
        }
      }
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

    // when autoplay fails, we mute the video and wait for the user
    // to interact with the page to unmute it again
    public autoUnmute() {
      this.unmute()

      // remove listeners
      document.removeEventListener('click', this.autoUnmute)
      this.control.removeListener('overlay.click', this.autoUnmute)
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

    public mobileKeyboardShow() {
      this._overlay.mobileKeyboardShow()
    }

    public mobileKeyboardHide() {
      this._overlay.mobileKeyboardHide()
    }

    public mobileKeyboardToggle() {
      if (this.state.mobile_keyboard_open) {
        this.mobileKeyboardHide()
      } else {
        this.mobileKeyboardShow()
      }
    }

    public setScreenSync(enabled: boolean = true, multiplier: number = 0, rate: number = 60) {
      Vue.set(this.state.screen.sync, 'enabled', enabled)
      Vue.set(this.state.screen.sync, 'multiplier', multiplier)
      Vue.set(this.state.screen.sync, 'rate', rate)
    }

    public setCursorDrawFunction(fn?: CursorDrawFunction) {
      Vue.set(this, 'cursorDrawFunction', fn)
    }

    public setInactiveCursorDrawFunction(fn?: InactiveCursorDrawFunction) {
      Vue.set(this, 'inactiveCursorDrawFunction', fn)
    }

    // TODO: Remove? Use REST API only?
    public setScreenSize(width: number, height: number, rate: number) {
      this.connection.websocket.send(EVENT.SCREEN_SET, { width, height, rate })
    }

    public setWebRTCVideo(video: string, bitrate: number = 0) {
      this.connection.setVideo(video, bitrate)
    }

    public addTrack(track: MediaStreamTrack, ...streams: MediaStream[]): RTCRtpSender {
      return this.connection.webrtc.addTrack(track, ...streams)
    }

    public removeTrack(sender: RTCRtpSender) {
      this.connection.webrtc.removeTrack(sender)
    }

    public sendUnicast(receiver: string, subject: string, body: any) {
      this.connection.websocket.send(EVENT.SEND_UNICAST, { receiver, subject, body })
    }

    public sendBroadcast(subject: string, body: any) {
      this.connection.websocket.send(EVENT.SEND_BROADCAST, { subject, body })
    }

    public sendMessage(event: string, payload?: any) {
      this.connection.websocket.send(event, payload)
    }

    public withApi<T>(c: new (configuration?: Configuration, basePath?: string, axios?: AxiosInstance) => T): T {
      return new c(this.api.config)
    }

    public control = new NekoControl(this.connection, this.state.control)

    public get room(): RoomApi {
      return this.api.room
    }

    public get members(): MembersApi {
      return this.api.members
    }

    async uploadDrop({ x, y, files }: { x: number; y: number; files: Array<File> }) {
      try {
        this.events.emit('upload.drop.started')

        await this.api.room.uploadDrop(x, y, files, {
          onUploadProgress: (progressEvent: ProgressEvent) => {
            this.events.emit('upload.drop.progress', progressEvent)
          },
        })

        this.events.emit('upload.drop.finished')
      } catch (err: any) {
        this.events.emit('upload.drop.finished', err)
      }
    }

    /////////////////////////////
    // Component lifecycle
    /////////////////////////////
    mounted() {
      // component size change
      this.observer.observe(this._component)

      // webrtc needs video tag to capture video snaps for fallback mode
      this.connection.webrtc.video = this._video

      // video events
      VideoRegister(this._video, this.state.video)

      this.connection.on('close', (error) => {
        this.events.emit('connection.closed', error)
        this.clear()
      })

      // when webrtc emits fallback event, it means it is about to reconnect
      // so we image that it provided (it is last frame of the video), we set
      // it to the screencast module and pause video in order to show fallback
      this.connection.webrtc.on('fallback', (image: string) => {
        this.fallbackImage = image

        // this ensures that fallback mode starts immediatly
        this.pause()
      })

      this.connection.webrtc.on('track', (event: RTCTrackEvent) => {
        const { track, streams } = event
        if (track.kind === 'audio') return

        // apply track only once it is unmuted
        track.addEventListener(
          'unmute',
          () => {
            // create stream
            if ('srcObject' in this._video) {
              this._video.srcObject = streams[0]
            } else {
              // @ts-ignore
              this._video.src = window.URL.createObjectURL(streams[0]) // for older browsers
            }

            if (this.autoplay || this.connection.activated) {
              this.play()
            }
          },
          { once: true },
        )
      })
    }

    beforeDestroy() {
      this.observer.disconnect()
      this.connection.destroy()
      this.clear()

      // removes users first interaction events
      this.autoUnmute()
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

    @Watch('state.screen.sync.enabled')
    onScreenSyncChange() {
      if (this.state.screen.sync.enabled) {
        this.syncScreenSize()
        window.addEventListener('resize', this.syncScreenSize)
      } else {
        window.removeEventListener('resize', this.syncScreenSize)
      }
    }

    syncScreenSizeTimeout = 0
    public syncScreenSize() {
      if (this.syncScreenSizeTimeout) {
        window.clearTimeout(this.syncScreenSizeTimeout)
      }
      this.syncScreenSizeTimeout = window.setTimeout(() => {
        const multiplier = this.state.screen.sync.multiplier || window.devicePixelRatio
        this.syncScreenSizeTimeout = 0
        const { offsetWidth, offsetHeight } = this._component
        this.setScreenSize(
          Math.round(offsetWidth * multiplier),
          Math.round(offsetHeight * multiplier),
          this.state.screen.sync.rate,
        )
      }, SCREEN_SYNC_THROTTLE)
    }

    @Watch('state.screen.size')
    onResize() {
      const { width, height } = this.state.screen.size
      const screenRatio = width / height

      const { offsetWidth, offsetHeight } = this._component
      const canvasRatio = offsetWidth / offsetHeight

      // vertical centering
      if (screenRatio > canvasRatio) {
        const vertical = offsetWidth / screenRatio
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
      else if (screenRatio < canvasRatio) {
        const horizontal = screenRatio * offsetHeight
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

    @Watch('screencast')
    @Watch('state.connection.webrtc.connected')
    updateConnectionType() {
      if (this.screencast) {
        Vue.set(this.state.connection, 'type', 'fallback')
      } else if (this.state.connection.webrtc.connected) {
        Vue.set(this.state.connection, 'type', 'webrtc')
      } else {
        Vue.set(this.state.connection, 'type', 'none')
      }
    }

    @Watch('state.connection.status')
    onConnectionStatusChange(status: 'connected' | 'connecting' | 'disconnected') {
      this.events.emit('connection.status', status)
    }

    @Watch('state.connection.type')
    onConnectionTypeChange(type: 'fallback' | 'webrtc' | 'none') {
      this.events.emit('connection.type', type)
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
      Vue.set(this.state.screen, 'size', { width: 1280, height: 720, rate: 30 })
      Vue.set(this.state.screen, 'configurations', [])
      Vue.set(this.state.screen, 'sync', false)
      Vue.set(this.state, 'session_id', null)
      Vue.set(this.state, 'sessions', {})
      Vue.set(this.state, 'settings', {
        private_mode: false,
        implicit_hosting: false,
        inactive_cursors: false,
        merciful_reconnect: false,
      })
      Vue.set(this.state, 'cursors', [])

      // webrtc
      Vue.set(this.state.connection.webrtc, 'stats', null)
      Vue.set(this.state.connection.webrtc, 'video', null)
      Vue.set(this.state.connection, 'type', 'none')
    }
  }
</script>
