<template>
  <ul>
    <li v-if="!implicitHosting && (!controlLocked || hosting)">
      <button
        type="button"
        :class="[
          !disabeld && shakeKbd ? 'shake' : '',
          disabeld && !hosting ? 'disabled' : '',
          !disabeld && !hosting ? 'faded' : '',
          'request',
        ]"
        v-tooltip="{
          content: !disabeld || hosting ? (hosting ? $t('controls.release') : $t('controls.request')) : '',
          placement: 'top',
          offset: 5,
          boundariesElement: 'body',
          delay: { show: 300, hide: 100 },
        }"
        :aria-label="!disabeld || hosting ? (hosting ? $t('controls.release') : $t('controls.request')) : ''"
        @click.stop.prevent="toggleControl"
      >
        <i class="fas fa-keyboard" aria-hidden="true" />
      </button>
    </li>
    <li class="no-pointer" v-if="implicitHosting">
      <span
        class="control-state"
        :class="[controlLocked ? 'disabled' : '', 'fas', 'fa-mouse-pointer']"
        v-tooltip="{
          content: controlLocked ? $t('controls.hasnot') : $t('controls.has'),
          placement: 'top',
          offset: 5,
          boundariesElement: 'body',
          delay: { show: 300, hide: 100 },
        }"
      />
    </li>
    <li v-if="implicitHosting || (!implicitHosting && (!controlLocked || hosting))">
      <label
        class="switch"
        v-tooltip="{
          content: hosting ? (locked ? $t('controls.unlock') : $t('controls.lock')) : '',
          placement: 'top',
          offset: 5,
          boundariesElement: 'body',
          delay: { show: 300, hide: 100 },
        }"
      >
        <input type="checkbox" v-model="locked" :disabled="!hosting || (implicitHosting && controlLocked)" />
        <span />
      </label>
    </li>
    <li>
      <button
        type="button"
        :class="[{ disabled: !playable }, 'play']"
        :aria-label="$t(playing ? 'ui.pause' : 'ui.play')"
        @click.stop.prevent="toggleMedia"
      >
        <i :class="[playing ? 'fa-pause-circle' : 'fa-play-circle', 'fas']" aria-hidden="true" />
      </button>
    </li>
    <li v-if="micAllowed">
      <button
        type="button"
        :class="[{ disabled: !playable }, microphoneActive ? '' : 'faded']"
        v-tooltip="{
          content: microphoneActive ? $t('controls.mic_off') : $t('controls.mic_on'),
          placement: 'top',
          offset: 5,
          boundariesElement: 'body',
          delay: { show: 300, hide: 100 },
        }"
        :aria-label="microphoneActive ? $t('controls.mic_off') : $t('controls.mic_on')"
        @click.stop.prevent="toggleMicrophone"
      >
        <i :class="[microphoneActive ? 'fa-microphone' : 'fa-microphone-slash', 'fas']" aria-hidden="true" />
      </button>
    </li>
    <li>
      <div class="volume">
        <button
          type="button"
          class="volume-button"
          :aria-label="$t(muted ? 'ui.unmute' : 'ui.mute')"
          @click.stop.prevent="toggleMute"
        >
          <i :class="[volume === 0 || muted ? 'fa-volume-mute' : 'fa-volume-up', 'fas']" aria-hidden="true" />
        </button>
        <input type="range" min="0" max="100" v-model="volume" :aria-label="$t('ui.volume')" />
      </div>
    </li>
  </ul>
</template>

