<template>
  <div class="video-player">
    <div ref="video" class="video">
      <div ref="container" class="video-container">
        <video
          ref="player"
          tabindex="0"
          @click.stop.prevent
          @contextmenu.stop.prevent
          @wheel.stop.prevent="onWheel"
          @mousemove.stop.prevent="onMouseMove"
          @mousedown.stop.prevent="onMouseDown"
          @mouseup.stop.prevent="onMouseUp"
          @mouseenter.stop.prevent="onMouseEnter"
          @mouseleave.stop.prevent="onMouseLeave"
          @keydown.stop.prevent="onKeyDown"
          @keyup.stop.prevent="onKeyUp"
        />
        <div v-if="!playing" class="video-overlay">
          <i @click.stop.prevent="toggleMedia" class="fas fa-play-circle" />
        </div>
        <div ref="aspect" class="aspect" />
      </div>
    </div>
    <div class="controls">
      <div class="neko">
        <img src="@/assets/logo.svg" alt="n.eko" />
        <span><b>n</b>.eko</span>
      </div>
      <ul>
        <li>
          <i
            alt="Request Control"
            :class="[{ enabled: controlling }, 'request', 'fas', 'fa-keyboard']"
            @click.stop.prevent="toggleControl"
          />
        </li>
        <li>
          <i
            alt="Play/Pause"
            :class="[playing ? 'fa-pause-circle' : 'fa-play-circle', 'play', 'fas']"
            @click.stop.prevent="toggleMedia"
          />
        </li>
        <li>
          <div class="volume">
            <input
              @input="setVolume"
              :class="[volume === 0 ? 'fa-volume-mute' : 'fa-volume-up', 'fas']"
              ref="volume"
              type="range"
              min="0"
              max="100"
            />
          </div>
        </li>
        <li>
          <i @click.stop.prevent="fullscreen" alt="Full Screen" class="fullscreen fas fa-expand-alt" />
        </li>
      </ul>
      <div class="right"></div>
    </div>
    <div class="connect" v-if="!connected">
      <div class="window">
        <div class="logo">
          <img src="@/assets/logo.svg" alt="n.eko" />
          <span><b>n</b>.eko</span>
        </div>
        <div class="message" v-if="!connecting">
          <span>Please enter the password:</span>
          <input type="password" v-model="password" />
          <span class="button" @click.stop.prevent="connect">
            Connect
          </span>
        </div>
        <div class="spinner" v-if="connecting">
          <div class="double-bounce1"></div>
          <div class="double-bounce2"></div>
        </div>
        <div class="loader" v-if="connecting" />
      </div>
    </div>
    <notifications group="neko" position="bottom left" />
  </div>
</template>

