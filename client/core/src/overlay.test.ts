// The keyboard/mouse/touch overlay on its own: what attach() puts into the box and detach() takes
// out, DOM events reaching client.input, the host's cursor drawn from the transport's events,
// and the mount() helper around it.
import { test } from 'vitest'
import assert from 'node:assert/strict'
import { tick, settle, reset, init, patch, FakeElement, resizeObservers, windowListeners } from './test/browser.ts'

const { NekoClient, Overlay, mount } = await import('./index.ts')
type Options = ConstructorParameters<typeof NekoClient>[0]
type Transport = NonNullable<Exclude<NonNullable<Options>['transport'], () => unknown>>

const LAYER = 'position:absolute;inset:0;width:100%;height:100%;'

// a transport with a recording input channel and events a test can emit
const fakeTransport = () => {
  const calls: unknown[][] = []
  const listeners = new Map<string, Set<(...a: any[]) => void>>()
  const noop = () => {}
  const record =
    (m: string) =>
    (...a: unknown[]) =>
      calls.push([m, ...a])
  const t = {
    kind: 'ws-mse' as const,
    element: null,
    input: {
      move: record('move'),
      scroll: record('scroll'),
      button: record('button'),
      key: record('key'),
      touch: record('touch'),
    },
    connect: async () => {},
    close: noop,
    attach: (box: HTMLElement) => {
      const video = new FakeElement('video')
      ;(box as unknown as FakeElement).append(video)
      return () => video.remove()
    },
    setPlaying: async () => {},
    setVolume: noop,
    setMuted: noop,
    on: (ev: string, cb: (...a: any[]) => void) => {
      ;(listeners.get(ev) ?? listeners.set(ev, new Set()).get(ev)!).add(cb)
      return () => listeners.get(ev)!.delete(cb)
    },
    emit: (ev: string, ...a: unknown[]) => listeners.get(ev)?.forEach((cb) => cb(...a)),
    subscribers: (ev: string) => listeners.get(ev)?.size ?? 0,
  }
  return { t: t as unknown as Transport, calls, emit: t.emit, subscribers: t.subscribers }
}

// a client in a connected room (s1 = us, s2 = Bob the host) with an area of 1000x500 holding a
// box with the transport's element; the overlay is not attached yet
const setup = () => {
  reset()
  const { t, calls, emit, subscribers } = fakeTransport()
  const client = new NekoClient({ transport: t })
  const p = init()
  const s = client.state
  client.store.setState({
    session_id: p.session_id,
    sessions: p.sessions,
    settings: p.settings,
    screen: { ...s.screen, size: p.screen_size },
    control: { ...s.control, host_id: 's2', touch: true },
    connection: { ...s.connection, status: 'connected' },
  })
  const area = new FakeElement('div')
  area.offsetWidth = 1000
  area.offsetHeight = 500
  const box = new FakeElement('div')
  area.append(box)
  t.attach(box as any)
  const overlay = new Overlay(client)
  const wrap = () => box.children[1]
  const canvas = () => wrap().children[0]
  const textarea = () => wrap().children[1]
  return { client, calls, emit, subscribers, area, box, overlay, wrap, canvas, textarea }
}
const host = async (client: InstanceType<typeof NekoClient>) => {
  patch(client, 'control', { host_id: 's1' })
  await settle()
  assert.equal(client.controlling, true)
}
// the canvas as laid out: 960x540 at (10, 20), half the 1920x1080 remote screen
const laidOut = (canvas: FakeElement) => (canvas.rect = { left: 10, top: 20, width: 960, height: 540 })

