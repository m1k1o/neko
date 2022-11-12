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
        <p class="file-size">{{ fileSize(item.size) }}</p>
        <i v-if="item.type !== 'dir'" class="fas fa-download download"
        @click="() => download(item)" />
      </div>
    </div>
    <div class="transfer-area">
      <div class="transfers" v-if="transfers.length > 0">
        <p>Downloads</p>
        <div v-for="download in downloads" :key="download.name" class="transfers-list-item">
          <div class="transfer-info">
            <p>{{ download.name }}</p>
            <p class="file-size">{{ Math.max(100, Math.round(download.progress / download.size * 100))}}%</p>
            <i class="fas fa-xmark remove-transfer" @click="() => removeTransfer(download)"></i>
          </div>
          <progress class="transfer-progress" :aria-label="download.name + ' progress'" :value="download.progress"
          :max="download.size"></progress>
        </div>
      </div>
      <div class="upload-area" @dragover.prevent @drop.prevent="onFileDrop">
        <i class="fas fa-file-arrow-up" />
        <p>Drag files here to upload</p>
      </div>
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

    .refresh {
      margin-left: auto;
    }

    .file-size {
      margin-left: auto;
      margin-right: 0.5em;
      color: rgba($color: #fff, $alpha: 0.40);
    }

    .refresh:hover, .download:hover, .remove-transfer:hover {
      cursor: pointer;
    }

    .transfer-area {
      margin-top: auto;
    }

    .transfers {
      margin: 10px 10px 10px 10px;
      background-color: rgba($color: #fff, $alpha: 0.05);
      border-radius: 5px;
    }

    .transfers > p {
      padding: 10px;
      font-weight: 600;
    }

    .transfer-info {
      display: flex;
      flex-direction: row;
      max-width: 100%;
      padding: 10px;
    }

    .transfer-progress {
      margin: 0px 10px 10px 10px;
      width: 95%;
    }

    .upload-area {
      display: flex;
      flex-direction: column;
      text-align: center;
      justify-content: center;
      margin: 10px 10px 10px 10px;
      background-color: rgba($color: #fff, $alpha: 0.05);
      border-radius: 5px;
    }

    .upload-area > i {
      font-size: 4em;
      margin: 10px 10px 10px 10px;
    }

    .upload-area > p {
      margin: 0px 10px 10px 10px;
    }

  }
</style>

<script lang="ts">

  import { Component, Vue } from 'vue-property-decorator'

  import Markdown from './markdown'
  import Content from './context.vue'
import { FileTransfer } from '~/neko/types'

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

    get transfers() {
      return this.$accessor.files.transfers
    }

    get downloads() {
      return this.$accessor.files.transfers.filter((t => t.direction === 'download'))
    }

    get uploads() {
      return this.$accessor.files.transfers.filter((t => t.direction === 'upload'))
    }
    
    refresh() {
      this.$accessor.files.refresh()
    }

    download(item: any) {
      const url = `/file?pwd=${this.$accessor.password}&filename=${item.name}`
      let transfer: FileTransfer = {
        id: Math.round(Math.random() * 10000),
        name: item.name,
        direction: 'download',
        // this is just an estimation, but for large files the content length
        // is not sent (chunked transfer)
        size: item.size,
        progress: 0,
        status: 'pending',
        axios: null,
        // TODO add support for aborting in progress requests, requires axios >=0.22
        abortController: null
      }
      transfer.axios = this.$http.get(url, {
        responseType: 'blob',
        onDownloadProgress: (x) => {
          transfer.progress = x.loaded

          if (x.lengthComputable) {
            transfer.size = x.total
          }
          if (transfer.progress === transfer.size) {
            transfer.status = 'completed'
          }
        }
      }).then((res) => {
        const url = window.URL
        .createObjectURL(new Blob([res.data]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', item.name)
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
      }).catch((err) => {
        this.$log.error(err)
      })
      this.$accessor.files.addTransfer(transfer)
    }

    removeTransfer(item: FileTransfer) {
      console.log(item)
      this.$accessor.files.removeTransfer(item)
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

    fileSize(size: number) {
      if (size < 1000) {
        return `${size} b`
      }
      if (size < 1000 ** 2) {
        return `${(size / 1000).toFixed(2)} kb`
      }
      if (size < 1000 ** 3) {
        return `${(size / 1000 ** 2).toFixed(2)} mb`
      }
      if (size < 1000 ** 4) {
        return `${(size / 1000 ** 3).toFixed(2)} gb`
      }
      return `${(size / 1000 ** 4).toFixed(3)} tb`
    }

    onFileDrop(e: any) {
      console.log('file dropped', e)
      console.log(e.dataTransfer.files)
    }
  }

</script>
