<template>
  <div class="files">
    <div class="files-cwd">
      <p>{{ cwd }}</p>
      <i class="fas fa-rotate-right refresh" @click="refresh" />
    </div>

    <!-- Filter bar -->
    <div class="files-filter">
      <span
        v-for="f in filterOptions"
        :key="f.key"
        class="filter-chip"
        :class="{ active: activeFilter === f.key }"
        @click="activeFilter = f.key"
      >
        <i :class="f.icon" />
        {{ f.label }}
      </span>
    </div>

    <div class="files-list">
      <div v-for="item in filteredFiles" :key="item.name" class="files-list-item">
        <i :class="fileIcon(item)" />
        <template v-if="renaming === item.name">
          <input
            class="rename-input"
            v-model="renameValue"
            @keydown.enter.stop.prevent="confirmRename(item)"
            @keydown.esc.stop.prevent="renaming = null"
            @blur="renaming = null"
            ref="renameInput"
          />
        </template>
        <template v-else>
          <p class="file-name" :title="item.name">{{ item.name }}</p>
          <p class="file-size">{{ fileSize(item.size) }}</p>
          <i
            v-if="isPreviewable(item)"
            class="fas fa-eye action-icon"
            :title="$t('files.preview')"
            @click="openPreview(item)"
          />
          <i
            v-if="item.type !== 'dir'"
            class="fas fa-download action-icon"
            :title="$t('files.download')"
            @click="download(item)"
          />
          <i
            class="fas fa-pen action-icon"
            :title="$t('files.rename')"
            @click="startRename(item)"
          />
          <i
            class="fas fa-trash action-icon"
            :title="$t('files.delete')"
            @click="deleteFile(item)"
          />
        </template>
      </div>
      <div v-if="filteredFiles.length === 0" class="files-empty">
        {{ $t('files.no_files') }}
      </div>
    </div>

    <!-- Preview modal -->
    <div v-if="preview" class="preview-overlay" @click.self="preview = null">
      <div class="preview-modal">
        <div class="preview-header">
          <span class="preview-title">{{ preview.name }}</span>
          <i class="fas fa-xmark preview-close" @click="preview = null" />
        </div>
        <div class="preview-body">
          <img v-if="isImage(preview)" :src="previewUrl" class="preview-image" />
          <video v-else-if="isVideo(preview)" :src="previewUrl" controls class="preview-video" />
          <audio v-else-if="isAudio(preview)" :src="previewUrl" controls class="preview-audio" />
          <iframe v-else-if="isPdf(preview)" :src="previewUrl" class="preview-pdf" />
          <pre v-else-if="previewText !== null" class="preview-text">{{ previewText }}</pre>
          <p v-else class="preview-unsupported">{{ $t('files.preview_unsupported') }}</p>
        </div>
      </div>
    </div>

    <div class="transfer-area">
      <div class="transfer-area-header" @click="transferAreaCollapsed = !transferAreaCollapsed">
        <i class="fas fa-arrows-up-down transfer-area-icon" />
        <span>{{ $t('files.transfers') }}</span>
        <i class="fas transfer-area-toggle" :class="transferAreaCollapsed ? 'fa-chevron-up' : 'fa-chevron-down'" />
      </div>
      <template v-if="!transferAreaCollapsed">
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
      </template>
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

    .files-filter {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
      margin: 6px 10px 0;

      .filter-chip {
        display: flex;
        align-items: center;
        gap: 4px;
        padding: 2px 8px;
        border-radius: 12px;
        font-size: 0.75em;
        cursor: pointer;
        background: rgba($color: #fff, $alpha: 0.07);
        color: rgba($color: #fff, $alpha: 0.6);
        user-select: none;

        &:hover { background: rgba($color: #fff, $alpha: 0.12); }
        &.active { background: rgba($color: #7289da, $alpha: 0.5); color: #fff; }
      }
    }

    .files-list {
      margin: 10px 10px 10px 10px;
      background-color: rgba($color: #fff, $alpha: 0.05);
      border-radius: 5px;
      overflow-y: scroll;
      scrollbar-width: thin;
      scrollbar-color: $background-tertiary transparent;

      &::-webkit-scrollbar { width: 8px; }
      &::-webkit-scrollbar-track { background-color: transparent; }
      &::-webkit-scrollbar-thumb {
        background-color: $background-tertiary;
        border: 2px solid $background-primary;
        border-radius: 4px;
      }
      &::-webkit-scrollbar-thumb:hover { background-color: $background-floating; }
    }

    .files-empty {
      padding: 1em;
      text-align: center;
      color: rgba($color: #fff, $alpha: 0.3);
      font-size: 0.85em;
    }

    .files-list-item {
      padding: 0.5em;
      border-bottom: 2px solid rgba($color: #fff, $alpha: 0.1);
      display: flex;
      flex-direction: row;
      line-height: 1.2;
      &:last-child { border-bottom: 0; }
    }

    .transfers-list-header {
      display: flex;
      justify-content: space-between;
      border-bottom: 2px solid rgba($color: #fff, $alpha: 0.1);
    }

    .file-icon, .transfer-status { width: 14px; margin-right: 0.5em; }

    .transfer-error {
      border: 1px solid $style-error;
      border-radius: 5px;
      padding: 10px;
    }

    .refresh { margin-left: auto; }

    .file-name {
      text-overflow: ellipsis;
      overflow: hidden;
      white-space: nowrap;
    }

    .rename-input {
      flex: 1;
      background: rgba($color: #fff, $alpha: 0.1);
      border: 1px solid rgba($color: #fff, $alpha: 0.3);
      border-radius: 3px;
      color: inherit;
      font-size: inherit;
      padding: 0 4px;
      outline: none;
      min-width: 0;
    }

    .file-size {
      margin-left: auto;
      margin-right: 0.5em;
      color: rgba($color: #fff, $alpha: 0.4);
      white-space: nowrap;
    }

    .action-icon { margin-left: 0.3em; width: 14px; flex-shrink: 0; }

    .refresh:hover, .action-icon:hover, .remove-transfer:hover { cursor: pointer; }

    .transfer-area {
      margin-top: auto;

      .transfer-area-header {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 6px 10px;
        cursor: pointer;
        font-size: 0.8em;
        color: rgba(255, 255, 255, 0.5);
        user-select: none;

        &:hover { color: rgba(255, 255, 255, 0.8); }

        span { flex: 1; }

        .transfer-area-icon { font-size: 0.9em; }
        .transfer-area-toggle { font-size: 0.75em; }
      }
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

      &::-webkit-scrollbar { width: 8px; }
      &::-webkit-scrollbar-track { background-color: transparent; }
      &::-webkit-scrollbar-thumb {
        background-color: $background-tertiary;
        border: 2px solid $background-primary;
        border-radius: 4px;
      }
      &::-webkit-scrollbar-thumb:hover { background-color: $background-floating; }
    }

    .transfers > p { padding: 10px; font-weight: 600; }

    .transfer-info {
      display: flex;
      flex-direction: row;
      max-width: 100%;
      padding: 10px;
    }

    .transfer-progress { margin: 0px 10px 10px 10px; width: 95%; }

    .upload-area {
      display: flex;
      flex-direction: column;
      text-align: center;
      justify-content: center;
      margin: 10px 10px 10px 10px;
      background-color: rgba($color: #fff, $alpha: 0.05);
      border-radius: 5px;
      cursor: pointer;

      &:hover, &.upload-area-drag { background-color: rgba($color: #fff, $alpha: 0.1); }
      > i { font-size: 4em; margin: 10px; }
      > p { margin: 0px 10px 10px 10px; }
    }

    .preview-overlay {
      position: fixed;
      inset: 0;
      background: rgba(0, 0, 0, 0.75);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 1000;
    }

    .preview-modal {
      background: $background-secondary;
      border-radius: 8px;
      display: flex;
      flex-direction: column;
      max-width: 90vw;
      max-height: 90vh;
      min-width: 320px;
      overflow: hidden;
    }

    .preview-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 10px 14px;
      border-bottom: 1px solid rgba($color: #fff, $alpha: 0.1);
      font-weight: 600;

      .preview-title {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: 80%;
      }

      .preview-close { cursor: pointer; opacity: 0.6; &:hover { opacity: 1; } }
    }

    .preview-body {
      overflow: auto;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 10px;
      flex: 1;
    }

    .preview-image { max-width: 80vw; max-height: 75vh; object-fit: contain; border-radius: 4px; }
    .preview-video { max-width: 80vw; max-height: 75vh; }
    .preview-audio { width: 360px; }
    .preview-pdf { width: 80vw; height: 75vh; border: none; }

    .preview-text {
      white-space: pre-wrap;
      word-break: break-all;
      font-size: 0.85em;
      max-width: 80vw;
      max-height: 75vh;
      overflow: auto;
      margin: 0;
      color: rgba($color: #fff, $alpha: 0.85);
    }

    .preview-unsupported { color: rgba($color: #fff, $alpha: 0.4); }
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'

  import Markdown from './markdown'
  import Content from './context.vue'
  import { FileTransfer, FileListItem } from '~/neko/types'

  const IMAGE_EXTS   = ['bmp', 'gif', 'jpeg', 'jpg', 'png', 'svg', 'tiff', 'webp', 'ico']
  const VIDEO_EXTS   = ['avi', 'mkv', 'mov', 'mpeg', 'mp4', 'webm']
  const AUDIO_EXTS   = ['aac', 'flac', 'mp3', 'ogg', 'wav', 'm4a']
  const DOC_EXTS     = ['pdf', 'txt', 'md', 'csv', 'json', 'xml', 'html', 'log', 'yaml', 'yml',
                        'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx']
  const ARCHIVE_EXTS = ['zip', 'rar', '7z', 'gz', 'tar', 'bz2']
  const TEXT_PREVIEW_EXTS = ['txt', 'md', 'csv', 'json', 'xml', 'html', 'log', 'yaml', 'yml',
                             'js', 'ts', 'py', 'sh', 'css', 'scss']

  function getExt(name: string): string {
    return (name.split('.').pop() || '').toLowerCase()
  }

  @Component({
    name: 'neko-files',
    components: {
      'neko-markdown': Markdown,
      'neko-context': Content,
    },
  })
  export default class extends Vue {
    public uploadAreaDrag = false
    public renaming: string | null = null
    public renameValue = ''
    public activeFilter = 'all'
    public preview: FileListItem | null = null
    public previewUrl = ''
    public previewText: string | null = null
    public transferAreaCollapsed = false

    readonly filterOptions = [
      { key: 'all',     label: '全部',   icon: 'fas fa-list' },
      { key: 'image',   label: '图片',   icon: 'fas fa-image' },
      { key: 'video',   label: '视频',   icon: 'fas fa-film' },
      { key: 'audio',   label: '音频',   icon: 'fas fa-music' },
      { key: 'doc',     label: '文档',   icon: 'fas fa-file-alt' },
      { key: 'archive', label: '压缩包', icon: 'fas fa-archive' },
      { key: 'other',   label: '其他',   icon: 'fas fa-file' },
    ]

    get cwd() { return this.$accessor.files.cwd }
    get files() { return this.$accessor.files.files }
    get transfers() { return this.$accessor.files.transfers }
    get downloads() { return this.transfers.filter((t) => t.direction === 'download') }
    get uploads() { return this.transfers.filter((t) => t.direction === 'upload') }

    get filteredFiles(): FileListItem[] {
      if (this.activeFilter === 'all') return this.files
      return this.files.filter((f) => this.fileCategory(f) === this.activeFilter)
    }

    fileCategory(file: FileListItem): string {
      if (file.type === 'dir') return 'other'
      const ext = getExt(file.name)
      if (IMAGE_EXTS.includes(ext))   return 'image'
      if (VIDEO_EXTS.includes(ext))   return 'video'
      if (AUDIO_EXTS.includes(ext))   return 'audio'
      if (DOC_EXTS.includes(ext))     return 'doc'
      if (ARCHIVE_EXTS.includes(ext)) return 'archive'
      return 'other'
    }

    isPreviewable(file: FileListItem): boolean {
      if (file.type === 'dir') return false
      const ext = getExt(file.name)
      return IMAGE_EXTS.includes(ext) || VIDEO_EXTS.includes(ext) ||
             AUDIO_EXTS.includes(ext) || ext === 'pdf' || TEXT_PREVIEW_EXTS.includes(ext)
    }

    isImage(file: FileListItem) { return IMAGE_EXTS.includes(getExt(file.name)) }
    isVideo(file: FileListItem) { return VIDEO_EXTS.includes(getExt(file.name)) }
    isAudio(file: FileListItem) { return AUDIO_EXTS.includes(getExt(file.name)) }
    isPdf(file: FileListItem)   { return getExt(file.name) === 'pdf' }

    async openPreview(item: FileListItem) {
      this.preview = item
      this.previewText = null
      this.previewUrl = ''
      const url = '/file?pwd=' + encodeURIComponent(this.$accessor.password) +
                  '&filename=' + encodeURIComponent(item.name)
      const ext = getExt(item.name)

      if (TEXT_PREVIEW_EXTS.includes(ext)) {
        try {
          const res = await this.$http.get(url, { responseType: 'text', withCredentials: false })
          this.previewText = res.data
        } catch {
          this.previewText = '加载失败'
        }
      } else {
        try {
          const res = await this.$http.get(url, { responseType: 'blob', withCredentials: false })
          this.previewUrl = URL.createObjectURL(res.data)
        } catch {
          this.preview = null
        }
      }
    }

    refresh() { this.$accessor.files.refresh() }

    download(item: FileListItem) {
      if (this.downloads.map((t) => t.name).includes(item.name)) return
      const url = '/file?pwd=' + encodeURIComponent(this.$accessor.password) +
                  '&filename=' + encodeURIComponent(item.name)
      const abortController = new AbortController()
      const transfer: FileTransfer = {
        id: Math.round(Math.random() * 10000),
        name: item.name, direction: 'download',
        size: item.size, progress: 0, status: 'pending', abortController,
      }
      this.$http.get(url, {
        responseType: 'blob', signal: abortController.signal, withCredentials: false,
        onDownloadProgress: (x) => {
          transfer.progress = x.loaded
          if (x.total && transfer.size !== x.total) transfer.size = x.total
          transfer.status = transfer.progress === transfer.size ? 'completed' : 'inprogress'
        },
      }).then((res) => {
        const a = document.createElement('a')
        a.href = URL.createObjectURL(new Blob([res.data]))
        a.setAttribute('download', item.name)
        document.body.appendChild(a); a.click(); document.body.removeChild(a)
        transfer.progress = transfer.size; transfer.status = 'completed'
      }).catch((error) => {
        this.$log.error(error); transfer.status = 'failed'; transfer.error = error.message
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
        const transfer: FileTransfer = {
          id: Math.round(Math.random() * 10000),
          name: file.name, direction: 'upload',
          size: file.size, progress: 0, status: 'pending', abortController,
        }
        this.$http.post(url, formdata, {
          signal: abortController.signal, withCredentials: false,
          onUploadProgress: (x: any) => {
            transfer.progress = x.loaded
            if (transfer.size !== x.total) transfer.size = x.total
            transfer.status = transfer.progress === transfer.size ? 'completed' : 'inprogress'
          },
        }).catch((error) => {
          this.$log.error(error); transfer.status = 'failed'; transfer.error = error.message
        })
        this.$accessor.files.addTransfer(transfer)
      }
    }

    openFileBrowser() {
      const input = document.createElement('input')
      input.type = 'file'
      input.setAttribute('multiple', 'true')
      input.onchange = (e: Event) => {
        if (!e) return
        const dt = new DataTransfer()
        const target = e.target as HTMLInputElement
        if (!target.files) return
        for (const f of target.files) dt.items.add(f)
        this.upload(dt)
      }
      input.click()
    }

    removeTransfer(transfer: FileTransfer) {
      if (transfer.status !== 'completed') transfer.abortController?.abort()
      this.$accessor.files.removeTransfer(transfer)
    }

    async deleteFile(item: FileListItem) {
      const url = '/file?pwd=' + encodeURIComponent(this.$accessor.password) +
                  '&filename=' + encodeURIComponent(item.name)
      try { await this.$http.delete(url, { withCredentials: false }) }
      catch (err: any) { this.$log.error(err) }
    }

    startRename(item: FileListItem) {
      this.renaming = item.name
      this.renameValue = item.name
      this.$nextTick(() => {
        const input = this.$refs.renameInput as HTMLInputElement | HTMLInputElement[]
        const el = Array.isArray(input) ? input[0] : input
        if (el) { el.focus(); el.select() }
      })
    }

    async confirmRename(item: FileListItem) {
      const newName = this.renameValue.trim()
      if (!newName || newName === item.name) { this.renaming = null; return }
      const url = '/file?pwd=' + encodeURIComponent(this.$accessor.password) +
                  '&filename=' + encodeURIComponent(item.name)
      try {
        await this.$http.patch(url, { new_name: newName }, { withCredentials: false })
        this.renaming = null
      } catch (err: any) { this.$log.error(err); this.renaming = null }
    }

    fileIcon(file: FileListItem) {
      const cls = 'file-icon fas '
      if (file.type === 'dir') return cls + 'fa-folder'
      const ext = getExt(file.name)
      if (IMAGE_EXTS.includes(ext))   return cls + 'fa-image'
      if (VIDEO_EXTS.includes(ext))   return cls + 'fa-film'
      if (AUDIO_EXTS.includes(ext))   return cls + 'fa-music'
      if (ARCHIVE_EXTS.includes(ext)) return cls + 'fa-archive'
      switch (ext) {
        case 'pdf':  return cls + 'fa-file-pdf'
        case 'txt': case 'md': case 'log': return cls + 'fa-file-text'
        case 'doc': case 'docx': return cls + 'fa-file-word'
        case 'xls': case 'xlsx': return cls + 'fa-file-excel'
        case 'ppt': case 'pptx': return cls + 'fa-file-powerpoint'
        default: return cls + 'fa-file'
      }
    }

    fileSize(size: number) {
      if (size < 1024) return size + ' B'
      if (size < 1024 ** 2) return Math.round(size / 1024) + ' KB'
      if (size < 1024 ** 3) return Math.round(size / 1024 ** 2) + ' MB'
      if (size < 1024 ** 4) return Math.round(size / 1024 ** 3) + ' GB'
      return Math.round(size / 1024 ** 4) + ' TB'
    }
  }
</script>
