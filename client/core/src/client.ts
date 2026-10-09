// Framework-free neko client for the v3 API: REST (server/openapi.yaml), the /api/ws event
// protocol, and the stream through a StreamTransport (WebRTC by default, see transport.ts).
// Written against master's server/pkg/types. UIs subscribe via `client.store` and read
// `client.state`; one-off happenings arrive on `client.events`.
import { Store, Emitter } from './store.ts'
import { NekoApi, ApiError } from './api.ts'
import { Overlay } from './overlay.ts'
import { WebRTCTransport } from './transport/webrtc.ts'
import { WebSocketInput } from './input/websocket.ts'
import type { InputChannel, SessionInfo, StreamTransport, TransportState } from './transport.ts'
import type { State, Settings, NekoEvents, LoginResponse } from './types.ts'

const RECONNECT_MAX = 10
const RECONNECT_BACKOFF_MS = 1500
// a socket that stays silent this long is dead (the server heartbeats every ~10s)
const STALE_MS = 25_000

export interface NekoClientOptions {
  // remember the session token in localStorage
  autologin?: boolean
  // connect as soon as setUrl() finds a valid session
  autoconnect?: boolean
  autoplay?: boolean
  inputMode?: 'auto' | 'touch' | 'mouse'
  // how the stream arrives; WebRTC unless given
  transport?: StreamTransport | (() => StreamTransport)
}

export interface Pos {
  x: number
  y: number
}

const defaultSettings = (): Settings => ({
  private_mode: false,
  locked_logins: false,
  locked_controls: false,
  control_protection: false,
  implicit_hosting: false,
  inactive_cursors: false,
  merciful_reconnect: false,
})

const initialState = (): State => ({
  authenticated: false,
  connection: { url: location.href, status: 'disconnected' },
  video: { playable: false, playing: false, volume: 1, muted: false, mutedByAutoplay: false },
  control: {
    host_id: null,
    locked: false,
    clipboard: null,
    scroll: { inverse: true, sensitivity: 0 },
    keyboard: { layout: 'us', variant: '' },
    touch: false,
  },
  screen: { size: { width: 1280, height: 720, rate: 30 }, configurations: [] },
  session_id: null,
  sessions: {},
  settings: defaultSettings(),
  mobile_keyboard_open: false,
})

export class NekoClient {
  readonly store = new Store<State>(initialState())
  readonly state = this.store.state
  readonly api = new NekoApi()
  readonly events = new Emitter<NekoEvents>()
  readonly transport: StreamTransport
  // where the overlay sends input: the transport's own channel, or the websocket
  readonly input: InputChannel

  canvasSize = { width: 0, height: 0 }

  private el: HTMLElement | null = null
  private container: HTMLElement | null = null
  private overlay: Overlay | null = null
  private observer = new ResizeObserver(() => this.onResize())

  private ws: WebSocket | null = null
  // every websocket message, for the transport's signalling
  private readonly messages = new Emitter<Record<string, (payload: any) => void>>()
  private wanted = false // user asked to be connected
  private attempts = 0
  private reconnectTimer = 0
  private reconnectGen = 0 // a close()+connect() while a reconnect was deciding must win
  private staleTimer = 0
  private lastMessage = 0
  private clipboardAt = 0 // when the remote clipboard last changed
  private localSeen = { text: '', at: 0 } // the local clipboard text last handed over, and when

  private readonly opts: NekoClientOptions

