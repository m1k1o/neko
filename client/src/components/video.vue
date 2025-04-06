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
        <textarea
          ref="overlay"
          class="overlay"
          spellcheck="false"
          tabindex="0"
          data-gramm="false"
          :style="{ pointerEvents: hosting ? 'auto' : 'none' }"
          @click.stop.prevent
          @contextmenu.stop.prevent
          @wheel.stop.prevent="onWheel"
          @mousemove.stop.prevent="onMouseMove"
          @mousedown.stop.prevent="onMouseDown"
          @mouseup.stop.prevent="onMouseUp"
          @mouseenter.stop.prevent="onMouseEnter"
          @mouseleave.stop.prevent="onMouseLeave"
          @touchmove.stop.prevent="onTouchHandler"
          @touchstart.stop.prevent="onTouchHandler"
          @touchend.stop.prevent="onTouchHandler"
        />
        <div v-if="!playing && playable" class="player-overlay" @click.stop.prevent="playAndUnmute">
          <i class="fas fa-play-circle" />
        </div>
        <div v-else-if="mutedOverlay && muted" class="player-overlay" @click.stop.prevent="unmute">
          <i class="fas fa-volume-up" />
        </div>
        <div ref="aspect" class="player-aspect" />
      </div>
      <ul v-if="!fullscreen && !hideControls" class="video-menu top">
        <li><i @click.stop.prevent="requestFullscreen" class="fas fa-expand"></i></li>
        <li v-if="admin"><i @click.stop.prevent="openResolution" class="fas fa-desktop"></i></li>
        <li v-if="!controlLocked && !implicitHosting" :class="extraControls || 'extra-control'">
          <i
            :class="[hosted && !hosting ? 'disabled' : '', !hosted && !hosting ? 'faded' : '', 'fas', 'fa-keyboard']"
            @click.stop.prevent="toggleControl"
          />
        </li>
      </ul>
      <ul v-if="!fullscreen && !hideControls" class="video-menu bottom">
        <li v-if="hosting && (!clipboard_read_available || !clipboard_write_available)">
          <i @click.stop.prevent="openClipboard" class="fas fa-clipboard"></i>
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
      background: #000;

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

          /* usually extra controls are only shown on mobile */
          &.extra-control {
            display: none;
          }
          @media (max-width: 768px) {
            &.extra-control {
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
        max-width: calc(16 / 9 * 100vh);

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
          cursor: pointer;

          i::before {
            font-size: 120px;
            text-align: center;
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
          cursor: default;
          outline: 0;
          border: 0;
          color: transparent;
          background: transparent;
          resize: none;
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
  import { mapState, mapMutations } from 'vuex' // Import Vuex helpers
  import ResizeObserver from 'resize-observer-polyfill'
  import { elementRequestFullscreen, onFullscreenChange, isFullscreen, lockKeyboard, unlockKeyboard } from '~/utils'
  import { BaseClient } from '~/neko/base' // Import BaseClient type

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
    @Ref('overlay') readonly _overlay!: HTMLTextAreaElement
    @Ref('aspect') readonly _aspect!: HTMLElement
    @Ref('player') readonly _player!: HTMLElement
    @Ref('video') readonly _video!: HTMLVideoElement
    @Ref('resolution') readonly _resolution!: Resolution
    @Ref('clipboard') readonly _clipboard!: Clipboard

    // all controls are hidden (e.g. for cast mode)
    @Prop(Boolean) readonly hideControls!: boolean
    // extra controls are shown (e.g. for embed mode)
    @Prop(Boolean) readonly extraControls!: boolean

    private keyboard = GuacamoleKeyboard()
    private observer = new ResizeObserver(this.onResize.bind(this))
    private focused = false
    private fullscreen = false
    private mutedOverlay = true

    // MSE related properties
    private mediaSource: MediaSource | null = null
    private sourceBuffer: SourceBuffer | null = null
    private wsBufferQueue: ArrayBuffer[] = []
    private isInitSegmentReceived = false
    private neko: BaseClient | null = null // To hold the Neko instance

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

    get implicitHosting() {
      return this.$accessor.remote.implicitHosting
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

    // server-side lock
    get controlLocked() {
      return 'control' in this.$accessor.locked && this.$accessor.locked['control'] && !this.$accessor.user.admin
    }

    get locked() {
      return this.$accessor.remote.locked || (this.controlLocked && (!this.hosting || this.implicitHosting))
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
      return (
        'clipboard' in navigator &&
        typeof navigator.clipboard.readText === 'function' &&
        // Firefox 122+ incorrectly reports that it can read the clipboard but it can't
        // instead it hangs when reading clipboard, until user clicks on the page
        // and the click itself is not handled by the page at all, also the clipboard
        // reads always fail with "Clipboard read operation is not allowed."
        navigator.userAgent.indexOf('Firefox') == -1
      )
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

    // Map Vuex state for streaming mode
    get streamingMode() {
      return this.$accessor.remote.streamingMode;
    }

    get isWsStreamingActive() {
      return this.$accessor.remote.isWsStreamingActive;
    }

    @Watch('width')
    onWidthChanged() {
      this.onResize()
    }

    @Watch('height')
    onHeightChanged() {
      this.onResize()
    }

    @Watch('volume')
    onVolumeChanged(volume: number) {
      volume /= 100

      if (this._video && this._video.volume != volume) {
        this._video.volume = volume
      }
    }

    @Watch('muted')
    onMutedChanged(muted: boolean) {
      if (this._video && this._video.muted != muted) {
        this._video.muted = muted

        if (!muted) {
          this.mutedOverlay = false
        }
      }
    }

    @Watch('stream')
    onStreamChanged(stream?: MediaStream) {
      // Only handle stream if in WebRTC mode
      if (this.streamingMode !== 'webrtc') return;

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
    async onPlayingChanged(playing: boolean) {
      if (this._video && this._video.paused && playing) {
        // if autoplay is disabled, play() will throw an error
        // and we need to properly save the state otherwise we
        // would be thinking we're playing when we're not
        try {
          await this._video.play()
        } catch (err: any) {
          if (!this._video.muted) {
            // video.play() can fail if audio is set due restrictive
            // browsers autoplay policy -> retry with muted audio
            try {
              this.$accessor.video.setMuted(true)
              this._video.muted = true
              await this._video.play()
            } catch (err: any) {
              // if it still fails, we're not playing anything
              this.$accessor.video.pause()
            }
          } else {
            this.$accessor.video.pause()
          }
        }
      }

      if (this._video && !this._video.paused && !playing) {
        this.pause()
      }
    }

    @Watch('clipboard')
    async onClipboardChanged(clipboard: string) {
      if (this.clipboard_write_available) {
        try {
          await navigator.clipboard.writeText(clipboard)
          this.$accessor.remote.setClipboard(clipboard)
        } catch (err: any) {
          this.$log.error(err)
        }
      }
    }

    @Watch('streamingMode')
    onStreamingModeChanged(newMode: string, oldMode: string) {
      this.$log.info(`Streaming mode changed from ${oldMode} to ${newMode}`);
      if (newMode === 'websocket' && oldMode !== 'websocket') {
        this.resetWebRTC(); // Make sure WebRTC resources are cleared
        this.setupWebSocketStreaming();
      } else if (newMode === 'webrtc' && oldMode !== 'webrtc') {
        this.resetWebSocketStreaming(); // Clean up MSE resources
        // WebRTC setup is likely handled elsewhere when stream is received
      }
    }

    mounted() {
      // Get the Neko client instance (assuming it's available as $client)
      this.neko = this.$client as BaseClient;
      if (!this.neko) {
        this.$log.error("Neko client instance ($client) not found in video component!");
        return;
      }

      // Listen for binary data from the Neko client
      this.neko.on('binary_data', this.handleWebSocketData);

      this._container.addEventListener('resize', this.onResize)
      this.onVolumeChanged(this.volume)
      this.onMutedChanged(this.muted)
      this.onStreamChanged(this.stream)
      this.onResize()

      this.observer.observe(this._component)

      onFullscreenChange(this._player, () => {
        this.fullscreen = isFullscreen()
        this.fullscreen ? lockKeyboard() : unlockKeyboard()
        this.onResize()
        this.$accessor.video.setPlayable(true)
        if (this.autoplay) {
          this.$nextTick(() => {
            this.$accessor.video.play()
          })
        }

        // Initial setup based on current mode
        if (this.streamingMode === 'websocket') {
          this.setupWebSocketStreaming();
        }
      })

      this._video.addEventListener('canplaythrough', () => {
        this.$accessor.video.setPlayable(true)
        if (this.autoplay) {
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

      this._video.addEventListener('volumechange', () => {
        this.$accessor.video.setMuted(this._video.muted)
        this.$accessor.video.setVolume(this._video.volume * 100)
      })

      this._video.addEventListener('playing', () => {
        this.$accessor.video.play()
      })

      this._video.addEventListener('pause', () => {
        this.$accessor.video.pause()
      })

      /* Initialize Guacamole Keyboard */
      this.keyboard.onkeydown = (key: number) => {
        if (!this.hosting || this.locked) {
          return true
        }

        this.$client.sendData('keydown', { key: this.keyMap(key) })
        return false
      }
      this.keyboard.onkeyup = (key: number) => {
        if (!this.hosting || this.locked) {
          return
        }

        this.$client.sendData('keyup', { key: this.keyMap(key) })
      }
      this.keyboard.listenTo(this._overlay)
    }

    beforeDestroy() {
      this.observer.disconnect()
      this.$accessor.video.setPlayable(false)
      // Remove listener
      this.neko?.off('binary_data', this.handleWebSocketData);
      // Clean up MSE resources
      this.resetWebSocketStreaming();
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

    async play() {
      if (!this._video.paused || !this.playable) {
        return
      }

      try {
        await this._video.play()
        this.onResize()
      } catch (err: any) {
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

    playAndUnmute() {
      this.$accessor.video.play()
      this.$accessor.video.setMuted(false)
    }

    unmute() {
      this.$accessor.video.setMuted(false)
    }

    toggleControl() {
      if (!this.playable) {
        return
      }

      this.$accessor.remote.toggle()
    }

    requestControl() {
      this.$accessor.remote.request()
    }

    requestFullscreen() {
      // try to fullscreen player element
      if (elementRequestFullscreen(this._player)) {
        this.onResize()
        return
      }

      // fallback to fullscreen video itself (on mobile devices)
      if (elementRequestFullscreen(this._video)) {
        this.onResize()
        return
      }
    }

    requestPictureInPicture() {
      //@ts-ignore
      this._video.requestPictureInPicture()
      this.onResize()
    }

    openResolution(event: MouseEvent) {
      this._resolution.open(event)
    }

    openClipboard() {
      this._clipboard.open()
    }

    async syncClipboard() {
      if (this.clipboard_read_available && window.document.hasFocus()) {
        try {
          const text = await navigator.clipboard.readText()
          if (this.clipboard !== text) {
            this.$accessor.remote.setClipboard(text)
            this.$accessor.remote.sendClipboard(text)
          }
        } catch (err: any) {
          this.$log.error(err)
        }
      }
    }

    sendMousePos(e: MouseEvent) {
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

      this.sendMousePos(e)

      if (!this.wheelThrottle) {
        this.wheelThrottle = true
        this.$client.sendData('wheel', { x, y })

        window.setTimeout(() => {
          this.wheelThrottle = false
        }, 100)
      }
    }

    onTouchHandler(e: TouchEvent) {
      let first = e.changedTouches[0]
      let type = ''
      switch (e.type) {
        case 'touchstart':
          type = 'mousedown'
          break
        case 'touchmove':
          type = 'mousemove'
          break
        case 'touchend':
          type = 'mouseup'
          break
        default:
          return
      }

      const simulatedEvent = new MouseEvent(type, {
        bubbles: true,
        cancelable: true,
        view: window,
        screenX: first.screenX,
        screenY: first.screenY,
        clientX: first.clientX,
        clientY: first.clientY,
      })
      first.target.dispatchEvent(simulatedEvent)
    }

    onMouseDown(e: MouseEvent) {
      if (!this.hosting) {
        this.$emit('control-attempt', e)
      }

      if (!this.hosting || this.locked) {
        return
      }

      this.sendMousePos(e)
      this.$client.sendData('mousedown', { key: e.button + 1 })
    }

    onMouseUp(e: MouseEvent) {
      if (!this.hosting || this.locked) {
        return
      }

      this.sendMousePos(e)
      this.$client.sendData('mouseup', { key: e.button + 1 })
    }

    onMouseMove(e: MouseEvent) {
      if (!this.hosting || this.locked) {
        return
      }

      this.sendMousePos(e)
    }

    onMouseEnter(e: MouseEvent) {
      if (this.hosting) {
        this.$accessor.remote.syncKeyboardModifierState({
          capsLock: e.getModifierState('CapsLock'),
          numLock: e.getModifierState('NumLock'),
          scrollLock: e.getModifierState('ScrollLock'),
        })

        this.syncClipboard()
      }

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

    onResize() {
      const { offsetWidth, offsetHeight } = !this.fullscreen ? this._component : document.body
      this._player.style.width = `${offsetWidth}px`
      this._player.style.height = `${offsetHeight}px`
      this._container.style.maxWidth = `${(this.horizontal / this.vertical) * offsetHeight}px`
      this._aspect.style.paddingBottom = `${(this.vertical / this.horizontal) * 100}%`
    }

    @Watch('focused')
    @Watch('hosting')
    @Watch('locked')
    onFocus() {
      // in order to capture key events, overlay must be focused
      if (this.focused && this.hosting && !this.locked) {
        this._overlay.focus()
      }
    }

    // --- WebSocket Streaming (MSE) Methods ---
    setupWebSocketStreaming() {
      if (!('MediaSource' in window)) {
        this.$log.error('Media Source Extensions (MSE) not supported by this browser.');
        // Handle error - maybe show a message to the user
        return;
      }
      if (this.isWsStreamingActive) return; // Already set up

      this.$log.info('Setting up WebSocket streaming (MSE)');
      this.$accessor.remote.setWsStreamingActive(true);
      this.isInitSegmentReceived = false;
      this.wsBufferQueue = [];

      // Reset video source if it was previously used by WebRTC
      if (this._video.srcObject) {
        this._video.srcObject = null;
      }
      if (this._video.src) {
        URL.revokeObjectURL(this._video.src);
        this._video.src = '';
      }
      this._video.load(); // Reset video element

      this.mediaSource = new MediaSource();
      this._video.src = URL.createObjectURL(this.mediaSource);

      this.mediaSource.addEventListener('sourceopen', this.handleSourceOpen);
      this.mediaSource.addEventListener('sourceended', () => this.$log.info('MediaSource sourceended'));
      this.mediaSource.addEventListener('sourceclose', () => this.$log.info('MediaSource sourceclose'));

      // Start playback attempt for MSE
      this.$accessor.video.setPlayable(true); // Assume playable once setup starts
      if (this.autoplay) {
        this.$nextTick(() => {
          this.$accessor.video.play();
        });
      }
    }

    handleSourceOpen() {
      if (!this.mediaSource) return;
      this.$log.info('MediaSource sourceopen event');

      // CRITICAL: Determine the correct MIME type
      const mimeCodec = 'video/mp4; codecs="avc1.42E01E, mp4a.40.2"'; // Adjust based on server!
      if (MediaSource.isTypeSupported(mimeCodec)) {
          try {
            this.sourceBuffer = this.mediaSource.addSourceBuffer(mimeCodec);
            this.$log.info('SourceBuffer created');

            this.sourceBuffer.addEventListener('updateend', () => {
                this.appendNextFromQueue(); // Process next queued segment
            });
            this.sourceBuffer.addEventListener('error', (e) => {
                this.$log.error('SourceBuffer error:', e);
            });
            this.sourceBuffer.addEventListener('abort', () => this.$log.warn('SourceBuffer abort event'));

            // Process any segments that arrived before the source buffer was ready
            this.appendNextFromQueue();

          } catch (e: any) {
            this.$log.error('Error creating SourceBuffer:', e);
            this.resetWebSocketStreaming();
          }
      } else {
        this.$log.error(`Unsupported MIME type or codec for MSE: ${mimeCodec}`);
        this.resetWebSocketStreaming();
      }
    }

    handleWebSocketData(data: ArrayBuffer /* ArrayBuffer */) {
        if (!this.isWsStreamingActive || !data || !this.mediaSource || this.mediaSource.readyState !== 'open') {
          //this.$log.debug('Ignoring WS data: streaming not active or MediaSource not open');
          return;
        }

        if (!this.sourceBuffer || this.sourceBuffer.updating) {
            // If buffer not ready or busy, queue the data
            this.wsBufferQueue.push(data);
            //this.$log.debug(`Queued segment (${this.wsBufferQueue.length})`);
            return;
        }

        this.appendBuffer(data);
    }

    appendBuffer(data: ArrayBuffer) {
        if (!this.sourceBuffer || this.sourceBuffer.updating || !this.isWsStreamingActive) {
          this.$log.warn('Attempted to append buffer while sourceBuffer not ready or not active.');
          this.wsBufferQueue.push(data); // Re-queue if possible
          return;
        }
        try {
            //this.$log.debug(`Appending buffer: ${data.byteLength} bytes. Init received: ${this.isInitSegmentReceived}`);
            this.sourceBuffer.appendBuffer(data);
            if (!this.isInitSegmentReceived) {
                // The first segment appended is the initialization segment
                this.isInitSegmentReceived = true;
                this.$log.info("Initialization segment appended.");
            }
        } catch (e: any) {
            this.$log.error('Error appending buffer:', e);
            if (e.name === 'QuotaExceededError') {
              this.$log.warn('QuotaExceededError: Buffer full? Attempting to clean...');
              // Basic cleanup: remove some time range from the beginning
              if (this.sourceBuffer && !this.sourceBuffer.updating && this._video.buffered.length > 0) {
                const removeEnd = this._video.buffered.start(0) + 5; // Remove first 5 seconds
                const removeStart = this._video.buffered.start(0);
                if (removeEnd > removeStart) {
                  try {
                    this.$log.info(`Attempting to remove buffer range: ${removeStart} - ${removeEnd}`);
                    this.sourceBuffer.remove(removeStart, removeEnd);
                    // Re-queue the failed segment after attempting removal
                    this.wsBufferQueue.unshift(data);
                  } catch (removeError: any) {
                    this.$log.error('Error removing buffer range:', removeError);
                    this.resetWebSocketStreaming(); // Fatal error if remove fails
                  }
                } else {
                   this.resetWebSocketStreaming(); // Cannot cleanup
                }
              } else {
                this.resetWebSocketStreaming(); // Cannot cleanup
              }
            } else {
              // Attempt to recover or reset for other errors
              this.resetWebSocketStreaming();
              // You might want to notify the user or attempt reconnection
            }
        }
    }

    appendNextFromQueue() {
        // Ensure sourceBuffer exists and is not updating
        if (this.wsBufferQueue.length > 0 && this.sourceBuffer && !this.sourceBuffer.updating) {
            const dataToAppend = this.wsBufferQueue.shift()!;
            //this.$log.debug(`Appending next from queue (${this.wsBufferQueue.length} left)`);
            this.appendBuffer(dataToAppend);
        }
    }

    resetWebSocketStreaming() {
      if (!this.isWsStreamingActive && !this.mediaSource) return; // Nothing to reset
      this.$log.info('Resetting WebSocket streaming (MSE)');
      this.$accessor.remote.setWsStreamingActive(false);
      this.isInitSegmentReceived = false;
      this.wsBufferQueue = [];

      if (this.sourceBuffer) {
        // Remove listeners before aborting or removing
          try {
            // Unregister listeners cleanly if possible (depends on how they were added)
            // Example: this.sourceBuffer.removeEventListener(...) 

            if (this.mediaSource && this.mediaSource.readyState === 'open') {
                if (!this.sourceBuffer.updating) {
                    this.sourceBuffer.abort(); // Abort current operation
                }
                this.mediaSource.removeSourceBuffer(this.sourceBuffer);
                this.$log.debug('SourceBuffer removed.');
            }
          } catch (e: any) {
            this.$log.warn("Error removing source buffer:", e);
          }
          this.sourceBuffer = null;
      }
      if (this.mediaSource) {
        if (this.mediaSource.readyState === 'open') {
          try {
            this.mediaSource.endOfStream();
            this.$log.debug('MediaSource endOfStream called.');
          } catch (e: any) {
              this.$log.warn('Error calling endOfStream:', e);
          }
        }
          // Unregister listeners cleanly
          this.mediaSource.removeEventListener('sourceopen', this.handleSourceOpen);
          // this.mediaSource.removeEventListener('sourceended', ...);
          // this.mediaSource.removeEventListener('sourceclose', ...);
          this.mediaSource = null;
      }
      // Clear video element source
      if (this._video && this._video.src) {
        URL.revokeObjectURL(this._video.src);
        this._video.removeAttribute('src');
        this._video.load(); // Reset video element state
        this.$log.debug('Video source cleared.');
      }
      this.$accessor.video.setPlayable(false); // Mark as not playable
    }

    // Helper to reset WebRTC related things if needed when switching to WS
    resetWebRTC() {
       this.$log.info('Resetting WebRTC stream/video source');
       if (this._video.srcObject) {
          (this._video.srcObject as MediaStream).getTracks().forEach(track => track.stop());
          this._video.srcObject = null;
       }
       // You might need to call a method in the Vuex store or Neko client
       // to properly close the PeerConnection if it's still active.
       // e.g., this.$client?.disconnectPeer(); // If such a method exists
       this.$accessor.video.setStream(undefined); // Clear WebRTC stream from store
       this.$accessor.video.setPlayable(false);
    }

    // --- End WebSocket Methods ---
  }
</script>
