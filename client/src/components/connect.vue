<template>
  <div class="connect">
    <div class="window">
      <div class="logo">
        <img src="@/assets/images/logo.svg" alt="n.eko" />
        <span><b>n</b>.eko</span>
      </div>
      <form class="message" v-if="!connecting" @submit.stop.prevent="connect">
        <span>{{ $t('connect.title') }}</span>
        <input type="text" :placeholder="$t('connect.displayname')" v-model="displayname" />
        <input type="password" :placeholder="$t('connect.password')" v-model="password" />
        <button type="submit" @click.stop.prevent="login">
          {{ $t('connect.connect') }}
        </button>
      </form>
      <div class="loader" v-if="connecting">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .connect {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba($color: $background-floating, $alpha: 0.8);

    display: flex;
    justify-content: center;
    align-items: center;

    .window {
      width: 300px;
      background: $background-secondary;
      border-radius: 5px;
      padding: 10px;

      .logo {
        width: 100%;
        display: flex;
        flex-direction: row;
        justify-content: center;
        align-items: center;

        img {
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
          display: block;
          text-align: center;
          text-transform: uppercase;
          line-height: 30px;
        }

        input {
          border: none;
          padding: 6px 8px;
          line-height: 20px;
          border-radius: 5px;
          margin: 5px 0;
          background: $background-tertiary;
          color: $text-normal;

          &::selection {
            background: $text-link;
          }
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
        }
      }

      .loader {
        width: 90px;
        height: 90px;
        position: relative;
        margin: 0 auto;

        .bounce1,
        .bounce2 {
          width: 100%;
          height: 100%;
          border-radius: 50%;
          background-color: $style-primary;
          opacity: 0.6;
          position: absolute;
          top: 0;
          left: 0;

          -webkit-animation: bounce 2s infinite ease-in-out;
          animation: bounce 2s infinite ease-in-out;
        }

        .bounce2 {
          -webkit-animation-delay: -1s;
          animation-delay: -1s;
        }
      }
    }

    @keyframes bounce {
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
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'
  import { get, set } from '~/utils/localstorage'

  @Component({ name: 'neko-connect' })
  export default class extends Vue {
    private displayname = ''
    private password = ''

    mounted() {
      if (this.$accessor.displayname !== '' && this.$accessor.password !== '') {
        this.$accessor.login({ displayname: this.$accessor.displayname, password: this.$accessor.password })
      }
    }

    get connecting() {
      return this.$accessor.connecting
    }

    login() {
      this.$accessor.login({ displayname: this.displayname, password: this.password })
    }
  }
</script>
