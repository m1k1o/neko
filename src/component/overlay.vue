<template>
  <div class="neko-overlay-wrap">
    <canvas ref="overlay" class="neko-overlay" tabindex="0" />
    <textarea
      ref="textarea"
      class="neko-overlay"
      :style="{ cursor }"
      v-model="textInput"
      @click.stop.prevent="control.emit('overlay.click', $event)"
      @contextmenu.stop.prevent="control.emit('overlay.contextmenu', $event)"
      @wheel.stop.prevent="onWheel"
      @mousemove.stop.prevent="onMouseMove"
      @mousedown.stop.prevent="onMouseDown"
      @mouseenter.stop.prevent="onMouseEnter"
      @mouseleave.stop.prevent="onMouseLeave"
      @dragenter.stop.prevent="onDragEnter"
      @dragleave.stop.prevent="onDragLeave"
      @dragover.stop.prevent="onDragOver"
      @drop.stop.prevent="onDrop"
    />
  </div>
</template>

<style lang="scss">
  /* hide elements around textarea if added by browsers extensions */
  .neko-overlay-wrap *:not(.neko-overlay) {
    display: none;
  }

  .neko-overlay {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 100%;
    height: 100%;
    font-size: 16px; /* at least 16px to avoid zooming on mobile */
    resize: none; /* hide textarea resize corner */
    caret-color: transparent; /* hide caret */
    outline: 0;
    border: 0;
    color: transparent;
    background: transparent;
  }
</style>

<script lang="ts" setup>
import { ref, watch, computed, onMounted, onBeforeUnmount } from 'vue'

import type { KeyboardInterface } from './utils/keyboard'
import type { GestureHandler } from './utils/gesturehandler'
import type { Session, Scroll } from './types/state'
import type { CursorPosition, CursorImage } from './types/webrtc'
import type { CursorDrawFunction, Dimension, KeyboardModifiers } from './types/cursors'

import { NewKeyboard } from './utils/keyboard'
import GestureHandlerInit from './utils/gesturehandler'
import { KeyTable, keySymsRemap } from './utils/keyboard-remapping'
import { getFilesFromDataTansfer } from './utils/file-upload'
import { NekoControl } from './internal/control'
import { NekoWebRTC } from './internal/webrtc'

// Wheel thresholds
const WHEEL_STEP = 53 // Pixels needed for one step
const WHEEL_LINE_HEIGHT = 19 // Assumed pixels for one line step

// Gesture thresholds
const GESTURE_ZOOMSENS = 75
const GESTURE_SCRLSENS = 50
const DOUBLE_TAP_TIMEOUT = 1000
const DOUBLE_TAP_THRESHOLD = 50

const MOUSE_MOVE_THROTTLE = 1000 / 60 // in ms, 60fps
const INACTIVE_CURSOR_INTERVAL = 1000 / 4 // in ms, 4fps

// refs

const overlay = ref<HTMLCanvasElement | null>(null)
const textarea = ref<HTMLTextAreaElement | null>(null)
const ctx = ref<CanvasRenderingContext2D | null>(null)

const canvasScale = ref(window.devicePixelRatio)

const keyboard = ref<KeyboardInterface | null>(null)
const gestureHandler = ref<GestureHandler | null>(null)
const textInput = ref('')

const focused = ref(false)

// props and emits

const props = defineProps<{
  control: NekoControl
  sessions: Record<string, Session>
  hostId: string
  webrtc: NekoWebRTC
  scroll: Scroll
  screenSize: Dimension
  canvasSize: Dimension
  cursorDraw: CursorDrawFunction | null
  isControling: boolean
  implicitControl: boolean
  inactiveCursors: boolean
  fps: number
  hasMobileKeyboard: boolean
}>()

const emit = defineEmits(['uploadDrop', 'updateKeyboardModifiers', 'mobileKeyboardOpen'])

// computed

const cursor = computed(() => {
  if (!props.isControling || !cursorImage.value) {
    return 'default'
  }

  const { uri, x, y } = cursorImage.value
  return 'url(' + uri + ') ' + x + ' ' + y + ', default'
})

// lifecycle

