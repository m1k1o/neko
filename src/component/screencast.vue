<template>
  <img :src="imageSrc" @load="onImageLoad" @error="onImageError" />
</template>

<script lang="ts">
  import { Vue, Component, Watch, Prop } from 'vue-property-decorator'
  import { RoomApi } from './api'

  const REFRESH_RATE = 1e3
  const ERROR_DELAY_MS = 2500

  @Component({
    name: 'neko-screencast',
  })
  export default class extends Vue {
    private imageSrc = ''
    private running = false
    private continue = false

    @Prop()
    private readonly image!: string

    @Watch('image')
    setImage(image: string) {
      this.imageSrc = image
    }

    @Prop()
    private readonly enabled!: boolean

    @Prop()
    private readonly api!: RoomApi

    async loop() {
      if (this.running) return
      this.running = true

      while (this.continue) {
        const lastLoad = Date.now()

        try {
          const res = await this.api.screenCastImage({ responseType: 'blob' })
          this.imageSrc = URL.createObjectURL(res.data)

          const delay = lastLoad - Date.now() + REFRESH_RATE
          if (delay > 0) {
            await new Promise((res) => setTimeout(res, delay))
          }
        } catch {
          await new Promise((res) => setTimeout(res, ERROR_DELAY_MS))
        }
      }

      this.running = false
      this.imageSrc = ''
    }

    mounted() {
      if (this.enabled) {
        this.start()
      }
    }

    beforeDestroy() {
      this.stop()
    }

    start() {
      this.continue = true

      if (!this.running) {
        setTimeout(this.loop, 0)
      }
    }

    stop() {
      this.continue = false
    }

    @Watch('enabled')
    onEnabledChanged(enabled: boolean) {
      if (enabled) {
        this.start()
      } else {
        this.stop()
      }
    }

    onImageLoad() {
      URL.revokeObjectURL(this.imageSrc)
      this.$emit('imageReady', this.running)
    }

    onImageError() {
      if (this.imageSrc) URL.revokeObjectURL(this.imageSrc)
      this.$emit('imageReady', false)
    }
  }
</script>
