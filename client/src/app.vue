<template>
  <div id="neko" class="app-shell" :class="[!videoOnly && side ? 'expanded' : '']">
    <template v-if="!$client.supported">
      <neko-unsupported />
    </template>
    <template v-else>
      <main class="neko-main">
        <div v-if="!videoOnly" class="header-container">
          <neko-header />
        </div>
        <div class="video-container" :class="{ 'video-only': videoOnly }">
          <div class="video-surface">
            <neko-video
              ref="video"
              :hideControls="hideControls"
              :extraControls="isEmbedMode"
              @control-attempt="controlAttempt"
            />
          </div>
        </div>
        <div v-if="!videoOnly" class="room-container">
          <neko-members />
          <div class="room-menu">
            <div class="settings">
              <neko-menu />
            </div>
            <div class="controls">
              <neko-controls :shakeKbd="shakeKbd" />
            </div>
            <div class="emotes">
              <neko-emotes />
            </div>
          </div>
        </div>
      </main>
      <div v-if="!videoOnly && side" class="panel-backdrop" @click.stop.prevent="closeSide" />
      <neko-side v-if="!videoOnly && side" />
      <neko-connect v-if="!connected" />
      <neko-about v-if="about" />
      <notifications
        v-if="!videoOnly"
        group="neko"
        position="top left"
        style="top: 50px; pointer-events: none"
        :ignoreDuplicates="true"
      />
    </template>
  </div>
</template>

