<template>
  <div class="neko-overlay-wrap">
    <canvas ref="overlay" class="neko-overlay" tabindex="0" />
    <textarea
      ref="textarea"
      class="neko-overlay"
      :style="{ cursor }"
      @click.stop.prevent="$emit('onAction', { action: 'click', target: $event })"
      @contextmenu.stop.prevent="$emit('onAction', { action: 'contextmenu', target: $event })"
      @input.stop.prevent="onInput"
      @wheel.stop.prevent="onWheel"
      @mousemove.stop.prevent="onMouseMove"
      @mousedown.stop.prevent="onMouseDown"
      @mouseup.stop.prevent="onMouseUp"
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
    font-size: 1px; /* chrome would not paste text if 0px */
    resize: none; /* hide textarea resize corner */
    outline: 0;
    border: 0;
    color: transparent;
    background: transparent;
  }
</style>

<script lang="ts">
  import { Vue, Component, Ref, Prop, Watch } from 'vue-property-decorator'

  import GuacamoleKeyboard from './utils/guacamole-keyboard'
  import { KeyTable, keySymsRemap } from './utils/keyboard-remapping'
  import { getFilesFromDataTansfer } from './utils/file-upload'
  import { NekoWebRTC } from './internal/webrtc'
  import { Session, Scroll } from './types/state'
  import { CursorPosition, CursorImage } from './types/webrtc'
  import { CursorDrawFunction, Dimension, KeyboardModifiers } from './types/cursors'

  const WHEEL_STEP = 53 // Delta threshold for a mouse wheel step
  const WHEEL_LINE_HEIGHT = 19

  const CANVAS_SCALE = 2

  const INACTIVE_CURSOR_INTERVAL = 250 // ms

  @Component({
    name: 'neko-overlay',
  })
  export default class extends Vue {
    @Ref('overlay') readonly _overlay!: HTMLCanvasElement
    @Ref('textarea') readonly _textarea!: HTMLTextAreaElement
    private _ctx!: CanvasRenderingContext2D

    private keyboard = GuacamoleKeyboard()
    private focused = false

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

    get cursor(): string {
      if (!this.isControling || !this.cursorImage) {
        return 'default'
      }

      const { uri, x, y } = this.cursorImage
      return 'url(' + uri + ') ' + x + ' ' + y + ', default'
    }

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

      let ctrlKey = 0
      let noKeyUp = {} as Record<number, boolean>

      // Initialize Guacamole Keyboard
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

        this.webrtc.send('keydown', { key })
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

        this.webrtc.send('keyup', { key })
      }
      this.keyboard.listenTo(this._textarea)

      this.webrtc.addListener('cursor-position', this.onCursorPosition)
      this.webrtc.addListener('cursor-image', this.onCursorImage)
      this.webrtc.addListener('disconnected', this.canvasClear)
      this.cursorElement.onload = this.canvasRequestRedraw
    }

    beforeDestroy() {
      // Guacamole Keyboard does not provide destroy functions

      this.webrtc.removeListener('cursor-position', this.onCursorPosition)
      this.webrtc.removeListener('cursor-image', this.onCursorImage)
      this.webrtc.removeListener('disconnected', this.canvasClear)
      this.cursorElement.onload = null

      // stop inactive cursor interval if exists
      if (this.inactiveCursorInterval !== null) {
        window.clearInterval(this.inactiveCursorInterval)
        this.inactiveCursorInterval = null
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
      Vue.set(this, 'cursorPosition', pos)
    }

    private wheelX = 0
    private wheelY = 0
    private wheelDate = Date.now()

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

    onInput(e: InputEvent) {
      this.$emit('clipboard', this._textarea.value)
      this._textarea.value = ''
    }

    onWheel(e: WheelEvent) {
      if (!this.isControling) {
        return
      }

      this.sendMousePos(e)

      const now = Date.now()
      const firstScroll = now - this.wheelDate > 250

      if (firstScroll) {
        this.wheelX = 0
        this.wheelY = 0
        this.wheelDate = now
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

      if (x != 0 || y != 0) {
        this.webrtc.send('wheel', { x, y })
      }
    }

    onMouseMove(e: MouseEvent) {
      if (this.isControling) {
        this.sendMousePos(e)
      }

      if (this.inactiveCursors) {
        this.saveInactiveMousePos(e)
      }
    }

    onMouseDown(e: MouseEvent) {
      if (!this.isControling) {
        this.implicitControlRequest(e)
        return
      }

      this.sendMousePos(e)
      this.webrtc.send('mousedown', { key: e.button + 1 })
    }

    onMouseUp(e: MouseEvent) {
      if (!this.isControling) {
        this.implicitControlRequest(e)
        return
      }

      this.sendMousePos(e)
      this.webrtc.send('mouseup', { key: e.button + 1 })
    }

    onMouseEnter(e: MouseEvent) {
      this._textarea.focus()
      this.focused = true

      if (this.isControling) {
        this.updateKeyboardModifiers(e)
      }
    }

    onMouseLeave(e: MouseEvent) {
      if (this.isControling) {
        this.keyboard.reset()

        // save current keyboard modifiers state
        Vue.set(this, 'keyboardModifiers', {
          capslock: e.getModifierState('CapsLock'),
          numlock: e.getModifierState('NumLock'),
        })
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

    @Watch('focused')
    @Watch('isControling')
    @Watch('inactiveCursors')
    restartInactiveCursorInterval() {
      // clear interval if exists
      if (this.inactiveCursorInterval !== null) {
        window.clearInterval(this.inactiveCursorInterval)
        this.inactiveCursorInterval = null
      }

      if (this.inactiveCursors && this.focused && !this.isControling) {
        this.inactiveCursorInterval = window.setInterval(this.sendInactiveMousePos.bind(this), INACTIVE_CURSOR_INTERVAL)
      }
    }

    saveInactiveMousePos(e: MouseEvent) {
      const pos = this.getMousePos(e.clientX, e.clientY)
      Vue.set(this, 'inactiveCursorPosition', pos)
    }

    sendInactiveMousePos() {
      if (this.inactiveCursorPosition != null) {
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
    private canvasRequestedFrame = false

    @Watch('canvasSize')
    onCanvasSizeChange({ width, height }: Dimension) {
      this._overlay.width = width * CANVAS_SCALE
      this._overlay.height = height * CANVAS_SCALE
      this.canvasRequestRedraw()
    }

    onCursorPosition(data: CursorPosition) {
      if (!this.isControling) {
        Vue.set(this, 'cursorPosition', data)
        this.canvasRequestRedraw()
      }
    }

    onCursorImage(data: CursorImage) {
      Vue.set(this, 'cursorImage', data)

      if (!this.isControling) {
        this.cursorElement.src = data.uri
      }
    }

    @Watch('hostId')
    @Watch('cursorDraw')
    canvasRequestRedraw() {
      // skip rendering if there is already in progress
      if (this.canvasRequestedFrame) return

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
      let { width, height } = this.canvasSize
      this._ctx.setTransform(CANVAS_SCALE, 0, 0, CANVAS_SCALE, 0, 0)

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
    }

    canvasClear() {
      const { width, height } = this._overlay
      this._ctx.clearRect(0, 0, width, height)
    }

    //
    // implicit hosting
    //

    private reqMouseDown: MouseEvent | null = null
    private reqMouseUp: MouseEvent | null = null

    @Watch('isControling')
    onControlChange(isControling: boolean) {
      Vue.set(this, 'keyboardModifiers', null)

      if (isControling && this.reqMouseDown) {
        this.updateKeyboardModifiers(this.reqMouseDown)
        this.sendMousePos(this.reqMouseDown)
        this.webrtc.send('mousedown', { key: this.reqMouseDown.button + 1 })
      }

      if (isControling && this.reqMouseUp) {
        this.webrtc.send('mouseup', { key: this.reqMouseUp.button + 1 })
      }

      this.canvasRequestRedraw()

      this.reqMouseDown = null
      this.reqMouseUp = null
    }

    implicitControlRequest(e: MouseEvent) {
      if (this.implicitControl && e.type === 'mousedown' && this.reqMouseDown == null) {
        this.reqMouseDown = e
        this.$emit('implicitControlRequest')
      }

      if (this.implicitControl && e.type === 'mouseup' && this.reqMouseUp == null) {
        this.reqMouseUp = e
      }
    }

    // unused
    implicitControlRelease() {
      if (this.implicitControl) {
        this.$emit('implicitControlRelease')
      }
    }
  }
</script>
