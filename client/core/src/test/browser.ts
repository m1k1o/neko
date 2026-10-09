// The browser, faked just far enough for the core to load and run under vitest: globals, a few
// elements, a websocket, a peer connection and fetch. Imported first; the core is then imported
// dynamically, after the globals exist. The clock is vitest's (see tick/settle).
import { vi } from 'vitest'

const g = globalThis as any

// elements: enough of a <video>, <div>, <canvas> and <textarea> for the transport and the overlay
export class FakeElement {
  readonly tag: string
  children: FakeElement[] = []
  parent: FakeElement | null = null
  style: Record<string, string> = {}
  playsInline = false
  spellcheck = true
  value = ''
  attrs: Record<string, string> = {}
  paused = true
  playError: Error | null = null // what play() rejects with, once
  plays = 0 // how often play() was called
  videoWidth = 0
  videoHeight = 0
  private listeners = new Map<string, Set<(e: any) => void>>()
  private _muted = false
  private _volume = 1
  private _srcObject: unknown = null

  constructor(tag: string) {
    this.tag = tag
  }
  get muted() {
    return this._muted
  }
  set muted(v: boolean) {
    this._muted = v
    this.dispatch('volumechange')
  }
  get volume() {
    return this._volume
  }
  set volume(v: number) {
    this._volume = v
    this.dispatch('volumechange')
  }
  get srcObject() {
    return this._srcObject
  }
  set srcObject(v: unknown) {
    const had = this._srcObject
    this._srcObject = v
    if (had && !v) this.dispatch('emptied')
  }
  addEventListener(ev: string, fn: (e: any) => void) {
    if (!this.listeners.has(ev)) this.listeners.set(ev, new Set())
    this.listeners.get(ev)!.add(fn)
  }
  removeEventListener(ev: string, fn: (e: any) => void) {
    this.listeners.get(ev)?.delete(fn)
  }
  dispatch(ev: string, e: any = {}) {
    for (const fn of [...(this.listeners.get(ev) ?? [])]) fn(e)
  }
  append(...els: FakeElement[]) {
    for (const el of els) {
      el.parent = this
      this.children.push(el)
    }
  }
  remove() {
    if (this.parent) this.parent.children = this.parent.children.filter((c) => c !== this)
    this.parent = null
  }
  setAttribute(k: string, v: string) {
    this.attrs[k] = v
  }
  async play() {
    this.plays++
    if (this.playError) {
      const err = this.playError
      this.playError = null
      throw err
    }
    this.paused = false
    this.dispatch('playing')
  }
  pause() {
    this.paused = true
    this.dispatch('pause')
  }
  getContext() {
    return {}
  }
  getBoundingClientRect() {
    return { left: 0, top: 0, width: 0, height: 0 }
  }
  focus() {}
  blur() {}
}

g.window = g
g.location = { href: 'http://neko.test/' }
Object.defineProperty(g, 'navigator', {
  value: { platform: 'Linux', userAgent: 'test', maxTouchPoints: 0 },
  configurable: true,
})
export const storage = new Map<string, string>()
g.localStorage = {
  getItem: (k: string) => storage.get(k) ?? null,
  setItem: (k: string, v: string) => storage.set(k, v),
  removeItem: (k: string) => storage.delete(k),
}
g.matchMedia = () => ({ matches: false })
g.ResizeObserver = class {
  observe() {}
  disconnect() {}
}
// what the overlay needs to construct (client.mount)
g.Image = class {}
g.devicePixelRatio = 1
g.requestAnimationFrame = (fn: () => void) => setTimeout(fn, 16)
g.addEventListener = g.removeEventListener = () => {}
export const documentListeners: Record<string, ((e: any) => void)[]> = {}
g.document = {
  createElement: (tag: string) => new FakeElement(tag),
  addEventListener: (ev: string, fn: (e: any) => void) => (documentListeners[ev] ??= []).push(fn),
  removeEventListener() {},
  hasFocus: () => true,
}

// the clock: timers and Date are fake, setImmediate stays real so promise chains can be drained
vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout', 'setInterval', 'clearInterval', 'Date'] })
// one macrotask turn drains every pending promise chain (the handshake, a fetch)
export const settle = () => new Promise<void>((r) => setImmediate(r))
export const tick = async (ms: number) => {
  await settle() // whatever is in flight may still schedule a timer
  await vi.advanceTimersByTimeAsync(ms) // fires the due timers in order, settling after each
}

// the network
export const sent: string[] = []
export const sentPayloads: { event: string; payload: unknown }[] = []
export const sockets: FakeSocket[] = []
export class FakeSocket {
  static OPEN = 1
  readyState = 1
  url: string
  closed = false
  code?: number
  onmessage?: (e: { data: string }) => void
  onclose?: () => void
  constructor(url: string) {
    this.url = url
    sockets.push(this)
  }
  send(data: string) {
    const m = JSON.parse(data)
    sent.push(m.event)
    sentPayloads.push(m)
  }
  close(code?: number) {
    this.closed = true
    this.code = code
  }
}
g.WebSocket = FakeSocket

