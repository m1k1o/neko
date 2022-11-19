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
        <p v-if="downloads.length > 0">{{ $t('files.downloads') }}</p>
        <div v-for="download in downloads" :key="download.id" class="transfers-list-item">
          <div class="transfer-info">
            <i class="fas transfer-status" :class="{ 'fa-arrows-rotate': download.status !== 'completed', 'fa-check': download.status === 'completed' }"></i>
            <p>{{ download.name }}</p>
            <p class="file-size">{{ Math.min(100, Math.round(download.progress / download.size * 100))}}%</p>
            <i class="fas fa-xmark remove-transfer" @click="() => removeTransfer(download)"></i>
          </div>
          <progress class="transfer-progress" :aria-label="download.name + ' progress'" :value="download.progress"
          :max="download.size"></progress>
        </div>
        <p v-if="uploads.length > 0">{{ $t('files.uploads' )}}</p>
        <div v-for="upload in uploads" :key="upload.id" class="transfers-list-item">
          <div class="transfer-info">
            <i class="fas transfer-status" :class="{ 'fa-arrows-rotate': upload.status !== 'completed', 'fa-check': upload.status === 'completed' }"></i>
            <p>{{ upload.name }}</p>
            <p class="file-size">{{ Math.min(100, Math.round(upload.progress / upload.size * 100))}}%</p>
            <i class="fas fa-xmark remove-transfer" @click="() => removeTransfer(upload)"></i>
          </div>
          <progress class="transfer-progress" :aria-label="upload.name + ' progress'" :value="upload.progress"
          :max="upload.size"></progress>
        </div>
      </div>
      <div class="upload-area" :class="{ 'upload-area-drag': uploadAreaDrag }"
      @dragover.prevent="() => uploadAreaDrag = true" @dragleave.prevent="() => uploadAreaDrag = false"
      @drop.prevent="(e) => upload(e.dataTransfer)" @click="openFileBrowser">
        <i class="fas fa-file-arrow-up" />
        <p>{{ $t('files.upload_here') }}</p>
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

    .file-icon, .transfer-status {
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
      max-height: 50vh;
      overflow-y: scroll;
      overflow-x: hidden;
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

    .upload-area:hover {
      cursor: pointer;
    }

    .upload-area-drag, .upload-area:hover {
      background-color: rgba($color: #fff, $alpha: 0.10);
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

    public uploadAreaDrag: boolean = false;

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
      if (this.downloads.map((t) => t.name).includes(item.name)) {
        return
      }

      const url = `/file?pwd=${this.$accessor.password}&filename=${item.name}`
      let transfer: FileTransfer = {
        id: Math.round(Math.random() * 10000),
        name: item.name,
        direction: 'download',
        // this may be smaller than the actual transfer amount, but for large files the
        // content length is not sent (chunked transfer)
        size: item.size,
        progress: 0,
        status: 'pending',
        axios: null,
        abortController: null
      }
      transfer.abortController = new AbortController()
      transfer.axios = this.$http.get(url, {
        responseType: 'blob',
        signal: transfer.abortController.signal,
        onDownloadProgress: (x) => {
          transfer.progress = x.loaded

          if (x.lengthComputable && transfer.size !== x.total) {
            transfer.size = x.total
          }
          if (transfer.progress === transfer.size) {
            transfer.status = 'completed'
          } else if (transfer.status !== 'inprogress') {
            transfer.status = 'inprogress'
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

        transfer.progress = transfer.size
        transfer.status = 'completed'
      }).catch((err) => {
        this.$log.error(err)
      })
      this.$accessor.files.addTransfer(transfer)
    }

    upload(dt: DataTransfer) {
      this.uploadAreaDrag = false

      for (const file of dt.files) {
        const formdata = new FormData()
        formdata.append("files", file, file.name)

        const url = `/file?pwd=${this.$accessor.password}`
        let transfer: FileTransfer = {
          id: Math.round(Math.random() * 10000),
          name: file.name,
          direction: 'upload',
          size: file.size,
          progress: 0,
          status: 'pending',
          axios: null,
          abortController: null
        }
        transfer.abortController = new AbortController()
        this.$http.post(url, formdata, {
          onUploadProgress: (x: any) => {
            transfer.progress = x.loaded

            if (transfer.size !== x.total) {
              transfer.size = x.total
            }
            if (transfer.progress === transfer.size) {
              transfer.status = 'completed'
            } else if (transfer.status !== 'inprogress') {
              transfer.status = 'inprogress'
            }
          }
        }).catch((err) => {
          this.$log.error(err)
        })
        this.$accessor.files.addTransfer(transfer)
      }
    }

    openFileBrowser() {
      const input = document.createElement('input')
      input.type = 'file'
      input.setAttribute('multiple', 'true')
      input.click()

      input.onchange = (e) => {
        if (e === null) {
          return
        }
        const dt = new DataTransfer()
        const target = e.target as any
        for (const f of target.files) {
          dt.items.add(f)
        }
        this.upload(dt)
      }
    }

    removeTransfer(transfer: FileTransfer) {
      if (transfer.status !== 'completed') {
        transfer.abortController?.abort()
      }
      this.$accessor.files.removeTransfer(transfer)
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
        case 'aac':
        case 'flac':
        case 'midi':
        case 'mp3':
        case 'ogg':
        case 'wav':
          className += 'fa-music'
          break
        case 'mkv':
        case 'mov':
        case 'mpeg':
        case 'mp4':
        case 'webm':
          className += 'fa-film'
          break
        case 'bmp':
        case 'gif':
        case 'jpeg':
        case 'jpg':
        case 'png':
        case 'svg':
        case 'tiff':
        case 'webp':
          className += 'fa-image'
          break;
        default:
          className += 'fa-file'
      }
      return className
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

  }

</script>