onMounted(() => {
  // register mouseup globally as user can release mouse button outside of overlay
  window.addEventListener('mouseup', onMouseUp, true)

  // get canvas overlay context
  const _ctx = overlay.value?.getContext('2d')
  if (_ctx != null) {
    ctx.value = _ctx
  }

  // synchronize intrinsic with extrinsic dimensions
  const { width, height } = overlay.value?.getBoundingClientRect() || { width: 0, height: 0 }
  canvasResize({ width, height })

  // react to pixel ratio changes
  onPixelRatioChange()

  let ctrlKey = 0
  let noKeyUp = {} as Record<number, boolean>

  // Initialize Keyboard
  keyboard.value = NewKeyboard()
  keyboard.value.onkeydown = (key: number) => {
    key = keySymsRemap(key)

    if (!props.isControling) {
      noKeyUp[key] = true
      return true
    }

    // ctrl+v is aborted
    if (ctrlKey != 0 && key == KeyTable.XK_v) {
      keyboard.value!.release(ctrlKey)
      noKeyUp[key] = true
      return true
    }

    // save information if it is ctrl key event
    const isCtrlKey = key == KeyTable.XK_Control_L || key == KeyTable.XK_Control_R
    if (isCtrlKey) ctrlKey = key

    props.control.keyDown(key)
    return isCtrlKey
  }
  keyboard.value.onkeyup = (key: number) => {
    key = keySymsRemap(key)

    if (key in noKeyUp) {
      delete noKeyUp[key]
      return
    }

    const isCtrlKey = key == KeyTable.XK_Control_L || key == KeyTable.XK_Control_R
    if (isCtrlKey) ctrlKey = 0

    props.control.keyUp(key)
  }
  keyboard.value.listenTo(textarea.value!)

  // Initialize GestureHandler
  gestureHandler.value = new GestureHandlerInit()

  // bind touch handler using @Watch on supportedTouchEvents
  // because we need to know if touch events are supported
  // by the server before we can bind touch handler

  // default value is false, so we can bind touch handler
  bindGestureHandler()

  props.webrtc.addListener('cursor-position', onCursorPosition)
  props.webrtc.addListener('cursor-image', onCursorImage)
  props.webrtc.addListener('disconnected', canvasClear)
  cursorElement.onload = canvasRequestRedraw
})

onBeforeUnmount(() => {
  window.removeEventListener('mouseup', onMouseUp, true)

  if (keyboard.value) {
    keyboard.value.removeListener()
  }

  // unbind touch handler
  unbindTouchHandler()

  // unbind gesture handler
  unbindGestureHandler()

  props.webrtc.removeListener('cursor-position', onCursorPosition)
  props.webrtc.removeListener('cursor-image', onCursorImage)
  props.webrtc.removeListener('disconnected', canvasClear)
  cursorElement.onload = null

  // stop inactive cursor interval if exists
  clearInactiveCursorInterval()

  // stop pixel ratio change listener
  if (unsubscribePixelRatioChange.value) {
    unsubscribePixelRatioChange.value()
  }
})

//
// touch handler for native touch events
//

function bindTouchHandler() {
  textarea.value?.addEventListener('touchstart', onTouchHandler, { passive: false })
  textarea.value?.addEventListener('touchmove', onTouchHandler, { passive: false })
  textarea.value?.addEventListener('touchend', onTouchHandler, { passive: false })
  textarea.value?.addEventListener('touchcancel', onTouchHandler, { passive: false })
}

function unbindTouchHandler() {
  textarea.value?.removeEventListener('touchstart', onTouchHandler)
  textarea.value?.removeEventListener('touchmove', onTouchHandler)
  textarea.value?.removeEventListener('touchend', onTouchHandler)
  textarea.value?.removeEventListener('touchcancel', onTouchHandler)
}

function onTouchHandler(ev: TouchEvent) {
  // we cannot use implicitControlRequest because we don't have mouse event
  if (!props.isControling) {
    // if implicitControl is enabled, request control
    if (props.implicitControl) {
      props.control.request()
    }
    // otherwise, ignore event
    return
  }

  ev.stopPropagation()
  ev.preventDefault()

  for (let i = 0; i < ev.changedTouches.length; i++) {
    const touch = ev.changedTouches[i]
    const pos = getMousePos(touch.clientX, touch.clientY)
    // force is float value between 0 and 1
    // pressure is integer value between 0 and 255
    const pressure = Math.round(touch.force * 255)

    switch (ev.type) {
      case 'touchstart':
        props.control.touchBegin(touch.identifier, pos, pressure)
        break
      case 'touchmove':
        props.control.touchUpdate(touch.identifier, pos, pressure)
        break
      case 'touchend':
      case 'touchcancel':
        props.control.touchEnd(touch.identifier, pos, pressure)
        break
    }
  }
}

