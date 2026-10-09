// WebRTC: the server's offer arrives over the main socket (signal/*), media on an
// RTCPeerConnection, input and the host's cursor on its data channel.
import { Emitter } from '../store.ts'
import { DataChannelInput } from '../input/datachannel.ts'
import type { SessionInfo, StreamTransport, TransportEvents, TransportState, TransportStatus } from '../transport.ts'

const RECONNECT_MAX = 10
const RECONNECT_BACKOFF_MS = 1500
// no offer within this long after asking for one: the server will not send media (e.g. can_watch
// is off), so the room is shown without video instead of a spinner for ever
const OFFER_TIMEOUT_MS = 8_000

export class WebRTCTransport implements StreamTransport {
  readonly kind = 'webrtc' as const
  readonly input = new DataChannelInput()

  static supported() {
    return typeof RTCPeerConnection !== 'undefined'
  }

  private readonly events = new Emitter<TransportEvents>()
  private readonly state: TransportState = {
    status: 'disconnected',
    size: { width: 0, height: 0 },
    video: { playable: false, playing: false, volume: 1, muted: false, mutedByAutoplay: false },
  }
  private video: HTMLVideoElement | null = null
  private session: SessionInfo | null = null
  private unsubs: (() => void)[] = []
  private pc: RTCPeerConnection | null = null
  // the server trickles its candidates before signal/provide; hold them until the offer is applied
  private candidates: RTCIceCandidateInit[] = []
  private failures = 0
  private retryTimer = 0
  private offerTimer = 0

  constructor() {
    this.input.onCursorPosition = (p) => this.events.emit('cursor.position', p)
    this.input.onCursorImage = (img) => this.events.emit('cursor.image', img)
  }

  get element() {
    return this.video
  }

  on<K extends keyof TransportEvents>(event: K, cb: TransportEvents[K]) {
    return this.events.on(event, cb)
  }

  /////////////////////////////
  // signalling + peer
  /////////////////////////////

  async connect(session: SessionInfo) {
    this.unsubscribe()
    this.session = session
    const on = (event: string, cb: (p: any) => void) => this.unsubs.push(session.on(event, cb))
    on('signal/provide', (p) => {
      clearTimeout(this.offerTimer)
      this.onProvide(p).catch((err) => console.error('[neko] webrtc setup failed', err))
    })
    const renegotiate = (p: { sdp: string }) =>
      this.onOffer(p.sdp).catch((err) => console.error('[neko] webrtc renegotiation failed', err))
    on('signal/offer', renegotiate)
    on('signal/restart', renegotiate)
    on('signal/answer', (p) => this.pc?.setRemoteDescription({ type: 'answer', sdp: p.sdp }).catch(() => {}))
    on('signal/candidate', (p) => {
      if (this.pc?.remoteDescription) this.pc.addIceCandidate(p).catch(() => {})
      else this.candidates.push(p)
    })
    // the server dropped its peer (it saw the media path fail first): ask for a new one,
    // with the same backoff as a failure the browser noticed itself
    on('signal/close', () => this.pc && this.onPeerState('failed'))
    this.setStatus('connecting')
    this.requestPeer()
  }

  suspend() {
    this.unsubscribe()
    this.session = null
    this.closePeer()
    this.setStatus('disconnected')
  }

  close() {
    this.suspend()
    this.failures = 0
    if (this.video) this.video.srcObject = null
  }

  private unsubscribe() {
    this.unsubs.splice(0).forEach((off) => off())
  }

  private requestPeer() {
    this.candidates = []
    this.session?.send('signal/request', { video: {}, audio: {} })
    clearTimeout(this.offerTimer)
    const noOffer = () => {
      if (!this.session || this.pc) return
      // only while the server is talking (heartbeats): after a lost network the media fails
      // first and the request goes into a dead socket, which is the stale check's business,
      // not a reason to show a frozen picture as "connected"
      if (Date.now() - this.session.lastMessage < OFFER_TIMEOUT_MS * 2) this.setStatus('unavailable')
      else this.offerTimer = window.setTimeout(noOffer, OFFER_TIMEOUT_MS)
    }
    this.offerTimer = window.setTimeout(noOffer, OFFER_TIMEOUT_MS)
  }

  private closePeer() {
    clearTimeout(this.retryTimer)
    clearTimeout(this.offerTimer)
    this.pc?.close()
    this.pc = null
    this.input.bind(null)
    this.events.emit('cursor.position', null)
  }

