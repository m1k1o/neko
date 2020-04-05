<template>
  <div class="settings">
    <ul>
      <li>
        <span>{{ $t('setting.scroll') }}</span>
        <label class="slider">
          <input type="range" min="5" max="100" v-model="scroll" />
        </label>
      </li>
      <li>
        <span>{{ $t('setting.scroll_invert') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="scroll_invert" />
          <span />
        </label>
      </li>
      <li>
        <span>{{ $t('setting.autoplay') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="autoplay" />
          <span />
        </label>
      </li>
      <li>
        <span>{{ $t('setting.ignore_emotes') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="ignore_emotes" />
          <span />
        </label>
      </li>
      <li>
        <span>{{ $t('setting.chat_sound') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="chat_sound" />
          <span />
        </label>
      </li>
      <li v-if="connected">
        <button @click.stop.prevent="logout">{{ $t('logout') }}</button>
      </li>
    </ul>
  </div>
</template>

<style lang="scss" scoped>
  .settings {
    flex: 1;
    display: flex;

    ul {
      flex: 1;
      display: flex;
      flex-direction: column;
      padding: 5px 20px;

      li {
        display: flex;
        flex-direction: row;
        align-content: center;
        justify-content: center;
        border-bottom: 1px solid $background-secondary;
        padding: 5px 0;
        white-space: nowrap;

        &:last-child {
          border-bottom: none;
        }

        span {
          margin-right: auto;
          height: 24px;
          line-height: 24px;
        }

        button {
          cursor: pointer;
          border-radius: 5px;
          padding: 4px;
          background: $style-primary;
          color: $text-normal;
          text-align: center;
          text-transform: uppercase;
          font-weight: bold;
          line-height: 30px;
          margin: 5px 0;
          border: none;
          display: block;
          width: 100%;
        }

        .switch {
          justify-self: flex-end;
          position: relative;
          width: 42px;
          height: 24px;

          input {
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
            background-color: $background-tertiary;
            transition: 0.2s;
            border-radius: 34px;

            &:before {
              position: absolute;
              content: '';
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
          }

          &:checked + span:before {
            transform: translateX(18px);
          }
        }

        .slider {
          white-space: nowrap;
          max-width: 120px;

          input[type='range'] {
            display: inline-block;
            background: transparent;
            appearance: none;
            height: 24px;
            max-width: 120px;

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
              appearance: none;
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
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-settings' })
  export default class extends Vue {
    get connected() {
      return this.$accessor.connected
    }

    get scroll() {
      return this.$accessor.settings.scroll.toString()
    }

    set scroll(value: string) {
      this.$accessor.settings.setScroll(parseInt(value))
    }

    get scroll_invert() {
      return this.$accessor.settings.scroll_invert
    }

    set scroll_invert(value: boolean) {
      this.$accessor.settings.setInvert(value)
    }

    get autoplay() {
      return this.$accessor.settings.autoplay
    }

    set autoplay(value: boolean) {
      this.$accessor.settings.setAutoplay(value)
    }

    get ignore_emotes() {
      return this.$accessor.settings.ignore_emotes
    }

    set ignore_emotes(value: boolean) {
      this.$accessor.settings.setIgnore(value)
    }

    get chat_sound() {
      return this.$accessor.settings.chat_sound
    }

    set chat_sound(value: boolean) {
      this.$accessor.settings.setSound(value)
    }

    logout() {
      this.$accessor.logout()
    }
  }
</script>
