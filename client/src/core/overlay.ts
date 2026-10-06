// Input layer over the video: mouse, wheel, keyboard, touch, file drop, and
// drawing the host's cursor for everyone who is not controlling.
import GuacamoleKeyboard from './keyboard/guacamole.js'
import type { NekoClient, Pos } from './client.ts'
import { OP, type CursorImage } from './types.ts'

const WHEEL_STEP = 53 // px of wheel delta per scroll step
const WHEEL_LINE_HEIGHT = 19 // px per line when the browser reports lines
const MOVE_THROTTLE_MS = 1000 / 60
const LAYER = 'position:absolute;inset:0;width:100%;height:100%;'

// X11 keysyms
const XK = {
  v: 0x0076,
  Control_L: 0xffe3,
  Control_R: 0xffe4,
  Meta_L: 0xffe7,
  Alt_L: 0xffe9,
  Alt_R: 0xffea,
  Super_L: 0xffeb,
  Super_R: 0xffec,
  Mode_switch: 0xff7e,
  ISO_Level3_Shift: 0xfe03,
}

// macOS: Cmd acts as Ctrl and Option as AltGr on the remote (same as noVNC / the legacy client)
const isMac = /Mac|iPhone|iPod|iPad/i.test(navigator.platform)
const MAC_REMAP: Record<number, number> = {
  [XK.Meta_L]: XK.Control_L,
  [XK.Super_L]: XK.Alt_L,
  [XK.Super_R]: XK.Super_L,
  [XK.Alt_L]: XK.Mode_switch,
  [XK.Alt_R]: XK.ISO_Level3_Shift,
}
const remap = (key: number) => (isMac && MAC_REMAP[key]) || key
const isCtrl = (key: number) => key === XK.Control_L || key === XK.Control_R

const TOUCH_OP: Record<string, number> = {
  touchstart: OP.TOUCH_BEGIN,
  touchmove: OP.TOUCH_UPDATE,
  touchend: OP.TOUCH_END,
  touchcancel: OP.TOUCH_END,
}

export class Overlay {
  readonly wrap = document.createElement('div')
  private readonly canvas = document.createElement('canvas')
  private readonly ctx = this.canvas.getContext('2d')!
  private readonly input = document.createElement('textarea')
  private readonly keyboard = new GuacamoleKeyboard()
  private readonly cleanup: (() => void)[] = []

  private cursor: CursorImage | null = null
  private readonly cursorImg = new Image()
  private cursorPos: Pos | null = null
  private drawQueued = false

  private wheel = { x: 0, y: 0, t: 0 }
  private lastMove = 0
  private moveTimer = 0
  private mouseDown = false
  private pendingPress: MouseEvent | null = null // implicit hosting: replayed once we are host
  private ctrlDown = 0
  private readonly noKeyUp = new Set<number>()

  private readonly client: NekoClient