  private async onProvide({ sdp, iceservers }: { sdp: string; iceservers?: RTCIceServer[] }) {
    this.closePeer()
    const pc = (this.pc = new RTCPeerConnection({ iceServers: iceservers ?? [] }))
    pc.onicecandidate = (e) => e.candidate && this.session?.send('signal/candidate', e.candidate.toJSON())
    pc.ontrack = (e) => this.onTrack(e)
    pc.ondatachannel = (e) => this.input.bind(e.channel)
    pc.onconnectionstatechange = () => this.pc === pc && this.onPeerState(pc.connectionState)
    // only fires after a local addTrack (microphone), the server offers everything else
    pc.onnegotiationneeded = async () => {
      if (pc.signalingState !== 'stable' || !pc.remoteDescription) return
      await pc.setLocalDescription(await pc.createOffer())
      this.session?.send('signal/offer', { sdp: pc.localDescription!.sdp })
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
      return this.events.emit('error', new Error(`this browser cannot play ${codec} video`))
    }
    this.session?.send('signal/answer', { sdp: pc.localDescription!.sdp })
  }

  private onPeerState(s: RTCPeerConnectionState) {
    if (s === 'connected') {
      this.failures = 0
      this.setStatus('connected')
    } else if (s === 'disconnected') {
      this.setStatus('connecting') // ICE may still recover
    } else if (s === 'failed') {
      this.setStatus('connecting')
      this.closePeer()
      // no media route (firewall, NAT1TO1, missing TURN): retry with backoff, then give up
      // instead of asking the server for a new peer several times a second
      if (++this.failures > RECONNECT_MAX)
        return this.events.emit('error', new Error('video connection failed (WebRTC)'))
      clearTimeout(this.retryTimer)
      this.retryTimer = window.setTimeout(() => this.requestPeer(), RECONNECT_BACKOFF_MS * Math.min(this.failures, 4))
    }
  }

  private onTrack({ track, streams, receiver }: RTCTrackEvent) {
    // ask the browser for the smallest jitter buffer: interactive desktop, not a movie
    if ('jitterBufferTarget' in receiver) (receiver as any).jitterBufferTarget = 0
    if (track.kind !== 'video' || !this.video) return
    this.video.srcObject = streams[0]
    if (this.session?.autoplay || this.state.video.playing) this.play().catch(() => {})
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
  // the element
  /////////////////////////////

  attach(container: HTMLElement) {
    this.detach()
    const video = (this.video = document.createElement('video'))
    video.playsInline = true
    video.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;background:transparent'
    const v = this.state.video
    const sync = () => this.emitState()
    video.addEventListener('canplaythrough', () => ((v.playable = true), sync()))
    video.addEventListener('playing', () => ((v.playing = true), sync()))
    video.addEventListener('pause', () => ((v.playing = false), sync()))
    video.addEventListener('emptied', () => ((v.playable = v.playing = false), sync()))
    video.addEventListener('volumechange', () => ((v.muted = video.muted), (v.volume = video.volume), sync()))
    video.addEventListener('resize', () => {
      this.state.size = { width: video.videoWidth, height: video.videoHeight }
      sync()
    })
    v.muted = video.muted
    v.volume = video.volume
    container.append(video)
    this.emitState()
  }

  detach() {
    this.video?.remove()
    this.video = null
  }

  setPlaying(on: boolean) {
    if (on) return this.play()
    this.video?.pause()
    return Promise.resolve()
  }

  setVolume(volume: number) {
    if (this.video) this.video.volume = volume
  }

  setMuted(on: boolean) {
    if (this.video) this.video.muted = on
    if (!on) {
      this.state.video.mutedByAutoplay = false
      this.emitState()
    }
  }

  private async play() {
    const video = this.video
    if (!video) return
    try {
      await video.play()
    } catch (err) {
      if (video.muted || (err as DOMException).name !== 'NotAllowedError') throw err
      // autoplay with sound is blocked until the user interacts: play muted,
      // unmute on the first click anywhere
      video.muted = true
      this.state.video.mutedByAutoplay = true
      this.emitState()
      await video.play()
      document.addEventListener('click', () => this.setMuted(false), { once: true })
    }
  }

  private setStatus(status: TransportStatus) {
    this.state.status = status
    this.emitState()
  }

  private emitState() {
    const { status, size, video } = this.state
    this.events.emit('state', { status, size: { ...size }, video: { ...video } })
  }
}
