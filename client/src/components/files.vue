<template>
  <div class="files">
    <div class="files-cwd">
      <p>{{ cwd }}</p>
      <i class="fas fa-rotate-right refresh" @click="refresh" />
    </div>
    <div class="files-list">
      <div v-for="item in files" :key="item.name" class="files-list-item">
        <i :class="fileIcon(item)" />
        <p class="file-name" :title="item.name">{{ item.name }}</p>
        <p class="file-size">{{ fileSize(item.size) }}</p>
        <i v-if="item.type !== 'dir'" class="fas fa-download download" @click="download(item)" />
      </div>
    </div>
    <div class="transfer-area">
      <div class="transfers" v-if="transfers.length > 0">
        <p v-if="downloads.length > 0" class="transfers-list-header">
          <span>{{ $t('files.downloads') }}</span>
          <i class="fas fa-xmark remove-transfer" @click="downloads.forEach((t) => removeTransfer(t))"></i>
        </p>
        <div v-for="download in downloads" :key="download.id" class="transfers-list-item">
          <div class="transfer-info">
            <i
              class="fas transfer-status"
              :class="{
                'fa-clock': download.status === 'pending',
                'fa-arrows-rotate': download.status === 'inprogress',
                'fa-check': download.status === 'completed',
                'fa-warning': download.status === 'failed',
              }"
            ></i>
            <p class="file-name" :title="download.name">{{ download.name }}</p>
            <p class="file-size">{{ Math.min(100, Math.round((download.progress / download.size) * 100)) }}%</p>
            <i class="fas fa-xmark remove-transfer" @click="removeTransfer(download)"></i>
          </div>
          <div v-if="download.status === 'failed'" class="transfer-error">{{ download.error }}</div>
          <progress
            v-else
            class="transfer-progress"
            :aria-label="download.name + ' progress'"
            :value="download.progress"
            :max="download.size"
          ></progress>
        </div>
        <p v-if="uploads.length > 0" class="transfers-list-header">
          <span>{{ $t('files.uploads') }}</span>
          <i class="fas fa-xmark remove-transfer" @click="uploads.forEach((t) => removeTransfer(t))"></i>
        </p>
        <div v-for="upload in uploads" :key="upload.id" class="transfers-list-item">
          <div class="transfer-info">
            <i
              class="fas transfer-status"
              :title="upload.status"
              :class="{
                'fa-clock': upload.status === 'pending',
                'fa-arrows-rotate': upload.status === 'inprogress',
                'fa-check': upload.status === 'completed',
                'fa-warning': upload.status === 'failed',
              }"
            ></i>
            <p class="file-name" :title="upload.name">{{ upload.name }}</p>
            <p class="file-size">{{ Math.min(100, Math.round((upload.progress / upload.size) * 100)) }}%</p>
            <i class="fas fa-xmark remove-transfer" @click="removeTransfer(upload)"></i>
          </div>
          <div v-if="upload.status === 'failed'" class="transfer-error">{{ upload.error }}</div>
          <progress
            v-else
            class="transfer-progress"
            :aria-label="upload.name + ' progress'"
            :value="upload.progress"
            :max="upload.size"
          ></progress>
        </div>
      </div>
      <div
        class="upload-area"
        :class="{ 'upload-area-drag': uploadAreaDrag }"
        @dragover.prevent="uploadAreaDrag = true"
        @dragleave.prevent="uploadAreaDrag = false"
        @drop.prevent="(e) => upload(e.dataTransfer)"
        @click="openFileBrowser"
      >
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
      border-bottom: 2px solid rgba($color: #fff, $alpha: 0.1);
      display: flex;
      flex-direction: row;
      line-height: 1.2;
    }

    .transfers-list-header {
      display: flex;
      justify-content: space-between;
      border-bottom: 2px solid rgba($color: #fff, $alpha: 0.1);
    }

    .file-icon,
    .transfer-status {
      width: 14px;
      margin-right: 0.5em;
    }

    .transfer-error {
      border: 1px solid $style-error;
      border-radius: 5px;
      padding: 10px;
    }

    .files-list-item:last-child {
      border-bottom: 0px;
    }

    .refresh {
      margin-left: auto;
    }

    .file-name {
      text-overflow: ellipsis;
      overflow: hidden;
      white-space: nowrap;
    }

    .file-size {
      margin-left: auto;
      margin-right: 0.5em;
      color: rgba($color: #fff, $alpha: 0.4);
      white-space: nowrap;
    }

    .refresh:hover,
    .download:hover,
    .remove-transfer:hover {
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

    .upload-area-drag,
    .upload-area:hover {
      background-color: rgba($color: #fff, $alpha: 0.1);
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
  import { FileTransfer, FileListItem } from '~/neko/types'

  @Component({
    name: 'neko-files',
    components: {
      'neko-markdown': Markdown,
      'neko-context': Content,
    },
  })
  export default class extends Vue {
    public uploadAreaDrag: boolean = false

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
      return this.$accessor.files.transfers.filter((t) => t.direction === 'download')
    }

    get uploads() {
      return this.$accessor.files.transfers.filter((t) => t.direction === 'upload')
    }

    refresh() {
      this.$accessor.files.refresh()
    }

    download(item: FileListItem) {
      if (this.downloads.map((t) => t.name).includes(item.name)) {
        return
      }

      const url =
        '/file?pwd=' + encodeURIComponent(this.$accessor.password) + '&filename=' + encodeURIComponent(item.name)
      const abortController = new AbortController()

      let transfer: FileTransfer = {
        id: Math.round(Math.random() * 10000),
        name: item.name,
        direction: 'download',
        // this may be smaller than the actual transfer amount, but for large files the
        // content length is not sent (chunked transfer)
        size: item.size,
        progress: 0,
        status: 'pending',
        abortController: abortController,
      }

      this.$http
        .get(url, {
          responseType: 'blob',
          signal: abortController.signal,
          withCredentials: false,
          onDownloadProgress: (x) => {
            transfer.progress = x.loaded

            if (x.total && transfer.size !== x.total) {
              transfer.size = x.total
            }
            if (transfer.progress === transfer.size) {
              transfer.status = 'completed'
            } else if (transfer.status !== 'inprogress') {
              transfer.status = 'inprogress'
            }
          },
        })
        .then((res) => {
          const url = window.URL.createObjectURL(new Blob([res.data]))
          const link = document.createElement('a')
          link.href = url
          link.setAttribute('download', item.name)
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)

          transfer.progress = transfer.size
          transfer.status = 'completed'
        })
        .catch((error) => {
          this.$log.error(error)

          transfer.status = 'failed'
          transfer.error = error.message
        })

      this.$accessor.files.addTransfer(transfer)
    }

    upload(dt: DataTransfer) {
      const url = '/file?pwd=' + encodeURIComponent(this.$accessor.password)
      this.uploadAreaDrag = false

      for (const file of dt.files) {
        const abortController = new AbortController()

        const formdata = new FormData()
        formdata.append('files', file, file.name)

        let transfer: FileTransfer = {
          id: Math.round(Math.random() * 10000),
          name: file.name,
          direction: 'upload',
          size: file.size,
          progress: 0,
          status: 'pending',
          abortController: abortController,
        }

        this.$http
          .post(url, formdata, {
            signal: abortController.signal,
            withCredentials: false,
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
            },
          })
          .catch((error) => {
            this.$log.error(error)

            transfer.status = 'failed'
            transfer.error = error.message
          })

        this.$accessor.files.addTransfer(transfer)
      }
    }

    openFileBrowser() {
      const input = document.createElement('input')
      input.type = 'file'
      input.setAttribute('multiple', 'true')
      input.onchange = (e: Event) => {
        if (e === null) return

        const dt = new DataTransfer()
        const target = e.target as HTMLInputElement
        if (target.files === null) return

        for (const f of target.files) {
          dt.items.add(f)
        }

        this.upload(dt)
      }
      input.click()
    }

    removeTransfer(transfer: FileTransfer) {
      if (transfer.status !== 'completed') {
        transfer.abortController?.abort()
      }
      this.$accessor.files.removeTransfer(transfer)
    }

    fileIcon(file: FileListItem) {
      let className = 'file-icon fas '
      // if is directory
      if (file.type === 'dir') {
        className += 'fa-folder'
        return className
      }
      // try to get file extension
      const ext = file.name.split('.').pop()
      if (ext === undefined) {
        className += 'fa-file'
        return className
      }
      // try to find icon
      switch (ext.toLowerCase()) {
        case 'txt':
        case 'md':
          className += 'fa-file-text'
          break
        case 'pdf':
          className += 'fa-file-pdf'
          break
        case 'zip':
        case 'rar':
        case '7z':
        case 'gz':
          className += 'fa-archive'
          break
        case 'aac':
        case 'flac':
        case 'midi':
        case 'mp3':
        case 'ogg':
        case 'wav':
          className += 'fa-music'
          break
        case 'avi':
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
          break
        default:
          className += 'fa-file'
      }
      return className
    }

    fileSize(size: number) {
      if (size < 1024) {
        return size + ' B'
      }
      if (size < 1024 * 1024) {
        return Math.round(size / 1024) + ' KB'
      }
      if (size < 1024 * 1024 * 1024) {
        return Math.round(size / (1024 * 1024)) + ' MB'
      }
      if (size < 1024 * 1024 * 1024 * 1024) {
        return Math.round(size / (1024 * 1024 * 1024)) + ' GB'
      }
      return Math.round(size / (1024 * 1024 * 1024 * 1024)) + ' TB'
    }
  }
</script>