  constructor(client: NekoClient, parent: HTMLElement) {
    this.client = client
    this.canvas.style.cssText = LAYER
    // a transparent textarea receives keys, IME composition and the mobile keyboard
    this.input.style.cssText =
      LAYER +
      'font-size:16px;resize:none;caret-color:transparent;outline:0;border:0;color:transparent;background:transparent;'
    this.input.spellcheck = false
    this.input.setAttribute('autocapitalize', 'off')
    this.input.setAttribute('autocomplete', 'off')
    this.input.setAttribute('data-gramm', 'false')
    this.wrap.append(this.canvas, this.input)
    parent.append(this.wrap)

    const on = (el: EventTarget, ev: string, fn: (e: any) => void, opts?: AddEventListenerOptions) => {
      el.addEventListener(ev, fn, opts)
      this.cleanup.push(() => el.removeEventListener(ev, fn, opts))
    }
    const stop = (fn: (e: any) => void) => (e: Event) => {
      e.preventDefault()
      e.stopPropagation()
      fn(e)
    }
    const input = this.input
    on(
      input,
      'click',
      stop((e) => client.events.emit('overlay.click', e)),
    )
    on(
      input,
      'contextmenu',
      stop(() => {}),
    )
    on(
      input,
      'wheel',
      stop((e) => this.onWheel(e)),
      { passive: false },
    )
    on(
      input,
      'mousemove',
      stop((e) => this.onMove(e)),
    )
    on(
      input,
      'mousedown',
      stop((e) => this.onButton(e, true)),
    )
    // mouseup anywhere: the button may be released outside the video
    on(window, 'mouseup', (e) => this.onButton(e, false), { capture: true })
    on(input, 'mouseenter', (e) => this.onEnter(e))
    on(input, 'mouseleave', () => {
      this.hovering = false
      this.keyboard.reset()
    })
    for (const ev of Object.keys(TOUCH_OP)) on(input, ev, (e) => this.onTouch(e), { passive: false })
    on(
      input,
      'dragover',
      stop((e) => this.onMove(e)),
    )
    on(
      input,
      'drop',
      stop((e) => this.onDrop(e)),
    )
    // text that arrives as input or composition events (IME, on-screen keyboards) is typed as
    // key events by the Guacamole keyboard; the textarea only has to be emptied afterwards
    on(input, 'input', (e) => !e.isComposing && (this.input.value = ''))
    on(input, 'compositionend', () => (this.input.value = ''))
    on(
      input,
      'paste',
      stop((e: ClipboardEvent) => this.onPaste(e.clipboardData?.getData('text/plain') ?? '')),
    )
    on(input, 'blur', () => (client.state.mobile_keyboard_open = false))

    this.keyboard.onkeydown = (key) => this.onKeyDown(remap(key))
    this.keyboard.onkeyup = (key) => this.onKeyUp(remap(key))
    this.keyboard.listenTo(input)

    this.cursorImg.onload = this.draw
    this.cleanup.push(
      client.store.watch(
        () => client.controlling,
        (c) => this.onControl(c),
      ),
      client.store.watch(() => client.state.control.host_id, this.draw),
      client.store.watch(
        () => this.active,
        () => this.focusIfActive(),
      ),
    )
    this.resize()
  }

  destroy() {
    clearTimeout(this.moveTimer)
    this.cleanup.forEach((fn) => fn())
    this.keyboard.reset()
    if (this.cursor) URL.revokeObjectURL(this.cursor.uri)
    this.wrap.remove()
  }

  // input only flows while we are host and have not locked ourselves out
  private get active() {
    return this.client.controlling && !this.client.state.control.locked
  }

  private pos(e: { clientX: number; clientY: number }): Pos {
    const r = this.canvas.getBoundingClientRect()
    const { width, height } = this.client.state.screen.size
    const clamp = (v: number, max: number) => Math.max(0, Math.min(Math.round(v), max))
    return {
      x: clamp((width / r.width) * (e.clientX - r.left), width),
      y: clamp((height / r.height) * (e.clientY - r.top), height),
    }
  }

  private move(p: Pos) {
    this.client.sendData(OP.MOVE, [2, p.x], [2, p.y])
  }

  private button(code: number, down: boolean) {
    this.client.sendData(down ? OP.BTN_DOWN : OP.BTN_UP, [4, code])
  }

  /////////////////////////////
  // mouse
  /////////////////////////////

  // throttled to MOVE_THROTTLE_MS, but the last position always goes out (trailing send),
  // otherwise the remote pointer stops short of where the local one came to rest
  private onMove(e: MouseEvent) {
    if (!this.active) return
    const pos = this.pos(e)
    clearTimeout(this.moveTimer)
    const wait = MOVE_THROTTLE_MS - (performance.now() - this.lastMove)
    if (wait <= 0) return this.sendMove(pos)
    this.moveTimer = window.setTimeout(() => this.sendMove(pos), wait)
  }

  private sendMove(pos: Pos) {
    this.lastMove = performance.now()
    if (this.active) this.move(pos)
  }

  private onButton(e: MouseEvent, down: boolean) {
    if (down) this.mouseDown = true
    else if (!this.mouseDown)
      return // only releases of presses that started on the video
    else this.mouseDown = false

    if (!this.client.controlling) {
      if (down && this.client.implicitControl) {
        this.pendingPress = e
        this.client.request()
      }
      return
    }
    if (this.client.state.control.locked) return
    this.move(this.pos(e))
    this.button(e.button + 1, down)
  }

