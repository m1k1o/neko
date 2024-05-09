<template>
  <ul>
    <li>
      <i
        :class="[!can_host ? 'disabled' : '', !hosting ? 'faded' : '', 'fas', 'fa-keyboard', 'request']"
        @click.stop.prevent="toggleControl"
      />
    </li>
    <li>
      <label class="switch">
        <input type="checkbox" v-model="locked" />
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
    <li>
      <i class="fa-sign-out-alt fas" @click.stop.prevent="disconnect" />
    </li>
  </ul>
</template>

<style lang="scss" scoped>
  @import '../assets/styles/_variables.scss';

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

<script lang="ts" setup>
import { computed } from 'vue'
import type Neko from '@/component/main.vue'

const props = defineProps<{
  neko: typeof Neko
}>()

const can_host = computed(() => props.neko.connected)
const hosting = computed(() => props.neko.controlling)
const volume = computed({
  get: () => props.neko.state.video.volume * 100,
  set: (volume: number) => {
    props.neko.setVolume(volume / 100)
  },
})
const muted = computed(() => props.neko.state.video.muted || props.neko.state.video.volume === 0)
const playing = computed(() => props.neko.state.video.playing)
const playable = computed(() => props.neko.state.video.playable)
const locked = computed({
  get: () => props.neko.state.control.locked,
  set: (lock: boolean) => {
    if (lock) {
      props.neko.control.lock()
    } else {
      props.neko.control.unlock()
    }
  },
})

function toggleControl() {
  if (can_host.value && hosting.value) {
    props.neko.room.controlRelease()
  }

  if (can_host.value && !hosting.value) {
    props.neko.room.controlRequest()
  }
}

function toggleMedia() {
  if (playable.value && playing.value) {
    props.neko.pause()
  }

  if (playable.value && !playing.value) {
    props.neko.play()
  }
}

function toggleMute() {
  if (playable.value && muted.value) {
    props.neko.unmute()
  }

  if (playable.value && !muted.value) {
    props.neko.mute()
  }
}

function disconnect() {
  props.neko.logout()
}
</script>
