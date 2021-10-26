<template>
  <canvas ref="overlay" class="neko-cursors" tabindex="0" />
</template>

<style lang="scss" scoped>
  .neko-cursors {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 100%;
    height: 100%;
    outline: 0;
  }
</style>

<script lang="ts">
  import { Vue, Component, Ref, Prop, Watch } from 'vue-property-decorator'

  import { SessionCursor } from './types/state'
  import { CursorDrawFunction, Dimension } from './types/cursors'

  const CANVAS_SCALE = 2

  @Component({
    name: 'neko-cursors',
  })
  export default class extends Vue {
    @Ref('overlay') readonly _overlay!: HTMLCanvasElement
    private _ctx!: CanvasRenderingContext2D

    @Prop()
    private readonly sessionId!: string

    @Prop()
    private readonly screenSize!: Dimension

    @Prop()
    private readonly canvasSize!: Dimension

    @Prop()
    private readonly cursors!: SessionCursor[]

    @Prop()
    private readonly cursorDraw!: CursorDrawFunction | null

    mounted() {
      // get canvas overlay context
      const ctx = this._overlay.getContext('2d')
      if (ctx != null) {
        this._ctx = ctx
      }

      // synchronize intrinsic with extrinsic dimensions
      const { width, height } = this._overlay.getBoundingClientRect()
      this._overlay.width = width * CANVAS_SCALE
      this._overlay.height = height * CANVAS_SCALE
    }

    beforeDestroy() {}

    private canvasRequestedFrame = false

    @Watch('canvasSize')
    onCanvasSizeChange({ width, height }: Dimension) {
      this._overlay.width = width * CANVAS_SCALE
      this._overlay.height = height * CANVAS_SCALE
      this.canvasRequestRedraw()
    }

    @Watch('cursors')
    @Watch('cursorDraw')
    canvasRequestRedraw() {
      // skip rendering if there is already in progress
      if (this.canvasRequestedFrame) return

      // request animation frame from a browser
      this.canvasRequestedFrame = true
      window.requestAnimationFrame(() => {
        this.canvasRedraw()
        this.canvasRequestedFrame = false
      })
    }

    canvasRedraw() {
      if (this.screenSize == null) return

      // clear drawings
      this.canvasClear()

      // get intrinsic dimensions
      let { width, height } = this.canvasSize
      this._ctx.setTransform(CANVAS_SCALE, 0, 0, CANVAS_SCALE, 0, 0)

      // draw cursors
      for (let { id, x, y } of this.cursors) {
        // ignore own cursor
        if (id == this.sessionId) continue

        // get cursor position
        x = Math.round((x / this.screenSize.width) * width)
        y = Math.round((y / this.screenSize.height) * height)

        // use custom draw function, if available
        if (this.cursorDraw) {
          this.cursorDraw(this._ctx, x, y, id)
          continue
        }

        this._ctx.save()
        this._ctx.font = '14px Arial, sans-serif'
        this._ctx.textBaseline = 'top'
        this._ctx.shadowColor = 'black'
        this._ctx.shadowBlur = 2
        this._ctx.lineWidth = 2
        this._ctx.fillStyle = 'black'
        this._ctx.strokeText(id, x, y)
        this._ctx.shadowBlur = 0
        this._ctx.fillStyle = 'white'
        this._ctx.fillText(id, x, y)
        this._ctx.restore()
      }
    }

    canvasClear() {
      const { width, height } = this._overlay
      this._ctx.clearRect(0, 0, width, height)
    }
  }
</script>
