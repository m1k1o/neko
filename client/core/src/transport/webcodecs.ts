// WebCodecs over a websocket: the server sends encoded samples on its own binary socket, the
// browser decodes them with WebCodecs and paints video to a canvas. Input goes over the main
// socket (the transport has no channel of its own), as does nothing else: the media socket
// carries only samples and this client's acks.
//
// Flow control is the point of this transport. The server keeps a small window of unacked frames
// and, once it is full, replaces the pending frame with the newest instead of queuing: a client
// that cannot keep up sees fewer frames, not later ones. This client holds up its end by acking
// each frame after it is painted, and by asking for a keyframe whenever it drops one, since a
// decoder cannot use delta frames across a gap.
import { Emitter } from '../emitter.ts'
import type { SessionInfo, StreamTransport, TransportEvents, TransportState, TransportStatus } from '../transport.ts'
import { encodeAck, encodeKeyframeRequest, parseMediaSample, type EncodedMediaSample } from './media-protocol.ts'

// how far the decoder may fall behind before frames are dropped. Small: a backlog here is latency
// the user sees, and the server is already coalescing on its side.
const MAX_VIDEO_DECODE_QUEUE = 4
// audio tolerates more: samples are short, independently decodable, and gaps are audible
const MAX_AUDIO_DECODE_QUEUE = 9
// audio is scheduled this far ahead of the context clock, and resynced if it drifts past the max
const AUDIO_LEAD_S = 0.03
const AUDIO_MAX_DRIFT_S = 0.15

const RECONNECT_MAX = 10
const RECONNECT_BACKOFF_MS = 1500

// what POST /api/media/ticket answers (server/internal/media/manager.go)
interface MediaTicket {
  ticket: string
  protocol: string
  video: { codec: string }
  audio: { codec: string; sample_rate: number; number_of_channels: number }
}

const PROTOCOL = 'webcodecs-ws-v1'

export class WebCodecsTransport implements StreamTransport {
  readonly kind = 'webcodecs-ws' as const
  // no input channel: the client falls back to sending input over the main websocket

  static supported() {
    return typeof VideoDecoder !== 'undefined' && typeof AudioDecoder !== 'undefined'
  }

  private readonly events = new Emitter<TransportEvents>()
  private readonly state: TransportState = {
    status: 'disconnected',
    size: { width: 0, height: 0 },
    video: { playable: false, playing: false, volume: 1, muted: false, mutedByAutoplay: false },
  }
  private canvas: HTMLCanvasElement | null = null
  private context: CanvasRenderingContext2D | null = null
  private session: SessionInfo | null = null
  private socket: WebSocket | null = null
  private videoDecoder: VideoDecoder | null = null
  private audioDecoder: AudioDecoder | null = null
  private audioContext: AudioContext | null = null
  private gain: GainNode | null = null
  private audioCursor = 0
  private failures = 0
  private retryTimer = 0
  private generation = 0 // bumps on every connect/close, so late async work is ignored

  // decode chain: a delta frame is undecodable until a keyframe re-establishes a reference
  private haveKeyframe = false
  private pendingSequence = 0
  private ackedSequence = 0
  private droppedFrames = 0
  private keyframeRequests = 0

  get element() {
    return this.canvas
  }

  on<K extends keyof TransportEvents>(event: K, cb: TransportEvents[K]) {
    return this.events.on(event, cb)
  }

  /////////////////////////////
  // connection
  /////////////////////////////

  // again on every reconnect: whatever a previous session left goes, and its last picture stays
  // on the canvas until a new frame replaces it
  async connect(session: SessionInfo) {
    this.closeStream()
    this.failures = 0
    this.session = session
    this.setStatus('connecting')
    await this.open()
  }

  // the stream is over for good: socket, decoders, audio, timers and the picture go
  close() {
    this.generation++
    this.session = null
    this.closeStream()
    this.failures = 0
    this.clearCanvas()
    this.state.video.playable = false
    this.setStatus('disconnected')
  }

