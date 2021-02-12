<template>
  <canvas
    ref="overlay"
    class="neko-overlay"
    :class="isControling ? 'neko-active' : ''"
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

    &.neko-active {
      outline: 2px solid red;
      box-sizing: border-box;
    }

    .cursor {
      position: relative;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Ref, Prop, Watch } from 'vue-property-decorator'

  import GuacamoleKeyboard from './utils/guacamole-keyboard'
  import { getFilesFromDataTansfer } from './utils/file-upload'
  import { NekoWebRTC } from './internal/webrtc'
  import { Control } from './types/state'

  const inactiveCursorWin10 =
    'url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAACEUlEQVR4nOzWz6sSURQH8O+89zJ5C32LKbAgktCSaPpBSL' +
    'uSNtHqLcOV+BeIGxei0oCtFME/wI0bF4GCK6mNuAghH7xFlBAO7bQoA/Vik3riyghTaCQzTsLzbIZZDPdzzj3nzt3Df44dYDsBRNSYTqcn5XL5KoADy1VERL' +
    'Is02g0+phIJG4BsFkOEEVxjhgOh59kWb5rKWIBWCAGg0EnFovdtgyhB+grkU6n7wA4ZzlgCWKzlVgGsLQnVgE2gVh7xvP5PH9ciUajFQDHyWTyHQDVKOS3+F' +
    'sF/pyOcDh83Uhj/nMFBEFANpuF0+nUQ92SJD0G8AXAdwAz0wE+nw8OhwPNZhPj8RiBQOC0Vqu9EgSBcrnc11Qq9R7AeW5cd/GVsdgCr9dLiqJQtVqdv/v9fm' +
    'KM9UVRfArgJoBrAC4DsJsOcLlc1Gg0qNVqVRljI0mS5oh6vU6lUukFgEta5gemLr4AFAoF6nQ6b20223G73X6ZyWTmgFAoRL1ej3f+DQ1gfqiq+qbf73/weD' +
    'zPADwoFouPut3uzO12UyQSoclkotrt9ocAHKZnr8UhAP4bvg/gIs+UMfaaMTZTFOUkHo8/B/AEwAWjl5pV+j1dZ//g4xUMBo8YY/cqlcqhNvffAJxq40dmA5' +
    'bFPoAjrev5EfwZQNfoKbju/u1ri/PvfgKYGMl+K2I7b8U7wA5wpgC/AgAA///Yyif1MZXzRQAAAABJRU5ErkJggg==) 4 4, crosshair'

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
    private readonly control!: Control

    @Prop()
    private readonly screenSize!: { width: number; height: number }

    @Prop()
    private readonly canvasSize!: { width: number; height: number }

    @Prop()
    private readonly isControling!: boolean

    @Prop()
    private readonly implicitControl!: boolean

    get cursor(): string {
      if (!this.isControling) {
        return inactiveCursorWin10
      }

      if (!this.control.cursor.image) {
        return 'auto'
      }

      const { uri, x, y } = this.control.cursor.image
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

        this.webrtc.send('keydown', { key })
        return false
      }
      this.keyboard.onkeyup = (key: number) => {
        if (!this.focused) {
          return
        }

        if (!this.isControling) {
          return
        }

        this.webrtc.send('keyup', { key })
      }
      this.keyboard.listenTo(this._overlay)
    }

    beforeDestroy() {
      // Guacamole Keyboard does not provide destroy functions
    }

    getMousePos(clientX: number, clientY: number) {
      const rect = this._overlay.getBoundingClientRect()

      return {
        x: Math.round((this.screenSize.width / rect.width) * (clientX - rect.left)),
        y: Math.round((this.screenSize.height / rect.height) * (clientY - rect.top)),
      }
    }

    private mousepos: { x: number; y: number } = { x: 0, y: 0 }
    setMousePos(e: MouseEvent) {
      const pos = this.getMousePos(e.clientX, e.clientY)
      this.webrtc.send('mousemove', pos)
      Vue.set(this, 'mousepos', pos)
    }

    onWheel(e: WheelEvent) {
      if (!this.isControling) {
        return
      }

      let x = e.deltaX
      let y = e.deltaY

      if (this.control.scroll.inverse) {
        x *= -1
        y *= -1
      }

      x = Math.min(Math.max(x, -this.control.scroll.sensitivity), this.control.scroll.sensitivity)
      y = Math.min(Math.max(y, -this.control.scroll.sensitivity), this.control.scroll.sensitivity)

      this.setMousePos(e)
      this.webrtc.send('wheel', { x, y })
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

      if (!this.isControling) {
        // TODO: Refactor
        //syncKeyboardModifierState({
        //  capsLock: e.getModifierState('CapsLock'),
        //  numLock: e.getModifierState('NumLock'),
        //  scrollLock: e.getModifierState('ScrollLock'),
        //})
      }
    }

    onMouseLeave(e: MouseEvent) {
      this._overlay.blur()
      this.focused = false

      if (this.isControling) {
        this.keyboard.reset()

        // TODO: Refactor
        //setKeyboardModifierState({
        //  capsLock: e.getModifierState('CapsLock'),
        //  numLock: e.getModifierState('NumLock'),
        //  scrollLock: e.getModifierState('ScrollLock'),
        //})
      }
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

    @Watch('canvasSize')
    onCanvasSizeChange({ width, height }: { width: number; height: number }) {
      this._overlay.width = width
      this._overlay.height = height
    }

    private cursorElem: HTMLImageElement = new Image()
    @Watch('control.cursor.image')
    onCursorImageChange({ uri }: { uri: string }) {
      this.cursorElem.src = uri
    }

    @Watch('control.cursor.position')
    onCursorPositionChange({ x, y }: { x: number; y: number }) {
      if (this.isControling || this.control.cursor.image == null) return

      // get intrinsic dimensions
      const { width, height } = this._overlay

      // redraw cursor
      this._ctx.clearRect(0, 0, width, height)
      this._ctx.drawImage(
        this.cursorElem,
        Math.round((x / this.screenSize.width) * width - this.control.cursor.image.x),
        Math.round((y / this.screenSize.height) * height - this.control.cursor.image.y),
        this.control.cursor.image.width,
        this.control.cursor.image.height,
      )
    }

    private reqMouseDown: any | null = null
    private reqMouseUp: any | null = null
    @Watch('isControling')
    onControlChange(isControling: boolean) {
      if (isControling && this.reqMouseDown) {
        this.setMousePos(this.reqMouseDown)
        this.webrtc.send('mousedown', { key: this.reqMouseDown.button + 1 })
      }

      if (isControling && this.reqMouseUp) {
        this.webrtc.send('mouseup', { key: this.reqMouseUp.button + 1 })
      }

      if (isControling) {
        const { width, height } = this._overlay
        this._ctx.clearRect(0, 0, width, height)
      } else {
        this.onCursorPositionChange(this.mousepos)
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
