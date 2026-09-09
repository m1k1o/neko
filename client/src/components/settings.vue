<template>
  <div class="settings">
    <div class="settings-scroll">
      <section class="settings-section">
        <h3>{{ $t('setting.group_playback') }}</h3>
        <ul>
          <li>
            <span>{{ $t('setting.scroll') }}</span>
            <label class="slider">
              <input type="range" min="1" max="100" v-model="scroll" />
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
        </ul>
      </section>
      <section class="settings-section">
        <h3>{{ $t('setting.group_chat') }}</h3>
        <ul>
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
          <li v-if="$accessor.openinapp.enabled">
            <span>{{ $t('setting.links_in_app') }}</span>
            <label class="switch">
              <input type="checkbox" v-model="links_in_app" />
              <span />
            </label>
          </li>
        </ul>
      </section>
      <section class="settings-section">
        <h3>{{ $t('setting.group_input') }}</h3>
        <ul>
          <li>
            <span>{{ $t('setting.keyboard_layout') }}</span>
            <label class="select">
              <select v-model="keyboard_layout">
                <option v-for="(name, code) in keyboard_layouts_list" :key="code" :value="code">{{ name }}</option>
              </select>
              <span />
            </label>
          </li>
        </ul>
      </section>
      <section class="settings-section" v-if="admin">
        <h3>{{ $t('setting.group_admin') }}</h3>
        <ul>
          <li class="broadcast">
            <div>
              <span>{{ $t('setting.broadcast_title') }}</span>
              <button
                v-if="!broadcast_is_active"
                @click.stop.prevent="$accessor.settings.broadcastCreate(broadcast_url)"
              >
                <i class="fas fa-play"></i>
              </button>
              <button v-else @click.stop.prevent="$accessor.settings.broadcastDestroy()" class="btn-red">
                <i class="fas fa-stop"></i>
              </button>
            </div>
            <input
              v-model="broadcast_url"
              :disabled="broadcast_is_active"
              class="input"
              :placeholder="$t('setting.broadcast_placeholder')"
            />
          </li>
        </ul>
      </section>
      <section class="settings-section" v-if="connected">
        <h3>{{ $t('setting.group_session') }}</h3>
        <ul>
          <li>
            <button @click.stop.prevent="logout">{{ $t('logout') }}</button>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .settings {
    flex: 1;
    display: flex;
    min-width: 0;

    .settings-scroll {
      flex: 1;
      min-width: 0;
      overflow-y: auto;
      padding: 2px 0 8px;
      scrollbar-width: thin;
    }

    .settings-section {
      margin-bottom: 14px;

      &:last-child {
        margin-bottom: 0;
      }

      h3 {
        margin: 6px 12px 4px;
        color: $text-muted;
        font-size: 10px;
        font-weight: 700;
        letter-spacing: 0.12em;
        line-height: 16px;
        text-transform: uppercase;
      }

      ul {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 2px;
        padding: 4px 6px;

        li {
          display: flex;
          flex-direction: row;
          align-content: center;
          justify-content: center;
          border-bottom: 1px solid rgba($text-normal, 0.08);
          padding: 10px 8px;
          white-space: nowrap;
          border-radius: 8px;
          transition: background 0.18s ease;

          &:hover {
            background: $background-modifier-hover;
          }

          &:last-child {
            border-bottom: none;
          }

          span {
            margin-right: auto;
            height: 24px;
            line-height: 24px;
            color: $interactive-normal;
          }

          button {
            cursor: pointer;
            border-radius: 9px;
            padding: 4px;
            background: $style-primary;
            color: $background-tertiary;
            text-align: center;
            text-transform: uppercase;
            font-weight: bold;
            line-height: 28px;
            margin: 4px 0;
            border: none;
            display: block;
            width: 100%;
            transition: background 0.18s ease, transform 0.18s ease;

            &:hover {
              background: lighten($style-primary, 5%);
              transform: translateY(-1px);
            }
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
              background-color: rgba($background-tertiary, 0.86);
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

          .select {
            max-width: 120px;
            text-align: right;

            select:hover {
              border: 1px solid $background-secondary;
            }

            select {
              -webkit-appearance: none;
              -moz-appearance: none;
              appearance: none;
              display: block;
              width: 100%;
              max-width: 100%;
              height: 30px;
              text-align: right;
              padding: 0 5px 0 10px;
              margin: 0;
              line-height: 30px;
              font-weight: bold;
              font-size: 12px;
              text-overflow: ellipsis;
              border: 1px solid transparent;
              border-radius: 5px;
              color: white;
              background-color: $background-tertiary;
              font-weight: lighter;
              cursor: pointer;

              option {
                font-weight: normal;
                color: $text-normal;
                background-color: $background-tertiary;
              }
            }
          }

          .input {
            display: block;
            height: 30px;
            text-align: right;
            padding: 0 10px;
            margin-left: 10px;
            line-height: 30px;
            text-overflow: ellipsis;
            border: 1px solid transparent;
            border-radius: 5px;
            color: white;
            background-color: $background-tertiary;
            font-weight: lighter;
            user-select: auto;

            &::selection {
              background: $text-normal;
            }

            &[disabled] {
              background: none;
            }
          }

          &.broadcast {
            display: flex;
            flex-direction: column;

            div {
              margin-bottom: 10px;
              display: flex;
              justify-content: space-between;

              button {
                flex-shrink: 1;
                width: auto !important;
                margin: 0;
                padding: 0 10px;

                &.btn-red {
                  background: #a62626;
                }
              }
            }

            .input {
              text-align: left;
              width: auto !important;
              margin: 0;
            }
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Watch, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-settings' })
  export default class extends Vue {
    private broadcast_url: string = ''

    get admin() {
      return this.$accessor.user.admin
    }

    get connected() {
      return this.$accessor.connection.connected
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

    get links_in_app() {
      return this.$accessor.settings.links_in_app
    }

    set links_in_app(value: boolean) {
      this.$accessor.settings.setLinksInApp(value)
    }

    get keyboard_layouts_list() {
      return this.$accessor.settings.keyboard_layouts_list
    }

    get keyboard_layout() {
      return this.$accessor.settings.keyboard_layout
    }

    get broadcast_is_active() {
      return this.$accessor.settings.broadcast_is_active
    }

    get broadcast_url_remote() {
      return this.$accessor.settings.broadcast_url
    }

    @Watch('broadcast_url_remote', { immediate: true })
    onBroadcastUrlChange() {
      this.broadcast_url = this.broadcast_url_remote
    }

    set keyboard_layout(value: string) {
      this.$accessor.settings.setKeyboardLayout(value)
      this.$accessor.remote.changeKeyboard()
    }

    logout() {
      this.$accessor.logout()
    }
  }
</script>