//
// gesture handler for emulated mouse events
//

function bindGestureHandler() {
  gestureHandler.value?.attach(textarea.value!)
  textarea.value?.addEventListener('gesturestart', onGestureHandler)
  textarea.value?.addEventListener('gesturemove', onGestureHandler)
  textarea.value?.addEventListener('gestureend', onGestureHandler)
}

function unbindGestureHandler() {
  gestureHandler.value?.detach()
  textarea.value?.removeEventListener('gesturestart', onGestureHandler)
  textarea.value?.removeEventListener('gesturemove', onGestureHandler)
  textarea.value?.removeEventListener('gestureend', onGestureHandler)
}

const gestureLastTapTime = ref<number | null>(null)
const gestureFirstDoubleTapEv = ref<any | null>(null)
const gestureLastMagnitudeX = ref(0)
const gestureLastMagnitudeY = ref(0)

function _handleTapEvent(ev: any, code: number) {
  let pos = getMousePos(ev.detail.clientX, ev.detail.clientY)

  // If the user quickly taps multiple times we assume they meant to
  // hit the same spot, so slightly adjust coordinates

  if (
    gestureLastTapTime.value !== null &&
    Date.now() - gestureLastTapTime.value < DOUBLE_TAP_TIMEOUT &&
    gestureFirstDoubleTapEv.value?.detail.type === ev.detail.type
  ) {
    const dx = gestureFirstDoubleTapEv.value.detail.clientX - ev.detail.clientX
    const dy = gestureFirstDoubleTapEv.value.detail.clientY - ev.detail.clientY
    const distance = Math.hypot(dx, dy)

    if (distance < DOUBLE_TAP_THRESHOLD) {
      pos = getMousePos(gestureFirstDoubleTapEv.value.detail.clientX, gestureFirstDoubleTapEv.value.detail.clientY)
    } else {
      gestureFirstDoubleTapEv.value = ev
    }
  } else {
    gestureFirstDoubleTapEv.value = ev
  }
  gestureLastTapTime.value = Date.now()

  props.control.buttonDown(code, pos)
  props.control.buttonUp(code, pos)
}