  constructor(opts: NekoClientOptions = {}) {
    this.opts = opts
    const t = opts.transport
    this.transport = typeof t === 'function' ? t() : (t ?? new WebRTCTransport())
    this.input = this.transport.input ?? new WebSocketInput(this)
    this.transport.on('state', (s) => this.onTransportState(s))
    this.transport.on('error', (err) => this.close(err))
    this.transport.on('cursor.position', (p) => (p ? this.overlay?.onCursorPosition(p) : this.overlay?.clearCursor()))
    this.transport.on('cursor.image', (img) => this.overlay?.onCursorImage(img))

    const { store, state } = this
    store.watch(
      () => this.controlling,
      () => this.sendKeyboardMap(),
    )
    store.watch(
      () => state.control.keyboard.layout + '/' + state.control.keyboard.variant,
      () => this.sendKeyboardMap(),
    )
    store.watch(
      () => state.screen.size,
      () => this.onResize(),
    )
    store.watch(
      () => state.connection.status,
      (s) => this.events.emit('connection.status', s),
    )
    store.watch(
      () => this.overlayVisible,
      () => this.syncOverlay(),
    )
    store.watch(
      () => this.overlayInteractive,
      () => this.syncOverlay(),
    )
  }

  /////////////////////////////
  // computed
  /////////////////////////////

  get connected() {
    return this.state.connection.status === 'connected'
  }

  get controlling() {
    return this.state.control.host_id !== null && this.state.control.host_id === this.state.session_id
  }

  get session() {
    const id = this.state.session_id
    return id ? (this.state.sessions[id] ?? null) : null
  }

  get isAdmin() {
    return !!this.session?.profile.is_admin
  }

  get privateModeEnabled() {
    return this.state.settings.private_mode && !this.isAdmin
  }

  get implicitControl() {
    return this.state.settings.implicit_hosting && !!this.session?.profile.can_host
  }

  get isTouchDevice() {
    if (this.opts.inputMode) return this.opts.inputMode === 'touch'
    return (
      ('ontouchstart' in window || navigator.maxTouchPoints > 0) &&
      !matchMedia('(pointer:fine)').matches &&
      !matchMedia('(hover:hover)').matches
    )
  }

  /////////////////////////////
  // auth
  /////////////////////////////

  // point at a server (defaults to this page); `?token=` in the url logs in with that token
  setUrl(url = location.href) {
    const u = new URL(url)
    const token = u.searchParams.get('token') || undefined
    const http = (u.origin + u.pathname).replace(/^ws/, 'http').replace(/\/(api\/ws\/?)?$/, '')

    this.close()
    this.api.url = http
    this.state.connection.url = http
    this.state.authenticated = false
    this.state.connection.token = token

    if (this.opts.autoconnect) {
      this.authenticate()
        .then(() => this.connect())
        .catch(() => {})
    }
  }

  async authenticate(token = this.state.connection.token) {
    if (!token && this.opts.autologin) token = localStorage.getItem('neko_session') ?? undefined
    if (token) this.setToken(token)
    await this.api.req('GET', '/whoami')
    this.state.authenticated = true
  }

  async login(username: string, password: string) {
    const res = await this.api.req<LoginResponse>('POST', '/login', { username, password })
    // a session resumed meanwhile (autologin racing an invite link) must not keep its websocket
    this.close()
    // token is only returned when the server does not use cookies
    if (res.token) this.setToken(res.token)
    this.state.authenticated = true
  }

  async logout() {
    this.close()
    try {
      await this.api.req('POST', '/logout')
    } finally {
      this.setToken('')
      this.state.authenticated = false
    }
  }

  private setToken(token: string) {
    this.api.token = token
    this.state.connection.token = token || undefined
    if (!this.opts.autologin) return
    try {
      token ? localStorage.setItem('neko_session', token) : localStorage.removeItem('neko_session')
    } catch {}
  }

  /////////////////////////////
  // connection: websocket + transport
  /////////////////////////////

  connect() {
    if (!this.state.authenticated) throw new Error('client not authenticated')
    if (this.wanted) return
    this.wanted = true
    this.attempts = 0
    this.state.connection.status = 'connecting'
    this.openSocket()
  }

  disconnect() {
    this.close()
  }

