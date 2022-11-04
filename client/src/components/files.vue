<template>
  <div class="files">
    <div class="files-cwd">
      <p>{{ cwd }}</p>
      <i class="fas fa-rotate-right refresh" @click="refresh" />
    </div>
    <div class="files-list">
      <div v-for="item in files" :key="item.name" class="files-list-item">
        <i :class="fileIcon(item)" />
        <p>{{ item.name }}</p>
        <i v-if="item.type !== 'dir'" class="fas fa-download download"
        @click="() => download(item)" />
      </div>
    </div>
    <div class="files-transfer" @dragover.prevent @drop.prevent="onFileDrop">
      <i class="fas fa-file-arrow-up" />
      <p>Drag files here to upload</p>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .files {
    flex: 1;
    flex-direction: column;
    display: flex;
    max-width: 100%;

    .files-cwd {
      display: flex;
      flex-direction: row;
      margin: 10px 10px 0px 10px;
      padding: 0.5em;
      font-weight: 600;
      background-color: rgba($color: #fff, $alpha: 0.05);
      border-radius: 5px;
    }

    .files-list {
      margin: 10px 10px 10px 10px;
      background-color: rgba($color: #fff, $alpha: 0.05);
      border-radius: 5px;
      overflow-y: scroll;
      scrollbar-width: thin;
      scrollbar-color: $background-tertiary transparent;

      &::-webkit-scrollbar {
        width: 8px;
      }

      &::-webkit-scrollbar-track {
        background-color: transparent;
      }

      &::-webkit-scrollbar-thumb {
        background-color: $background-tertiary;
        border: 2px solid $background-primary;
        border-radius: 4px;
      }

      &::-webkit-scrollbar-thumb:hover {
        background-color: $background-floating;
      }
    }

    .files-list-item {
      padding: 0.5em;
      border-bottom: 2px solid rgba($color: #fff, $alpha: 0.10);
      display: flex;
      flex-direction: row;
    }

    .file-icon {
      width: 14px;
      margin-right: 0.5em;
    }

    .files-list-item:last-child {
      border-bottom: 0px;
    }

    .refresh, .download {
      margin-left: auto;
    }

    .refresh:hover, .download:hover {
      cursor: pointer;
    }

    .files-transfer {
      display: flex;
      flex-direction: column;
      text-align: center;
      justify-content: center;
      margin: auto 10px 10px 10px;
      background-color: rgba($color: #fff, $alpha: 0.05);
      border-radius: 5px;
    }

    .files-transfer > i {
      font-size: 4em;
      margin: 10px 10px 10px 10px;
    }

    .files-transfer > p {
      margin: 0px 10px 10px 10px;
    }

  }
</style>

<script lang="ts">

  import { Component, Vue } from 'vue-property-decorator'

  import Markdown from './markdown'
  import Content from './context.vue'

  @Component({
    name: 'neko-files',
    components: {
      'neko-markdown': Markdown,
      'neko-context': Content,
    }
  })
  export default class extends Vue {

    get cwd() {
      return this.$accessor.files.cwd
    }

    get files() {
      return this.$accessor.files.files
    }
    
    refresh() {
      this.$accessor.files.refresh()
    }

    download(item: any) {
      console.log(item.name);
    }

    fileIcon(file: any) {
      let className = 'file-icon fas '
      if (file.type === 'dir') {
        className += 'fa-folder'
        return className
      }
      const parts = file.name.split('.')
      if (!parts) {
        className += 'fa-file'
        return className
      }
      const ext = parts[parts.length - 1]
      switch (ext) {
        case 'mp3':
        case 'flac':
          className += 'fa-music'
          break;
        case 'webm':
        case 'mp4':
        case 'mkv':
          className += 'fa-film'
          break;
        default:
          className += 'fa-file'
      }
      return className;
    }

    onFileDrop(e: any) {
      console.log('file dropped', e)
      console.log(e.dataTransfer.files)
    }
  }

</script>