function onGestureHandler(ev: any) {
  // we cannot use implicitControlRequest because we don't have mouse event
  if (!props.isControling) {
    // if implicitControl is enabled, request control
    if (props.implicitControl) {
      props.control.request()
    }
    // otherwise, ignore event
    return
  }

  const pos = getMousePos(ev.detail.clientX, ev.detail.clientY)

  let magnitude
  switch (ev.type) {
    case 'gesturestart':
      switch (ev.detail.type) {
        case 'onetap':
          _handleTapEvent(ev, 1)
          break
        case 'twotap':
          _handleTapEvent(ev, 3)
          break
        case 'threetap':
          _handleTapEvent(ev, 2)
          break
        case 'drag':
          props.control.buttonDown(1, pos)
          break
        case 'longpress':
          props.control.buttonDown(3, pos)
          break

        case 'twodrag':
          gestureLastMagnitudeX.value = ev.detail.magnitudeX
          gestureLastMagnitudeY.value = ev.detail.magnitudeY
          props.control.move(pos)
          break
        case 'pinch':
          gestureLastMagnitudeX.value = Math.hypot(ev.detail.magnitudeX, ev.detail.magnitudeY)
          props.control.move(pos)
          break
      }
      break

    case 'gesturemove':
      switch (ev.detail.type) {
        case 'onetap':
        case 'twotap':
        case 'threetap':
          break
        case 'drag':
        case 'longpress':
          props.control.move(pos)
          break
        case 'twodrag':
          // Always scroll in the same position.
          // We don't know if the mouse was moved so we need to move it
          // every update.
          props.control.move(pos)
          while (ev.detail.magnitudeY - gestureLastMagnitudeY.value > GESTURE_SCRLSENS) {
            props.control.scroll({ delta_x: 0, delta_y: 1 })
            gestureLastMagnitudeY.value += GESTURE_SCRLSENS
          }
          while (ev.detail.magnitudeY - gestureLastMagnitudeY.value < -GESTURE_SCRLSENS) {
            props.control.scroll({ delta_x: 0, delta_y: -1 })
            gestureLastMagnitudeY.value -= GESTURE_SCRLSENS
          }
          while (ev.detail.magnitudeX - gestureLastMagnitudeX.value > GESTURE_SCRLSENS) {
            props.control.scroll({ delta_x: 1, delta_y: 0 })
            gestureLastMagnitudeX.value += GESTURE_SCRLSENS
          }
          while (ev.detail.magnitudeX - gestureLastMagnitudeX.value < -GESTURE_SCRLSENS) {
            props.control.scroll({ delta_x: -1, delta_y: 0 })
            gestureLastMagnitudeX.value -= GESTURE_SCRLSENS
          }
          break
        case 'pinch':
          // Always scroll in the same position.
          // We don't know if the mouse was moved so we need to move it
          // every update.
          props.control.move(pos)
          magnitude = Math.hypot(ev.detail.magnitudeX, ev.detail.magnitudeY)
          if (Math.abs(magnitude - gestureLastMagnitudeX.value) > GESTURE_ZOOMSENS) {
            while (magnitude - gestureLastMagnitudeX.value > GESTURE_ZOOMSENS) {
              props.control.scroll({ delta_x: 0, delta_y: 1, control_key: true })
              gestureLastMagnitudeX.value += GESTURE_ZOOMSENS
            }
            while (magnitude - gestureLastMagnitudeX.value < -GESTURE_ZOOMSENS) {
              props.control.scroll({ delta_x: 0, delta_y: -1, control_key: true })
              gestureLastMagnitudeX.value -= GESTURE_ZOOMSENS
            }
          }
          break
      }
      break

    case 'gestureend':
      switch (ev.detail.type) {
        case 'onetap':
        case 'twotap':
        case 'threetap':
        case 'pinch':
        case 'twodrag':
          break
        case 'drag':
          props.control.buttonUp(1, pos)
          break
        case 'longpress':
          props.control.buttonUp(3, pos)
          break
      }
      break
  }
}


//
// touch and gesture handlers cannot be used together
//

function onTouchEventsChange() {
  unbindGestureHandler()
  unbindTouchHandler()

  if (!props.control.enabledTouchEvents) {
    return
  }

  if (props.control.supportedTouchEvents) {
    bindTouchHandler()
  } else {
    bindGestureHandler()
  }
}

watch(() => props.control.enabledTouchEvents, onTouchEventsChange)
watch(() => props.control.supportedTouchEvents, onTouchEventsChange)


function getModifierState(e: MouseEvent): KeyboardModifiers {
  // we can only use locks, because when someone holds key outside
  // of the renderer, and releases it inside, keyup event is not fired
  // by guacamole keyboard and modifier state is not updated

  return {
    //shift: e.getModifierState('Shift'),
    capslock: e.getModifierState('CapsLock'),
    //control: e.getModifierState('Control'),
    //alt: e.getModifierState('Alt'),
    numlock: e.getModifierState('NumLock'),
    //meta: e.getModifierState('Meta'),
    //super: e.getModifierState('Super'),
    //altgr: e.getModifierState('AltGraph'),
  }
}

function getMousePos(clientX: number, clientY: number) {
  const rect = overlay.value!.getBoundingClientRect()

  return {
    x: Math.round((props.screenSize.width / rect.width) * (clientX - rect.left)),
    y: Math.round((props.screenSize.height / rect.height) * (clientY - rect.top)),
  }
}

function sendMousePos(e: MouseEvent) {
  const pos = getMousePos(e.clientX, e.clientY)
  // not using NekoControl here because we want to avoid
  // sending mousemove events over websocket
  if (props.webrtc.connected) {
    props.webrtc.send('mousemove', pos)
  } // otherwise, no events are sent
  cursorPosition.value = pos
}