  private async open() {
    const session = this.session
    if (!session) return
    const generation = ++this.generation

    let ticket: MediaTicket
    try {
      ticket = await this.fetchTicket(session)
    } catch (err) {
      // the server will not stream to this session (can_watch off, codec unsupported): show the
      // room without video rather than retrying for ever
      if (this.generation === generation) this.setStatus('unavailable')
      return
    }
    if (this.generation !== generation) return

    if (ticket.protocol !== PROTOCOL) {
      return this.fail(new Error(`unsupported media protocol ${ticket.protocol}`))
    }
    if (!(await this.configureDecoders(ticket))) return
    if (this.generation !== generation) return

    const url = new URL(session.url.replace(/^http/, 'ws'))
    url.pathname = url.pathname.replace(/\/+$/, '') + '/media/ws'
    url.searchParams.set('ticket', ticket.ticket)

    const socket = (this.socket = new WebSocket(url.toString()))
    socket.binaryType = 'arraybuffer'
    socket.onmessage = ({ data }) => {
      if (this.socket !== socket) return
      try {
        if (!(data instanceof ArrayBuffer)) throw new Error('media websocket sent a non-binary message')
        this.onSample(parseMediaSample(data))
      } catch (err) {
        this.fail(err instanceof Error ? err : new Error(String(err)))
      }
    }
    socket.onopen = () => {
      if (this.socket !== socket) return
      this.failures = 0
      this.setStatus('connected')
    }
    socket.onclose = () => this.socket === socket && this.onDisconnected()
    socket.onerror = () => this.socket === socket && this.onDisconnected()
  }

  private async fetchTicket(session: SessionInfo): Promise<MediaTicket> {
    const headers: Record<string, string> = { 'Content-Type': 'application/json' }
    if (session.token) headers['Authorization'] = 'Bearer ' + session.token
    const res = await fetch(session.url.replace(/\/+$/, '') + '/api/media/ticket', {
      method: 'POST',
      credentials: 'include',
      headers,
      body: JSON.stringify({ video_id: '' }),
    })
    if (!res.ok) throw new Error('media ticket rejected: ' + res.status)
    return JSON.parse(await res.text()) as MediaTicket
  }

  private async configureDecoders(ticket: MediaTicket) {
    const generation = this.generation
    try {
      const [video, audio] = await Promise.all([
        VideoDecoder.isConfigSupported({ codec: ticket.video.codec, optimizeForLatency: true }),
        AudioDecoder.isConfigSupported({
          codec: ticket.audio.codec,
          sampleRate: ticket.audio.sample_rate,
          numberOfChannels: ticket.audio.number_of_channels,
        }),
      ])
      if (this.generation !== generation) return false
      if (!video.supported) throw new Error(`this browser cannot decode ${ticket.video.codec} video`)
      if (!audio.supported) throw new Error(`this browser cannot decode ${ticket.audio.codec} audio`)

      this.videoDecoder = new VideoDecoder({ output: this.onVideoFrame, error: this.onDecoderError })
      this.videoDecoder.configure({ codec: ticket.video.codec, optimizeForLatency: true })
      this.audioDecoder = new AudioDecoder({ output: this.onAudioData, error: this.onDecoderError })
      this.audioDecoder.configure({
        codec: ticket.audio.codec,
        sampleRate: ticket.audio.sample_rate,
        numberOfChannels: ticket.audio.number_of_channels,
      })
      return true
    } catch (err) {
      // a codec this browser cannot play will not work on a retry either
      this.fail(err instanceof Error ? err : new Error(String(err)))
      return false
    }
  }

  private onDisconnected() {
    this.closeStream()
    if (!this.session) return
    this.setStatus('connecting')
    if (++this.failures > RECONNECT_MAX) {
      return this.events.emit('error', new Error('video connection failed (websocket)'))
    }
    clearTimeout(this.retryTimer)
    this.retryTimer = setTimeout(
      () => this.session && this.open(),
      RECONNECT_BACKOFF_MS * Math.min(this.failures, 4),
    ) as unknown as number
  }

