<template>
  <div class="header">
    <a
      href="https://github.com/m1k1o/neko"
      :title="$t('ui.github_repository')"
      target="_blank"
      rel="noopener noreferrer"
      class="neko"
    >
      <span class="brand-mark"><img src="@/assets/images/logo.svg" alt="" /></span>
      <span class="brand-copy">
        <span class="brand-name"><b>n</b>.eko</span>
        <span class="brand-caption">{{ $t('ui.remote_browser') }}</span>
      </span>
    </a>
    <div class="connection-indicator" :class="[connectionState, networkQuality]" role="status" :title="connectionTitle">
      <span class="status-dot" aria-hidden="true" />
      <span class="connection-copy">
        <span class="connection-label">{{ connectionLabel }}</span>
        <span v-if="networkRtt !== null" class="connection-rtt">{{ networkRtt }} ms</span>
      </span>
      <span
        v-if="connected && networkQuality !== 'unknown'"
        class="quality-bars"
        :class="networkQuality"
        aria-hidden="true"
      >
        <i v-for="bar in 3" :key="bar" />
      </span>
    </div>
    <ul class="menu" role="toolbar" :aria-label="$t('ui.room_controls')">
      <li>
        <button
          type="button"
          :class="[{ disabled: !admin }, { locked: isLocked('control') }, 'icon-button']"
          @click="toggleLock('control')"
          :aria-label="lockedTooltip('control')"
          v-tooltip="{
            content: lockedTooltip('control'),
            placement: 'bottom',
            offset: 5,
            boundariesElement: 'body',
            delay: { show: 300, hide: 100 },
          }"
        >
          <i class="fas fa-mouse" aria-hidden="true" />
        </button>
      </li>
      <li>
        <button
          type="button"
          :class="[{ disabled: !admin }, { locked: isLocked('login') }, 'icon-button']"
          @click="toggleLock('login')"
          :aria-label="lockedTooltip('login')"
          v-tooltip="{
            content: lockedTooltip('login'),
            placement: 'bottom',
            offset: 5,
            boundariesElement: 'body',
            delay: { show: 300, hide: 100 },
          }"
        >
          <i :class="[locked ? 'fa-lock' : 'fa-lock-open', 'fas']" aria-hidden="true" />
        </button>
      </li>
      <li v-if="fileTransfer">
        <button
          type="button"
          :class="[{ disabled: !admin }, { locked: isLocked('file_transfer') }, 'icon-button']"
          @click="toggleLock('file_transfer')"
          :aria-label="lockedTooltip('file_transfer')"
          v-tooltip="{
            content: lockedTooltip('file_transfer'),
            placement: 'bottom',
            offset: 5,
            boundariesElement: 'body',
            delay: { show: 300, hide: 100 },
          }"
        >
          <i class="fas fa-file" aria-hidden="true" />
        </button>
      </li>
      <li>
        <span v-if="showBadge" class="badge">&bull;</span>
        <button type="button" class="icon-button toggle" :aria-label="$t('ui.toggle_room_panel')" @click="toggleMenu">
          <i class="fas fa-bars" aria-hidden="true" />
        </button>
      </li>
    </ul>
  </div>
</template>