const wheelX = ref(0)
const wheelY = ref(0)
const wheelTimeStamp = ref(0)

// negative sensitivity can be acheived using increased step value
const wheelStep = computed(() => {
  let x = WHEEL_STEP

  if (props.scroll.sensitivity < 0) {
    x *= Math.abs(props.scroll.sensitivity) + 1
  }

  return x
})


// sensitivity can only be positive
const wheelSensitivity = computed(() => {
  let x = 1

  if (props.scroll.sensitivity > 0) {
    x = Math.abs(props.scroll.sensitivity) + 1
  }

  if (props.scroll.inverse) {
    x *= -1
  }

  return x
})

// use v-model instead of @input because v-model
// doesn't get updated during IME composition
function onTextInputChange() {
  if (textInput.value == '') return
  props.control.paste(textInput.value)
  textInput.value = ''
}

watch(() => textInput.value, onTextInputChange)

function onWheel(e: WheelEvent) {
  if (!props.isControling) {
    return
  }

  // when the last scroll was more than 250ms ago
  const firstScroll = e.timeStamp - wheelTimeStamp.value > 250

  if (firstScroll) {
    wheelX.value = 0
    wheelY.value = 0
    wheelTimeStamp.value = e.timeStamp
  }

  let dx = e.deltaX
  let dy = e.deltaY

  if (e.deltaMode !== 0) {
    dx *= WHEEL_LINE_HEIGHT
    dy *= WHEEL_LINE_HEIGHT
  }

  wheelX.value += dx
  wheelY.value += dy

  let x = 0
  if (Math.abs(wheelX.value) >= wheelStep.value || firstScroll) {
    if (wheelX.value < 0) {
      x = wheelSensitivity.value * -1
    } else if (wheelX.value > 0) {
      x = wheelSensitivity.value
    }

    if (!firstScroll) {
      wheelX.value = 0
    }
  }

  let y = 0
  if (Math.abs(wheelY.value) >= wheelStep.value || firstScroll) {
    if (wheelY.value < 0) {
      y = wheelSensitivity.value * -1
    } else if (wheelY.value > 0) {
      y = wheelSensitivity.value
    }

    if (!firstScroll) {
      wheelY.value = 0
    }
  }

  // skip if not scrolled
  if (x == 0 && y == 0) return

  // TODO: add position for precision scrolling
  props.control.scroll({
    delta_x: x,
    delta_y: y,
    control_key: e.ctrlKey,
  })
}

const lastMouseMove = ref(0)

function onMouseMove(e: MouseEvent) {
  // throttle mousemove events
  if (e.timeStamp - lastMouseMove.value < MOUSE_MOVE_THROTTLE) return
  lastMouseMove.value = e.timeStamp

  if (props.isControling) {
    sendMousePos(e)
  }

  if (props.inactiveCursors) {
    saveInactiveMousePos(e)
  }
}

const isMouseDown = ref(false)

function onMouseDown(e: MouseEvent) {
  isMouseDown.value = true

  if (!props.isControling) {
    implicitControlRequest(e)
    return
  }

  const key = e.button + 1
  const pos = getMousePos(e.clientX, e.clientY)
  props.control.buttonDown(key, pos)
}

function onMouseUp(e: MouseEvent) {
  // only if we are the one who started the mouse down
  if (!isMouseDown.value) return
  isMouseDown.value = false

  if (!props.isControling) {
    implicitControlRequest(e)
    return
  }

  const key = e.button + 1
  const pos = getMousePos(e.clientX, e.clientY)
  props.control.buttonUp(key, pos)
}

function onMouseEnter(e: MouseEvent) {
  // focus opens the keyboard on mobile (only for android)
  if (!props.hasMobileKeyboard) {
    textarea.value?.focus()
  }

  focused.value = true

  if (props.isControling) {
    updateKeyboardModifiers(e)
  }
}

function onMouseLeave(e: MouseEvent) {
  if (props.isControling) {
    // save current keyboard modifiers state
    keyboardModifiers.value = getModifierState(e)
  }

  focused.value = false
}

function onDragEnter(e: DragEvent) {
  onMouseEnter(e as MouseEvent)
}

function onDragLeave(e: DragEvent) {
  onMouseLeave(e as MouseEvent)
}