  private closeStream() {
    clearTimeout(this.retryTimer)
    const socket = this.socket
    this.socket = null
    if (socket) {
      socket.onmessage = socket.onclose = socket.onerror = socket.onopen = null
      socket.close()
    }
    this.closeDecoder(this.videoDecoder)
    this.videoDecoder = null
    this.closeDecoder(this.audioDecoder)
    this.audioDecoder = null
    this.haveKeyframe = false
    this.pendingSequence = this.ackedSequence = 0
    this.audioCursor = 0
  }

  private closeDecoder(decoder: VideoDecoder | AudioDecoder | null) {
    try {
      if (decoder && decoder.state !== 'closed') decoder.close()
    } catch {
      // closing a decoder that already failed throws; nothing to do about it
    }
  }

  private fail(error: Error) {
    this.closeStream()
    this.session = null
    this.events.emit('error', error)
  }

  private onDecoderError = (error: DOMException) => this.fail(new Error(error.message || 'media decoder failed'))

  /////////////////////////////
  // samples
  /////////////////////////////

  private onSample(sample: EncodedMediaSample) {
    if (sample.track === 'video') this.onVideoSample(sample)
    else this.onAudioSample(sample)
  }

  private onVideoSample(sample: EncodedMediaSample) {
    const decoder = this.videoDecoder
    if (!decoder || decoder.state !== 'configured') return

    // nothing is decodable until a keyframe arrives
    if (!this.haveKeyframe && !sample.keyframe) return

    // the decoder is behind: drop the frame to keep latency bounded. That breaks the decode chain,
    // so wait for a keyframe and ask for one, rather than closing the connection.
    if (decoder.decodeQueueSize >= MAX_VIDEO_DECODE_QUEUE) {
      this.haveKeyframe = false
      this.droppedFrames++
      this.keyframeRequests++
      this.send(encodeKeyframeRequest())
      // ack it anyway: the server's window must not stall on a frame that will never be painted
      this.ack(sample.sequence)
      return
    }

    if (sample.keyframe) this.haveKeyframe = true
    this.pendingSequence = sample.sequence
    decoder.decode(
      new EncodedVideoChunk({
        type: sample.keyframe ? 'key' : 'delta',
        timestamp: sample.timestamp,
        duration: sample.duration,
        data: sample.data,
        transfer: [sample.data.buffer as ArrayBuffer],
      }),
    )
  }

  private onAudioSample(sample: EncodedMediaSample) {
    const decoder = this.audioDecoder
    if (!decoder || decoder.state !== 'configured') return
    // audio samples decode independently, so a backed up decoder can skip one
    if (decoder.decodeQueueSize >= MAX_AUDIO_DECODE_QUEUE) return
    decoder.decode(
      new EncodedAudioChunk({
        type: 'key',
        timestamp: sample.timestamp,
        duration: sample.duration,
        data: sample.data,
        transfer: [sample.data.buffer as ArrayBuffer],
      }),
    )
  }

  // acks are cumulative, so a stale one is dropped rather than moving the window backwards
  private ack(sequence: number) {
    if (sequence <= this.ackedSequence) return
    this.ackedSequence = sequence
    this.send(encodeAck(sequence))
  }

  private send(data: ArrayBuffer) {
    if (this.socket?.readyState === WebSocket.OPEN) this.socket.send(data)
  }

  private onVideoFrame = (frame: VideoFrame) => {
    try {
      if (!this.canvas || !this.context) return
      const { displayWidth: width, displayHeight: height } = frame
      if (this.canvas.width !== width || this.canvas.height !== height) {
        this.canvas.width = width
        this.canvas.height = height
        this.state.size = { width, height }
        this.emitState()
      }
      if (!this.state.video.playable) {
        this.state.video.playable = true
        this.emitState()
      }
      // a paused stream keeps its last picture rather than freezing mid-frame
      if (this.state.video.playing) {
        this.context.drawImage(frame as unknown as CanvasImageSource, 0, 0, width, height)
      }
    } finally {
      frame.close()
      // ack after painting: the server's window should reflect what the user has seen, not what
      // arrived
      this.ack(this.pendingSequence)
    }
  }

