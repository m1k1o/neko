<template>
  <div ref="component" class="component">
    <div ref="container" class="player-container">
      <video ref="video" :autoplay="autoplay" :muted="autoplay" />
      <neko-overlay
        :webrtc="webrtc"
        :control="state.control"
        :screenWidth="state.screen.size.width"
        :screenHeight="state.screen.size.height"
        :isControling="controlling && watching"
        :implicitControl="state.control.implicit_hosting && state.members[state.member_id].profile.can_host"
        @implicit-control-request="websocket.send('control/request')"
        @implicit-control-release="websocket.send('control/release')"
        @drop-files="uploadDrop($event)"
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
  export * as ApiModels from './api/models'
  export * as StateModels from './types/state'

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

    @Prop({ type: Boolean })
    private readonly autoplay!: boolean

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
        },
        type: 'none',
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
        cursor: null,
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
    public setUrl(url: string) {
      const httpURL = url.replace(/^ws/, 'http').replace(/\/$|\/ws\/?$/, '')
      this.api.setUrl(httpURL)
      this.websocket.setUrl(httpURL)
    }

    public async login(id: string, secret: string) {
      if (this.authenticated) {
        throw new Error('client already authenticated')
      }

      await this.api.session.login({ id, secret })
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
      if (!this.authenticated) {
        throw new Error('client not authenticated')
      }

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

      this.websocket.send('signal/request')
    }

    public webrtcDisconnect() {
      if (!this.connected) {
        throw new Error('client not connected to websocket')
      }

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

    public setScreenSize(width: number, height: number, rate: number) {
      //this.api.room.screenConfigurationChange({ screenConfiguration: { width, height, rate } })
      this.websocket.send('screen/set', { width, height, rate })
    }

    public setWebRTCVideo(video: string) {
      if (!this.state.connection.webrtc.videos.includes(video)) {
        throw new Error('VideoID not found.')
      }

      this.websocket.send('signal/video', { video: video })
    }

    public sendUnicast(receiver: string, subject: string, body: any) {
      this.websocket.send('send/unicast', { receiver, subject, body })
    }

    public sendBroadcast(subject: string, body: any) {
      this.websocket.send('send/broadcast', { subject, body })
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
          case 'signal/candidate':
            this.webrtc.setCandidate(payload)
            break
          case 'system/disconnect':
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
        this.webrtcConnect()
      })
      this.websocket.on('disconnected', () => {
        Vue.set(this.state.connection, 'websocket', 'disconnected')
        this.events.emit('connection.websocket', 'disconnected')

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

        if (this.autoplay) {
          this._video.play()
        }
      })
      this.webrtc.on('candidate', (candidate: RTCIceCandidateInit) => {
        this.websocket.send('signal/candidate', candidate)
      })
      this.webrtc.on('stats', (stats: WebRTCStats) => {
        Vue.set(this.state.connection.webrtc, 'stats', stats)
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
      this.webrtc.on('disconnected', () => {
        Vue.set(this.state.connection.webrtc, 'status', 'disconnected')
        Vue.set(this.state.connection.webrtc, 'stats', null)
        Vue.set(this.state.connection.webrtc, 'video', null)
        Vue.set(this.state.connection.webrtc, 'videos', [])
        Vue.set(this.state.connection, 'type', 'none')
        this.events.emit('connection.webrtc', 'disconnected')

        if (!this._video) return

        // destroy stream
        if ('srcObject' in this._video) {
          this._video.srcObject = null
        } else {
          // @ts-ignore
          this._video.removeAttribute('src')
        }

        // reconnect WebRTC
        if (this.connected) {
          setTimeout(() => {
            try {
              this.webrtcConnect()
            } catch (e) {}
          }, 1000)
        }
      })

      // check if is user logged in
      this.api.session.whoami().then(() => {
        Vue.set(this.state.connection, 'authenticated', true)
        this.websocket.connect()
      })

      // unmute on users first interaction
      if (this.autoplay) {
        document.addEventListener('click', () => this.unmute(), { once: true })
      }
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
      Vue.set(this.state.control, 'cursor', null)
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