test('attach: cursor canvas and textarea after the transport’s element, as the client used to make them; detach takes them out and unsubscribes', async () => {
  const { client, box, overlay, wrap, canvas, textarea, subscribers } = setup()
  assert.equal(box.children.length, 1, 'only the stream before attach')
  const detach = overlay.attach(box as any)
  assert.deepEqual(
    box.children.map((c) => c.tag),
    ['video', 'div'],
  )
  assert.deepEqual(
    wrap().children.map((c) => c.tag),
    ['canvas', 'textarea'],
  )
  assert.equal(canvas().style.cssText, LAYER)
  assert.equal(
    textarea().style.cssText,
    LAYER +
      'font-size:16px;resize:none;caret-color:transparent;outline:0;border:0;color:transparent;background:transparent;',
  )
  assert.equal(textarea().spellcheck, false)
  assert.deepEqual(textarea().attrs, { autocapitalize: 'off', autocomplete: 'off', 'data-gramm': 'false' })
  assert.deepEqual(
    [wrap().style.display, wrap().style.pointerEvents],
    ['', 'auto'],
    'shown, and can_host takes the mouse',
  )

  // the box is letterboxed inside the area (1000x500 for 16:9 -> 888.9x500, centred), the canvas sized with it
  const w = (500 * 1920) / 1080
  assert.deepEqual(box.style, { width: `${w}px`, height: '500px', marginTop: '0px', marginLeft: `${(1000 - w) / 2}px` })
  assert.deepEqual([(canvas() as any).width, (canvas() as any).height], [w, 500], 'the canvas backing store follows')
  assert.equal(resizeObservers.at(-1)?.observed, box.parent, 'the area is observed')
  patch(client, 'screen', { size: { width: 1000, height: 1000, rate: 30 } }) // a square screen: 500x500 in the middle
  await settle()
  assert.deepEqual(box.style, { width: '500px', height: '500px', marginTop: '0px', marginLeft: '250px' })
  box.parent!.offsetWidth = 400 // the area shrinks: the observer reports it
  resizeObservers.at(-1)!.cb()
  assert.deepEqual(box.style, { width: '400px', height: '400px', marginTop: '50px', marginLeft: '0px' })

  // private mode hides it from users, a local lock or no can_host takes the mouse away
  patch(client, 'settings', { private_mode: true })
  await settle()
  assert.equal(wrap().style.display, 'none')
  patch(client, 'settings', { private_mode: false })
  patch(client, 'control', { locked: true })
  await settle()
  assert.deepEqual([wrap().style.display, wrap().style.pointerEvents], ['', 'none'])
  patch(client, 'control', { locked: false })
  await settle()

  assert.equal(windowListeners.mouseup?.size, 1)
  assert.deepEqual([subscribers('cursor.position'), subscribers('cursor.image')], [1, 1])
  const w1 = wrap()
  detach()
  assert.deepEqual(
    box.children.map((c) => c.tag),
    ['video'],
  )
  assert.equal(windowListeners.mouseup?.size, 0)
  assert.deepEqual([subscribers('cursor.position'), subscribers('cursor.image')], [0, 0])
  assert.ok(resizeObservers.at(-1)!.disconnected)
  patch(client, 'control', { locked: true }) // the store subscriptions are gone too
  patch(client, 'screen', { size: { width: 1920, height: 1080, rate: 60 } })
  await settle()
  assert.equal(w1.style.pointerEvents, 'auto')
  assert.equal(box.style.width, '400px')
  detach() // twice: nothing left to do
})

test('attaching twice replaces the first attachment: one set of elements, listeners and subscriptions', () => {
  const { box, overlay, wrap, subscribers } = setup()
  overlay.attach(box as any)
  const first = resizeObservers.length
  overlay.attach(box as any)
  assert.deepEqual(
    box.children.map((c) => c.tag),
    ['video', 'div'],
  )
  assert.equal(wrap().children.length, 2)
  assert.equal(windowListeners.mouseup.size, 1)
  assert.deepEqual([subscribers('cursor.position'), subscribers('cursor.image')], [1, 1])
  assert.ok(resizeObservers[first - 1].disconnected, 'the first observer is gone')
  assert.equal(resizeObservers.at(-1)!.disconnected, false)
  const other = new FakeElement('div')
  box.parent!.append(other)
  const detach = overlay.attach(other as any) // and it can move to another box
  assert.deepEqual([box.children.length, other.children.length], [1, 1])
  detach()
  assert.deepEqual([box.children.length, other.children.length], [1, 0])
  assert.equal(windowListeners.mouseup.size, 0)
})

test('keyboard: the browser’s keydown/keypress/keyup on the textarea reach client.input.key as the X11 keysym, only while hosting and unlocked', async () => {
  const { client, calls, box, overlay, textarea } = setup()
  overlay.attach(box as any)
  // what the browser fires for the letter a
  const press = () => {
    textarea().dispatch('keydown', { keyCode: 65, key: 'a', location: 0 })
    textarea().dispatch('keypress', { keyCode: 97, charCode: 97, key: 'a', location: 0 })
  }
  const release = () => textarea().dispatch('keyup', { keyCode: 65, key: 'a', location: 0 })
  press()
  release()
  assert.deepEqual(calls, [], 'a viewer’s keys go nowhere')
  await host(client)
  press()
  assert.deepEqual(calls, [['key', 0x61, true]])
  release()
  assert.deepEqual(calls, [
    ['key', 0x61, true],
    ['key', 0x61, false],
  ])
  client.lock()
  await settle()
  press()
  release()
  assert.equal(calls.length, 2, 'locked: nothing is sent')
})