export const OFFER =
  'v=0\r\nm=audio 9 RTP/AVP 0\r\na=rtpmap:0 opus/48000\r\nm=video 9 RTP/AVP 96\r\na=rtpmap:96 VP8/90000\r\n'
export const ANSWER = 'v=0\r\nm=audio 9 RTP/AVP 0\r\nm=video 9 RTP/AVP 96\r\n'
export class FakePeer {
  static all: FakePeer[] = []
  static answerSdp = ANSWER
  config: unknown
  connectionState = 'new'
  signalingState = 'stable'
  remoteDescription: unknown = null
  localDescription: { sdp: string } | null = null
  candidates: unknown[] = []
  senders: unknown[] = []
  closed = false
  onconnectionstatechange?: () => void
  ontrack?: (e: unknown) => void
  ondatachannel?: (e: { channel: unknown }) => void
  onicecandidate?: (e: { candidate: { toJSON(): unknown } | null }) => void
  onnegotiationneeded?: () => Promise<void>
  constructor(config?: unknown) {
    this.config = config
    FakePeer.all.push(this)
  }
  async setRemoteDescription(d: unknown) {
    this.remoteDescription = d
  }
  async createAnswer() {
    return { type: 'answer', sdp: FakePeer.answerSdp }
  }
  async createOffer() {
    return { type: 'offer', sdp: 'local-offer' }
  }
  async setLocalDescription(d: { sdp: string }) {
    this.localDescription = d
  }
  async addIceCandidate(c: unknown) {
    this.candidates.push(c)
  }
  addTrack(track: unknown) {
    const sender = { track }
    this.senders.push(sender)
    return sender
  }
  removeTrack(sender: unknown) {
    if (!this.senders.includes(sender)) throw new Error('InvalidAccessError')
    this.senders = this.senders.filter((s) => s !== sender)
  }
  getSenders() {
    return this.senders
  }
  close() {
    this.closed = true
  }
  // what the browser's ICE agent would report
  transition(s: string) {
    this.connectionState = s
    this.onconnectionstatechange?.()
  }
}
g.RTCPeerConnection = FakePeer

export const fetches: { url: string; init: RequestInit }[] = []
// what fetch answers; tests replace net.respond
export const net = {
  respond: (_url: string, _init: RequestInit): { status: number; body?: unknown } => ({ status: 200 }),
}
g.fetch = async (url: string, init: RequestInit) => {
  fetches.push({ url, init })
  const { status, body } = net.respond(url, init)
  const text = body === undefined ? '' : JSON.stringify(body)
  return {
    ok: status < 300,
    status,
    statusText: 'status ' + status,
    json: async () => JSON.parse(text),
    text: async () => text,
  }
}

// a data channel as the peer would hand it over
export class FakeChannel {
  readyState = 'connecting'
  binaryType = ''
  frames: number[][] = []
  onmessage?: (e: { data: ArrayBuffer }) => void
  send(b: ArrayBuffer) {
    this.frames.push([...new Uint8Array(b)])
  }
}

// server payloads
export const profile = (over = {}) => ({
  name: 'Alice',
  is_admin: false,
  can_login: true,
  can_connect: true,
  can_watch: true,
  can_host: true,
  can_share_media: false,
  can_access_clipboard: true,
  sends_inactive_cursor: false,
  can_see_inactive_cursors: false,
  ...over,
})
export const session = (id: string, over = {}) => ({
  id,
  profile: profile(over),
  state: { is_connected: true, is_watching: false },
})
export const settings = (over = {}) => ({
  private_mode: false,
  locked_logins: false,
  locked_controls: false,
  control_protection: false,
  implicit_hosting: true,
  inactive_cursors: false,
  merciful_reconnect: true,
  ...over,
})
export const init = () => ({
  session_id: 's1',
  sessions: { s1: session('s1'), s2: session('s2', { name: 'Bob' }) },
  settings: settings(),
  screen_size: { width: 1920, height: 1080, rate: 60 },
  control_host: { has_host: true, host_id: 's2' },
  touch_events: true,
})
export const receive = (event: string, payload?: unknown, ws = sockets.at(-1)!) =>
  ws.onmessage!({ data: JSON.stringify({ event, payload }) })

// a clean slate for a test
export const reset = () => {
  sent.length = sentPayloads.length = sockets.length = fetches.length = FakePeer.all.length = 0
  for (const k of Object.keys(documentListeners)) delete documentListeners[k]
  vi.clearAllTimers()
  storage.clear()
  net.respond = () => ({ status: 200 })
  FakePeer.answerSdp = ANSWER
}
