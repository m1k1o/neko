<template>
  <aside class="neko-menu">
    <nav class="tabs-container" aria-label="Room panel">
      <ul>
        <li>
          <button
            type="button"
            role="tab"
            :aria-selected="tab === 'chat'"
            :class="{ active: tab === 'chat' }"
            @click.stop.prevent="change('chat')"
          >
            <i class="fas fa-comment-alt" aria-hidden="true" />
            <span>{{ $t('side.chat') }}</span>
          </button>
        </li>
        <li v-if="filetransferAllowed">
          <button
            type="button"
            role="tab"
            :aria-selected="tab === 'files'"
            :class="{ active: tab === 'files' }"
            @click.stop.prevent="change('files')"
          >
            <i class="fas fa-file" aria-hidden="true" />
            <span>{{ $t('side.files') }}</span>
          </button>
        </li>
        <li>
          <button
            type="button"
            role="tab"
            :aria-selected="tab === 'settings'"
            :class="{ active: tab === 'settings' }"
            @click.stop.prevent="change('settings')"
          >
            <i class="fas fa-sliders-h" aria-hidden="true" />
            <span>{{ $t('side.settings') }}</span>
          </button>
        </li>
      </ul>
    </nav>
    <div class="page-container">
      <neko-chat v-if="tab === 'chat'" />
      <neko-files v-if="tab === 'files'" />
      <neko-settings v-if="tab === 'settings'" />
    </div>
  </aside>
</template>

<style lang="scss">
  .neko-menu {
    width: $side-width;
    background: linear-gradient(180deg, rgba($background-primary, 0.98), rgba($background-secondary, 0.98));
    border-left: 1px solid rgba($text-normal, 0.08);
    box-shadow: -16px 0 40px rgba(2, 8, 18, 0.18);
    flex-shrink: 0;
    max-height: 100%;
    max-width: 100%;
    display: flex;
    flex-direction: column;

    .tabs-container {
      background: rgba($background-tertiary, 0.66);
      min-height: $menu-height;
      max-height: 100%;
      max-width: 100%;
      display: flex;
      flex-shrink: 0;

      ul {
        width: 100%;
        height: 100%;
        display: flex;
        align-items: stretch;
        padding: 10px 12px;
        gap: 6px;

        li {
          flex: 1;

          button {
            width: 100%;
            height: 100%;
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 7px;
            border: 1px solid transparent;
            border-radius: 9px;
            padding: 8px 10px;
            color: $text-muted;
            background: transparent;
            font-weight: 600;
            cursor: pointer;
            transition: color 0.18s ease, background 0.18s ease, border-color 0.18s ease;

            i {
              font-size: 12px;
            }

            &:hover,
            &:focus-visible {
              color: $interactive-hover;
              background: $background-modifier-hover;
            }

            &.active {
              color: $style-primary;
              background: $background-modifier-selected;
              border-color: rgba($style-primary, 0.2);
            }
          }
        }
      }
    }

    .page-container {
      max-height: 100%;
      flex-grow: 1;
      display: flex;
      overflow: auto;
      padding: 10px 12px 12px;
      min-height: 0;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Watch } from 'vue-property-decorator'

  import Settings from '~/components/settings.vue'
  import Chat from '~/components/chat.vue'
  import Files from '~/components/files.vue'

  @Component({
    name: 'neko-side',
    components: {
      'neko-settings': Settings,
      'neko-chat': Chat,
      'neko-files': Files,
    },
  })
  export default class extends Vue {
    get canDownload() {
      return this.$accessor.user.admin || this.$accessor.files.userDownload
    }

    get canUpload() {
      return this.$accessor.user.admin || this.$accessor.files.userUpload
    }

    get filetransferAllowed() {
      return (
        this.$accessor.remote.fileTransfer &&
        (this.$accessor.user.admin || !this.$accessor.isLocked('file_transfer')) &&
        (this.canDownload || this.canUpload)
      )
    }

    get tab() {
      return this.$accessor.client.tab
    }

    @Watch('tab', { immediate: true })
    @Watch('filetransferAllowed', { immediate: true })
    onTabChange() {
      // do not show the files tab if file transfer is disabled
      if (this.tab === 'files' && !this.filetransferAllowed) {
        this.change('chat')
      }
    }

    @Watch('filetransferAllowed')
    onFileTransferAllowedChange() {
      if (this.filetransferAllowed) {
        this.$accessor.files.refresh()
      }
    }

    change(tab: string) {
      this.$accessor.client.setTab(tab)
    }
  }
</script>
