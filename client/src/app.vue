<template>
  <div id="neko" :class="[side ? 'expanded' : '']">
    <template v-if="!$client.supported">
      <neko-unsupported />
    </template>
    <template v-else>
      <main class="neko-main">
        <div v-if="!hideControls" class="header-container">
          <neko-header />
        </div>
        <div class="video-container">
          <neko-video ref="video" :hideControls="hideControls" @control-attempt="controlAttempt" />
        </div>
        <div v-if="!hideControls" class="room-container">
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
      <neko-side v-if="!hideControls && side" />
      <neko-connect v-if="!connected" />
      <neko-about v-if="about" />
      <notifications
        v-if="!hideControls"
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

    .neko-main {
      min-width: 360px;
      max-width: 100%;
      flex-grow: 1;
      flex-direction: column;
      display: flex;
      overflow: auto;

      .header-container {
        background: $background-tertiary;
        height: $menu-height;
        flex-shrink: 0;
        display: flex;
      }

      .video-container {
        background: rgba($color: #000, $alpha: 0.4);
        max-width: 100%;
        flex-grow: 1;
        display: flex;
      }

      .room-container {
        background: $background-tertiary;
        height: $controls-height;
        max-width: 100%;
        flex-shrink: 0;
        flex-direction: column;
        display: flex;

        .room-menu {
          max-width: 100%;
          flex: 1;
          display: flex;

          .settings {
            margin-left: 10px;
            flex: 1;
            justify-content: flex-start;
            align-items: center;
            display: flex;
          }

          .controls {
            flex: 1;
            justify-content: center;
            align-items: center;
            display: flex;
          }

          .emotes {
            margin-right: 10px;
            flex: 1;
            justify-content: flex-end;
            align-items: center;
            display: flex;
          }
        }
      }
    }
  }

  @media only screen and (max-width: 600px) {
    #neko.expanded {
      .neko-main {
        transform: translateX(calc(-100% + 65px));

        video {
          display: none;
        }
      }

      .neko-menu {
        position: absolute;
        top: 0;
        right: 0;
        bottom: 0;
        left: 65px;
        width: calc(100% - 65px);
      }
    }
  }

  @media only screen and (max-width: 768px) {
    #neko .neko-main .room-container {
      display: none;
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

    get hideControls() {
      return !!new URL(location.href).searchParams.get('cast')
    }

    @Watch('hideControls', { immediate: true })
    onHideControls() {
      this.$accessor.video.setMuted(false)
      this.$accessor.settings.setSound(false)
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
      return this.$accessor.connected
    }
  }
</script>