<style lang="scss">
  #neko {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    max-width: 100vw;
    max-height: 100vh;
    flex-direction: row;
    display: flex;
    background: radial-gradient(circle at 20% -10%, rgba($style-primary, 0.1), transparent 36%), $background-tertiary;
    color: $text-normal;

    .neko-main {
      min-width: 360px;
      max-width: 100%;
      flex-grow: 1;
      min-height: 0;
      flex-direction: column;
      display: flex;
      overflow: hidden;
      background: rgba($background-primary, 0.58);

      .header-container {
        background: rgba($background-tertiary, 0.86);
        border-bottom: 1px solid rgba($text-normal, 0.08);
        height: $menu-height;
        flex-shrink: 0;
        display: flex;
        position: relative;
        z-index: 2;
      }

      .video-container {
        background: linear-gradient(145deg, rgba($background-tertiary, 0.92), rgba($background-primary, 0.58));
        max-width: 100%;
        flex-grow: 1;
        min-height: 0;
        min-width: 0;
        padding: 16px;
        display: flex;

        .video-surface {
          position: relative;
          flex: 1;
          min-width: 0;
          min-height: 0;
          display: flex;
          overflow: hidden;
          border: 1px solid rgba($text-normal, 0.1);
          border-radius: 16px;
          background: #050a12;
          box-shadow: $elevation-high;
        }

        &.video-only {
          padding: 0;
          background: #000;

          .video-surface {
            border: 0;
            border-radius: 0;
            box-shadow: none;
          }
        }
      }

      .room-container {
        background: rgba($background-tertiary, 0.92);
        border-top: 1px solid rgba($text-normal, 0.08);
        height: $controls-height;
        min-height: $controls-height;
        max-width: 100%;
        flex-shrink: 0;
        flex-direction: column;
        display: flex;
        padding: 8px 16px 12px;
        overflow: visible;

        > .members {
          flex: 0 0 64px;
          height: 64px;
          min-height: 64px;
          position: relative;
          z-index: 1;
        }

        .room-menu {
          max-width: 100%;
          flex: 1 1 auto;
          min-height: 56px;
          display: flex;
          gap: 12px;
          position: relative;
          z-index: 2;

          .settings {
            margin-left: 0;
            flex: 1;
            min-width: 0;
            justify-content: flex-start;
            align-items: center;
            display: flex;
          }

          .controls {
            flex: 1;
            min-width: 220px;
            min-height: 56px;
            padding: 0 12px;
            border: 1px solid rgba($text-normal, 0.08);
            border-radius: 14px;
            background: rgba($background-primary, 0.64);
            justify-content: center;
            align-items: center;
            display: flex;
            overflow-x: auto;
            overflow-y: hidden;

            > ul {
              display: flex;
              flex: 0 0 auto;
              min-width: max-content;
              justify-content: center;
            }
          }

          .emotes {
            margin-right: 0;
            flex: 1;
            min-width: 0;
            justify-content: flex-end;
            align-items: center;
            display: flex;
          }
        }
      }
    }
  }

  .panel-backdrop {
    display: none;
  }

  @media only screen and (max-width: 1024px) {
    html,
    body {
      overflow-y: auto !important;
      width: auto !important;
      height: auto !important;
    }

    body > p {
      display: none;
    }

    #neko {
      position: relative;
      flex-direction: column;
      max-height: initial !important;

      .neko-main {
        height: 100vh;
      }

      .video-container {
        padding: 10px;
      }

      .neko-menu {
        position: fixed;
        top: 0;
        right: 0;
        bottom: 0;
        z-index: 10;
        height: 100vh;
        width: min(100%, 420px) !important;
      }

      .panel-backdrop {
        display: block;
        position: fixed;
        inset: 0;
        z-index: 9;
        background: rgba($background-tertiary, 0.68);
        backdrop-filter: blur(3px);
      }
    }
  }

  @media only screen and (max-width: 1024px) and (orientation: portrait) {
    #neko {
      &.expanded .neko-main {
        height: 100vh;
      }

      &.expanded > .neko-menu {
        height: 100vh;
        width: min(100%, 420px) !important;
      }
    }
  }

  @media only screen and (max-width: 768px) {
    #neko .neko-main .room-container {
      display: none;
    }

    #neko .neko-main .video-container {
      padding: 8px;
    }

    #neko .neko-main .video-container .video-surface {
      border-radius: 12px;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Ref, Watch } from 'vue-property-decorator'

  import Connect from '~/components/connect.vue'
  import Video from '~/components/video.vue'
  import Menu from '~/components/menu.vue'
  import Side from '~/components/side.vue'
  import Controls from '~/components/controls.vue'
  import Members from '~/components/members.vue'
  import Emotes from '~/components/emotes.vue'
  import About from '~/components/about.vue'
  import Header from '~/components/header.vue'
  import Unsupported from '~/components/unsupported.vue'

  @Component({
    name: 'neko',
    components: {
      'neko-connect': Connect,
      'neko-video': Video,
      'neko-menu': Menu,
      'neko-side': Side,
      'neko-controls': Controls,
      'neko-members': Members,
      'neko-emotes': Emotes,
      'neko-about': About,
      'neko-header': Header,
      'neko-unsupported': Unsupported,
    },
  })
  export default class extends Vue {
    @Ref('video') video!: Video

    shakeKbd = false

    get volume() {
      const numberParam = parseFloat(new URL(location.href).searchParams.get('volume') || '1.0')
      return Math.max(0.0, Math.min(!isNaN(numberParam) ? numberParam * 100 : 100, 100))
    }

    get scroll() {
      const numberParam = parseInt(new URL(location.href).searchParams.get('scroll') || '', 10)
      return Math.max(1, Math.min(!isNaN(numberParam) ? numberParam : 10, 100))
    }

    get isCastMode() {
      return !!new URL(location.href).searchParams.get('cast')
    }

    get isEmbedMode() {
      return !!new URL(location.href).searchParams.get('embed')
    }

    get hideControls() {
      return this.isCastMode
    }

    get videoOnly() {
      return this.isCastMode || this.isEmbedMode
    }

    @Watch('volume', { immediate: true })
    onVolume(volume: number) {
      if (new URL(location.href).searchParams.has('volume')) {
        this.$accessor.video.setVolume(volume)
      }
    }

    @Watch('scroll', { immediate: true })
    onScroll(scroll: number) {
      if (new URL(location.href).searchParams.has('scroll')) {
        this.$accessor.settings.setScroll(scroll)
      }
    }

    @Watch('hideControls', { immediate: true })
    onHideControls(enabled: boolean) {
      if (enabled) {
        this.$accessor.video.setMuted(false)
        this.$accessor.settings.setSound(false)
      }
    }

    @Watch('side')
    onSide(side: boolean) {
      if (side) {
        console.log('side enabled')
        // scroll to the side
        this.$nextTick(() => {
          const side = document.querySelector('aside')
          if (side) {
            side.scrollIntoView({ behavior: 'smooth', block: 'start' })
          }
        })
      }
    }

    controlAttempt() {
      if (this.shakeKbd || this.$accessor.remote.hosted) return

      this.shakeKbd = true
      window.setTimeout(() => (this.shakeKbd = false), 5000)
    }

    get about() {
      return this.$accessor.client.about
    }

    get side() {
      return this.$accessor.client.side
    }

    get connected() {
      return this.$accessor.connection.connected
    }

    closeSide() {
      this.$accessor.client.setSide(false)
    }
  }
</script>