<style lang="scss" scoped>
  .header {
    flex: 1;
    min-width: 0;
    padding: 0 18px;
    display: flex;
    flex-direction: row;
    align-items: center;

    .neko {
      flex: 1;
      min-width: 0;
      display: flex;
      justify-content: flex-start;
      align-items: center;
      color: $text-normal;
      text-decoration: none;
      gap: 10px;

      .brand-mark {
        width: 34px;
        height: 34px;
        display: grid;
        place-items: center;
        border: 1px solid rgba($style-primary, 0.35);
        border-radius: 10px;
        background: rgba($style-primary, 0.12);

        img {
          display: block;
          height: 24px;
        }
      }

      .brand-copy {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 2px;

        .brand-name {
          font-size: 20px;
          font-weight: 600;
          line-height: 20px;
          letter-spacing: 0.01em;

          b {
            color: $style-primary;
            font-weight: 800;
          }
        }

        .brand-caption {
          color: $text-muted;
          font-size: 9px;
          font-weight: 700;
          letter-spacing: 0.14em;
          line-height: 10px;
        }
      }
    }

    .connection-indicator {
      min-width: 118px;
      display: flex;
      align-items: center;
      gap: 8px;
      margin-right: 14px;
      padding: 7px 10px;
      border: 1px solid rgba($text-normal, 0.08);
      border-radius: 10px;
      background: rgba($background-primary, 0.52);

      .status-dot {
        width: 8px;
        height: 8px;
        flex: 0 0 auto;
        border-radius: 50%;
        background: $interactive-muted;
        box-shadow: 0 0 0 3px rgba($interactive-muted, 0.12);
      }

      .connection-copy {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 2px;

        .connection-label {
          overflow: hidden;
          color: $interactive-normal;
          font-size: 11px;
          font-weight: 600;
          line-height: 13px;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        .connection-rtt {
          color: $text-muted;
          font-size: 9px;
          line-height: 10px;
        }
      }

      &.connected .status-dot {
        background: $style-primary;
        box-shadow: 0 0 0 3px rgba($style-primary, 0.14);
      }

      &.connecting .status-dot,
      &.reconnecting .status-dot {
        background: $style-warning;
        box-shadow: 0 0 0 3px rgba($style-warning, 0.14);
        animation: status-pulse 1.5s ease-in-out infinite;
      }

      &.poor .status-dot {
        background: $style-error;
        box-shadow: 0 0 0 3px rgba($style-error, 0.14);
      }

      .quality-bars {
        height: 16px;
        display: flex;
        align-items: flex-end;
        gap: 2px;
        margin-left: auto;

        i {
          width: 3px;
          border-radius: 2px;
          background: $interactive-muted;

          &:nth-child(1) {
            height: 5px;
          }

          &:nth-child(2) {
            height: 9px;
          }

          &:nth-child(3) {
            height: 13px;
          }
        }

        &.good i {
          background: $style-primary;
        }

        &.fair i:nth-child(-n + 2) {
          background: $style-warning;
        }

        &.poor i:nth-child(1) {
          background: $style-error;
        }
      }
    }

    .menu {
      justify-self: flex-end;
      margin-right: 10px;
      white-space: nowrap;

      li {
        display: inline-block;
        margin-left: 8px;

        .icon-button {
          display: grid;
          place-items: center;
          width: 36px;
          height: 36px;
          border: 1px solid transparent;
          border-radius: 10px;
          color: $interactive-normal;
          background: transparent;
          cursor: pointer;
          transition: color 0.18s ease, background 0.18s ease, border-color 0.18s ease, transform 0.18s ease;

          &:hover:not(.disabled),
          &:focus-visible:not(.disabled) {
            color: $interactive-hover;
            background: $background-modifier-hover;
            border-color: rgba($text-normal, 0.12);
            transform: translateY(-1px);
          }

          i {
            font-size: 15px;
          }
        }

        .disabled {
          cursor: default;
          opacity: 0.38;
        }

        .locked {
          color: rgba($color: $style-error, $alpha: 0.5);
        }

        .toggle {
          color: $style-primary;
          background: rgba($style-primary, 0.12) !important;
          border-color: rgba($style-primary, 0.2) !important;
        }

        .badge {
          position: absolute;
          background: red;
          font-weight: bold;
          font-size: 1.25em;
          border-radius: 50%;
          width: 18px;
          height: 18px;
          text-align: center;
          line-height: 18px;
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

    @keyframes status-pulse {
      50% {
        opacity: 0.35;
      }
    }
  }

  @media only screen and (max-width: 600px) {
    .header {
      padding: 0 10px;

      .brand-caption {
        display: none !important;
      }

      .connection-indicator {
        min-width: 34px;
        width: 34px;
        height: 34px;
        justify-content: center;
        margin-right: 4px;
        padding: 0;

        .connection-copy {
          display: none;
        }

        .quality-bars {
          display: none;
        }
      }

      .menu li {
        margin-left: 2px;
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'
  import { AdminLockResource } from '~/neko/messages'

  @Component({ name: 'neko-header' })
  export default class extends Vue {
    get connectionState() {
      return this.$accessor.connectionState
    }

    get networkQuality() {
      return this.$accessor.networkQuality
    }

    get networkRtt() {
      return this.$accessor.networkRtt
    }

    get connected() {
      return this.$accessor.connected
    }

    get connectionLabel() {
      if (this.connectionState === 'connected') {
        return this.$t('connection.connected')
      }
      if (this.connectionState === 'connecting' || this.connectionState === 'reconnecting') {
        return this.$t('connection.reconnecting')
      }
      return this.$t('connection.disconnected')
    }

    get connectionTitle() {
      if (this.networkRtt !== null && this.networkQuality !== 'unknown') {
        return `${this.connectionLabel} · ${this.networkRtt} ms · ${this.$t(
          'connection.network_' + this.networkQuality,
        )}`
      }
      return this.connectionLabel as string
    }

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

    get fileTransfer() {
      return this.$accessor.remote.fileTransfer
    }

    toggleLock(resource: AdminLockResource) {
      this.$accessor.toggleLock(resource)
    }

    isLocked(resource: AdminLockResource): boolean {
      return this.$accessor.isLocked(resource)
    }

    readTexts: number = 0
    toggleMenu() {
      this.$accessor.client.toggleSide()
      this.readTexts = this.texts
    }

    lockedTooltip(resource: AdminLockResource) {
      if (this.admin) {
        return this.$t(`locks.${resource}.` + (this.isLocked(resource) ? `unlock` : `lock`))
      }

      return this.$t(`locks.${resource}.` + (this.isLocked(resource) ? `locked` : `unlocked`))
    }
  }
</script>