  // stop everything; emits connection.closed if we were connected or connecting
  private close(error?: Error) {
    const was = this.wanted
    this.wanted = false
    clearTimeout(this.reconnectTimer)
    clearInterval(this.staleTimer)
    this.reconnectGen++
    const ws = this.ws
    this.ws = null
    // a normal closure: without a status code the server assumes a reconnect is coming and
    // keeps the session connected (and host) for another 5 s
    ws?.close(1000)
    this.transport.close()
    this.clear()
    if (was) this.events.emit('connection.closed', error)
  }

  private openSocket() {
    const { url, token } = this.state.connection
    const ws = (this.ws = new WebSocket(
      url.replace(/^http/, 'ws') + '/api/ws' + (token ? '?token=' + encodeURIComponent(token) : ''),
    ))
    this.lastMessage = Date.now()
    ws.onmessage = (e) => {
      this.lastMessage = Date.now()
      const { event, payload } = JSON.parse(e.data)
      this.onMessage(event, payload)
    }
    ws.onclose = () => this.ws === ws && this.onSocketLost()
    // a dropped network often never closes the socket; give up on it and reconnect
    clearInterval(this.staleTimer)
    this.staleTimer = window.setInterval(() => {
      if (this.ws === ws && Date.now() - this.lastMessage > STALE_MS) {
        this.ws = null
        ws.close()
        this.onSocketLost()
      }
    }, 5000)
  }

  private onSocketLost() {
    this.ws = null
    this.transport.suspend()
    if (!this.wanted) return
    this.state.connection.status = 'connecting'
    if (++this.attempts > RECONNECT_MAX) return this.close(new Error('connection lost'))

    // a deleted session never comes back; anything else is worth retrying. The lookup is bounded:
    // on a black-holed network a fetch can otherwise hang for minutes
    const gen = this.reconnectGen
    this.api
      .req('GET', '/whoami', undefined, { signal: AbortSignal.timeout(5000) })
      .then(
        () => true,
        (err) => !(err instanceof ApiError && err.status === 401),
      )
      .then((retry) => {
        if (!this.wanted || gen !== this.reconnectGen) return
        if (!retry) return this.close(new Error('session expired'))
        this.reconnectTimer = window.setTimeout(
          () => this.openSocket(),
          RECONNECT_BACKOFF_MS * Math.min(this.attempts, 4),
        )
      })
  }

  send(event: string, payload?: any) {
    if (this.ws?.readyState === WebSocket.OPEN) this.ws.send(JSON.stringify({ event, payload }))
  }

  // the stream: connected only once the transport is; its status while the socket is up
  private onTransportState(s: TransportState) {
    Object.assign(this.state.video, s.video)
    if (!this.wanted) return
    if (s.status === 'connected') this.attempts = 0
    this.state.connection.status = s.status === 'connected' || s.status === 'unavailable' ? 'connected' : 'connecting'
  }

  // what the transport gets of this session, once the server has introduced it (system/init)
  private sessionInfo(init: SessionInfo['init']): SessionInfo {
    const self = this
    return {
      url: this.state.connection.url,
      token: this.state.connection.token,
      init,
      send: (event, payload) => this.send(event, payload),
      on: (event, cb) => this.messages.on(event, cb),
      get lastMessage() {
        return self.lastMessage
      },
      get autoplay() {
        return !!self.opts.autoplay
      },
    }
  }

  /////////////////////////////
  // websocket events -> state
  /////////////////////////////

