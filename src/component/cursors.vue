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
const ctx = ref<CanvasRenderingContext2D | null>(null)
const canvasScale = ref(window.devicePixelRatio)

let unsubscribePixelRatioChange = null as (() => void) | null

onMounted(() => {
  // get canvas overlay context
  const canvas = overlay.value
  if (canvas != null) {
    ctx.value = canvas.getContext('2d')

    // synchronize intrinsic with extrinsic dimensions
    const { width, height } = canvas.getBoundingClientRect()
    canvasResize({ width, height })
  }

  // react to pixel ratio changes
  onPixelRatioChange()

  // store last drawing points
  last_points.value = {}
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

  canvasScale.value = window.devicePixelRatio
  onCanvasSizeChange(props.canvasSize)
}

function onCanvasSizeChange({ width, height }: Dimension) {
  canvasResize({ width, height })
  canvasUpdateCursors()
}

watch(() => props.canvasSize, onCanvasSizeChange)

function canvasResize({ width, height }: Dimension) {
  overlay.value!.width = width * canvasScale.value
  overlay.value!.height = height * canvasScale.value
  ctx.value?.setTransform(canvasScale.value, 0, 0, canvasScale.value, 0, 0)
}

// start as undefined to prevent jumping
const last_animation_time = ref<number>(0)
// current animation progress (0-1)
const percent = ref<number>(0)
// points to be animated for each session
const points = ref<SessionCursors[]>([])
// last points coordinates for each session
const last_points = ref<Record<string, Cursor>>({})

function canvasAnimateFrame(now: number = NaN) {
  // request another frame
  if (percent.value <= 1) window.requestAnimationFrame(canvasAnimateFrame)

  // calc elapsed time since last loop
  const elapsed = now - last_animation_time.value

  // skip if fps is set and elapsed time is less than fps
  if (props.fps > 0 && elapsed < 1000 / props.fps) return

  // calc current animation progress
  const delta = elapsed / POS_INTERVAL_MS
  last_animation_time.value = now

  // skip very first delta to prevent jumping
  if (isNaN(delta)) return

  // set the animation position
  percent.value += delta

  // draw points for current frame
  canvasDrawPoints(percent.value)
}

function canvasDrawPoints(percent: number = 1) {
  // clear canvas
  canvasClear()

  // draw current position
  for (const p of points.value) {
    const { x, y } = getMovementXYatPercent(p.cursors, percent)
    canvasDrawCursor(x, y, p.id)
  }
}

function canvasUpdateCursors() {
  let new_last_points = {} as Record<string, Cursor>

  // track unchanged cursors
  let unchanged = 0

  // create points for animation
  points.value = []
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
    if (id in last_points.value) {
      const last_point = last_points.value[id]

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
    points.value.push(pos)
  }

  // apply new last points
  last_points.value = new_last_points

  // no cursors to animate
  if (points.value.length == 0) {
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
  const p = percent.value
  percent.value = 0
  if (p > 1 || !p) {
    canvasAnimateFrame()
  }
}

watch(() => props.hostId, canvasUpdateCursors)
watch(() => props.cursors, canvasUpdateCursors)

function canvasDrawCursor(x: number, y: number, id: string) {
  // get intrinsic dimensions
  const { width, height } = props.canvasSize
  x = Math.round((x / props.screenSize.width) * width)
  y = Math.round((y / props.screenSize.height) * height)

  // reset transformation, X and Y will be 0 again
  ctx.value!.setTransform(canvasScale.value, 0, 0, canvasScale.value, 0, 0)

  // use custom draw function, if available
  if (props.cursorDraw) {
    props.cursorDraw(ctx.value!, x, y, id)
    return
  }

  // get cursor tag
  const cursorTag = props.sessions[id]?.profile.name || ''

  // draw inactive cursor tag
  ctx.value!.font = '14px Arial, sans-serif'
  ctx.value!.textBaseline = 'top'
  ctx.value!.shadowColor = 'black'
  ctx.value!.shadowBlur = 2
  ctx.value!.lineWidth = 2
  ctx.value!.fillStyle = 'black'
  ctx.value!.strokeText(cursorTag, x, y)
  ctx.value!.shadowBlur = 0
  ctx.value!.fillStyle = 'white'
  ctx.value!.fillText(cursorTag, x, y)
}

function canvasClear() {
  // reset transformation, X and Y will be 0 again
  ctx.value?.setTransform(canvasScale.value, 0, 0, canvasScale.value, 0, 0)

  const { width, height } = props.canvasSize
  ctx.value?.clearRect(0, 0, width, height)
}
</script>
