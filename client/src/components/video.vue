<template>
  <div ref="component" class="video">
    <div ref="player" class="player">
      <div ref="container" class="player-container">
        <video ref="video" playsinline />
        <div class="emotes">
          <template v-for="(emote, index) in emotes">
            <neko-emote :id="index" :key="index" />
          </template>
        </div>
        <div
          ref="overlay"
          class="overlay"
          tabindex="0"
          @click.stop.prevent
          @contextmenu.stop.prevent
          @wheel.stop.prevent="onWheel"
          @mousemove.stop.prevent="onMouseMove"
          @mousedown.stop.prevent="onMouseDown"
          @mouseup.stop.prevent="onMouseUp"
          @mouseenter.stop.prevent="onMouseEnter"
          @mouseleave.stop.prevent="onMouseLeave"
        />
        <div v-if="!playing" class="player-overlay">
          <i @click.stop.prevent="toggle" v-if="playable" class="fas fa-play-circle" />
        </div>
        <div ref="aspect" class="player-aspect" />
      </div>
      <ul v-if="!fullscreen && !hideControls" class="video-menu top">
        <li><i @click.stop.prevent="requestFullscreen" class="fas fa-expand"></i></li>
        <li v-if="admin"><i @click.stop.prevent="onResolution" class="fas fa-desktop"></i></li>
        <li class="request-control">
          <i
            :class="[hosted && !hosting ? 'disabled' : '', !hosted && !hosting ? 'faded' : '', 'fas', 'fa-keyboard']"
            @click.stop.prevent="toggleControl"
          />
        </li>
      </ul>
      <ul v-if="!fullscreen && !hideControls" class="video-menu bottom">
        <li v-if="hosting && (!clipboard_read_available || !clipboard_write_available)">
          <i @click.stop.prevent="onClipboard" class="fas fa-clipboard"></i>
        </li>
        <li>
          <i
            v-if="pip_available"
            @click.stop.prevent="requestPictureInPicture"
            v-tooltip="{ content: 'Picture-in-Picture', placement: 'left', offset: 5, boundariesElement: 'body' }"
            class="fas fa-external-link-alt"
          />
        </li>
      </ul>
      <neko-resolution ref="resolution" v-if="admin" />
      <neko-clipboard ref="clipboard" v-if="hosting && (!clipboard_read_available || !clipboard_write_available)" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .video {
    width: 100%;
    height: 100%;

    .player {
      position: absolute;
      display: flex;
      justify-content: center;
      align-items: center;

      .video-menu {
        position: absolute;
        right: 20px;

        &.top {
          top: 15px;
        }

        &.bottom {
          bottom: 15px;
        }

        li {
          margin: 0 0 10px 0;

          i {
            width: 30px;
            height: 30px;
            background: rgba($color: #fff, $alpha: 0.2);
            border-radius: 5px;
            line-height: 30px;
            font-size: 16px;
            text-align: center;
            color: rgba($color: #fff, $alpha: 0.6);
            cursor: pointer;

            &.faded {
              color: rgba($color: $text-normal, $alpha: 0.4);
            }

            &.disabled {
              color: rgba($color: $style-error, $alpha: 0.4);
            }
          }

          &.request-control {
            display: none;
          }

          @media (max-width: 768px) {
            &.request-control {
              display: inline-block;
            }
          }

          &:last-child {
            margin: 0;
          }
        }
      }

      .player-container {
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

        .player-overlay,
        .emotes {
          position: absolute;
          top: 0;
          bottom: 0;
          width: 100%;
          height: 100%;
          overflow: hidden;
        }

        .player-overlay {
          background: rgba($color: #000, $alpha: 0.2);
          display: flex;
          justify-content: center;
          align-items: center;

          i {
            cursor: pointer;
            &::before {
              font-size: 120px;
              text-align: center;
            }
          }

          &.hidden {
            display: none;
          }
        }

        .overlay {
          position: absolute;
          top: 0;
          bottom: 0;
          width: 100%;
          height: 100%;
        }

        .player-aspect {
          display: block;
          padding-bottom: 56.25%;
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue, Prop } from 'vue-property-decorator'
  import ResizeObserver from 'resize-observer-polyfill'

  import Emote from './emote.vue'
  import Resolution from './resolution.vue'
  import Clipboard from './clipboard.vue'

  // @ts-ignore
  import GuacamoleKeyboard from '~/utils/guacamole-keyboard.ts'

  const WHEEL_LINE_HEIGHT = 19

  @Component({
    name: 'neko-video',
    components: {
      'neko-emote': Emote,
      'neko-resolution': Resolution,
      'neko-clipboard': Clipboard,
    },
  })
  export default class extends Vue {
    @Ref('component') readonly _component!: HTMLElement
    @Ref('container') readonly _container!: HTMLElement
    @Ref('overlay') readonly _overlay!: HTMLElement
    @Ref('aspect') readonly _aspect!: HTMLElement
    @Ref('player') readonly _player!: HTMLElement
    @Ref('video') readonly _video!: HTMLVideoElement
    @Ref('resolution') readonly _resolution!: any
    @Ref('clipboard') readonly _clipboard!: any

    @Prop(Boolean) readonly hideControls = false

    private keyboard = GuacamoleKeyboard()
    private observer = new ResizeObserver(this.onResise.bind(this))
    private focused = false
    private fullscreen = false
    private startsMuted = true

    get admin() {
      return this.$accessor.user.admin
    }

    get connected() {
      return this.$accessor.connected
    }

    get connecting() {
      return this.$accessor.connecting
    }

    get hosting() {
      return this.$accessor.remote.hosting
    }

    get hosted() {
      return this.$accessor.remote.hosted
    }

    get volume() {
      return this.$accessor.video.volume
    }

    get muted() {
      return this.$accessor.video.muted
    }

    get stream() {
      return this.$accessor.video.stream
    }

    get playing() {
      return this.$accessor.video.playing
    }

    get playable() {
      return this.$accessor.video.playable
    }

    get emotes() {
      return this.$accessor.chat.emotes
    }

    get autoplay() {
      return this.$accessor.settings.autoplay
    }

    get locked() {
      return this.$accessor.remote.locked
    }

    get scroll() {
      return this.$accessor.settings.scroll
    }

    get scroll_invert() {
      return this.$accessor.settings.scroll_invert
    }

    get pip_available() {
      //@ts-ignore
      return typeof document.createElement('video').requestPictureInPicture === 'function'
    }

    get clipboard_read_available() {
      return 'clipboard' in navigator && typeof navigator.clipboard.readText === 'function'
    }

    get clipboard_write_available() {
      return 'clipboard' in navigator && typeof navigator.clipboard.writeText === 'function'
    }

    get clipboard() {
      return this.$accessor.remote.clipboard
    }

    get width() {
      return this.$accessor.video.width
    }

    get height() {
      return this.$accessor.video.height
    }

    get rate() {
      return this.$accessor.video.rate
    }

    get vertical() {
      return this.$accessor.video.vertical
    }

    get horizontal() {
      return this.$accessor.video.horizontal
    }

    @Watch('width')
    onWidthChanged(width: number) {
      this.onResise()
    }

    @Watch('height')
    onHeightChanged(height: number) {
      this.onResise()
    }

    @Watch('volume')
    onVolumeChanged(volume: number) {
      if (this._video) {
        this._video.volume = this.volume / 100
      }
    }

    @Watch('muted')
    onMutedChanged(muted: boolean) {
      if (this._video) {
        this._video.muted = muted
        this.startsMuted = muted
      }
    }

    @Watch('stream')
    onStreamChanged(stream?: MediaStream) {
      if (!this._video || !stream) {
        return
      }

      if ('srcObject' in this._video) {
        this._video.srcObject = stream
      } else {
        // @ts-ignore
        this._video.src = window.URL.createObjectURL(this.stream) // for older browsers
      }
    }

    @Watch('playing')
    onPlayingChanged(playing: boolean) {
      if (playing) {
        this.play()
      } else {
        this.pause()
      }
    }

    @Watch('clipboard')
    onClipboardChanged(clipboard: string) {
      if (this.clipboard_write_available) {
        navigator.clipboard.writeText(clipboard).catch(console.error)
      }
    }

    mounted() {
      this._container.addEventListener('resize', this.onResise)
      this.onVolumeChanged(this.volume)
      this.onStreamChanged(this.stream)
      this.onResise()

      this.observer.observe(this._component)

      this._player.addEventListener('fullscreenchange', () => {
        this.fullscreen = document.fullscreenElement !== null
        this.onResise()
      })

      this._video.addEventListener('canplaythrough', () => {
        this.$accessor.video.setPlayable(true)
        if (this.autoplay) {
          if (this.startsMuted && (!document.hasFocus() || !this.$accessor.active)) {
            this.$accessor.video.setMuted(true)
            this._video.muted = true
          }

          this.$nextTick(() => {
            this.$accessor.video.play()
          })
        }
      })

      this._video.addEventListener('ended', () => {
        this.$accessor.video.setPlayable(false)
      })

      this._video.addEventListener('error', (event) => {
        this.$log.error(event.error)
        this.$accessor.video.setPlayable(false)
      })

      document.addEventListener('focusin', this.onFocus.bind(this))

      /* Initialize Guacamole Keyboard */
      this.keyboard.onkeydown = (key: number) => {
        if (!this.focused || !this.hosting || this.locked) {
          return true
        }

        this.$client.sendData('keydown', { key: this.keyMap(key) })
        return false
      }
      this.keyboard.onkeyup = (key: number) => {
        if (!this.focused || !this.hosting || this.locked) {
          return
        }

        this.$client.sendData('keyup', { key: this.keyMap(key) })
      }
      this.keyboard.listenTo(this._overlay)
    }

    beforeDestroy() {
      this.observer.disconnect()
      this.$accessor.video.setPlayable(false)
      document.removeEventListener('focusin', this.onFocus.bind(this))
      /* Guacamole Keyboard does not provide destroy functions */
    }

    get hasMacOSKbd() {
      return /(Mac|iPhone|iPod|iPad)/i.test(navigator.platform)
    }

    KeyTable = {
      XK_ISO_Level3_Shift: 0xfe03, // AltGr
      XK_Mode_switch: 0xff7e, // Character set switch
      XK_Control_L: 0xffe3, // Left control
      XK_Control_R: 0xffe4, // Right control
      XK_Meta_L: 0xffe7, // Left meta
      XK_Meta_R: 0xffe8, // Right meta
      XK_Alt_L: 0xffe9, // Left alt
      XK_Alt_R: 0xffea, // Right alt
      XK_Super_L: 0xffeb, // Left super
      XK_Super_R: 0xffec, // Right super
    }

    keyMap(key: number): number {
      // Alt behaves more like AltGraph on macOS, so shuffle the
      // keys around a bit to make things more sane for the remote
      // server. This method is used by noVNC, RealVNC and TigerVNC
      // (and possibly others).
      if (this.hasMacOSKbd) {
        switch (key) {
          case this.KeyTable.XK_Meta_L:
            key = this.KeyTable.XK_Control_L
            break
          case this.KeyTable.XK_Super_L:
            key = this.KeyTable.XK_Alt_L
            break
          case this.KeyTable.XK_Super_R:
            key = this.KeyTable.XK_Super_L
            break
          case this.KeyTable.XK_Alt_L:
            key = this.KeyTable.XK_Mode_switch
            break
          case this.KeyTable.XK_Alt_R:
            key = this.KeyTable.XK_ISO_Level3_Shift
            break
        }
      }

      return key
    }

    play() {
      if (!this._video.paused || !this.playable) {
        return
      }

      try {
        this._video
          .play()
          .then(() => {
            this.onResise()
          })
          .catch((err) => this.$log.error)
      } catch (err) {
        this.$log.error(err)
      }
    }

    pause() {
      if (this._video.paused || !this.playable) {
        return
      }

      this._video.pause()
    }

    toggle() {
      if (!this.playable) {
        return
      }

      if (!this.playing) {
        this.$accessor.video.play()
      } else {
        this.$accessor.video.pause()
      }
    }

    toggleControl() {
      if (!this.playable) {
        return
      }

      this.$accessor.remote.toggle()
    }

    _elementRequestFullscreen(el: HTMLElement) {
      if (typeof el.requestFullscreen === 'function') {
        el.requestFullscreen()
        //@ts-ignore
      } else if (typeof el.webkitRequestFullscreen === 'function') {
        //@ts-ignore
        el.webkitRequestFullscreen()
        //@ts-ignore
      } else if (typeof el.webkitEnterFullscreen === 'function') {
        //@ts-ignore
        el.webkitEnterFullscreen()
        //@ts-ignore
      } else if (typeof el.msRequestFullScreen === 'function') {
        //@ts-ignore
        el.msRequestFullScreen()
      } else {
        return false
      }

      return true
    }

    requestFullscreen() {
      // try to fullscreen player element
      if (this._elementRequestFullscreen(this._player)) {
        this.onResise()
        return
      }

      // fallback to fullscreen video itself (on mobile devices)
      if (this._elementRequestFullscreen(this._video)) {
        this.onResise()
        return
      }
    }

    requestPictureInPicture() {
      //@ts-ignore
      this._video.requestPictureInPicture()
      this.onResise()
    }

    onFocus() {
      if (!document.hasFocus() || !this.$accessor.active) {
        return
      }

      if (this.hosting && this.clipboard_read_available) {
        navigator.clipboard
          .readText()
          .then((text) => {
            if (this.clipboard !== text) {
              this.$accessor.remote.setClipboard(text)
              this.$accessor.remote.sendClipboard(text)
            }
          })
          .catch(this.$log.error)
      }
    }

    onMousePos(e: MouseEvent) {
      const { w, h } = this.$accessor.video.resolution
      const rect = this._overlay.getBoundingClientRect()
      this.$client.sendData('mousemove', {
        x: Math.round((w / rect.width) * (e.clientX - rect.left)),
        y: Math.round((h / rect.height) * (e.clientY - rect.top)),
      })
    }

    wheelThrottle = false
    onWheel(e: WheelEvent) {
      if (!this.hosting || this.locked) {
        return
      }
      this.onMousePos(e)

      let x = e.deltaX
      let y = e.deltaY

      // Pixel units unless it's non-zero.
      // Note that if deltamode is line or page won't matter since we aren't
      // sending the mouse wheel delta to the server anyway.
      // The difference between pixel and line can be important however since
      // we have a threshold that can be smaller than the line height.
      if (e.deltaMode !== 0) {
        x *= WHEEL_LINE_HEIGHT
        y *= WHEEL_LINE_HEIGHT
      }

      if (this.scroll_invert) {
        x = x * -1
        y = y * -1
      }

      x = Math.min(Math.max(x, -this.scroll), this.scroll)
      y = Math.min(Math.max(y, -this.scroll), this.scroll)

      if (!this.wheelThrottle) {
        this.wheelThrottle = true
        this.$client.sendData('wheel', { x, y })

        window.setTimeout(() => {
          this.wheelThrottle = false
        }, 100)
      }
    }

    onMouseDown(e: MouseEvent) {
      if (!this.hosting) {
        this.$emit('control-attempt', e)
      }

      if (!this.hosting || this.locked) {
        return
      }

      this.onMousePos(e)
      this.$client.sendData('mousedown', { key: e.button + 1 })
    }

    onMouseUp(e: MouseEvent) {
      if (!this.hosting || this.locked) {
        return
      }

      this.onMousePos(e)
      this.$client.sendData('mouseup', { key: e.button + 1 })
    }

    onMouseMove(e: MouseEvent) {
      if (!this.hosting || this.locked) {
        return
      }

      this.onMousePos(e)
    }

    onMouseEnter(e: MouseEvent) {
      if (this.hosting) {
        this.$accessor.remote.syncKeyboardModifierState({
          capsLock: e.getModifierState('CapsLock'),
          numLock: e.getModifierState('NumLock'),
          scrollLock: e.getModifierState('ScrollLock'),
        })
      }

      this._overlay.focus()
      this.onFocus()
      this.focused = true
    }

    onMouseLeave(e: MouseEvent) {
      if (this.hosting) {
        this.$accessor.remote.setKeyboardModifierState({
          capsLock: e.getModifierState('CapsLock'),
          numLock: e.getModifierState('NumLock'),
          scrollLock: e.getModifierState('ScrollLock'),
        })
      }

      this.keyboard.reset()
      this.focused = false
    }

    onResise() {
      let height = 0
      if (!this.fullscreen) {
        const { offsetWidth, offsetHeight } = this._component
        this._player.style.width = `${offsetWidth}px`
        this._player.style.height = `${offsetHeight}px`
        height = offsetHeight
      } else {
        const { offsetWidth, offsetHeight } = this._player
        height = offsetHeight
      }

      this._container.style.maxWidth = `${(this.horizontal / this.vertical) * height}px`
      this._aspect.style.paddingBottom = `${(this.vertical / this.horizontal) * 100}%`
    }

    onResolution(event: MouseEvent) {
      this._resolution.open(event)
    }

    onClipboard(event: MouseEvent) {
      this._clipboard.open(event)
    }
  }
</script>