  private onMessage(event: string, p: any) {
    this.messages.emit(event, p)
    if (event.startsWith('signal/')) return // the transport's signalling
    const { state, events } = this
    switch (event) {
      case 'system/init':
        state.session_id = p.session_id
        state.control.touch = !!p.touch_events
        state.screen.size = p.screen_size
        state.sessions = p.sessions ?? {}
        state.settings = p.settings
        state.control.host_id = p.control_host?.has_host ? p.control_host.host_id : null
        this.transport.connect(this.sessionInfo(p)).catch((err) => this.close(err))
        break
      case 'system/admin':
        state.screen.configurations = [...(p.screen_sizes_list ?? [])].sort(
          (a, b) => b.width - a.width || b.height - a.height || b.rate - a.rate,
        )
        events.emit('room.broadcast.status', p.broadcast_status?.is_active, p.broadcast_status?.url)
        break
      case 'system/settings': {
        const { id, ...settings } = p
        state.settings = settings
        events.emit('room.settings.updated', settings, id)
        break
      }
      case 'system/disconnect':
        this.close(new Error(p.message))
        break
      case 'system/heartbeat':
        // lets proxies see client->server traffic; the server ignores the payload
        this.send('client/heartbeat')
        break

      case 'session/created':
        state.sessions[p.id] = p
        events.emit('session.created', p.id)
        break
      case 'session/deleted':
        delete state.sessions[p.id]
        events.emit('session.deleted', p.id)
        break
      case 'session/profile': {
        const { id, ...profile } = p
        if (!state.sessions[id]) return
        state.sessions[id].profile = profile
        events.emit('session.updated', id)
        break
      }
      case 'session/state': {
        const { id, ...st } = p
        if (!state.sessions[id]) return
        state.sessions[id].state = st
        events.emit('session.updated', id)
        break
      }
      case 'session/cursors':
        break // inactive cursors: not shown by this client

      case 'control/host':
        state.control.host_id = p.has_host ? p.host_id : null
        events.emit('room.control.host', p.has_host, p.host_id, p.id)
        break
      case 'control/request':
        events.emit('room.control.request', p.id)
        break

      case 'screen/updated': {
        const { id, ...size } = p
        state.screen.size = size
        events.emit('room.screen.updated', size.width, size.height, size.rate, id)
        break
      }
      case 'clipboard/updated':
        state.control.clipboard = { text: p.text }
        this.clipboardAt = Date.now()
        navigator.clipboard?.writeText(p.text).catch(() => {}) // only over https
        events.emit('room.clipboard.updated', p.text)
        break
      case 'broadcast/status':
        events.emit('room.broadcast.status', p.is_active, p.url)
        break
      case 'send/unicast':
        events.emit('receive.unicast', p.sender, p.subject, p.body)
        break
      case 'send/broadcast':
        events.emit('receive.broadcast', p.sender, p.subject, p.body)
        break

      default:
        events.emit('message', event, p)
    }
  }

  private clear() {
    const { state } = this
    state.connection.status = 'disconnected'
    state.control.host_id = null
    state.control.clipboard = null
    state.screen.configurations = []
    state.session_id = null
    state.sessions = {}
    state.settings = defaultSettings()
  }

  /////////////////////////////
  // control
  /////////////////////////////

  request() {
    this.send('control/request')
  }

  release() {
    this.send('control/release')
  }

  // local lock: keep control but stop sending input
  lock() {
    this.state.control.locked = true
  }

  unlock() {
    this.state.control.locked = false
  }

  // types text remotely via the server's clipboard
  paste(text: string) {
    this.send('control/paste', { text })
  }

  // Before a paste keystroke, with the text the browser's own paste event handed over: the side
  // that copied most recently wins. The remote reports every change of its clipboard; the local
  // clipboard can only be looked at when the browser hands it over, so if the remote changed since
  // the local text was last seen, the remote is taken as fresher and nothing is sent (on https the
  // remote text is also written to the local clipboard, so the two then match). Otherwise the
  // local text becomes the remote clipboard, and this resolves once the server has confirmed it.
  async preparePaste(text: string) {
    const stale = this.localSeen.at < this.clipboardAt
    const fresh = text !== '' && text !== this.state.control.clipboard?.text && !stale
    this.localSeen = { text, at: Date.now() }
    if (!fresh) return
    this.send('clipboard/set', { text })
    await new Promise<void>((done) => {
      const off = this.events.on('room.clipboard.updated', () => (off(), done()))
      setTimeout(() => (off(), done()), 500)
    })
  }

  setScrollInverse(value = true) {
    this.state.control.scroll.inverse = value
  }

  setScrollSensitivity(value: number) {
    this.state.control.scroll.sensitivity = value
  }

