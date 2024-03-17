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

<script lang="ts" setup>

import { ref, onMounted, onBeforeUnmount, watch } from 'vue'

import type { SessionCursors, Cursor, Session } from './types/state'
import type { InactiveCursorDrawFunction, Dimension } from './types/cursors'
import { getMovementXYatPercent } from './utils/canvas-movement'

// How often are position data arriving
const POS_INTERVAL_MS = 750
// How many pixel change is considered as movement
const POS_THRESHOLD_PX = 20

const props = defineProps<{
  sessions: Record<string, Session>
  sessionId: string
  hostId: string | null
  screenSize: Dimension
  canvasSize: Dimension
  cursors: SessionCursors[]
  cursorDraw: InactiveCursorDrawFunction | null
  fps: number
}>()

const overlay = ref<HTMLCanvasElement | null>(null)

let ctx: CanvasRenderingContext2D | null = null
let canvasScale = window.devicePixelRatio
let unsubscribePixelRatioChange = null as (() => void) | null

onMounted(() => {
  // get canvas overlay context
  const canvas = overlay.value
  if (canvas != null) {
    ctx = canvas.getContext('2d')

    // synchronize intrinsic with extrinsic dimensions
    const { width, height } = canvas.getBoundingClientRect()
    canvasResize({ width, height })
  }

  // react to pixel ratio changes
  onPixelRatioChange()

  // store last drawing points
  last_points = {}
})

onBeforeUnmount(() => {
  // stop pixel ratio change listener
  if (unsubscribePixelRatioChange) {
    unsubscribePixelRatioChange()
  }
})

function onPixelRatioChange() {
  if (unsubscribePixelRatioChange) {
    unsubscribePixelRatioChange()
  }

  const media = window.matchMedia(`(resolution: ${window.devicePixelRatio}dppx)`)
  media.addEventListener('change', onPixelRatioChange)
  unsubscribePixelRatioChange = () => {
    media.removeEventListener('change', onPixelRatioChange)
  }

  canvasScale = window.devicePixelRatio
  onCanvasSizeChange(props.canvasSize)
}

function onCanvasSizeChange({ width, height }: Dimension) {
  canvasResize({ width, height })
  canvasUpdateCursors()
}

watch(() => props.canvasSize, onCanvasSizeChange)

function canvasResize({ width, height }: Dimension) {
  overlay.value!.width = width * canvasScale
  overlay.value!.height = height * canvasScale
  ctx?.setTransform(canvasScale, 0, 0, canvasScale, 0, 0)
}

// start as undefined to prevent jumping
let last_animation_time = 0
// current animation progress (0-1)
let percent = 0
// points to be animated for each session
let points: SessionCursors[] = []
// last points coordinates for each session
let last_points: Record<string, Cursor> = {}

function canvasAnimateFrame(now: number = NaN) {
  // request another frame
  if (percent <= 1) window.requestAnimationFrame(canvasAnimateFrame)

  // calc elapsed time since last loop
  const elapsed = now - last_animation_time

  // skip if fps is set and elapsed time is less than fps
  if (props.fps > 0 && elapsed < 1000 / props.fps) return

  // calc current animation progress
  const delta = elapsed / POS_INTERVAL_MS
  last_animation_time = now

  // skip very first delta to prevent jumping
  if (isNaN(delta)) return

  // set the animation position
  percent += delta

  // draw points for current frame
  canvasDrawPoints(percent)
}

function canvasDrawPoints(percent: number = 1) {
  // clear canvas
  canvasClear()

  // draw current position
  for (const p of points) {
    const { x, y } = getMovementXYatPercent(p.cursors, percent)
    canvasDrawCursor(x, y, p.id)
  }
}

function canvasUpdateCursors() {
  let new_last_points = {} as Record<string, Cursor>

  // track unchanged cursors
  let unchanged = 0

  // create points for animation
  points = []
  for (const { id, cursors } of props.cursors) {
    if (
      // if there are no positions
      cursors.length == 0 ||
      // ignore own cursor
      id == props.sessionId ||
      // ignore host's cursor
      id == props.hostId
    ) {
      unchanged++
      continue
    }

    // get last point
    const new_last_point = cursors[cursors.length - 1]

    // add last cursor position to cursors (if available)
    let pos = { id } as SessionCursors
    if (id in last_points) {
      const last_point = last_points[id]

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
    points.push(pos)
  }

  // apply new last points
  last_points = new_last_points

  // no cursors to animate
  if (points.length == 0) {
    canvasClear()
    return
  }

  // if all cursors are unchanged
  if (unchanged == props.cursors.length) {
    // draw only last known position without animation
    canvasDrawPoints()
    return
  }

  // start animation if not running
  const p = percent
  percent = 0
  if (p > 1 || !p) {
    canvasAnimateFrame()
  }
}

watch(() => props.hostId, canvasUpdateCursors)
watch(() => props.cursors, canvasUpdateCursors)

function canvasDrawCursor(x: number, y: number, id: string) {
  if (!ctx) return

  // get intrinsic dimensions
  const { width, height } = props.canvasSize
  x = Math.round((x / props.screenSize.width) * width)
  y = Math.round((y / props.screenSize.height) * height)

  // reset transformation, X and Y will be 0 again
  ctx.setTransform(canvasScale, 0, 0, canvasScale, 0, 0)

  // use custom draw function, if available
  if (props.cursorDraw) {
    props.cursorDraw(ctx, x, y, id)
    return
  }

  // get cursor tag
  const cursorTag = props.sessions[id]?.profile.name || ''

  // draw inactive cursor tag
  ctx.font = '14px Arial, sans-serif'
  ctx.textBaseline = 'top'
  ctx.shadowColor = 'black'
  ctx.shadowBlur = 2
  ctx.lineWidth = 2
  ctx.fillStyle = 'black'
  ctx.strokeText(cursorTag, x, y)
  ctx.shadowBlur = 0
  ctx.fillStyle = 'white'
  ctx.fillText(cursorTag, x, y)
}

function canvasClear() {
  if (!ctx) return

  // reset transformation, X and Y will be 0 again
  ctx.setTransform(canvasScale, 0, 0, canvasScale, 0, 0)

  const { width, height } = props.canvasSize
  ctx.clearRect(0, 0, width, height)
}
</script>
