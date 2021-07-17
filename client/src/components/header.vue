<template>
  <div class="header">
    <div class="neko">
      <img src="@/assets/images/logo.svg" alt="n.eko" />
      <span><b>n</b>.eko</span>
    </div>
    <ul class="menu">
      <li>
        <i
          :class="[{ disabled: !admin }, { 'fa-lock-open': !locked }, { 'fa-lock': locked }, 'fas', 'lock']"
          @click="toggleLock"
          v-tooltip="{
            content: admin
              ? locked
                ? $t('room.unlock')
                : $t('room.lock')
              : locked
              ? $t('room.locked')
              : $t('room.unlocked'),
            placement: 'bottom',
            offset: 5,
            boundariesElement: 'body',
            delay: { show: 300, hide: 100 },
          }"
        />
      </li>
      <li>
        <span v-if="showBadge" class="badge">&bull;</span>
        <i class="fas fa-bars toggle" @click="toggleMenu" />
      </li>
    </ul>
  </div>
</template>

<style lang="scss" scoped>
  .header {
    flex: 1;
    display: flex;
    flex-direction: row;
    align-items: center;

    .neko {
      flex: 1;
      display: flex;
      justify-content: flex-start;
      align-items: center;
      width: 150px;
      margin-left: 20px;

      img {
        display: block;
        float: left;
        height: 30px;
        margin-right: 10px;
      }

      span {
        font-size: 30px;
        line-height: 30px;

        b {
          font-weight: 900;
        }
      }
    }

    .menu {
      justify-self: flex-end;
      margin-right: 10px;
      white-space: nowrap;

      li {
        display: inline-block;
        margin-right: 10px;

        i {
          display: block;
          width: 30px;
          height: 30px;
          text-align: center;
          line-height: 32px;
          border-radius: 3px;
          cursor: pointer;
        }

        .disabled {
          cursor: default;
          opacity: 0.8;
        }

        .fa-lock {
          color: rgba($color: $style-error, $alpha: 0.5);
        }

        .toggle {
          background: $background-primary;
        }

        .badge {
          position: absolute;
          background: red;
          font-weight: bold;
          font-size: 1.25em;
          border-radius: 50%;
          width: 20px;
          height: 20px;
          text-align: center;
          line-height: 20px;
          pointer-events: none;

          transform: translate(-50%, -25%) scale(1);
          box-shadow: 0 0 0 0 rgba(0, 0, 0, 1);
          animation: badger-pulse 2s infinite;
        }

        @keyframes badger-pulse {
          0% {
            transform: translate(-50%, -25%) scale(0.85);
            box-shadow: 0 0 0 0 rgba(0, 0, 0, 0.7);
          }

          70% {
            transform: translate(-50%, -25%) scale(1);
            box-shadow: 0 0 0 10px rgba(0, 0, 0, 0);
          }

          100% {
            transform: translate(-50%, -25%) scale(0.85);
            box-shadow: 0 0 0 0 rgba(0, 0, 0, 0);
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
    get admin() {
      return this.$accessor.user.admin
    }

    get locked() {
      return this.$accessor.locked
    }

    get side() {
      return this.$accessor.client.side
    }

    get texts() {
      return this.$accessor.chat.texts
    }

    get showBadge() {
      return !this.side && this.readTexts != this.texts
    }

    readTexts: number = 0
    toggleMenu() {
      this.$accessor.client.toggleSide()
      this.readTexts = this.texts
    }

    toggleLock() {
      if (this.admin) {
        if (this.locked) {
          this.$accessor.unlock()
        } else {
          this.$accessor.lock()
        }
      }
    }
  }
</script>