  private onWheel(e: WheelEvent) {
    if (!this.active) return

    // the first event of a gesture scrolls immediately, then one step per WHEEL_STEP px
    const first = e.timeStamp - this.wheel.t > 250
    if (first) this.wheel = { x: 0, y: 0, t: e.timeStamp }
    const k = e.deltaMode !== 0 ? WHEEL_LINE_HEIGHT : 1
    this.wheel.x += e.deltaX * k
    this.wheel.y += e.deltaY * k

    // sensitivity -5..5: negative widens the step, positive scrolls more per step
    const { sensitivity, inverse } = this.client.state.control.scroll
    const step = sensitivity < 0 ? WHEEL_STEP * (1 - sensitivity) : WHEEL_STEP
    const amount = (sensitivity > 0 ? sensitivity + 1 : 1) * (inverse ? -1 : 1)
    const axis = (d: 'x' | 'y') => {
      const v = this.wheel[d]
      if (!first && Math.abs(v) < step) return 0
      if (!first) this.wheel[d] = 0
      return Math.sign(v) * amount
    }
    const dx = axis('x')
    const dy = axis('y')
    if (dx || dy) this.client.sendData(OP.SCROLL, [-2, dx], [-2, dy], [1, e.ctrlKey ? 1 : 0])
  }

  private hovering = false

  // keys go to the hidden input only while it is useful: a viewer typing in the chat must not lose
  // the focus by brushing the video (touch devices: focusing would pop up the on-screen keyboard)
  private focusIfActive() {
    if (this.hovering && this.active && !this.client.isTouchDevice) this.input.focus()
  }

  private onEnter(e: MouseEvent) {
    this.hovering = true
    this.focusIfActive()
    if (this.client.controlling) {
      this.client.send('keyboard/modifiers', {
        capslock: e.getModifierState('CapsLock'),
        numlock: e.getModifierState('NumLock'),
      })
    }
  }

  private onDrop(e: DragEvent) {
    if (!this.client.controlling && !this.client.implicitControl) return
    // ponytail: dropped folders are skipped; walk webkitGetAsEntry() if anyone needs them
    const files = [...(e.dataTransfer?.files ?? [])]
    if (files.length) this.client.uploadDrop({ ...this.pos(e), files })
  }

  /////////////////////////////
  // keyboard
  /////////////////////////////

  private onKeyDown(key: number): boolean {
    if (!this.active) {
      this.noKeyUp.add(key)
      return true
    }
    // Ctrl+V: the keystroke stays with the browser, whose paste event hands over the local
    // clipboard text (see onPaste). Ctrl itself was sent and stays held, so the next key has it.
    if (this.ctrlDown && key === XK.v) {
      this.noKeyUp.add(key)
      return true
    }
    if (isCtrl(key)) this.ctrlDown = key
    this.client.sendData(OP.KEY_DOWN, [4, key])
    return isCtrl(key) // ctrl must reach the browser for the paste shortcut above
  }

  private onKeyUp(key: number) {
    if (this.noKeyUp.delete(key)) return
    if (isCtrl(key)) this.ctrlDown = 0
    this.client.sendData(OP.KEY_UP, [4, key])
  }

  // the remote gets its paste keystroke once the clipboard is settled. By then the user may have
  // let go of Ctrl, so the chord is completed with our own Ctrl when needed
  private async onPaste(text: string) {
    if (!this.active) return
    await this.client.preparePaste(text)
    if (!this.active) return
    const ctrl = this.ctrlDown
    if (!ctrl) this.client.sendData(OP.KEY_DOWN, [4, XK.Control_L])
    this.client.sendData(OP.KEY_DOWN, [4, XK.v])
    this.client.sendData(OP.KEY_UP, [4, XK.v])
    if (!ctrl) this.client.sendData(OP.KEY_UP, [4, XK.Control_L])
  }

