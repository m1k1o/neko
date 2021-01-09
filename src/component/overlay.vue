<template>
  <div
    ref="overlay"
    class="overlay"
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
  .overlay {
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
  import { getFilesFromDataTansfer } from './utils/file-upload'
  import { NekoWebRTC } from './internal/webrtc'
  import { Control } from './types/state'

  @Component({
    name: 'neko-overlay',
  })
  export default class extends Vue {
    @Ref('overlay') readonly _overlay!: HTMLElement

    private keyboard = GuacamoleKeyboard()
    private focused = false

    @Prop()
    private readonly webrtc!: NekoWebRTC

    @Prop()
    private readonly control!: Control

    @Prop()
    private readonly screenWidth!: number

    @Prop()
    private readonly screenHeight!: number

    @Prop()
    private readonly isControling!: boolean

    @Prop()
    private readonly implicitControl!: boolean

    get cursor(): string {
      if (!this.control.cursor) return 'auto'

      const { uri, hot_x, hot_y } = this.control.cursor
      return 'url(' + uri + ') ' + hot_x + ' ' + hot_y + ', auto'
    }

    mounted() {
      // Initialize Guacamole Keyboard
      this.keyboard.onkeydown = (key: number) => {
        if (!this.focused) {
          return true
        }

        if (!this.isControling) {
          this.implicitControlRequest()
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
          this.implicitControlRequest()
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
        x: Math.round((this.screenWidth / rect.width) * (clientX - rect.left)),
        y: Math.round((this.screenHeight / rect.height) * (clientY - rect.top)),
      }
    }

    setMousePos(e: MouseEvent) {
      this.webrtc.send('mousemove', this.getMousePos(e.clientX, e.clientY))
    }

    onWheel(e: WheelEvent) {
      if (!this.isControling) {
        this.implicitControlRequest()
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
        this.implicitControlRequest()
        return
      }

      this.setMousePos(e)
    }

    onMouseDown(e: MouseEvent) {
      if (!this.isControling) {
        this.implicitControlRequest()
        return
      }

      this.setMousePos(e)
      this.webrtc.send('mousedown', { key: e.button + 1 })
    }

    onMouseUp(e: MouseEvent) {
      if (!this.isControling) {
        this.implicitControlRequest()
        return
      }

      this.setMousePos(e)
      this.webrtc.send('mouseup', { key: e.button + 1 })
    }

    onMouseEnter(e: MouseEvent) {
      this._overlay.focus()
      this.focused = true

      if (!this.isControling) {
        this.implicitControlRequest()
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
        this.implicitControlRelease()

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

    isRequesting = false
    @Watch('isControling')
    onControlChange(isControling: boolean) {
      this.isRequesting = false
    }

    implicitControlRequest() {
      if (!this.isRequesting && this.implicitControl) {
        this.isRequesting = true
        this.$emit('implicit-control-request')
      }
    }

    implicitControlRelease() {
      if (this.implicitControl) {
        this.$emit('implicit-control-release')
      }
    }
  }
</script>
