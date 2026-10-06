// Framework-free neko client for the v3 API: REST (server/openapi.yaml), the
// /api/ws event protocol and WebRTC with a binary data channel for input.
// Written against master's server/pkg/types. UIs subscribe via `client.store`
// and read `client.state`; one-off happenings arrive on `client.events`.
import { Store, Emitter } from './store.ts'
import { NekoApi, ApiError, type Schemas } from './api.ts'
import { Overlay } from './overlay.ts'
import { OP, type State, type Settings, type NekoEvents } from './types.ts'

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

  video: HTMLVideoElement | null = null
  canvasSize = { width: 0, height: 0 }

  private el: HTMLElement | null = null
  private container: HTMLElement | null = null
  private overlay: Overlay | null = null
  private observer = new ResizeObserver(() => this.onResize())

  private ws: WebSocket | null = null
  private pc: RTCPeerConnection | null = null
  private dc: RTCDataChannel | null = null
  // the server trickles its candidates before signal/provide; hold them until the offer is applied
  private candidates: RTCIceCandidateInit[] = []
  private wanted = false // user asked to be connected
  private attempts = 0
  private reconnectTimer = 0
  private peerFailures = 0
  private peerTimer = 0
  private staleTimer = 0
  private lastMessage = 0

  private readonly opts: NekoClientOptions

  constructor(opts: NekoClientOptions = {}) {
    this.opts = opts
    const { store, state } = this
    store.watch(
      () => this.controlling,
      () => this.sendKeyboardMap(),
    )
    store.watch(
      () => state.control.keyboard,
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
    const res = await this.api.req<Schemas['SessionLoginResponse']>('POST', '/login', { username, password })
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
  // connection: websocket + webrtc
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
    clearTimeout(this.peerTimer)
    clearInterval(this.staleTimer)
    this.peerFailures = 0
    const ws = this.ws
    this.ws = null
    // a normal closure: without a status code the server assumes a reconnect is coming and
    // keeps the session connected (and host) for another 5 s
    ws?.close(1000)
    this.closePeer()
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
    this.closePeer()
    if (!this.wanted) return
    this.state.connection.status = 'connecting'
    if (++this.attempts > RECONNECT_MAX) return this.close(new Error('connection lost'))

    // a deleted session never comes back; anything else is worth retrying
    this.api
      .req('GET', '/whoami')
      .then(
        () => true,
        (err) => !(err instanceof ApiError && err.status === 401),
      )
      .then((retry) => {
        if (!this.wanted) return
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

  private requestPeer() {
    this.candidates = []
    this.send('signal/request', { video: {}, audio: {} })
  }

  private closePeer() {
    this.pc?.close()
    this.pc = null
    this.dc = null
    this.overlay?.clearCursor()
  }

  private async onProvide({ sdp, iceservers }: { sdp: string; iceservers?: RTCIceServer[] }) {
    this.closePeer()
    const pc = (this.pc = new RTCPeerConnection({ iceServers: iceservers ?? [] }))
    pc.onicecandidate = (e) => e.candidate && this.send('signal/candidate', e.candidate.toJSON())
    pc.ontrack = (e) => this.onTrack(e)
    pc.ondatachannel = (e) => this.bindDataChannel(e.channel)
    pc.onconnectionstatechange = () => this.pc === pc && this.onPeerState(pc.connectionState)
    // only fires after a local addTrack (microphone), the server offers everything else
    pc.onnegotiationneeded = async () => {
      if (pc.signalingState !== 'stable' || !pc.remoteDescription) return
      await pc.setLocalDescription(await pc.createOffer())
      this.send('signal/offer', { sdp: pc.localDescription!.sdp })
    }
    await this.onOffer(sdp)
  }

  private async onOffer(sdp: string) {
    const pc = this.pc
    if (!pc) return
    await pc.setRemoteDescription({ type: 'offer', sdp })
    for (const c of this.candidates.splice(0)) pc.addIceCandidate(c).catch(() => {})
    await pc.setLocalDescription(await pc.createAnswer())
    // no video codec in common (e.g. an H264 stream and a Firefox without the OpenH264 plugin):
    // the browser rejects the video section, the server cannot start the track, retrying won't help
    if (/^m=video 0 /m.test(pc.localDescription!.sdp)) {
      const codec = sdp.match(/^m=video[\s\S]*?a=rtpmap:\d+ ([\w-]+)/m)?.[1] ?? 'the stream'
      return this.close(new Error(`this browser cannot play ${codec} video`))
    }
    this.send('signal/answer', { sdp: pc.localDescription!.sdp })
  }

  private onPeerState(s: RTCPeerConnectionState) {
    if (s === 'connected') {
      this.attempts = 0
      this.peerFailures = 0
      this.state.connection.status = 'connected'
    } else if (s === 'disconnected') {
      this.state.connection.status = 'connecting' // ICE may still recover
    } else if (s === 'failed') {
      this.state.connection.status = 'connecting'
      this.closePeer()
      // no media route (firewall, NAT1TO1, missing TURN): retry with backoff, then give up
      // instead of asking the server for a new peer several times a second
      if (++this.peerFailures > RECONNECT_MAX) return this.close(new Error('video connection failed (WebRTC)'))
      clearTimeout(this.peerTimer)
      this.peerTimer = window.setTimeout(
        () => this.requestPeer(),
        RECONNECT_BACKOFF_MS * Math.min(this.peerFailures, 4),
      )
    }
  }

  private onTrack({ track, streams, receiver }: RTCTrackEvent) {
    // ask the browser for the smallest jitter buffer: interactive desktop, not a movie
    if ('jitterBufferTarget' in receiver) (receiver as any).jitterBufferTarget = 0
    if (track.kind !== 'video' || !this.video) return
    this.video.srcObject = streams[0]
    if (this.opts.autoplay || this.state.video.playing) this.play().catch(() => {})
  }

  private bindDataChannel(dc: RTCDataChannel) {
    dc.binaryType = 'arraybuffer'
    dc.onmessage = (e) => {
      const v = new DataView(e.data)
      const op = v.getUint8(0)
      if (op === OP.CURSOR_POSITION) {
        this.overlay?.onCursorPosition({ x: v.getUint16(3), y: v.getUint16(5) })
      } else if (op === OP.CURSOR_IMAGE) {
        const uri = URL.createObjectURL(new Blob([e.data.slice(11)], { type: 'image/png' }))
        this.overlay?.onCursorImage({
          width: v.getUint16(3),
          height: v.getUint16(5),
          x: v.getUint16(7),
          y: v.getUint16(9),
          uri,
        })
      }
    }
    this.dc = dc
  }

  // binary input; fields are [bytes, value] pairs (2 = u16, 4 = u32, -2 = i16, -4 = i32, 1 = u8)
  sendData(op: number, ...fields: [number, number][]) {
    if (this.dc?.readyState !== 'open') return
    const len = fields.reduce((n, [b]) => n + Math.abs(b), 0)
    const v = new DataView(new ArrayBuffer(3 + len))
    v.setUint8(0, op)
    v.setUint16(1, len)
    let o = 3
    for (const [b, val] of fields) {
      if (b === 1) v.setUint8(o, val)
      else if (b === 2) v.setUint16(o, val)
      else if (b === -2) v.setInt16(o, val)
      else if (b === 4) v.setUint32(o, val)
      else v.setInt32(o, val)
      o += Math.abs(b)
    }
    this.dc.send(v.buffer)
  }

  /////////////////////////////
  // websocket events -> state
  /////////////////////////////

  private onMessage(event: string, p: any) {
    const { state, events } = this
    switch (event) {
      case 'system/init':
        state.session_id = p.session_id
        state.control.touch = !!p.touch_events
        state.screen.size = p.screen_size
        state.sessions = p.sessions ?? {}
        state.settings = p.settings
        state.control.host_id = p.control_host?.has_host ? p.control_host.host_id : null
        this.requestPeer()
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

      case 'signal/provide':
        this.onProvide(p).catch((err) => console.error('[neko] webrtc setup failed', err))
        break
      case 'signal/offer':
      case 'signal/restart':
        this.onOffer(p.sdp).catch((err) => console.error('[neko] webrtc renegotiation failed', err))
        break
      case 'signal/answer':
        this.pc?.setRemoteDescription({ type: 'answer', sdp: p.sdp }).catch(() => {})
        break
      case 'signal/candidate':
        if (this.pc?.remoteDescription) this.pc.addIceCandidate(p).catch(() => {})
        else this.candidates.push(p)
        break
      case 'signal/close':
        // the server dropped its peer (it saw the media path fail first): ask for a new one,
        // with the same backoff as a failure the browser noticed itself
        if (this.pc) this.onPeerState('failed')
        break
      case 'signal/video':
      case 'signal/audio':
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
    if (this.video) this.video.srcObject = null
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

  // types text remotely via the server's clipboard (used for IME / mobile input)
  paste(text: string) {
    this.send('control/paste', { text })
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
  // media
  /////////////////////////////

  async play() {
    const video = this.video!
    try {
      await video.play()
    } catch (err) {
      if (video.muted) throw err
      // autoplay with sound is blocked until the user interacts: play muted,
      // unmute on the first click anywhere
      video.muted = true
      this.state.video.mutedByAutoplay = true
      await video.play()
      document.addEventListener('click', () => this.unmute(), { once: true })
    }
  }

  pause() {
    this.video?.pause()
  }

  mute() {
    if (this.video) this.video.muted = true
  }

  unmute() {
    if (this.video) this.video.muted = false
    this.state.video.mutedByAutoplay = false
  }

  setVolume(value: number) {
    if (this.video) this.video.volume = Math.max(0, Math.min(1, value))
  }

  // share local media (microphone); renegotiation happens via onnegotiationneeded
  addTrack(track: MediaStreamTrack, ...streams: MediaStream[]): RTCRtpSender {
    if (!this.pc) throw new Error('not connected')
    return this.pc.addTrack(track, ...streams)
  }

  removeTrack(sender: RTCRtpSender) {
    // a sender from a previous peer connection is already gone; removing it from the current one throws
    if (this.pc?.getSenders().includes(sender)) this.pc.removeTrack(sender)
  }

  /////////////////////////////
  // DOM
  /////////////////////////////

  mount(el: HTMLElement) {
    if (this.el) throw new Error('client already mounted')
    this.el = el
    const container = (this.container = document.createElement('div'))
    container.style.position = 'relative'

    const video = (this.video = document.createElement('video'))
    video.playsInline = true
    video.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;background:transparent'
    const v = this.state.video
    video.addEventListener('canplaythrough', () => (v.playable = true))
    video.addEventListener('playing', () => (v.playing = true))
    video.addEventListener('pause', () => (v.playing = false))
    video.addEventListener('emptied', () => (v.playable = v.playing = false))
    video.addEventListener('volumechange', () => ((v.muted = video.muted), (v.volume = video.volume)))
    v.muted = video.muted
    v.volume = video.volume

    container.append(video)
    el.append(container)
    this.overlay = new Overlay(this, container)
    this.syncOverlay()
    this.observer.observe(el)
    this.onResize()
  }

  unmount() {
    this.observer.disconnect()
    this.overlay?.destroy()
    this.container?.remove()
    this.el = this.container = this.video = this.overlay = null
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