test('mouse: a mousedown reaches client.input.move and .button with the position in remote screen coordinates; the release may land anywhere', async () => {
  const { client, calls, box, overlay, canvas, textarea } = setup()
  overlay.attach(box as any)
  laidOut(canvas())
  const press = { clientX: 10 + 480, clientY: 20 + 270, button: 0 }
  textarea().dispatch('mousedown', press)
  assert.deepEqual(calls, [], 'not hosting: the press asks for control (implicit hosting) and waits')
  await host(client) // ...to be replayed once we are host
  assert.deepEqual(calls, [
    ['move', 960, 540],
    ['button', 1, true],
  ])
  for (const fn of windowListeners.mouseup) fn({ clientX: 10 + 960 + 50, clientY: 20 - 50, button: 0 }) // outside: clamped
  assert.deepEqual(calls.slice(2), [
    ['move', 1920, 0],
    ['button', 1, false],
  ])
  for (const fn of windowListeners.mouseup) fn(press) // a release without a press on the video: nothing
  assert.equal(calls.length, 4)
  textarea().dispatch('mousedown', { ...press, button: 2 }) // hosting: straight through
  assert.deepEqual(calls.slice(4), [
    ['move', 960, 540],
    ['button', 3, true],
  ])
})

test('touch: a touchstart reaches client.input.touch with the id, position and pressure', async () => {
  const { client, calls, box, overlay, canvas, textarea } = setup()
  overlay.attach(box as any)
  laidOut(canvas())
  await host(client)
  textarea().dispatch('touchstart', {
    changedTouches: [{ identifier: 7, clientX: 10 + 96, clientY: 20 + 54, force: 0.5 }],
  })
  assert.deepEqual(calls, [['touch', 'begin', 7, 192, 108, 128]])
  patch(client, 'control', { touch: false }) // a server without touch events: the first finger is the left button
  textarea().dispatch('touchend', { changedTouches: [{ identifier: 7, clientX: 10 + 96, clientY: 20 + 54, force: 0 }] })
  assert.deepEqual(calls.slice(1), [
    ['move', 192, 108],
    ['button', 1, false],
  ])
})

test('cursor: the host’s cursor from the transport is drawn on the canvas with their name, cleared on null, and is our own pointer while hosting', async () => {
  const { client, box, overlay, canvas, textarea, emit } = setup()
  overlay.attach(box as any)
  const ctx = canvas().ctx
  const drawn = () => ctx.calls.filter((c: unknown[]) => c[0] !== 'setTransform').map((c: unknown[]) => c[0])
  const image = { width: 16, height: 16, x: 1, y: 2, uri: 'blob:cursor' }
  emit('cursor.image', image)
  assert.equal(textarea().style.cursor, 'default', 'a viewer keeps their own pointer')
  emit('cursor.position', { x: 960, y: 540 })
  await tick(16) // the next frame
  assert.deepEqual(drawn(), ['clearRect', 'drawImage', 'strokeText', 'fillText'])
  const w = (500 * 1920) / 1080
  const [, img, x, y, dw, dh] = ctx.calls.find((c: unknown[]) => c[0] === 'drawImage')!
  assert.equal((img as { src: string }).src, 'blob:cursor')
  assert.deepEqual([x, y, dw, dh], [Math.round(w / 2) - 1, 250 - 2, 16, 16])
  assert.deepEqual(ctx.calls.find((c: unknown[]) => c[0] === 'fillText')!.slice(1), [
    'Bob',
    Math.round(w / 2) + 16,
    250 + 16,
  ])

  ctx.calls.length = 0
  emit('cursor.position', null) // the stream dropped its peer
  await tick(16)
  assert.deepEqual(drawn(), ['clearRect'])

  emit('cursor.position', { x: 1, y: 1 })
  await host(client)
  await tick(16)
  assert.equal(textarea().style.cursor, 'url(blob:cursor) 1 2, default')
  assert.equal(drawn().includes('drawImage'), false, 'the host sees it as their real pointer, not drawn')
})

