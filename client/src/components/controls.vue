<template>
  <ul>
    <li v-if="!isTouch">
      <i
        :class="[
          hosted && !hosting ? 'disabled' : '',
          !hosted && !hosting ? 'faded' : '',
          'fas',
          'fa-keyboard',
          'request',
        ]"
        v-tooltip="{
          content: !hosted || hosting ? (hosting ? $t('controls.release') : $t('controls.request')) : '',
          placement: 'top',
          offset: 5,
          boundariesElement: 'body',
          delay: { show: 300, hide: 100 },
        }"
        @click.stop.prevent="toggleControl"
      />
    </li>
    <li>
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
        <input type="checkbox" v-model="locked" :disabled="!hosting" />
        <span />
      </label>
    </li>
    <li>
      <i
        :class="[{ disabled: !playable }, playing ? 'fa-pause-circle' : 'fa-play-circle', 'fas', 'play']"
        @click.stop.prevent="toggleMedia"
      />
    </li>
    <li>
      <div class="volume">
        <i
          :class="[volume === 0 || muted ? 'fa-volume-mute' : 'fa-volume-up', 'fas']"
          @click.stop.prevent="toggleMute"
        />
        <input type="range" min="0" max="100" v-model="volume" />
      </div>
    </li>
  </ul>
</template>

<style lang="scss" scoped>
  ul {
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    list-style: none;

    li {
      font-size: 24px;
      cursor: pointer;

      i {
        padding: 0 5px;

        &.faded {
          color: rgba($color: $text-normal, $alpha: 0.4);
        }

        &.disabled {
          color: rgba($color: $style-error, $alpha: 0.4);
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

        input[type='range'] {
          width: 100%;
          background: transparent;
          width: 150px;
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
            font-family: 'Font Awesome 5 Free';
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
  import { Vue, Component } from 'vue-property-decorator'

  @Component({ name: 'neko-controls' })
  export default class extends Vue {
    get isTouch() {
      return (
        (typeof navigator.maxTouchPoints !== 'undefined' ? navigator.maxTouchPoints < 0 : false) ||
        'ontouchstart' in document.documentElement
      )
    }

    get hosted() {
      return this.$accessor.remote.hosted
    }

    get hosting() {
      return this.$accessor.remote.hosting
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
  }
</script>