<style lang="scss" scoped>
  .video-player {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;

    .video {
      position: absolute;
      top: 60px;
      left: 0;
      right: 0;
      bottom: 0;

      display: flex;
      justify-content: center;
      align-items: center;

      .video-container {
        position: relative;
        width: 100%;
        max-width: 16 / 9 * 100vh;

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

        .video-overlay {
          position: absolute;
          top: 0;
          bottom: 0;
          width: 100%;
          height: 100%;
          background: rgba($color: $style-darker, $alpha: 0.2);
          display: flex;
          justify-content: center;
          align-items: center;

          i {
            cursor: pointer;
            &::before {
              font-size: 120px;
              color: rgba($color: $style-light, $alpha: 0.4);
              text-align: center;
            }
          }

          &.hidden {
            display: none;
          }
        }

        .aspect {
          display: block;
          padding-bottom: 56.25%;
        }
      }
    }

    .controls {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 60px;
      background: $style-darker;
      padding: 0 50px;
      display: flex;

      .neko {
        flex: 1; /* shorthand for: flex-grow: 1, flex-shrink: 1, flex-basis: 0 */
        display: flex;
        justify-content: flex-start;
        align-items: center;
        width: 150px;

        img {
          display: block;
          float: left;
          height: 54px;
          margin-right: 10px;
        }

        span {
          color: $style-light;
          font-size: 30px;
          line-height: 56px;

          b {
            font-weight: 900;
          }
        }
      }

      ul {
        flex: 1;
        display: flex;
        flex-direction: row;
        justify-content: center;
        align-items: center;
        list-style: none;

        li {
          padding: 0 10px;
          color: $style-light;
          font-size: 20px;
          cursor: pointer;

          .request {
            color: rgba($color: $style-light, $alpha: 0.5);

            &.enabled {
              color: $style-light;
            }
          }

          .volume {
            display: block;
            margin-top: 3px;

            input[type='range'] {
              -webkit-appearance: none;
              width: 100%;
              background: transparent;
              width: 200px;
              height: 20px;

              &::-webkit-slider-thumb {
                -webkit-appearance: none;
                height: 12px;
                width: 12px;
                border-radius: 12px;
                background: $style-light;
                cursor: pointer;
                margin-top: -4px;
              }

              &::-webkit-slider-runnable-track {
                width: 100%;
                height: 4px;
                cursor: pointer;
                background: $style-primary;
                border-radius: 2px;
                margin-bottom: 2px;
              }

              &::before {
                color: $style-light;
                text-align: center;
                margin-right: 5px;
              }
            }
          }
        }
      }

      .right {
        flex: 1;
        display: flex;
        justify-content: flex-end;
        align-items: center;
      }
    }

    .connect {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba($color: $style-darker, $alpha: 0.8);

      display: flex;
      justify-content: center;
      align-items: center;

      .window {
        width: 300px;
        background: $style-light;
        border-radius: 5px;
        padding: 10px;

        .logo {
          color: $style-darker;
          width: 100%;
          display: flex;
          flex-direction: row;
          justify-content: center;
          align-items: center;

          img {
            filter: invert(100%);
            height: 90px;
            margin-right: 10px;
          }

          span {
            font-size: 30px;
            line-height: 56px;

            b {
              font-weight: 900;
            }
          }
        }

        .message {
          display: flex;
          flex-direction: column;

          span {
            text-align: center;
            text-transform: uppercase;
            margin: 5px 0;
          }

          input {
            border: solid 1px rgba($color: $style-darker, $alpha: 0.4);
            padding: 3px;
            line-height: 20px;
            border-radius: 5px;
            margin: 5px 0;
          }

          .button {
            cursor: pointer;
            border-radius: 5px;
            padding: 4px;
            background: $style-primary;
            color: $style-light;
            text-align: center;
            text-transform: uppercase;
            font-weight: bold;
            line-height: 30px;
            margin: 5px 0;
          }
        }

        .spinner {
          width: 90px;
          height: 90px;
          position: relative;
          margin: 0 auto;

          .double-bounce1,
          .double-bounce2 {
            width: 100%;
            height: 100%;
            border-radius: 50%;
            background-color: $style-primary;
            opacity: 0.6;
            position: absolute;
            top: 0;
            left: 0;

            -webkit-animation: sk-bounce 2s infinite ease-in-out;
            animation: sk-bounce 2s infinite ease-in-out;
          }

          .double-bounce2 {
            -webkit-animation-delay: -1s;
            animation-delay: -1s;
          }
        }
      }
    }
  }

  @keyframes sk-bounce {
    0%,
    100% {
      transform: scale(0);
      -webkit-transform: scale(0);
    }
    50% {
      transform: scale(1);
      -webkit-transform: scale(1);
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'

  const MOUSE_MOVE = 0x01
  const MOUSE_UP = 0x02
  const MOUSE_DOWN = 0x03
  const MOUSE_CLK = 0x04
  const KEY_DOWN = 0x05
  const KEY_UP = 0x06
  const KEY_CLK = 0x07

  @Component({ name: 'stream-video' })
  export default class extends Vue {
    @Ref('player') readonly _player!: HTMLVideoElement
    @Ref('container') readonly _container!: HTMLElement
    @Ref('aspect') readonly _aspect!: HTMLElement
    @Ref('video') readonly _video!: HTMLElement
    @Ref('volume') readonly _volume!: HTMLInputElement

    private focused = false
    private connected = false
    private connecting = false
    private controlling = false
    private playing = false
    private volume = 0
    private width = 1280
    private height = 720
    private state: RTCIceConnectionState = 'disconnected'
    private password = ''

    private ws?: WebSocket
    private peer?: RTCPeerConnection
    private channel?: RTCDataChannel
    private id?: string
    private stream?: MediaStream
    private timeout?: number

    @Watch('volume')
    onVolumeChanged(volume: number) {
      if (this._player) {
        this._player.volume = this.volume / 100
      }
    }

    mounted() {
      window.addEventListener('resize', this.onResise)
      this.onResise()
      this.volume = this._player.volume * 100
      this._volume.value = `${this.volume}`
    }

    beforeDestroy() {
      window.removeEventListener('resize', this.onResise)
    }

    toggleControl() {
      if (!this.ws) {
        return
      }
      if (this.controlling) {
        this.ws.send(JSON.stringify({ event: 'control/release' }))
      } else {
        this.ws.send(JSON.stringify({ event: 'control/request' }))
      }
    }

    toggleMedia() {
      if (!this.playing) {
        this._player
          .play()
          .then(() => {
            this.playing = true
            this.width = this._player.videoWidth
            this.height = this._player.videoHeight
            this.onResise()
          })
          .catch(err => {
            console.error(err)
          })
      } else {
        this._player.pause()
        this.playing = false
      }
    }

    setVolume() {
      this.volume = parseInt(this._volume.value)
    }

    fullscreen() {
      this._video.requestFullscreen()
    }

    connect() {
      this.ws = new WebSocket(
        process.env.NODE_ENV === 'development' ?  `ws://${process.env.NEKO_DEV}/ws?password=${this.password}` : `${/https/gi.test(location.protocol) ? 'wss' : 'ws'}://${location.host}/ws?password=${this.password}` ,
      )

      this.ws.onmessage = this.onMessage.bind(this)
      this.ws.onerror = event => console.error((event as ErrorEvent).error)
      this.ws.onclose = event => this.onClose.bind(this)
      this.onConnecting()
      this.timeout = setTimeout(this.onTimeout.bind(this), 5000)
    }

    createPeer() {
      if (!this.ws) {
        return
      }

      this.peer = new RTCPeerConnection({ iceServers: [{ urls: 'stun:stun.l.google.com:19302' }] })
      this.peer.onicecandidate = event => {
        if (event.candidate === null && this.peer!.localDescription) {
          this.ws!.send(
            JSON.stringify({
              event: 'sdp/provide',
              sdp: this.peer!.localDescription.sdp,
            }),
          )
        }
      }

      this.peer.oniceconnectionstatechange = event => {
        this.state = this.peer!.iceConnectionState

        switch (this.state) {
          case 'connected':
            this.onConnected()
            break
          case 'disconnected':
            this.onClose()
            break
        }
      }
      this.peer.ontrack = this.onTrack.bind(this)
      this.peer.addTransceiver('audio', { direction: 'recvonly' })
      this.peer.addTransceiver('video', { direction: 'recvonly' })

      this.channel = this.peer.createDataChannel('data')

      this.peer
        .createOffer()
        .then(d => this.peer!.setLocalDescription(d))
        .catch(err => console.log(err))
    }

    updateControles(event: 'wheel', data: { x: number; y: number }): void
    updateControles(event: 'mousemove', data: { x: number; y: number; rect: DOMRect }): void
    updateControles(event: 'mousedown' | 'mouseup' | 'keydown' | 'keyup', data: { key: number }): void
    updateControles(event: string, data: any) {
      if (!this.controlling) {
        return
      }

      let buffer: ArrayBuffer
      let payload: DataView
      switch (event) {
        case 'mousemove':
          buffer = new ArrayBuffer(7)
          payload = new DataView(buffer)
          payload.setUint8(0, MOUSE_MOVE)
          payload.setUint16(1, 4, true)
          payload.setUint16(3, Math.round((this.width / data.rect.width) * (data.x - data.rect.left)), true)
          payload.setUint16(5, Math.round((this.height / data.rect.height) * (data.y - data.rect.top)), true)
          break
        case 'wheel':
          buffer = new ArrayBuffer(4)
          payload = new DataView(buffer)
          payload.setUint8(0, MOUSE_CLK)
          payload.setUint16(1, 1, true)

          const ydir = Math.sign(data.y)
          const xdir = Math.sign(data.x)

          if ((!xdir && !ydir) || (xdir && ydir)) return
          if (ydir && ydir < 0) payload.setUint8(3, 4)
          if (ydir && ydir > 0) payload.setUint8(3, 5)
          if (xdir && xdir < 0) payload.setUint8(3, 6)
          if (xdir && xdir > 0) payload.setUint8(3, 7)
          break
        case 'mousedown':
          buffer = new ArrayBuffer(4)
          payload = new DataView(buffer)
          payload.setUint8(0, MOUSE_DOWN)
          payload.setUint16(1, 1, true)
          payload.setUint8(3, data.key)
          break
        case 'mouseup':
          buffer = new ArrayBuffer(4)
          payload = new DataView(buffer)
          payload.setUint8(0, MOUSE_UP)
          payload.setUint16(1, 1, true)
          payload.setUint8(3, data.key)
          break
        case 'keydown':
          buffer = new ArrayBuffer(5)
          payload = new DataView(buffer)
          payload.setUint8(0, KEY_DOWN)
          payload.setUint16(1, 2, true)
          payload.setUint16(3, data.key, true)
          break
        case 'keyup':
          buffer = new ArrayBuffer(5)
          payload = new DataView(buffer)
          payload.setUint8(0, KEY_UP)
          payload.setUint16(1, 2, true)
          payload.setUint16(3, data.key, true)
          break
      }

      // @ts-ignore
      if (this.channel && typeof buffer !== 'undefined') {
        this.channel.send(buffer)
      }
    }

    getAspect() {
      const { width, height } = this

      if ((height == 0 && width == 0) || (height == 0 && width != 0) || (height != 0 && width == 0)) {
        return null
      }

      if (height == width) {
        return {
          horizontal: 1,
          vertical: 1,
        }
      }

      let dividend = width
      let divisor = height
      let gcd = -1

      if (height > width) {
        dividend = height
        divisor = width
      }

      while (gcd == -1) {
        const remainder = dividend % divisor
        if (remainder == 0) {
          gcd = divisor
        } else {
          dividend = divisor
          divisor = remainder
        }
      }

      return {
        horizontal: width / gcd,
        vertical: height / gcd,
      }
    }

    onMousePos(e: MouseEvent) {
      const rect = this._player.getBoundingClientRect() as DOMRect
      this.updateControles('mousemove', {
        x: e.clientX,
        y: e.clientY,
        rect,
      })
    }

    onWheel(e: WheelEvent) {
      this.onMousePos(e)
      this.updateControles('wheel', { x: e.deltaX, y: e.deltaY })
    }

    onMouseDown(e: MouseEvent) {
      this.onMousePos(e)
      this.updateControles('mousedown', { key: e.button })
    }

    onMouseUp(e: MouseEvent) {
      this.onMousePos(e)
      this.updateControles('mouseup', { key: e.button })
    }

    onMouseMove(e: MouseEvent) {
      this.onMousePos(e)
    }

    onMouseEnter(e: MouseEvent) {
      this._player.focus()
      this.focused = true
    }

    onMouseLeave(e: MouseEvent) {
      this.focused = false
    }

    onKeyDown(e: KeyboardEvent) {
      if (!this.focused) {
        return
      }
      this.updateControles('keydown', { key: e.keyCode })
    }

    onKeyUp(e: KeyboardEvent) {
      if (!this.focused) {
        return
      }
      this.updateControles('keyup', { key: e.keyCode })
    }

    onResise() {
      const aspect = this.getAspect()
      if (!aspect) {
        return
      }
      const { horizontal, vertical } = aspect
      this._container.style.maxWidth = `${(horizontal / vertical) * this._video.offsetHeight}px`
      this._aspect.style.paddingBottom = `${(vertical / horizontal) * 100}%`
    }

    onMessage(e: MessageEvent) {
      const { event, ...payload } = JSON.parse(e.data)

      switch (event) {
        case 'sdp/reply':
          if (!this.peer) {
            return
          }
          this.peer.setRemoteDescription(new RTCSessionDescription({ type: 'answer', sdp: payload.sdp }))
          break
        case 'identity/provide':
          this.id = payload.id
          this.createPeer()
          break
        case 'control/requesting':
          this.controlling = true
          this.$notify({
            group: 'neko',
            type: 'info',
            title: 'Another user is requesting the controls',
            duration: 3000,
            speed: 1000,
          })
          break
        case 'control/give':
          this.controlling = true
          this.$notify({
            group: 'neko',
            type: 'info',
            title: 'You have the controls',
            duration: 5000,
            speed: 1000,
          })
          break
        case 'control/locked':
          this.controlling = false
          this.$notify({
            group: 'neko',
            type: 'info',
            title: 'Another user has the controls',
            duration: 3000,
            speed: 1000,
          })
          break
        case 'control/given':
          this.controlling = false
          this.$notify({
            group: 'neko',
            type: 'info',
            title: 'Someone has taken the controls',
            duration: 5000,
            speed: 1000,
          })
          break
        case 'control/release':
          this.controlling = false
          this.$notify({
            group: 'neko',
            type: 'info',
            title: 'You released the controls',
            duration: 5000,
            speed: 1000,
          })
          break
        case 'control/released':
          this.controlling = false
          this.$notify({
            group: 'neko',
            type: 'info',
            title: 'The controls have been released',
            duration: 5000,
            speed: 1000,
          })
          break
        default:
          console.warn(`[NEKO] Unknown message event ${event}`)
      }
    }

    onTrack(event: RTCTrackEvent) {
      if (event.track.kind === 'audio') {
        return
      }

      this.stream = event.streams[0]
      if (!this.stream) {
        return
      }

      if ('srcObject' in this._player) {
        this._player.srcObject = this.stream
      } else {
        // @ts-ignore
        this._player.src = window.URL.createObjectURL(this.stream) // for older browsers
      }

      if (this._player.paused) {
        this.toggleMedia()
      }
    }

    onTimeout() {
      this.connected = false
      this.connecting = false
      this.$notify({
        group: 'neko',
        type: 'error',
        title: 'Unable to connect to server!',
        duration: 5000,
        speed: 1000,
      })
    }

    onConnecting() {
      this.connecting = true
    }

    onConnected() {
      this.connected = true
      this.connecting = false
      this.$notify({
        group: 'neko',
        type: 'success',
        title: 'Successfully connected!',
        duration: 5000,
        speed: 1000,
      })
      if (this.timeout) {
        clearTimeout(this.timeout)
      }
    }

    onClose() {
      this.controlling = false
      this.connected = false
      this.connecting = false
      this.ws = undefined
      this.peer = undefined
      if (this.playing) {
        this.toggleMedia()
      }
      this.$notify({
        group: 'neko',
        type: 'error',
        title: 'Disconnected from server!',
        duration: 5000,
        speed: 1000,
      })
    }
  }
</script>
