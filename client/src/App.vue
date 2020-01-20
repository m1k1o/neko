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
            :class="[{ enabled: hosting }, 'request', 'fas', 'fa-keyboard']"
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
        <form class="message" v-if="!connecting" @submit.stop.prevent="connect">
          <span>Please enter the password:</span>
          <input type="text" placeholder="Username" v-model="username" />
          <input type="password" placeholder="Password" v-model="password" />
          <button type="submit" class="button" @click.stop.prevent="connect">
            Connect
          </button>
        </form>
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
  import { EVENT } from '~/client/events'

  @Component({ name: 'stream-video' })
  export default class extends Vue {
    @Ref('player') readonly _player!: HTMLVideoElement
    @Ref('container') readonly _container!: HTMLElement
    @Ref('aspect') readonly _aspect!: HTMLElement
    @Ref('video') readonly _video!: HTMLElement
    @Ref('volume') readonly _volume!: HTMLInputElement

    private focused = false
    private playing = false
    private username = ''
    private password = ''

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

    get stream() {
      return this.$accessor.video.stream
    }

    @Watch('volume')
    onVolumeChanged(volume: number) {
      if (this._player) {
        this._player.volume = this.volume / 100
      }
    }

    @Watch('stream')
    onStreamChanged(stream?: MediaStream) {
      if (!this._player || !stream) {
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

    mounted() {
      window.addEventListener('resize', this.onResise)
      this.onResise()
      this._player.volume = this.volume / 100
      this._volume.value = `${this.volume}`
      this.onStreamChanged(this.stream)
    }

    beforeDestroy() {
      window.removeEventListener('resize', this.onResise)
    }

    toggleMedia() {
      if (!this.playing) {
        this._player
          .play()
          .then(() => {
            const { videoWidth, videoHeight } = this._player
            this.$accessor.video.setResolution({ width: videoWidth, height: videoHeight })
            this.playing = true
            this.onResise()
          })
          .catch(err => {})
      } else {
        this._player.pause()
        this.playing = false
      }
    }

    connect() {
      this.$client.connect(this.password, this.username)
    }

    toggleControl() {
      if (!this.connected) {
        return
      }

      if (!this.hosting) {
        this.$client.sendMessage(EVENT.CONTROL.REQUEST)
      } else {
        this.$client.sendMessage(EVENT.CONTROL.RELEASE)
      }
    }

    setVolume() {
      this.$accessor.video.setVolume(parseInt(this._volume.value))
    }

    fullscreen() {
      this._video.requestFullscreen()
    }

    onMousePos(e: MouseEvent) {
      const { w, h } = this.$accessor.video.resolution
      const rect = this._player.getBoundingClientRect()
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
      this.$client.sendData('wheel', {
        x: (e.deltaX * -1) / 10,
        y: (e.deltaY * -1) / 10,
      }) // TODO: Add user settings
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
      this._player.focus()
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
      const aspect = this.$accessor.video.aspect
      if (!aspect) {
        return
      }
      const { horizontal, vertical } = aspect
      this._container.style.maxWidth = `${(horizontal / vertical) * this._video.offsetHeight}px`
      this._aspect.style.paddingBottom = `${(vertical / horizontal) * 100}%`
    }
  }
</script>
