<template>
  <canvas
    ref="overlay"
    class="neko-overlay"
    tabindex="0"
    :style="{ cursor }"
    @click.stop.prevent
    @contextmenu.stop.prevent
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
</template>

<style lang="scss" scoped>
  .neko-overlay {
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

  import GuacamoleKeyboard from './utils/guacamole-keyboard'
  import { keySymsRemap } from './utils/keyboard-remapping'
  import { getFilesFromDataTansfer } from './utils/file-upload'
  import { NekoWebRTC } from './internal/webrtc'
  import { Scroll } from './types/state'

  const inactiveCursorWin10 =
    'url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAACEUlEQVR4nOzWz6sSURQH8O+89zJ5C32LKbAgktCSaPpBSL' +
    'uSNtHqLcOV+BeIGxei0oCtFME/wI0bF4GCK6mNuAghH7xFlBAO7bQoA/Vik3riyghTaCQzTsLzbIZZDPdzzj3nzt3Df44dYDsBRNSYTqcn5XL5KoADy1VERL' +
    'Is02g0+phIJG4BsFkOEEVxjhgOh59kWb5rKWIBWCAGg0EnFovdtgyhB+grkU6n7wA4ZzlgCWKzlVgGsLQnVgE2gVh7xvP5PH9ciUajFQDHyWTyHQDVKOS3+F' +
    'sF/pyOcDh83Uhj/nMFBEFANpuF0+nUQ92SJD0G8AXAdwAz0wE+nw8OhwPNZhPj8RiBQOC0Vqu9EgSBcrnc11Qq9R7AeW5cd/GVsdgCr9dLiqJQtVqdv/v9fm' +
    'KM9UVRfArgJoBrAC4DsJsOcLlc1Gg0qNVqVRljI0mS5oh6vU6lUukFgEta5gemLr4AFAoF6nQ6b20223G73X6ZyWTmgFAoRL1ej3f+DQ1gfqiq+qbf73/weD' +
    'zPADwoFouPut3uzO12UyQSoclkotrt9ocAHKZnr8UhAP4bvg/gIs+UMfaaMTZTFOUkHo8/B/AEwAWjl5pV+j1dZ//g4xUMBo8YY/cqlcqhNvffAJxq40dmA5' +
    'bFPoAjrev5EfwZQNfoKbju/u1ri/PvfgKYGMl+K2I7b8U7wA5wpgC/AgAA///Yyif1MZXzRQAAAABJRU5ErkJggg==) 4 4, crosshair'

  const WHEEL_STEP = 50 // Delta threshold for a mouse wheel step
  const WHEEL_LINE_HEIGHT = 19

  @Component({
    name: 'neko-overlay',
  })
  export default class extends Vue {
    @Ref('overlay') readonly _overlay!: HTMLCanvasElement
    private _ctx: any = null

    private keyboard = GuacamoleKeyboard()
    private focused = false

    @Prop()
    private readonly webrtc!: NekoWebRTC

    @Prop()
    private readonly scroll!: Scroll

    @Prop()
    private readonly screenSize!: { width: number; height: number }

    @Prop()
    private readonly canvasSize!: { width: number; height: number }

    @Prop()
    private readonly cursorTag!: string

    @Prop()
    private readonly isControling!: boolean

    @Prop()
    private readonly implicitControl!: boolean

    get cursor(): string {
      if (!this.isControling) {
        return inactiveCursorWin10
      }

      if (!this.cursorImage) {
        return 'auto'
      }

      const { uri, x, y } = this.cursorImage
      return 'url(' + uri + ') ' + x + ' ' + y + ', auto'
    }

    mounted() {
      this._ctx = this._overlay.getContext('2d')

      // synchronize intrinsic with extrinsic dimensions
      const { width, height } = this._overlay.getBoundingClientRect()
      this._overlay.width = width
      this._overlay.height = height

      // Initialize Guacamole Keyboard
      this.keyboard.onkeydown = (key: number) => {
        if (!this.focused) {
          return true
        }

        if (!this.isControling) {
          return true
        }

        this.webrtc.send('keydown', {
          key: keySymsRemap(key),
        })
        return false
      }
      this.keyboard.onkeyup = (key: number) => {
        if (!this.focused) {
          return
        }

        if (!this.isControling) {
          return
        }

        this.webrtc.send('keyup', {
          key: keySymsRemap(key),
        })
      }
      this.keyboard.listenTo(this._overlay)

      this.webrtc.addListener('cursor-position', this.onCursorPosition)
      this.webrtc.addListener('cursor-image', this.onCursorImage)
      this.webrtc.addListener('disconnected', this.canvasClear)
      this.cursorElement.onload = this.canvasRedraw
    }

    beforeDestroy() {
      // Guacamole Keyboard does not provide destroy functions

      this.webrtc.removeListener('cursor-position', this.onCursorPosition)
      this.webrtc.removeListener('cursor-image', this.onCursorImage)
      this.webrtc.removeListener('disconnected', this.canvasClear)
      this.cursorElement.onload = null
    }

    getMousePos(clientX: number, clientY: number) {
      const rect = this._overlay.getBoundingClientRect()

      return {
        x: Math.round((this.screenSize.width / rect.width) * (clientX - rect.left)),
        y: Math.round((this.screenSize.height / rect.height) * (clientY - rect.top)),
      }
    }

    setMousePos(e: MouseEvent) {
      const pos = this.getMousePos(e.clientX, e.clientY)
      this.webrtc.send('mousemove', pos)
      Vue.set(this, 'cursorPosition', pos)
    }

    private wheelX = 0
    private wheelY = 0

    // negative sensitivity can be acheived using increased step value
    get wheelStep() {
      let x = 20 * (window.devicePixelRatio || 1)

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

    onWheel(e: WheelEvent) {
      if (!this.isControling) {
        return
      }

      this.setMousePos(e)

      let dx = e.deltaX
      let dy = e.deltaY

      if (e.deltaMode !== 0) {
        dx *= WHEEL_LINE_HEIGHT
        dy *= WHEEL_LINE_HEIGHT
      }

      this.wheelX += dx
      this.wheelY += dy

      console.log(typeof dx, dx, typeof dy, dy, this.wheelX, this.wheelY)

      let x = 0
      if (Math.abs(this.wheelX) >= this.wheelStep) {
        if (this.wheelX < 0) {
          x = this.wheelSensitivity * -1
        } else if (this.wheelX > 0) {
          x = this.wheelSensitivity
        }

        this.wheelX = 0
      }

      let y = 0
      if (Math.abs(this.wheelY) >= this.wheelStep) {
        if (this.wheelY < 0) {
          y = this.wheelSensitivity * -1
        } else if (this.wheelY > 0) {
          y = this.wheelSensitivity
        }

        this.wheelY = 0
      }

      if (x != 0 || y != 0) {
        this.webrtc.send('wheel', { x, y })
      }
    }

    onMouseMove(e: MouseEvent) {
      if (!this.isControling) {
        return
      }

      this.setMousePos(e)
    }

    onMouseDown(e: MouseEvent) {
      if (!this.isControling) {
        this.implicitControlRequest(e)
        return
      }

      this.setMousePos(e)
      this.webrtc.send('mousedown', { key: e.button + 1 })
    }

    onMouseUp(e: MouseEvent) {
      if (!this.isControling) {
        this.implicitControlRequest(e)
        return
      }

      this.setMousePos(e)
      this.webrtc.send('mouseup', { key: e.button + 1 })
    }

    onMouseEnter(e: MouseEvent) {
      this._overlay.focus()
      this.focused = true

      if (this.isControling) {
        this.updateKbdModifiers(e)
      }
    }

    onMouseLeave(e: MouseEvent) {
      if (this.isControling) {
        this.keyboard.reset()

        // save current kbd modifiers state
        Vue.set(this, 'kbdModifiers', {
          capslock: e.getModifierState('CapsLock'),
          numlock: e.getModifierState('NumLock'),
        })
      }

      this._overlay.blur()
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

        this.$emit('drop-files', { ...this.getMousePos(e.clientX, e.clientY), files })
      }
    }

    //
    // keyboard modifiers
    //

    private kbdModifiers: { capslock: boolean; numlock: boolean } | null = null

    updateKbdModifiers(e: MouseEvent) {
      const capslock = e.getModifierState('CapsLock')
      const numlock = e.getModifierState('NumLock')

      if (
        this.kbdModifiers === null ||
        this.kbdModifiers.capslock !== capslock ||
        this.kbdModifiers.numlock !== numlock
      ) {
        this.$emit('update-kbd-modifiers', { capslock, numlock })
      }
    }

    //
    // canvas
    //

    private cursorImage: { width: number; height: number; x: number; y: number; uri: string } | null = null
    private cursorElement: HTMLImageElement = new Image()
    private cursorPosition: { x: number; y: number } | null = null

    @Watch('cursorTag')
    onCursorTagChange() {
      if (!this.isControling) {
        this.canvasRedraw()
      }
    }

    @Watch('canvasSize')
    onCanvasSizeChange({ width, height }: { width: number; height: number }) {
      this._overlay.width = width
      this._overlay.height = height

      if (this.isControling) {
        this.canvasClear()
      } else {
        this.canvasRedraw()
      }
    }

    onCursorPosition(data: { x: number; y: number }) {
      if (!this.isControling) {
        Vue.set(this, 'cursorPosition', data)
        this.canvasRedraw()
      }
    }

    onCursorImage(data: { width: number; height: number; x: number; y: number; uri: string }) {
      Vue.set(this, 'cursorImage', data)

      if (!this.isControling) {
        this.cursorElement.src = data.uri
      }
    }

    canvasRedraw() {
      if (this.cursorPosition == null || this.screenSize == null || this.cursorImage == null) return

      // get intrinsic dimensions
      const { width, height } = this._overlay

      // clear drawings
      this._ctx.clearRect(0, 0, width, height)

      // ignore hidden cursor
      if (this.cursorImage.width <= 1 && this.cursorImage.height <= 1) return

      // redraw cursor
      let x = Math.round((this.cursorPosition.x / this.screenSize.width) * width - this.cursorImage.x)
      let y = Math.round((this.cursorPosition.y / this.screenSize.height) * height - this.cursorImage.y)
      this._ctx.drawImage(this.cursorElement, x, y, this.cursorImage.width, this.cursorImage.height)

      // redraw cursor tag
      if (this.cursorTag) {
        x += this.cursorImage.width + this.cursorImage.x
        y += this.cursorImage.height + this.cursorImage.y

        this._ctx.save()
        this._ctx.font = '14px Arial, sans-serif'
        this._ctx.textBaseline = 'top'
        this._ctx.shadowColor = 'black'
        this._ctx.shadowBlur = 2
        this._ctx.lineWidth = 2
        this._ctx.fillStyle = 'black'
        this._ctx.strokeText(this.cursorTag, x, y)
        this._ctx.shadowBlur = 0
        this._ctx.fillStyle = 'white'
        this._ctx.fillText(this.cursorTag, x, y)
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

    private reqMouseDown: any | null = null
    private reqMouseUp: any | null = null

    @Watch('isControling')
    onControlChange(isControling: boolean) {
      Vue.set(this, 'kbdModifiers', null)

      if (isControling && this.reqMouseDown) {
        this.updateKbdModifiers(this.reqMouseDown)
        this.setMousePos(this.reqMouseDown)
        this.webrtc.send('mousedown', { key: this.reqMouseDown.button + 1 })
      }

      if (isControling && this.reqMouseUp) {
        this.webrtc.send('mouseup', { key: this.reqMouseUp.button + 1 })
      }

      if (isControling) {
        this.canvasClear()
      } else {
        this.canvasRedraw()
      }

      this.reqMouseDown = null
      this.reqMouseUp = null
    }

    implicitControlRequest(e: MouseEvent) {
      if (this.implicitControl && e.type === 'mousedown' && this.reqMouseDown == null) {
        this.reqMouseDown = e
        this.$emit('implicit-control-request')
      }

      if (this.implicitControl && e.type === 'mouseup' && this.reqMouseUp == null) {
        this.reqMouseUp = e
      }
    }

    // unused
    implicitControlRelease() {
      if (this.implicitControl) {
        this.$emit('implicit-control-release')
      }
    }
  }
</script>