  private onAudioData = (data: AudioData) => {
    try {
      const context = this.audioContext
      if (!context || !this.gain || !this.state.video.playing) return
      const buffer = context.createBuffer(data.numberOfChannels, data.numberOfFrames, data.sampleRate)
      for (let channel = 0; channel < data.numberOfChannels; channel++) {
        data.copyTo(buffer.getChannelData(channel), { planeIndex: channel, format: 'f32-planar' })
      }
      // drifted behind the clock or too far ahead: resync instead of accumulating latency
      const now = context.currentTime
      if (this.audioCursor < now || this.audioCursor > now + AUDIO_MAX_DRIFT_S) {
        this.audioCursor = now + AUDIO_LEAD_S
      }
      const source = context.createBufferSource()
      source.buffer = buffer
      source.connect(this.gain)
      source.start(this.audioCursor)
      this.audioCursor += buffer.duration
    } finally {
      data.close()
    }
  }

  /////////////////////////////
  // the element
  /////////////////////////////

  attach(container: HTMLElement) {
    this.canvas?.remove() // a second attach replaces the element
    const canvas = (this.canvas = document.createElement('canvas'))
    canvas.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;background:transparent'
    this.context = canvas.getContext('2d', { alpha: false }) as CanvasRenderingContext2D | null
    container.append(canvas)
    this.emitState()
    return () => {
      canvas.remove()
      if (this.canvas === canvas) {
        this.canvas = null
        this.context = null
      }
    }
  }

  async setPlaying(on: boolean) {
    this.state.video.playing = on
    if (!on) {
      this.emitState()
      return
    }
    this.ensureAudio()
    // autoplay policy can leave resume() pending until a gesture, so playback is not held up by it
    if (this.audioContext && this.audioContext.state !== 'running') {
      this.audioContext.resume().catch(() => {})
      // sound is blocked until the user interacts: play on, unmute at the first click
      if (!this.state.video.muted && !this.state.video.mutedByAutoplay) {
        this.state.video.mutedByAutoplay = true
        this.applyGain()
        document.addEventListener('click', () => this.setMuted(false), { once: true })
      }
    }
    this.emitState()
  }

  setVolume(volume: number) {
    this.state.video.volume = volume
    this.applyGain()
    this.emitState()
  }

  setMuted(on: boolean) {
    this.state.video.muted = on
    if (!on) {
      this.state.video.mutedByAutoplay = false
      this.audioContext?.resume().catch(() => {})
    }
    this.applyGain()
    this.emitState()
  }

  private ensureAudio() {
    if (this.audioContext) return
    this.audioContext = new AudioContext({ latencyHint: 'interactive' })
    this.gain = this.audioContext.createGain()
    this.gain.connect(this.audioContext.destination)
    this.applyGain()
  }

  private applyGain() {
    const { muted, mutedByAutoplay, volume } = this.state.video
    if (this.gain) this.gain.gain.value = muted || mutedByAutoplay ? 0 : volume
  }

  private clearCanvas() {
    if (this.canvas && this.context) this.context.clearRect(0, 0, this.canvas.width, this.canvas.height)
  }

  /////////////////////////////
  // state
  /////////////////////////////

  // what the window is doing, for tuning it: frames dropped here, the decoder backlog, and the
  // socket's own send buffer
  stats() {
    return {
      droppedFrames: this.droppedFrames,
      keyframeRequests: this.keyframeRequests,
      videoDecodeQueue: this.videoDecoder?.decodeQueueSize ?? 0,
      audioDecodeQueue: this.audioDecoder?.decodeQueueSize ?? 0,
      bufferedAmount: this.socket?.bufferedAmount ?? 0,
      ackedSequence: this.ackedSequence,
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
