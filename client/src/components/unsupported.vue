<template>
  <div class="unsupported">
    <div class="window" role="alert">
      <div class="logo">
        <span class="logo-mark"><img src="@/assets/images/logo.svg" alt="" /></span>
        <span class="brand-name"><b>n</b>.eko</span>
      </div>
      <div class="message">
        <span class="eyebrow">{{ $t('ui.browser_check') }}</span>
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
      width: min(100% - 40px, 390px);
      background: linear-gradient(155deg, rgba($background-secondary, 0.98), rgba($background-primary, 0.98));
      border: 1px solid rgba($text-normal, 0.12);
      border-radius: 20px;
      padding: 28px;
      box-shadow: $elevation-high;

      .logo {
        width: 100%;
        display: flex;
        flex-direction: row;
        justify-content: flex-start;
        align-items: center;
        gap: 12px;
        margin-bottom: 26px;

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

        .eyebrow {
          color: $text-muted;
          font-size: 10px;
          font-weight: 700;
          letter-spacing: 0.16em;
          margin-bottom: 10px;
        }

        > span:last-child {
          display: block;
          color: $interactive-normal;
          text-align: left;
          line-height: 22px;
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
