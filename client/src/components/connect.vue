<template>
  <div class="connect">
    <div class="window" role="dialog" aria-modal="true" aria-labelledby="connect-title">
      <div class="window-topline">
        <span class="eyebrow">{{ $t('ui.room_eyebrow') }}</span>
        <span class="secure-badge"><i class="fas fa-shield-halved" aria-hidden="true" /> {{ $t('ui.secure') }}</span>
      </div>
      <div class="logo" :title="$t('ui.about')" @click.stop.prevent="about">
        <span class="logo-mark"><img src="@/assets/images/logo.svg" alt="" /></span>
        <span class="brand-name"><b>n</b>.eko</span>
      </div>
      <form class="message" v-if="!connecting" @submit.stop.prevent="login">
        <h1 id="connect-title">{{ autoPassword ? $t('connect.invitation_title') : $t('connect.login_title') }}</h1>
        <label class="field">
          <span>{{ $t('connect.displayname') }}</span>
          <input
            type="text"
            :placeholder="$t('connect.displayname')"
            v-model="displayname"
            autocomplete="nickname"
            autofocus
            :aria-invalid="loginError ? 'true' : 'false'"
            :class="{ invalid: loginError }"
            @input="loginError = ''"
          />
        </label>
        <label class="field" v-if="!autoPassword">
          <span>{{ $t('connect.password') }}</span>
          <div class="password-field">
            <input
              :type="showPassword ? 'text' : 'password'"
              :placeholder="$t('connect.password')"
              v-model="password"
              autocomplete="current-password"
              @input="loginError = ''"
            />
            <button
              type="button"
              class="password-toggle"
              :aria-label="$t(showPassword ? 'connect.hide_password' : 'connect.show_password')"
              @click.stop.prevent="showPassword = !showPassword"
            >
              <i :class="[showPassword ? 'fa-eye-slash' : 'fa-eye', 'fas']" aria-hidden="true" />
            </button>
          </div>
        </label>
        <p v-if="loginError" class="field-error" role="alert">
          <i class="fas fa-circle-exclamation" aria-hidden="true" />
          {{ loginError }}
        </p>
        <button class="primary-button" type="submit">
          <span>{{ $t('connect.connect') }}</span>
          <i class="fas fa-arrow-right" aria-hidden="true" />
        </button>
        <button class="about-link" type="button" @click.stop.prevent="about">{{ $t('ui.about') }}</button>
      </form>
      <div class="loader" v-if="connecting" role="status" aria-live="polite">
        <div class="spinner" />
        <span>{{ $t('connection.reconnecting') }}</span>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .connect {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 20;
    padding: 20px;
    background: rgba($color: $background-tertiary, $alpha: 0.76);
    backdrop-filter: blur(16px);

    display: flex;
    justify-content: center;
    align-items: center;

    .window {
      width: min(100%, 390px);
      background: linear-gradient(155deg, rgba($background-secondary, 0.98), rgba($background-primary, 0.98));
      border: 1px solid rgba($text-normal, 0.12);
      border-radius: 20px;
      padding: 28px;
      box-shadow: $elevation-high;

      .window-topline {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 28px;

        .eyebrow {
          color: $text-muted;
          font-size: 10px;
          font-weight: 700;
          letter-spacing: 0.16em;
        }

        .secure-badge {
          color: $style-primary;
          font-size: 11px;
          font-weight: 600;

          i {
            margin-right: 4px;
          }
        }
      }

      .logo {
        width: 100%;
        display: flex;
        flex-direction: row;
        justify-content: flex-start;
        align-items: center;
        cursor: pointer;
        gap: 12px;
        margin-bottom: 30px;

        .logo-mark {
          width: 48px;
          height: 48px;
          display: grid;
          place-items: center;
          border-radius: 14px;
          background: rgba($style-primary, 0.14);
          border: 1px solid rgba($style-primary, 0.3);

          img {
            height: 32px;
          }
        }

        .brand-name {
          color: $text-normal;
          font-size: 28px;
          font-weight: 600;
          line-height: 32px;

          b {
            color: $style-primary;
            font-weight: 800;
          }
        }
      }

      .message {
        display: flex;
        flex-direction: column;

        h1 {
          color: $interactive-hover;
          font-size: 24px;
          font-weight: 600;
          letter-spacing: -0.02em;
          line-height: 30px;
          margin-bottom: 8px;
        }

        .subtitle {
          color: $text-muted;
          line-height: 20px;
          margin-bottom: 20px;
        }

        .field {
          display: flex;
          flex-direction: column;
          gap: 7px;
          margin-bottom: 14px;

          span {
            color: $interactive-normal;
            font-size: 12px;
            font-weight: 600;
          }

          input {
            width: 100%;
            border: 1px solid rgba($text-normal, 0.12);
            padding: 11px 13px;
            line-height: 20px;
            border-radius: 10px;
            background: rgba($background-tertiary, 0.75);
            color: $text-normal;
            transition: border-color 0.18s ease, box-shadow 0.18s ease;

            &:hover {
              border-color: rgba($text-normal, 0.24);
            }

            &:focus {
              border-color: $style-primary;
              box-shadow: 0 0 0 3px rgba($style-primary, 0.14);
              outline: none;
            }

            &::placeholder {
              color: $text-muted;
            }

            &.invalid {
              border-color: $style-error;
              box-shadow: 0 0 0 3px rgba($style-error, 0.12);
            }
          }

          .password-field {
            position: relative;

            input {
              padding-right: 42px;
            }

            .password-toggle {
              position: absolute;
              top: 50%;
              right: 8px;
              width: 30px;
              height: 30px;
              display: grid;
              place-items: center;
              transform: translateY(-50%);
              border: 0;
              border-radius: 7px;
              color: $text-muted;
              background: transparent;
              cursor: pointer;

              &:hover,
              &:focus-visible {
                color: $interactive-hover;
                background: $background-modifier-hover;
              }
            }
          }
        }

        .field-error {
          display: flex;
          align-items: center;
          gap: 7px;
          margin: -2px 0 6px;
          color: $style-error;
          font-size: 12px;
          line-height: 18px;
        }

        .primary-button {
          width: 100%;
          display: flex;
          justify-content: center;
          align-items: center;
          gap: 10px;
          cursor: pointer;
          border-radius: 10px;
          padding: 11px 14px;
          background: $style-primary;
          color: $background-tertiary;
          font-size: 14px;
          font-weight: 700;
          line-height: 20px;
          margin: 8px 0 12px;
          border: none;
          transition: transform 0.18s ease, box-shadow 0.18s ease, background 0.18s ease;

          &:hover {
            background: lighten($style-primary, 5%);
            box-shadow: 0 8px 20px rgba($style-primary, 0.18);
            transform: translateY(-1px);
          }
        }

        .about-link {
          display: block;
          margin: 0 auto;
          border: 0;
          color: $text-muted;
          background: transparent;
          cursor: pointer;
          font-size: 12px;

          &:hover {
            color: $text-link;
          }
        }
      }

      .loader {
        min-height: 250px;
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        gap: 16px;
        color: $text-muted;

        .spinner {
          width: 44px;
          height: 44px;
          border: 3px solid rgba($style-primary, 0.2);
          border-top-color: $style-primary;
          border-radius: 50%;
          animation: spin 0.9s linear infinite;
        }
      }
    }

    @keyframes spin {
      to {
        transform: rotate(360deg);
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-connect' })
  export default class extends Vue {
    private autoPassword: string | null = new URL(location.href).searchParams.get('pwd')

    private displayname: string = ''
    private password: string = ''
    private loginError: string = ''
    private showPassword: boolean = false

    mounted() {
      // auto-password fill
      let password = this.$accessor.password
      if (this.autoPassword !== null) {
        this.removeUrlParam('pwd')
        password = this.autoPassword
      }

      // auto-user fill
      let displayname = this.$accessor.displayname
      const usr = new URL(location.href).searchParams.get('usr')
      if (usr) {
        this.removeUrlParam('usr')
        displayname = this.$accessor.displayname || usr
      }

      if (displayname !== '' && password !== '') {
        this.$accessor.login({ displayname, password })
        this.autoPassword = null
      }
    }

    get connecting() {
      return this.$accessor.connecting
    }

    removeUrlParam(param: string) {
      let url = document.location.href
      let urlparts = url.split('?')

      if (urlparts.length >= 2) {
        let urlBase = urlparts.shift()
        let queryString = urlparts.join('?')

        let prefix = encodeURIComponent(param) + '='
        let pars = queryString.split(/[&;]/g)
        for (let i = pars.length; i-- > 0; ) {
          if (pars[i].lastIndexOf(prefix, 0) !== -1) {
            pars.splice(i, 1)
          }
        }

        url = urlBase + (pars.length > 0 ? '?' + pars.join('&') : '')
        window.history.pushState('', document.title, url)
      }
    }

    login() {
      let password = this.password
      if (this.autoPassword !== null) {
        password = this.autoPassword
      }

      if (this.displayname == '') {
        this.loginError = this.$t('connect.empty_displayname') as string
        return
      }

      this.loginError = ''
      this.$accessor.login({ displayname: this.displayname, password })
      this.autoPassword = null
    }

    about() {
      this.$accessor.client.toggleAbout()
    }
  }
</script>