function onDragOver(e: DragEvent) {
  onMouseMove(e as MouseEvent)
}

async function onDrop(e: DragEvent) {
  if (props.isControling || props.implicitControl) {
    const dt = e.dataTransfer
    if (!dt) return

    const files = await getFilesFromDataTansfer(dt)
    if (files.length === 0) return

    const pos = getMousePos(e.clientX, e.clientY)
    emit('uploadDrop', { ...pos, files })
  }
}

//
// inactive cursor position
//

const inactiveCursorInterval = ref<number | null>(null)
const inactiveCursorPosition = ref<CursorPosition | null>(null)

function clearInactiveCursorInterval() {
  if (inactiveCursorInterval.value) {
    window.clearInterval(inactiveCursorInterval.value)
    inactiveCursorInterval.value = null
  }
}

function restartInactiveCursorInterval() {
  // clear interval if exists
  clearInactiveCursorInterval()

  if (props.inactiveCursors && focused.value && !props.isControling) {
    inactiveCursorInterval.value = window.setInterval(sendInactiveMousePos, INACTIVE_CURSOR_INTERVAL)
  }
}

watch([focused, props.isControling, props.inactiveCursors], restartInactiveCursorInterval)

function saveInactiveMousePos(e: MouseEvent) {
  const pos = getMousePos(e.clientX, e.clientY)
  inactiveCursorPosition.value = pos
}

function sendInactiveMousePos() {
  if (inactiveCursorPosition.value && props.webrtc.connected) {
    // not using NekoControl here, because inactive cursors are
    // treated differently than moving the mouse while controling
    props.webrtc.send('mousemove', inactiveCursorPosition.value)
  } // if webrtc is not connected, we don't need to send anything
}

//
// keyboard modifiers
//

const keyboardModifiers = ref<KeyboardModifiers | null>(null)

function updateKeyboardModifiers(e: MouseEvent) {
  const mods = getModifierState(e)
  const newMods = Object.values(mods).join()
  const oldMods = Object.values(keyboardModifiers.value || {}).join()

  // update keyboard modifiers only if they changed
  if (newMods !== oldMods) {
    emit('updateKeyboardModifiers', mods)
  }
}

//
// canvas
//

const cursorImage = ref<CursorImage | null>(null)
const cursorElement = new Image()
const cursorPosition = ref<CursorPosition | null>(null)
const cursorLastTime = ref(0)
const canvasRequestedFrame = ref(false)
const canvasRenderTimeout = ref<number | null>(null)

const unsubscribePixelRatioChange = ref<(() => void) | null>(null)

function onPixelRatioChange() {
  if (unsubscribePixelRatioChange.value) {
    unsubscribePixelRatioChange.value()
  }

  const media = window.matchMedia(`(resolution: ${window.devicePixelRatio}dppx)`)
  media.addEventListener('change', onPixelRatioChange)
  unsubscribePixelRatioChange.value = () => {
    media.removeEventListener('change', onPixelRatioChange)
  }

  canvasScale.value = window.devicePixelRatio
  onCanvasSizeChange(props.canvasSize)
}

function onCanvasSizeChange({ width, height }: Dimension) {
  canvasResize({ width, height })
  canvasRequestRedraw()
}

watch(() => props.canvasSize, onCanvasSizeChange)

function onCursorPosition(data: CursorPosition) {
  if (!props.isControling) {
    cursorPosition.value = data
    canvasRequestRedraw()
  }
}

function onCursorImage(data: CursorImage) {
  cursorImage.value = data

  if (!props.isControling) {
    cursorElement.src = data.uri
  }
}

function canvasResize({ width, height }: Dimension) {
  overlay.value!.width = width * canvasScale.value
  overlay.value!.height = height * canvasScale.value
  ctx.value!.setTransform(canvasScale.value, 0, 0, canvasScale.value, 0, 0)
}

function canvasRequestRedraw() {
  if (canvasRequestedFrame.value) return

  if (props.fps > 0) {
    if (canvasRenderTimeout.value) {
      window.clearTimeout(canvasRenderTimeout.value)
      canvasRenderTimeout.value = null
    }

    const now = Date.now()
    if (now - cursorLastTime.value < 1000 / props.fps) {
      canvasRenderTimeout.value = window.setTimeout(canvasRequestRedraw, 1000 / props.fps)
      return
    }

    cursorLastTime.value = now
  }

  canvasRequestedFrame.value = true
  window.requestAnimationFrame(() => {
    if (props.isControling) {
      canvasClear()
    } else {
      canvasRedraw()
    }

    canvasRequestedFrame.value = false
  })
}

