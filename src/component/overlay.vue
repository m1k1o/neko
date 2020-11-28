<template>
  <div
    ref="overlay"
    class="overlay"
    tabindex="0"
    @click.stop.prevent
    @contextmenu.stop.prevent
    @wheel.stop.prevent="onWheel"
    @mousemove.stop.prevent="onMouseMove"
    @mousedown.stop.prevent="onMouseDown"
    @mouseup.stop.prevent="onMouseUp"
    @mouseenter.stop.prevent="onMouseEnter"
    @mouseleave.stop.prevent="onMouseLeave"
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
  import { Vue, Component, Ref, Prop } from 'vue-property-decorator'

  import GuacamoleKeyboard from './utils/guacamole-keyboard'
  import { NekoWebRTC } from './internal/webrtc'

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
    private readonly screenWidth!: number

    @Prop()
    private readonly screenHeight!: number

    @Prop()
    private readonly scrollSensitivity!: number

    @Prop()
    private readonly scrollInvert!: boolean

    @Prop()
    private readonly isControling!: boolean

    mounted() {
      // Initialize Guacamole Keyboard
      this.keyboard.onkeydown = (key: number) => {
        if (!this.focused || !this.isControling) {
          return true
        }

        this.webrtc.send('keydown', { key })
        return false
      }
      this.keyboard.onkeyup = (key: number) => {
        if (!this.focused || !this.isControling) {
          return
        }

        this.webrtc.send('keyup', { key })
      }
      this.keyboard.listenTo(this._overlay)
    }

    beforeDestroy() {
      // Guacamole Keyboard does not provide destroy functions
    }

    setMousePos(e: MouseEvent) {
      const rect = this._overlay.getBoundingClientRect()

      this.webrtc.send('mousemove', {
        x: Math.round((this.screenWidth / rect.width) * (e.clientX - rect.left)),
        y: Math.round((this.screenHeight / rect.height) * (e.clientY - rect.top)),
      })
    }

    onWheel(e: WheelEvent) {
      if (!this.isControling) {
        return
      }

      let x = e.deltaX
      let y = e.deltaY

      if (this.scrollInvert) {
        x *= -1
        y *= -1
      }

      x = Math.min(Math.max(x, -this.scrollSensitivity), this.scrollSensitivity)
      y = Math.min(Math.max(y, -this.scrollSensitivity), this.scrollSensitivity)

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
        return
      }

      this.setMousePos(e)
      this.webrtc.send('mousedown', { key: e.button + 1 })
    }

    onMouseUp(e: MouseEvent) {
      if (!this.isControling) {
        return
      }

      this.setMousePos(e)
      this.webrtc.send('mouseup', { key: e.button + 1 })
    }

    onMouseEnter(e: MouseEvent) {
      if (!this.isControling) {
        return
      }

      // TODO: Refactor
      //syncKeyboardModifierState({
      //  capsLock: e.getModifierState('CapsLock'),
      //  numLock: e.getModifierState('NumLock'),
      //  scrollLock: e.getModifierState('ScrollLock'),
      //})

      this._overlay.focus()
      this.focused = true
    }

    onMouseLeave(e: MouseEvent) {
      if (!this.isControling) {
        return
      }

      // TODO: Refactor
      //setKeyboardModifierState({
      //  capsLock: e.getModifierState('CapsLock'),
      //  numLock: e.getModifierState('NumLock'),
      //  scrollLock: e.getModifierState('ScrollLock'),
      //})

      this.keyboard.reset()
      this._overlay.blur()
      this.focused = false
    }
  }
</script>
