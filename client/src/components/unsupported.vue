<template>
  <div class="unsupported">
    <div class="window">
      <div class="logo">
        <img src="@/assets/images/logo.svg" alt="n.eko" />
        <span><b>n</b>.eko</span>
      </div>
      <div class="message">
        <span>{{ $t('unsupported') }}</span>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .unsupported {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;

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
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-unsupported' })
  export default class extends Vue {}
</script>