<style lang="scss" scoped>
  .shake {
    animation: shake 1.25s cubic-bezier(0, 0, 0, 1);
  }

  @keyframes shake {
    0% {
      transform: scale(1) translate(0px, 0) rotate(0);
    }
    10% {
      transform: scale(1.25) translate(-2px, -2px) rotate(-20deg);
    }
    20% {
      transform: scale(1.5) translate(4px, -4px) rotate(20deg);
    }
    30% {
      transform: scale(1.75) translate(-4px, -6px) rotate(-20deg);
    }
    40% {
      transform: scale(2) translate(6px, -8px) rotate(20deg);
    }
    50% {
      transform: scale(2.25) translate(-6px, -10px) rotate(-20deg);
    }
    60% {
      transform: scale(2) translate(6px, -8px) rotate(20deg);
    }
    70% {
      transform: scale(1.75) translate(-4px, -6px) rotate(-20deg);
    }
    80% {
      transform: scale(1.5) translate(4px, -4px) rotate(20deg);
    }
    90% {
      transform: scale(1.25) translate(-2px, -2px) rotate(-20deg);
    }
    100% {
      transform: scale(1) translate(0px, 0) rotate(0);
    }
  }

  ul {
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    list-style: none;

    li {
      font-size: 24px;
      cursor: pointer;

      &.no-pointer {
        cursor: default;
      }

      > button {
        width: 42px;
        height: 42px;
        display: grid;
        place-items: center;
        padding: 0;
        border: 1px solid transparent;
        border-radius: 12px;
        color: $interactive-normal;
        background: transparent;
        cursor: pointer;
        transition: color 0.18s ease, background 0.18s ease, border-color 0.18s ease, transform 0.18s ease;

        &:hover:not(.disabled),
        &:focus-visible:not(.disabled) {
          color: $interactive-hover;
          background: $background-modifier-hover;
          border-color: rgba($text-normal, 0.12);
          transform: translateY(-1px);
        }

        &.faded {
          color: rgba($text-normal, 0.42);
        }

        &.disabled {
          color: rgba($style-error, 0.5);
          cursor: default;
        }

        i {
          padding: 0;
        }
      }

      > .control-state {
        padding: 0 5px;

        &.disabled {
          color: rgba($style-error, 0.5);
        }
      }

      .volume {
        white-space: nowrap;
        display: block;
        display: flex;
        flex-direction: row;
        justify-content: center;
        align-items: center;
        list-style: none;

        .volume-button {
          width: 38px;
          height: 38px;
          padding: 0;
          border: 0;
          display: grid;
          place-items: center;
          color: $interactive-normal;
          background: transparent;
          cursor: pointer;

          &:hover,
          &:focus-visible {
            color: $interactive-hover;
          }

          i {
            padding: 0;
          }
        }

        input[type='range'] {
          width: 100%;
          background: transparent;
          width: clamp(5rem, 12vw, 9.375rem);
          height: 20px;
          -webkit-appearance: none;

          &::-moz-range-thumb {
            height: 12px;
            width: 12px;
            border-radius: 12px;
            background: #fff;
            cursor: pointer;
          }

          &::-moz-range-track {
            width: 100%;
            height: 4px;
            cursor: pointer;
            background: $style-primary;
            border-radius: 2px;
          }

          &::-webkit-slider-thumb {
            -webkit-appearance: none;
            height: 12px;
            width: 12px;
            border-radius: 12px;
            background: #fff;
            cursor: pointer;
            margin-top: -4px;
          }

          &::-webkit-slider-runnable-track {
            width: 100%;
            height: 4px;
            cursor: pointer;
            background: $style-primary;
            border-radius: 2px;
          }
        }
      }

      .switch {
        margin: 0 5px;
        display: block;
        position: relative;
        width: 42px;
        height: 24px;

        input[type='checkbox'] {
          opacity: 0;
          width: 0;
          height: 0;
        }

        span {
          position: absolute;
          cursor: pointer;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background-color: $background-secondary;
          transition: 0.2s;
          border-radius: 34px;

          &:before {
            color: $background-tertiary;
            font-weight: 900;
            font-family: 'Font Awesome 6 Free';
            content: '\f3c1';
            font-size: 8px;
            line-height: 18px;
            text-align: center;
            position: absolute;
            height: 18px;
            width: 18px;
            left: 3px;
            bottom: 3px;
            background-color: white;
            transition: 0.3s;
            border-radius: 50%;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
          }
        }
      }

      input[type='checkbox'] {
        &:checked + span {
          background-color: $style-primary;

          &:before {
            content: '\f023';
            transform: translateX(18px);
          }
        }

        &:disabled + span {
          &:before {
            content: '';
            background-color: rgba($color: $text-normal, $alpha: 0.4);
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Prop, Watch } from 'vue-property-decorator'

  @Component({ name: 'neko-controls' })
  export default class extends Vue {
    @Prop(Boolean) readonly shakeKbd!: boolean

    get controlLocked() {
      return (
        'control' in this.$accessor.session.locked &&
        this.$accessor.session.locked['control'] &&
        !this.$accessor.user.admin
      )
    }

    get disabeld() {
      return this.$accessor.remote.hosted
    }

    get hosting() {
      return this.$accessor.remote.hosting
    }

    get controlling() {
      return this.$accessor.remote.controlling
    }

    get implicitHosting() {
      return this.$accessor.remote.implicitHosting
    }

    // Microphone is allowed when the user is actively controlling (has host).
    // With implicit hosting, the controlling getter is true only when the user
    // has actually been assigned as host (clicked inside the video), not for
    // everyone by default. This prevents multiple users from sharing their
    // microphone simultaneously — only the person in control can.
    get micAllowed() {
      return this.controlling
    }

    get volume() {
      return this.$accessor.video.volume
    }

    set volume(volume: number) {
      this.$accessor.video.setVolume(volume)
    }

    get muted() {
      return this.$accessor.video.muted || this.volume === 0
    }

    get playing() {
      return this.$accessor.video.playing
    }

    get playable() {
      return this.$accessor.video.playable
    }

    get locked() {
      return this.$accessor.remote.locked && this.$accessor.remote.hosting
    }

    set locked(locked: boolean) {
      this.$accessor.remote.setLocked(locked)
    }

    toggleControl() {
      if (!this.playable) {
        return
      }
      this.$accessor.remote.toggle()
    }

    toggleMedia() {
      if (!this.playable) {
        return
      }
      this.$accessor.video.togglePlay()
    }

    toggleMute() {
      this.$accessor.video.toggleMute()
    }

    microphoneActive = false

    // Auto-disable microphone when the user loses control (e.g. another user
    // takes host, or admin releases control). This ensures the mic track is
    // cleaned up and the server-side audio input is freed for the new host.
    @Watch('controlling')
    onControllingChanged(isControlling: boolean) {
      if (!isControlling && this.microphoneActive) {
        this.$client.disableMicrophone()
        this.microphoneActive = false
      }
    }

    async toggleMicrophone() {
      if (!this.playable || !this.micAllowed) {
        return
      }

      if (this.microphoneActive) {
        this.$client.disableMicrophone()
        this.microphoneActive = false
      } else {
        try {
          await this.$client.enableMicrophone()
          this.microphoneActive = true
        } catch (err: any) {
          this.$swal({
            title: this.$t('controls.mic_error') as string,
            text: err.message,
            icon: 'error',
          })
        }
      }
    }
  }
</script>