  mobileKeyboardToggle() {
    // the blur handler clears the flag before the next line runs, so decide from the element
    const open = document.activeElement === this.input
    if (open) this.input.blur()
    else this.input.focus()
    this.client.state.mobile_keyboard_open = !open
  }

  /////////////////////////////
  // touch
  /////////////////////////////

  private onTouch(e: TouchEvent) {
    if (!this.client.controlling) {
      if (e.type === 'touchstart' && this.client.implicitControl) this.client.request()
      return // let the page scroll
    }
    e.preventDefault()
    e.stopPropagation()
    if (this.client.state.control.locked) return

    if (this.client.state.control.touch) {
      for (const t of e.changedTouches) {
        const p = this.pos(t)
        this.client.sendData(TOUCH_OP[e.type], [4, t.identifier], [-4, p.x], [-4, p.y], [1, Math.round(t.force * 255)])
      }
      return
    }
    // server without touch support: the first finger drives the left mouse button
    const t = e.changedTouches[0]
    this.move(this.pos(t))
    if (e.type === 'touchstart') this.button(1, true)
    else if (e.type !== 'touchmove') this.button(1, false)
  }

  /////////////////////////////
  // host changes + remote cursor
  /////////////////////////////

  private onControl(controlling: boolean) {
    this.ctrlDown = 0
    this.noKeyUp.clear()
    this.updateCursorStyle()
    this.draw()

    const e = this.pendingPress
    this.pendingPress = null
    if (controlling && e) {
      this.move(this.pos(e))
      this.button(e.button + 1, true)
      if (!this.mouseDown) this.button(e.button + 1, false) // released while we waited
    }
  }

  onCursorImage(img: CursorImage) {
    const old = this.cursor
    this.cursor = img
    this.cursorImg.src = img.uri
    this.updateCursorStyle()
    if (old) URL.revokeObjectURL(old.uri)
  }

  onCursorPosition(p: Pos) {
    this.cursorPos = p
    this.draw()
  }

  clearCursor() {
    this.cursorPos = null
    this.draw()
  }

  // the host sees the remote cursor as their real mouse pointer
  private updateCursorStyle() {
    const c = this.cursor
    this.input.style.cursor = this.client.controlling && c ? `url(${c.uri}) ${c.x} ${c.y}, default` : 'default'
  }

  // ponytail: devicePixelRatio is read on resize only; listen to matchMedia(resolution) if
  // cursors look blurry after dragging the window to another monitor
  resize() {
    const { width, height } = this.client.canvasSize
    this.canvas.width = width * devicePixelRatio
    this.canvas.height = height * devicePixelRatio
    this.draw()
  }

  private draw = () => {
    if (this.drawQueued) return
    this.drawQueued = true
    requestAnimationFrame(() => {
      this.drawQueued = false
      this.paint()
    })
  }

  // everyone but the host sees the host's cursor drawn here, tagged with their name
  private paint() {
    const { ctx, cursor, cursorPos } = this
    const { width, height } = this.client.canvasSize
    ctx.setTransform(devicePixelRatio, 0, 0, devicePixelRatio, 0, 0)
    ctx.clearRect(0, 0, width, height)
    if (this.client.controlling || !cursor || !cursorPos || (cursor.width <= 1 && cursor.height <= 1)) return

    const screen = this.client.state.screen.size
    let x = Math.round((cursorPos.x / screen.width) * width)
    let y = Math.round((cursorPos.y / screen.height) * height)
    ctx.drawImage(this.cursorImg, x - cursor.x, y - cursor.y, cursor.width, cursor.height)

    const hostId = this.client.state.control.host_id
    const name = hostId ? this.client.state.sessions[hostId]?.profile.name : ''
    if (!name) return
    x += cursor.width
    y += cursor.height
    ctx.font = '14px Arial, sans-serif'
    ctx.textBaseline = 'top'
    ctx.shadowColor = 'black'
    ctx.shadowBlur = 2
    ctx.lineWidth = 2
    ctx.fillStyle = 'black'
    ctx.strokeText(name, x, y)
    ctx.shadowBlur = 0
    ctx.fillStyle = 'white'
    ctx.fillText(name, x, y)
  }
}