watch(() => props.hostId, canvasRequestRedraw)
watch(() => props.cursorDraw, canvasRequestRedraw)

function canvasRedraw() {
  if (!cursorPosition.value || !props.screenSize || !cursorImage.value) return

  // clear drawings
  canvasClear()

  // ignore hidden cursor
  if (cursorImage.value.width <= 1 && cursorImage.value.height <= 1) return

  // get intrinsic dimensions
  const { width, height } = props.canvasSize

  // reset transformation, X and Y will be 0 again
  ctx.value!.setTransform(canvasScale.value, 0, 0, canvasScale.value, 0, 0)

  // get cursor position
  let x = Math.round((cursorPosition.value.x / props.screenSize.width) * width)
  let y = Math.round((cursorPosition.value.y / props.screenSize.height) * height)

  // use custom draw function, if available
  if (props.cursorDraw) {
    props.cursorDraw(ctx.value!, x, y, cursorElement, cursorImage.value, props.hostId)
    return
  }

  // draw cursor image
  ctx.value!.drawImage(
    cursorElement,
    x - cursorImage.value.x,
    y - cursorImage.value.y,
    cursorImage.value.width,
    cursorImage.value.height,
  )

  // draw cursor tag
  const cursorTag = props.sessions[props.hostId]?.profile.name || ''
  if (cursorTag) {
    x += cursorImage.value.width
    y += cursorImage.value.height

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
}

function canvasClear() {
  // reset transformation, X and Y will be 0 again
  ctx.value!.setTransform(canvasScale.value, 0, 0, canvasScale.value, 0, 0)

  const { width, height } = props.canvasSize
  ctx.value!.clearRect(0, 0, width, height)
}

//
// implicit hosting
//

const reqMouseDown = ref<MouseEvent | null>(null)
const reqMouseUp = ref<MouseEvent | null>(null)

function onControlChange(isControling: boolean) {
  keyboardModifiers.value = null

  if (isControling && reqMouseDown.value) {
    updateKeyboardModifiers(reqMouseDown.value)
    onMouseDown(reqMouseDown.value)
  }

  if (isControling && reqMouseUp.value) {
    onMouseUp(reqMouseUp.value)
  }

  canvasRequestRedraw()

  reqMouseDown.value = null
  reqMouseUp.value = null
}

watch(() => props.isControling, onControlChange)

function implicitControlRequest(e: MouseEvent) {
  if (props.implicitControl && e.type === 'mousedown') {
    reqMouseDown.value = e
    reqMouseUp.value = null
    props.control.request()
  }

  if (props.implicitControl && e.type === 'mouseup') {
    reqMouseUp.value = e
  }
}

// unused
function implicitControlRelease() {
  if (props.implicitControl) {
    props.control.release()
  }
}

//
// mobile keyboard
//

const kbdShow = ref(false)
const kbdOpen = ref(false)

function mobileKeyboardShow() {
  // skip if not a touch device
  if (!props.hasMobileKeyboard) return

  kbdShow.value = true
  kbdOpen.value = false

  textarea.value!.focus()
  window.visualViewport?.addEventListener('resize', onVisualViewportResize)
  emit('mobileKeyboardOpen', true)
}

function mobileKeyboardHide() {
  // skip if not a touch device
  if (!props.hasMobileKeyboard) return

  kbdShow.value = false
  kbdOpen.value = false

  emit('mobileKeyboardOpen', false)
  window.visualViewport?.removeEventListener('resize', onVisualViewportResize)
  textarea.value!.blur()
}

// visual viewport resize event is fired when keyboard is opened or closed
// android does not blur textarea when keyboard is closed, so we need to do it manually
function onVisualViewportResize() {
  if (!kbdShow.value) return

  if (!kbdOpen.value) {
    kbdOpen.value = true
  } else {
    mobileKeyboardHide()
  }
}

</script>
