<template>
  <img ref="image" />
</template>

<script lang="ts">
  import { Vue, Component, Ref, Watch, Prop } from 'vue-property-decorator'
  import { RoomApi } from './api'

  const REFRESH_RATE = 1e3

  @Component({
    name: 'neko-screencast',
  })
  export default class extends Vue {
    @Ref('image') readonly _image!: HTMLImageElement
    private running = false
    private continue = false

    @Prop()
    private readonly enabled!: boolean

    @Prop()
    private readonly api!: RoomApi

    async loop() {
      if (this.running) return
      this.running = true

      while (this.continue) {
        const lastLoad = Date.now()

        const res = await this.api.screenCastImage({ responseType: 'blob' })
        const image = URL.createObjectURL(res.data)

        if (this._image) {
          this._image.src = image
        }

        const delay = lastLoad - Date.now() + REFRESH_RATE
        if (delay > 0) {
          await new Promise((res) => setTimeout(res, delay))
        }

        URL.revokeObjectURL(image)
      }

      this.running = false
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
      if (!this.running) {
        this.continue = true
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
  }
</script>
