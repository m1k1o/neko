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
    active = false

    @Prop()
    private readonly api!: RoomApi

    async loop() {
      while (this.active) {
        const lastLoad = Date.now()

        if (this._image.src) {
          URL.revokeObjectURL(this._image.src)
        }

        const res = await this.api.screenCastImage({ responseType: 'blob' })
        this._image.src = URL.createObjectURL(res.data)

        const delay = lastLoad - Date.now() + REFRESH_RATE
        if (delay > 0) {
          await new Promise((res) => setTimeout(res, delay))
        }
      }
    }

    async mounted() {
      this.active = true

      setTimeout(this.loop, 0)
    }

    beforeDestroy() {
      this.active = false

      if (this._image.src) {
        URL.revokeObjectURL(this._image.src)
      }
    }
  }
</script>
