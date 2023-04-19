<template>
  <canvas ref="overlay" class="neko-cursors" tabindex="0" />
</template>

<style lang="scss">
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

  // How often are position data arriving
  const POS_INTERVAL_MS = 750
  // How many pixel change is considered as movement
  const POS_THRESHOLD_PX = 20

  @Component({
    name: 'neko-cursors',
  })
  export default class extends Vue {
    @Ref('overlay') readonly _overlay!: HTMLCanvasElement
    private _ctx!: CanvasRenderingContext2D

    private canvasScale = window.devicePixelRatio

    @Prop()
    private readonly sessions!: Record<string, Session>

    @Prop()
    private readonly sessionId!: string

    @Prop()
    private readonly hostId!: string | null

    @Prop()
    private readonly screenSize!: Dimension

    @Prop()
    private readonly canvasSize!: Dimension

    @Prop()
    private readonly cursors!: SessionCursors[]

    @Prop()
    private readonly cursorDraw!: InactiveCursorDrawFunction | null

    @Prop()
    private readonly fps!: number

    mounted() {
      // get canvas overlay context
      const ctx = this._overlay.getContext('2d')
      if (ctx != null) {
        this._ctx = ctx
      }

      // synchronize intrinsic with extrinsic dimensions
      const { width, height } = this._overlay.getBoundingClientRect()
      this.canvasResize({ width, height })

      // react to pixel ratio changes
      this.onPixelRatioChange()

      // store last drawing points
      this._last_points = {}
    }

    beforeDestroy() {
      // stop pixel ratio change listener
      if (this.unsubscribePixelRatioChange) {
        this.unsubscribePixelRatioChange()
      }
    }

    private unsubscribePixelRatioChange?: () => void
    private onPixelRatioChange() {
      if (this.unsubscribePixelRatioChange) {
        this.unsubscribePixelRatioChange()
      }

      const media = matchMedia(`(resolution: ${window.devicePixelRatio}dppx)`)
      media.addEventListener('change', this.onPixelRatioChange)
      this.unsubscribePixelRatioChange = () => {
        media.removeEventListener('change', this.onPixelRatioChange)
      }

      this.canvasScale = window.devicePixelRatio
      this.onCanvasSizeChange(this.canvasSize)
    }

    @Watch('canvasSize')
    onCanvasSizeChange({ width, height }: Dimension) {
      this.canvasResize({ width, height })
      this.canvasUpdateCursors()
    }

    canvasResize({ width, height }: Dimension) {
      this._overlay.width = width * this.canvasScale
      this._overlay.height = height * this.canvasScale
      this._ctx.setTransform(this.canvasScale, 0, 0, this.canvasScale, 0, 0)
    }

    // start as undefined to prevent jumping
    private _last_animation_time!: number
    // current animation progress (0-1)
    private _percent!: number
    // points to be animated for each session
    private _points!: SessionCursors[]
    // last points coordinates for each session
    private _last_points!: Record<string, Cursor>

    canvasAnimateFrame(now: number = NaN) {
      // request another frame
      if (this._percent <= 1) window.requestAnimationFrame(this.canvasAnimateFrame)

      // calc elapsed time since last loop
      const elapsed = now - this._last_animation_time

      // skip if fps is set and elapsed time is less than fps
      if (this.fps > 0 && elapsed < 1000 / this.fps) return

      // calc current animation progress
      const delta = elapsed / POS_INTERVAL_MS
      this._last_animation_time = now

      // skip very first delta to prevent jumping
      if (isNaN(delta)) return

      // set the animation position
      this._percent += delta

      // draw points for current frame
      this.canvasDrawPoints(this._percent)
    }

    canvasDrawPoints(percent: number = 1) {
      // clear canvas
      this.canvasClear()

      // draw current position
      for (const p of this._points) {
        const { x, y } = getMovementXYatPercent(p.cursors, percent)
        this.canvasDrawCursor(x, y, p.id)
      }
    }

    @Watch('hostId')
    @Watch('cursors')
    canvasUpdateCursors() {
      let new_last_points = {} as Record<string, Cursor>

      // track unchanged cursors
      let unchanged = 0

      // create points for animation
      this._points = []
      for (const { id, cursors } of this.cursors) {
        if (
          // if there are no positions
          cursors.length == 0 ||
          // ignore own cursor
          id == this.sessionId ||
          // ignore host's cursor
          id == this.hostId
        ) {
          unchanged++
          continue
        }

        // get last point
        const new_last_point = cursors[cursors.length - 1]

        // add last cursor position to cursors (if available)
        let pos = { id } as SessionCursors
        if (id in this._last_points) {
          const last_point = this._last_points[id]

          // if cursor did not move considerably
          if (
            Math.abs(last_point.x - new_last_point.x) < POS_THRESHOLD_PX &&
            Math.abs(last_point.y - new_last_point.y) < POS_THRESHOLD_PX
          ) {
            // we knew that this cursor did not change, but other
            // might, so we keep only one point to be drawn
            pos.cursors = [new_last_point]
            // and increase unchanged counter
            unchanged++
          } else {
            // if cursor moved, we want to include last point
            // in the animation, so that movement can be seamless
            pos.cursors = [last_point, ...cursors]
          }
        } else {
          // if cursor does not have last point, it is not
          // displayed in canvas and it should be now
          pos.cursors = [...cursors]
        }

        new_last_points[id] = new_last_point
        this._points.push(pos)
      }

      // apply new last points
      this._last_points = new_last_points

      // no cursors to animate
      if (this._points.length == 0) {
        this.canvasClear()
        return
      }

      // if all cursors are unchanged
      if (unchanged == this.cursors.length) {
        // draw only last known position without animation
        this.canvasDrawPoints()
        return
      }

      // start animation if not running
      const percent = this._percent
      this._percent = 0
      if (percent > 1 || !percent) {
        this.canvasAnimateFrame()
      }
    }

    canvasDrawCursor(x: number, y: number, id: string) {
      // get intrinsic dimensions
      const { width, height } = this.canvasSize
      x = Math.round((x / this.screenSize.width) * width)
      y = Math.round((y / this.screenSize.height) * height)

      // reset transformation, X and Y will be 0 again
      this._ctx.setTransform(this.canvasScale, 0, 0, this.canvasScale, 0, 0)

      // use custom draw function, if available
      if (this.cursorDraw) {
        this.cursorDraw(this._ctx, x, y, id)
        return
      }

      // get cursor tag
      const cursorTag = this.sessions[id]?.profile.name || ''

      // draw inactive cursor tag
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
    }

    canvasClear() {
      // reset transformation, X and Y will be 0 again
      this._ctx.setTransform(this.canvasScale, 0, 0, this.canvasScale, 0, 0)

      const { width, height } = this.canvasSize
      this._ctx.clearRect(0, 0, width, height)
    }
  }
</script>
