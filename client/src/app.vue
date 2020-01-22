<template>
  <div id="neko">
    <main class="neko-main">
      <div class="header-container">
        <div class="neko">
          <img src="@/assets/logo.svg" alt="n.eko" />
          <span><b>n</b>.eko</span>
        </div>
        <i class="fas fa-bars toggle" @click="toggle" />
      </div>
      <div class="video-container">
        <neko-video ref="video" />
      </div>
      <div class="room-container">
        <neko-members />
        <div class="room-menu">
          <div class="settings">
            <neko-menu />
          </div>
          <div class="controls">
            <neko-controls />
          </div>
          <div class="emoji">
            <neko-emoji />
          </div>
        </div>
      </div>
    </main>
    <neko-side v-if="side" />
    <neko-connect v-if="!connected" />
    <neko-about v-if="about" />
    <notifications group="neko" position="top left" />
  </div>
</template>

<style lang="scss">
  #neko {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    flex-direction: row;
    display: flex;

    .neko-main {
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

        .toggle {
          display: block;
          width: 30px;
          height: 30px;
          text-align: center;
          line-height: 32px;
          background: $background-primary;
          justify-self: flex-end;
          border-radius: 3px;
          margin: 5px 10px 0 0;
          cursor: pointer;
        }

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

          .emoji {
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
</style>

<script lang="ts">
  import { Vue, Component, Ref } from 'vue-property-decorator'

  import Connect from '~/components/connect.vue'
  import Video from '~/components/video.vue'
  import Menu from '~/components/menu.vue'
  import Side from '~/components/side.vue'
  import Controls from '~/components/controls.vue'
  import Members from '~/components/members.vue'
  import Emoji from '~/components/emoji.vue'
  import About from '~/components/about.vue'

  @Component({
    name: 'neko',
    components: {
      'neko-connect': Connect,
      'neko-video': Video,
      'neko-menu': Menu,
      'neko-side': Side,
      'neko-controls': Controls,
      'neko-members': Members,
      'neko-emoji': Emoji,
      'neko-about': About,
    },
  })
  export default class extends Vue {
    @Ref('video') video!: Video

    get about() {
      return this.$accessor.client.about
    }

    get side() {
      return this.$accessor.client.side
    }

    get connected() {
      return this.$accessor.connected
    }

    toggle() {
      this.$accessor.client.toggleSide()
    }
  }
</script>
