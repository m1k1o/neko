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

  import { SessionCursors, Cursor, Session } from './types/state'
  import { InactiveCursorDrawFunction, Dimension } from './types/cursors'
  import { getMovementXYatPercent } from './utils/canvas-movement'

  const CANVAS_SCALE = 2
  const POS_INTERVAL_MS = 750

  @Component({
    name: 'neko-cursors',
  })
  export default class extends Vue {
    @Ref('overlay') readonly _overlay!: HTMLCanvasElement
    private _ctx!: CanvasRenderingContext2D

    @Prop()
    private readonly sessions!: Record<string, Session>

    @Prop()
    private readonly sessionId!: string

    @Prop()
    private readonly screenSize!: Dimension

    @Prop()
    private readonly canvasSize!: Dimension

    @Prop()
    private readonly cursors!: SessionCursors[]

    @Prop()
    private readonly cursorDraw!: InactiveCursorDrawFunction | null

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

      this._last_points = {}
    }

    beforeDestroy() {}

    @Watch('canvasSize')
    onCanvasSizeChange({ width, height }: Dimension) {
      this._overlay.width = width * CANVAS_SCALE
      this._overlay.height = height * CANVAS_SCALE
    }

    // start as undefined to prevent jumping
    private _prev_time!: number
    private _percent!: number
    private _points!: SessionCursors[]
    private _last_points!: Record<string, Cursor>

    canvasAnimate(now: number = NaN) {
      // request another frame
      if (this._percent <= 1) requestAnimationFrame(this.canvasAnimate)

      // calculate factor
      const delta = (now - this._prev_time) / POS_INTERVAL_MS
      this._prev_time = now

      // skip very first delta to prevent jumping
      if (isNaN(delta)) return

      // set the animation position
      this._percent += delta

      this.canvasClear()

      // scale
      this._ctx.setTransform(CANVAS_SCALE, 0, 0, CANVAS_SCALE, 0, 0)

      // draw current position
      for (const p of this._points) {
        const { x, y } = getMovementXYatPercent(p.cursors, this._percent)
        this.canvasRedraw(x, y, p.id)
      }
    }

    @Watch('cursors')
    canvasSetPosition(e: SessionCursors[]) {
      console.log('consuming', e)

      // clear on no cursor
      if (e.length == 0) {
        this._last_points = {}
        this.canvasClear()
        return
      }

      // create points for animation
      this._points = []
      for (const { id, cursors } of e) {
        let pos = { id } as SessionCursors
        if (id in this._last_points) {
          pos.cursors = [this._last_points[id], ...cursors]
        } else {
          pos.cursors = [...cursors]
        }
        this._last_points[id] = cursors[cursors.length - 1]
        this._points.push(pos)
      }

      // no cursors to animate
      if (this._points.length == 0) {
        return
      }

      // start animation if not running
      const percent = this._percent
      this._percent = 0
      if (percent > 1 || percent == 0) {
        this.canvasAnimate()
      }
    }

    canvasRedraw(x: number, y: number, id: string) {
      // get intrinsic dimensions
      let { width, height } = this.canvasSize
      x = Math.round((x / this.screenSize.width) * width)
      y = Math.round((y / this.screenSize.height) * height)

      // get cursor tag
      const cursorTag = this.sessions[id]?.profile.name || ''

      // use custom draw function, if available
      if (this.cursorDraw) {
        this.cursorDraw(this._ctx, x, y, cursorTag)
        return
      }

      this._ctx.save()
      this._ctx.font = '14px Arial, sans-serif'
      this._ctx.textBaseline = 'top'
      this._ctx.shadowColor = 'black'
      this._ctx.shadowBlur = 2
      this._ctx.lineWidth = 2
      this._ctx.fillStyle = 'black'
      this._ctx.strokeText(cursorTag, x, y)
      this._ctx.shadowBlur = 0
      this._ctx.fillStyle = 'white'
      this._ctx.fillText(cursorTag, x, y)
      this._ctx.restore()
    }

    canvasClear() {
      const { width, height } = this._overlay
      this._ctx.clearRect(0, 0, width, height)
    }
  }
</script>
