<template>
  <ul>
    <li>
      <i
        :class="[
          hosted && !hosting ? 'disabled' : '',
          !hosted && !hosting ? 'faded' : '',
          'fas',
          'fa-keyboard',
          'request',
        ]"
        @click.stop.prevent="toggleControl"
      />
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
        padding: 0 10px;

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
          width: 200px;
          height: 20px;
          -webkit-appearance: none;

          &::-moz-range-thumb {
            height: 12px;
            width: 12px;
            border-radius: 12px;
            background: $interactive-active;
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
            background: $interactive-active;
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
    }
  }
</style>

<script lang="ts">
  import { Vue, Component } from 'vue-property-decorator'

  @Component({ name: 'neko-controls' })
  export default class extends Vue {
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