  setKeyboard(layout: string, variant = '') {
    this.state.control.keyboard = { layout, variant }
  }

  private sendKeyboardMap() {
    if (this.controlling && this.state.control.keyboard.layout) this.send('keyboard/map', this.state.control.keyboard)
  }

  setScreenSize(width: number, height: number, rate: number) {
    this.send('screen/set', { width, height, rate })
  }

  sendBroadcast(subject: string, body: any) {
    this.send('send/broadcast', { subject, body })
  }

  sendUnicast(receiver: string, subject: string, body: any) {
    this.send('send/unicast', { receiver, subject, body })
  }

  async uploadDrop({ x, y, files }: { x: number; y: number; files: File[] }) {
    const form = new FormData()
    form.append('x', String(x))
    form.append('y', String(y))
    for (const f of files) form.append('files', f)
    try {
      await this.api.upload('/room/upload/drop', form, (p) => this.events.emit('upload.drop.progress', p))
      this.events.emit('upload.drop.finished')
    } catch (err: any) {
      this.events.emit('upload.drop.finished', err)
    }
  }

  mobileKeyboardToggle() {
    this.overlay?.mobileKeyboardToggle()
  }

  /////////////////////////////
  // media (through the transport)
  /////////////////////////////

  // when the browser refuses sound, playback starts muted (video.mutedByAutoplay) and unmutes on
  // the first click
  play() {
    return this.transport.setPlaying(true)
  }

  pause() {
    this.transport.setPlaying(false)
  }

  mute() {
    this.transport.setMuted(true)
  }

  unmute() {
    this.transport.setMuted(false)
  }

  setVolume(value: number) {
    this.transport.setVolume(Math.max(0, Math.min(1, value)))
  }

  // share local media (microphone), on transports that can
  addTrack(track: MediaStreamTrack, ...streams: MediaStream[]): RTCRtpSender {
    if (!this.transport.addTrack) throw new Error('transport cannot send media')
    return this.transport.addTrack(track, ...streams)
  }

  removeTrack(sender: RTCRtpSender) {
    this.transport.removeTrack?.(sender)
  }

  /////////////////////////////
  // DOM
  /////////////////////////////

  // the stream and the input overlay fill `el`, letterboxed to the remote screen
  mount(el: HTMLElement) {
    if (this.el) throw new Error('client already mounted')
    this.el = el
    const container = (this.container = document.createElement('div'))
    container.style.position = 'relative'
    this.transport.attach(container)
    el.append(container)
    this.overlay = new Overlay(this, container)
    this.syncOverlay()
    this.observer.observe(el)
    this.onResize()
  }

  unmount() {
    this.observer.disconnect()
    this.overlay?.destroy()
    this.transport.detach()
    this.container?.remove()
    this.el = this.container = this.overlay = null
  }

  private get overlayVisible() {
    return !this.privateModeEnabled && this.state.connection.status !== 'disconnected'
  }

  private get overlayInteractive() {
    return !this.state.control.locked && !!this.session?.profile.can_host
  }

  private syncOverlay() {
    if (!this.overlay) return
    this.overlay.wrap.style.display = this.overlayVisible ? '' : 'none'
    this.overlay.wrap.style.pointerEvents = this.overlayInteractive ? 'auto' : 'none'
  }

  // letterbox the video area to the remote screen's aspect ratio
  private onResize() {
    if (!this.el || !this.container) return
    const { width, height } = this.state.screen.size
    const { offsetWidth: W, offsetHeight: H } = this.el
    const w = Math.min(W, (H * width) / height)
    const h = Math.min(H, (W * height) / width)
    Object.assign(this.container.style, {
      width: `${w}px`,
      height: `${h}px`,
      marginTop: `${(H - h) / 2}px`,
      marginLeft: `${(W - w) / 2}px`,
    })
    this.canvasSize = { width: w, height: h }
    this.overlay?.resize()
  }
}
