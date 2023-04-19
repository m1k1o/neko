<template>
  <div class="neko-overlay-wrap">
    <canvas ref="overlay" class="neko-overlay" tabindex="0" />
    <textarea
      ref="textarea"
      class="neko-overlay"
      :style="{ cursor }"
      v-model="textInput"
      @click.stop.prevent="wsControl.emit('overlay.click', $event)"
      @contextmenu.stop.prevent="wsControl.emit('overlay.contextmenu', $event)"
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

<script lang="ts">
  import { Vue, Component, Ref, Prop, Watch } from 'vue-property-decorator'

  import { KeyboardInterface, NewKeyboard } from './utils/keyboard'
  import { KeyTable, keySymsRemap } from './utils/keyboard-remapping'
  import { getFilesFromDataTansfer } from './utils/file-upload'
  import { NekoControl } from './internal/control'
  import { NekoWebRTC } from './internal/webrtc'
  import { Session, Scroll } from './types/state'
  import { CursorPosition, CursorImage } from './types/webrtc'
  import { CursorDrawFunction, Dimension, KeyboardModifiers } from './types/cursors'

  const WHEEL_STEP = 53 // Delta threshold for a mouse wheel step
  const WHEEL_LINE_HEIGHT = 19

  const MOUSE_MOVE_THROTTLE = 1000 / 60 // in ms, 60fps
  const INACTIVE_CURSOR_INTERVAL = 1000 / 4 // in ms, 4fps

  @Component({
    name: 'neko-overlay',
  })
  export default class extends Vue {
    @Ref('overlay') readonly _overlay!: HTMLCanvasElement
    @Ref('textarea') readonly _textarea!: HTMLTextAreaElement
    private _ctx!: CanvasRenderingContext2D

    private canvasScale = window.devicePixelRatio

    private keyboard!: KeyboardInterface
    private textInput = ''

    private focused = false

    @Prop()
    private readonly wsControl!: NekoControl

    @Prop()
    private readonly sessions!: Record<string, Session>

    @Prop()
    private readonly hostId!: string

    @Prop()
    private readonly webrtc!: NekoWebRTC

    @Prop()
    private readonly scroll!: Scroll

    @Prop()
    private readonly screenSize!: Dimension

    @Prop()
    private readonly canvasSize!: Dimension

    @Prop()
    private readonly cursorDraw!: CursorDrawFunction | null

    @Prop()
    private readonly isControling!: boolean

    @Prop()
    private readonly implicitControl!: boolean

    @Prop()
    private readonly inactiveCursors!: boolean

    @Prop()
    private readonly fps!: number

    @Prop()
    private readonly hasMobileKeyboard!: boolean

    get cursor(): string {
      if (!this.isControling || !this.cursorImage) {
        return 'default'
      }

      const { uri, x, y } = this.cursorImage
      return 'url(' + uri + ') ' + x + ' ' + y + ', default'
    }

    mounted() {
      // register mouseup globally as user can release mouse button outside of overlay
      window.addEventListener('mouseup', this.onMouseUp, true)

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

      let ctrlKey = 0
      let noKeyUp = {} as Record<number, boolean>

      // Initialize Keyboard
      this.keyboard = NewKeyboard()
      this.keyboard.onkeydown = (key: number) => {
        key = keySymsRemap(key)

        if (!this.isControling) {
          noKeyUp[key] = true
          return true
        }

        // ctrl+v is aborted
        if (ctrlKey != 0 && key == KeyTable.XK_v) {
          this.keyboard.release(ctrlKey)
          noKeyUp[key] = true
          return true
        }

        // save information if it is ctrl key event
        const isCtrlKey = key == KeyTable.XK_Control_L || key == KeyTable.XK_Control_R
        if (isCtrlKey) ctrlKey = key

        if (this.webrtc.connected) {
          this.webrtc.send('keydown', { key })
        } else {
          this.wsControl.keyDown(key)
        }

        return isCtrlKey
      }
      this.keyboard.onkeyup = (key: number) => {
        key = keySymsRemap(key)

        if (key in noKeyUp) {
          delete noKeyUp[key]
          return
        }

        const isCtrlKey = key == KeyTable.XK_Control_L || key == KeyTable.XK_Control_R
        if (isCtrlKey) ctrlKey = 0

        if (this.webrtc.connected) {
          this.webrtc.send('keyup', { key })
        } else {
          this.wsControl.keyUp(key)
        }
      }
      this.keyboard.listenTo(this._textarea)

      this.webrtc.addListener('cursor-position', this.onCursorPosition)
      this.webrtc.addListener('cursor-image', this.onCursorImage)
      this.webrtc.addListener('disconnected', this.canvasClear)
      this.cursorElement.onload = this.canvasRequestRedraw
    }

    beforeDestroy() {
      window.removeEventListener('mouseup', this.onMouseUp, true)

      if (this.keyboard) {
        this.keyboard.removeListener()
      }

      this.webrtc.removeListener('cursor-position', this.onCursorPosition)
      this.webrtc.removeListener('cursor-image', this.onCursorImage)
      this.webrtc.removeListener('disconnected', this.canvasClear)
      this.cursorElement.onload = null

      // stop inactive cursor interval if exists
      this.clearInactiveCursorInterval()

      // stop pixel ratio change listener
      if (this.unsubscribePixelRatioChange) {
        this.unsubscribePixelRatioChange()
      }
    }

    getMousePos(clientX: number, clientY: number) {
      const rect = this._overlay.getBoundingClientRect()

      return {
        x: Math.round((this.screenSize.width / rect.width) * (clientX - rect.left)),
        y: Math.round((this.screenSize.height / rect.height) * (clientY - rect.top)),
      }
    }

    sendMousePos(e: MouseEvent) {
      const pos = this.getMousePos(e.clientX, e.clientY)
      this.webrtc.send('mousemove', pos)
      this.cursorPosition = pos
    }

    private wheelX = 0
    private wheelY = 0
    private wheelTimeStamp = 0

    // negative sensitivity can be acheived using increased step value
    get wheelStep() {
      let x = WHEEL_STEP

      if (this.scroll.sensitivity < 0) {
        x *= Math.abs(this.scroll.sensitivity) + 1
      }

      return x
    }

    // sensitivity can only be positive
    get wheelSensitivity() {
      let x = 1

      if (this.scroll.sensitivity > 0) {
        x = Math.abs(this.scroll.sensitivity) + 1
      }

      if (this.scroll.inverse) {
        x *= -1
      }

      return x
    }

    // use v-model instead of @input because v-model
    // doesn't get updated during IME composition
    @Watch('textInput')
    onTextInputChange() {
      if (this.textInput == '') return
      this.wsControl.paste(this.textInput)
      this.textInput = ''
    }

    onWheel(e: WheelEvent) {
      if (!this.isControling) {
        return
      }

      // when the last scroll was more than 250ms ago
      const firstScroll = e.timeStamp - this.wheelTimeStamp > 250

      if (firstScroll) {
        this.wheelX = 0
        this.wheelY = 0
        this.wheelTimeStamp = e.timeStamp
      }

      let dx = e.deltaX
      let dy = e.deltaY

      if (e.deltaMode !== 0) {
        dx *= WHEEL_LINE_HEIGHT
        dy *= WHEEL_LINE_HEIGHT
      }

      this.wheelX += dx
      this.wheelY += dy

      let x = 0
      if (Math.abs(this.wheelX) >= this.wheelStep || firstScroll) {
        if (this.wheelX < 0) {
          x = this.wheelSensitivity * -1
        } else if (this.wheelX > 0) {
          x = this.wheelSensitivity
        }

        if (!firstScroll) {
          this.wheelX = 0
        }
      }

      let y = 0
      if (Math.abs(this.wheelY) >= this.wheelStep || firstScroll) {
        if (this.wheelY < 0) {
          y = this.wheelSensitivity * -1
        } else if (this.wheelY > 0) {
          y = this.wheelSensitivity
        }

        if (!firstScroll) {
          this.wheelY = 0
        }
      }

      // skip if not scrolled
      if (x == 0 && y == 0) return

      if (this.webrtc.connected) {
        this.sendMousePos(e)
        this.webrtc.send('wheel', { x, y })
      } else {
        this.wsControl.scroll({ x, y })
      }
    }

    lastMouseMove = 0

    onMouseMove(e: MouseEvent) {
      // throttle mousemove events
      if (e.timeStamp - this.lastMouseMove < MOUSE_MOVE_THROTTLE) return
      this.lastMouseMove = e.timeStamp

      if (this.isControling) {
        this.sendMousePos(e)
      }

      if (this.inactiveCursors) {
        this.saveInactiveMousePos(e)
      }
    }

    isMouseDown = false

    onMouseDown(e: MouseEvent) {
      this.isMouseDown = true

      if (!this.isControling) {
        this.implicitControlRequest(e)
        return
      }

      const key = e.button + 1
      if (this.webrtc.connected) {
        this.sendMousePos(e)
        this.webrtc.send('mousedown', { key })
      } else {
        const pos = this.getMousePos(e.clientX, e.clientY)
        this.wsControl.buttonDown(key, pos)
      }
    }

    onMouseUp(e: MouseEvent) {
      // only if we are the one who started the mouse down
      if (!this.isMouseDown) return
      this.isMouseDown = false

      if (!this.isControling) {
        this.implicitControlRequest(e)
        return
      }

      const key = e.button + 1
      if (this.webrtc.connected) {
        this.sendMousePos(e)
        this.webrtc.send('mouseup', { key })
      } else {
        const pos = this.getMousePos(e.clientX, e.clientY)
        this.wsControl.buttonUp(key, pos)
      }
    }

    onMouseEnter(e: MouseEvent) {
      // focus opens the keyboard on mobile (only for android)
      if (!this.hasMobileKeyboard) {
        this._textarea.focus()
      }

      this.focused = true

      if (this.isControling) {
        this.updateKeyboardModifiers(e)
      }
    }

    onMouseLeave(e: MouseEvent) {
      if (this.isControling) {
        // save current keyboard modifiers state
        this.keyboardModifiers = {
          capslock: e.getModifierState('CapsLock'),
          numlock: e.getModifierState('NumLock'),
        }
      }

      this.focused = false
    }

    onDragEnter(e: DragEvent) {
      this.onMouseEnter(e as MouseEvent)
    }

    onDragLeave(e: DragEvent) {
      this.onMouseLeave(e as MouseEvent)
    }

    onDragOver(e: DragEvent) {
      this.onMouseMove(e as MouseEvent)
    }

    async onDrop(e: DragEvent) {
      if (this.isControling || this.implicitControl) {
        let dt = e.dataTransfer
        if (!dt) return

        const files = await getFilesFromDataTansfer(dt)
        if (files.length === 0) return

        this.$emit('uploadDrop', { ...this.getMousePos(e.clientX, e.clientY), files })
      }
    }

    //
    // inactive cursor position
    //

    private inactiveCursorInterval: number | null = null
    private inactiveCursorPosition: CursorPosition | null = null

    clearInactiveCursorInterval() {
      if (this.inactiveCursorInterval) {
        window.clearInterval(this.inactiveCursorInterval)
        this.inactiveCursorInterval = null
      }
    }

    @Watch('focused')
    @Watch('isControling')
    @Watch('inactiveCursors')
    restartInactiveCursorInterval() {
      // clear interval if exists
      this.clearInactiveCursorInterval()

      if (this.inactiveCursors && this.focused && !this.isControling) {
        this.inactiveCursorInterval = window.setInterval(this.sendInactiveMousePos.bind(this), INACTIVE_CURSOR_INTERVAL)
      }
    }

    saveInactiveMousePos(e: MouseEvent) {
      const pos = this.getMousePos(e.clientX, e.clientY)
      this.inactiveCursorPosition = pos
    }

    sendInactiveMousePos() {
      if (this.inactiveCursorPosition) {
        this.webrtc.send('mousemove', this.inactiveCursorPosition)
      }
    }

    //
    // keyboard modifiers
    //

    private keyboardModifiers: KeyboardModifiers | null = null

    updateKeyboardModifiers(e: MouseEvent) {
      const capslock = e.getModifierState('CapsLock')
      const numlock = e.getModifierState('NumLock')

      if (
        this.keyboardModifiers === null ||
        this.keyboardModifiers.capslock !== capslock ||
        this.keyboardModifiers.numlock !== numlock
      ) {
        this.$emit('updateKeyboardModifiers', { capslock, numlock })
      }
    }

    //
    // canvas
    //

    private cursorImage: CursorImage | null = null
    private cursorElement: HTMLImageElement = new Image()
    private cursorPosition: CursorPosition | null = null
    private cursorLastTime = 0
    private canvasRequestedFrame = false
    private canvasRenderTimeout: number | null = null

    private unsubscribePixelRatioChange?: () => void
    private onPixelRatioChange() {
      if (this.unsubscribePixelRatioChange) {
        this.unsubscribePixelRatioChange()
      }

      const media = window.matchMedia(`(resolution: ${window.devicePixelRatio}dppx)`)
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
      this.canvasRequestRedraw()
    }

    onCursorPosition(data: CursorPosition) {
      if (!this.isControling) {
        this.cursorPosition = data
        this.canvasRequestRedraw()
      }
    }

    onCursorImage(data: CursorImage) {
      this.cursorImage = data

      if (!this.isControling) {
        this.cursorElement.src = data.uri
      }
    }

    canvasResize({ width, height }: Dimension) {
      this._overlay.width = width * this.canvasScale
      this._overlay.height = height * this.canvasScale
      this._ctx.setTransform(this.canvasScale, 0, 0, this.canvasScale, 0, 0)
    }

    @Watch('hostId')
    @Watch('cursorDraw')
    canvasRequestRedraw() {
      // skip rendering if there is already in progress
      if (this.canvasRequestedFrame) return

      // throttle rendering according to fps
      if (this.fps > 0) {
        if (this.canvasRenderTimeout) {
          window.clearTimeout(this.canvasRenderTimeout)
          this.canvasRenderTimeout = null
        }

        const now = Date.now()
        if (now - this.cursorLastTime < 1000 / this.fps) {
          // ensure that last frame is rendered
          this.canvasRenderTimeout = window.setTimeout(this.canvasRequestRedraw, 1000 / this.fps)
          return
        }

        this.cursorLastTime = now
      }

      // request animation frame from a browser
      this.canvasRequestedFrame = true
      window.requestAnimationFrame(() => {
        if (this.isControling) {
          this.canvasClear()
        } else {
          this.canvasRedraw()
        }

        this.canvasRequestedFrame = false
      })
    }

    canvasRedraw() {
      if (!this.cursorPosition || !this.screenSize || !this.cursorImage) return

      // clear drawings
      this.canvasClear()

      // ignore hidden cursor
      if (this.cursorImage.width <= 1 && this.cursorImage.height <= 1) return

      // get intrinsic dimensions
      const { width, height } = this.canvasSize

      // reset transformation, X and Y will be 0 again
      this._ctx.setTransform(this.canvasScale, 0, 0, this.canvasScale, 0, 0)

      // get cursor position
      let x = Math.round((this.cursorPosition.x / this.screenSize.width) * width)
      let y = Math.round((this.cursorPosition.y / this.screenSize.height) * height)

      // use custom draw function, if available
      if (this.cursorDraw) {
        this.cursorDraw(this._ctx, x, y, this.cursorElement, this.cursorImage, this.hostId)
        return
      }

      // draw cursor image
      this._ctx.drawImage(
        this.cursorElement,
        x - this.cursorImage.x,
        y - this.cursorImage.y,
        this.cursorImage.width,
        this.cursorImage.height,
      )

      // draw cursor tag
      const cursorTag = this.sessions[this.hostId]?.profile.name || ''
      if (cursorTag) {
        x += this.cursorImage.width
        y += this.cursorImage.height

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
    }

    canvasClear() {
      // reset transformation, X and Y will be 0 again
      this._ctx.setTransform(this.canvasScale, 0, 0, this.canvasScale, 0, 0)

      const { width, height } = this.canvasSize
      this._ctx.clearRect(0, 0, width, height)
    }

    //
    // implicit hosting
    //

    private reqMouseDown: MouseEvent | null = null
    private reqMouseUp: MouseEvent | null = null

    @Watch('isControling')
    onControlChange(isControling: boolean) {
      this.keyboardModifiers = null

      if (isControling && this.reqMouseDown) {
        this.updateKeyboardModifiers(this.reqMouseDown)
        this.onMouseDown(this.reqMouseDown)
      }

      if (isControling && this.reqMouseUp) {
        this.onMouseUp(this.reqMouseUp)
      }

      this.canvasRequestRedraw()

      this.reqMouseDown = null
      this.reqMouseUp = null
    }

    implicitControlRequest(e: MouseEvent) {
      if (this.implicitControl && e.type === 'mousedown') {
        this.reqMouseDown = e
        this.reqMouseUp = null
        this.wsControl.request()
      }

      if (this.implicitControl && e.type === 'mouseup') {
        this.reqMouseUp = e
      }
    }

    // unused
    implicitControlRelease() {
      if (this.implicitControl) {
        this.wsControl.release()
      }
    }

    //
    // mobile keyboard
    //

    public kbdShow = false
    public kbdOpen = false

    public mobileKeyboardShow() {
      // skip if not a touch device
      if (!this.hasMobileKeyboard) return

      this.kbdShow = true
      this.kbdOpen = false

      this._textarea.focus()
      window.visualViewport?.addEventListener('resize', this.onVisualViewportResize)
      this.$emit('mobileKeyboardOpen', true)
    }

    public mobileKeyboardHide() {
      // skip if not a touch device
      if (!this.hasMobileKeyboard) return

      this.kbdShow = false
      this.kbdOpen = false

      this.$emit('mobileKeyboardOpen', false)
      window.visualViewport?.removeEventListener('resize', this.onVisualViewportResize)
      this._textarea.blur()
    }

    // visual viewport resize event is fired when keyboard is opened or closed
    // android does not blur textarea when keyboard is closed, so we need to do it manually
    onVisualViewportResize() {
      if (!this.kbdShow) return

      if (!this.kbdOpen) {
        this.kbdOpen = true
      } else {
        this.mobileKeyboardHide()
      }
    }
  }
</script>