test('mobileKeyboardToggle: focuses the textarea to open the on-screen keyboard, blurs it to close', () => {
  const { client, box, overlay, textarea } = setup()
  overlay.attach(box as any)
  overlay.mobileKeyboardToggle()
  assert.deepEqual([textarea().focused, client.state.mobile_keyboard_open], [true, true])
  overlay.mobileKeyboardToggle()
  assert.deepEqual([textarea().focused, client.state.mobile_keyboard_open], [false, false])
  textarea().focused = true
  client.store.setState({ mobile_keyboard_open: true })
  textarea().dispatch('blur') // the keyboard was dismissed by the user
  assert.equal(client.state.mobile_keyboard_open, false)
})

test('mount(client, el): the stream and the overlay in a letterboxed box inside el; the returned function takes it all out', () => {
  const { client, area, box } = setup()
  box.remove()
  const unmount = mount(client, area as any)
  assert.equal(area.children.length, 1)
  const mounted = area.children[0]
  assert.equal(mounted.style.position, 'relative')
  assert.deepEqual(
    mounted.children.map((c) => c.tag),
    ['video', 'div'],
  )
  assert.deepEqual(
    mounted.children[1].children.map((c) => c.tag),
    ['canvas', 'textarea'],
  )
  assert.equal(mounted.style.width, `${(500 * 1920) / 1080}px`)
  assert.equal(windowListeners.mouseup.size, 1)
  unmount()
  assert.deepEqual(area.children, [])
  assert.deepEqual(mounted.children, [], 'the transport’s element and the overlay are detached, not just the box')
  assert.equal(windowListeners.mouseup.size, 0)
})

test('store subscriptions: resize on the screen size, sync on visibility and interactivity, control and focus on hosting, a redraw on the host; once per change, not for the rest, and one update with several fields once', async () => {
  const { client, box, overlay } = setup()
  const calls = { resize: 0, sync: 0, control: [] as boolean[], focus: 0, draw: 0 }
  const o = overlay as any
  o.onResize = () => calls.resize++
  o.sync = () => calls.sync++
  o.onControl = (c: boolean) => calls.control.push(c)
  o.focusIfActive = () => calls.focus++
  o.draw = () => calls.draw++ // before attach: the host subscription takes the function itself
  overlay.attach(box as any)
  Object.assign(calls, { resize: 0, sync: 0, draw: 0 }) // attach syncs and sizes once by itself

  patch(client, 'control', { clipboard: { text: 'x' } }) // not watched
  client.store.setState({ mobile_keyboard_open: true, authenticated: true })
  patch(client, 'settings', { locked_logins: true })
  assert.deepEqual(calls, { resize: 0, sync: 0, control: [], focus: 0, draw: 0 })

  patch(client, 'screen', { size: { width: 800, height: 600, rate: 30 } })
  assert.deepEqual(calls, { resize: 1, sync: 0, control: [], focus: 0, draw: 0 })
  patch(client, 'settings', { private_mode: true }) // hidden from a user
  assert.deepEqual(calls, { resize: 1, sync: 1, control: [], focus: 0, draw: 0 })
  patch(client, 'control', { host_id: 's1' }) // we are host: control, focus, and the drawn cursor goes
  assert.deepEqual(calls, { resize: 1, sync: 1, control: [true], focus: 1, draw: 1 })
  client.lock() // interactive and active flip, controlling does not
  assert.deepEqual(calls, { resize: 1, sync: 2, control: [true], focus: 2, draw: 1 })
  client.unlock()
  patch(client, 'control', { host_id: 's2' })
  assert.deepEqual(calls, { resize: 1, sync: 3, control: [true, false], focus: 4, draw: 2 })

  // system/init changes the screen size, the sessions (we become admin), the settings (private
  // mode on) and the host in one update: every subscriber runs once, on the final state. Visible
  // stays true (an admin sees through private mode), so sync does not run at all; applied field
  // by field, settings before sessions, it would flip to false and back.
  patch(client, 'settings', { private_mode: false }) // visible again
  const before = JSON.parse(JSON.stringify(calls))
  client.connect()
  const p = init()
  p.sessions.s1.profile.is_admin = true
  p.settings.private_mode = true
  p.control_host = { has_host: true, host_id: 's1' }
  p.screen_size = { width: 640, height: 480, rate: 30 }
  ;(client as any).onMessage('system/init', p)
  assert.deepEqual(calls, {
    resize: before.resize + 1,
    sync: before.sync,
    control: [...before.control, true],
    focus: before.focus + 1,
    draw: before.draw + 1,
  })
  assert.deepEqual([o.visible, o.interactive, o.active], [true, true, true])
})
