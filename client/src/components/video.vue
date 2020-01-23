<template>
  <div ref="component" class="video">
    <div ref="player" class="player">
      <div ref="container" class="player-container">
        <video ref="video" />
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
          @keydown.stop.prevent="onKeyDown"
          @keyup.stop.prevent="onKeyUp"
        />
        <div v-if="!playing" class="player-overlay">
          <i @click.stop.prevent="toggle" v-if="playable" class="fas fa-play-circle" />
        </div>
        <div ref="aspect" class="player-aspect" />
      </div>
      <i v-if="!fullscreen" @click.stop.prevent="requestFullscreen" class="expand fas fa-expand"></i>
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

      .expand {
        position: absolute;
        right: 20px;
        top: 15px;
        width: 30px;
        height: 30px;
        background: rgba($color: #fff, $alpha: 0.2);
        border-radius: 5px;
        line-height: 30px;
        font-size: 16px;
        text-align: center;
        color: rgba($color: #fff, $alpha: 0.6);
        cursor: pointer;
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
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'
  import ResizeObserver from 'resize-observer-polyfill'

  import Emote from './emote.vue'

  @Component({
    name: 'neko-video',
    components: {
      'neko-emote': Emote,
    },
  })
  export default class extends Vue {
    @Ref('component') readonly _component!: HTMLElement
    @Ref('container') readonly _container!: HTMLElement
    @Ref('overlay') readonly _overlay!: HTMLElement
    @Ref('aspect') readonly _aspect!: HTMLElement
    @Ref('player') readonly _player!: HTMLElement
    @Ref('video') readonly _video!: HTMLVideoElement

    private observer = new ResizeObserver(this.onResise.bind(this))
    private focused = false
    private fullscreen = false

    get connected() {
      return this.$accessor.connected
    }

    get connecting() {
      return this.$accessor.connecting
    }

    get hosting() {
      return this.$accessor.remote.hosting
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

    get scroll() {
      return this.$accessor.settings.scroll
    }

    get scroll_invert() {
      return this.$accessor.settings.scroll_invert
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
          this.$accessor.video.play()
        }
      })

      this._video.addEventListener('ended', () => {
        this.$accessor.video.setPlayable(false)
      })

      this._video.addEventListener('error', event => {
        console.error(event.error)
        this.$accessor.video.setPlayable(false)
      })
    }

    beforeDestroy() {
      this.observer.disconnect()
      this.$accessor.video.setPlayable(false)
    }

    play() {
      if (!this._video.paused || !this.playable) {
        return
      }

      this._video
        .play()
        .then(() => {
          const { videoWidth, videoHeight } = this._video
          this.$accessor.video.setResolution({ width: videoWidth, height: videoHeight })
          this.onResise()
        })
        .catch(err => console.log(err))
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

    requestFullscreen() {
      this._player.requestFullscreen()
      this.onResise()
    }

    onMousePos(e: MouseEvent) {
      const { w, h } = this.$accessor.video.resolution
      const rect = this._overlay.getBoundingClientRect()
      this.$client.sendData('mousemove', {
        x: Math.round((w / rect.width) * (e.clientX - rect.left)),
        y: Math.round((h / rect.height) * (e.clientY - rect.top)),
      })
    }

    onWheel(e: WheelEvent) {
      if (!this.hosting) {
        return
      }
      this.onMousePos(e)

      let x = e.deltaX
      let y = e.deltaY

      if (this.scroll_invert) {
        x = x * -1
        y = y * -1
      }

      x = Math.min(Math.max(x, -this.scroll), this.scroll)
      y = Math.min(Math.max(y, -this.scroll), this.scroll)

      this.$client.sendData('wheel', { x, y })
    }

    onMouseDown(e: MouseEvent) {
      if (!this.hosting) {
        return
      }
      this.onMousePos(e)
      this.$client.sendData('mousedown', { key: e.button })
    }

    onMouseUp(e: MouseEvent) {
      if (!this.hosting) {
        return
      }
      this.onMousePos(e)
      this.$client.sendData('mouseup', { key: e.button })
    }

    onMouseMove(e: MouseEvent) {
      if (!this.hosting) {
        return
      }
      this.onMousePos(e)
    }

    onMouseEnter(e: MouseEvent) {
      this._overlay.focus()
      this.focused = true
    }

    onMouseLeave(e: MouseEvent) {
      this.focused = false
    }

    onKeyDown(e: KeyboardEvent) {
      if (!this.focused || !this.hosting) {
        return
      }
      this.$client.sendData('keydown', { key: e.keyCode })
    }

    onKeyUp(e: KeyboardEvent) {
      if (!this.focused || !this.hosting) {
        return
      }
      this.$client.sendData('keyup', { key: e.keyCode })
    }

    onResise() {
      const { horizontal, vertical } = this.$accessor.video

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

      this._container.style.maxWidth = `${(horizontal / vertical) * height}px`
      this._aspect.style.paddingBottom = `${(vertical / horizontal) * 100}%`
    }
  }
</script>
