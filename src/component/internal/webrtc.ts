import EventEmitter from 'eventemitter3'
import type { WebRTCStats, CursorPosition, CursorImage } from '../types/webrtc'
import { Logger } from '../utils/logger'
import { videoSnap } from '../utils/video-snap'

const maxUint32 = 2 ** 32 - 1

export const OPCODE = {
  MOVE: 0x01,
  SCROLL: 0x02,
  KEY_DOWN: 0x03,
  KEY_UP: 0x04,
  BTN_DOWN: 0x05,
  BTN_UP: 0x06,
  PING: 0x07,
  // touch events
  TOUCH_BEGIN: 0x08,
  TOUCH_UPDATE: 0x09,
  TOUCH_END: 0x0a,
} as const

export interface ICEServer {
  urls: string[]
  username: string
  credential: string
}

export interface NekoWebRTCEvents {
  connected: () => void
  disconnected: (error?: Error) => void
  track: (event: RTCTrackEvent) => void
  candidate: (candidate: RTCIceCandidateInit) => void
  negotiation: (description: RTCSessionDescriptionInit) => void
  stable: (isStable: boolean) => void
  stats: (stats: WebRTCStats) => void
  fallback: (image: string) => void // send last frame image URL as fallback
  ['cursor-position']: (data: CursorPosition) => void
  ['cursor-image']: (data: CursorImage) => void
}

export class NekoWebRTC extends EventEmitter<NekoWebRTCEvents> {
  // used for creating snaps from video for fallback mode
  public video!: HTMLVideoElement
  // information for WebRTC that server video has been paused, 0fps is expected
  public paused = false

  private _peer?: RTCPeerConnection
  private _channel?: RTCDataChannel
  private _track?: MediaStreamTrack
  private _state: RTCIceConnectionState = 'disconnected'
  private _connected = false
  private _candidates: RTCIceCandidateInit[] = []
  private _statsStop?: () => void
  private _requestLatency = 0
  private _responseLatency = 0

  // eslint-disable-next-line
  constructor(
    private readonly _log: Logger = new Logger('webrtc'),
  ) {
    super()
  }

  get supported() {
    return typeof RTCPeerConnection !== 'undefined' && typeof RTCPeerConnection.prototype.addTransceiver !== 'undefined'
  }

  get open() {
    return (
      typeof this._peer !== 'undefined' && typeof this._channel !== 'undefined' && this._channel.readyState == 'open'
    )
  }

  get connected() {
    return this.open && ['connected', 'checking', 'completed'].includes(this._state)
  }

  public async setCandidate(candidate: RTCIceCandidateInit) {
    if (!this._peer) {
      this._candidates.push(candidate)
      return
    }

    await this._peer.addIceCandidate(candidate)
    this._log.debug(`adding remote ICE candidate`, { candidate })
  }

  public async connect(iceServers: ICEServer[]) {
    if (!this.supported) {
      throw new Error('browser does not support webrtc')
    }

    if (this.connected) {
      throw new Error('attempting to create peer while connected')
    }

    this._log.info(`connecting`)

    this._connected = false
    this._peer = new RTCPeerConnection({ iceServers })

    if (iceServers.length == 0) {
      this._log.warn(`iceservers are empty`)
    }

    this._peer.onicecandidate = (event: RTCPeerConnectionIceEvent) => {
      if (!event.candidate) {
        this._log.debug(`sent all local ICE candidates`)
        return
      }

      const init = event.candidate.toJSON()
      this.emit('candidate', init)
      this._log.debug(`sending local ICE candidate`, { init })
    }

    this._peer.onicecandidateerror = (event: Event) => {
      const e = event as RTCPeerConnectionIceErrorEvent
      const fields = { error: e.errorText, code: e.errorCode, port: e.port, url: e.url }
      this._log.warn(`ICE candidate error`, fields)
    }

    this._peer.onconnectionstatechange = () => {
      if (!this._peer) {
        this._log.warn(`attempting to call 'onconnectionstatechange' for nonexistent peer`)
        return
      }

      const state = this._peer.connectionState
      this._log.info(`peer connection state changed`, { state })

      switch (state) {
        case 'connected':
          this.onConnected()
          break
        // Chrome sends failed state change only for connectionState and not iceConnectionState, and firefox
        // does not support connectionState at all.
        case 'closed':
        case 'failed':
          this.onDisconnected(new Error('peer ' + state))
          break
      }
    }

    this._peer.oniceconnectionstatechange = () => {
      if (!this._peer) {
        this._log.warn(`attempting to call 'oniceconnectionstatechange' for nonexistent peer`)
        return
      }

      this._state = this._peer.iceConnectionState
      this._log.info(`ice connection state changed`, { state: this._state })

      switch (this._state) {
        case 'connected':
          this.onConnected()
          // Connected event makes connection stable.
          this.emit('stable', true)
          break
        case 'disconnected':
          // Disconnected event makes connection unstable,
          // may go back to a connected state after some time.
          this.emit('stable', false)
          break
        // We don't watch the disconnected signaling state here as it can indicate temporary issues and may
        // go back to a connected state after some time. Watching it would close the video call on any temporary
        // network issue.
        case 'closed':
        case 'failed':
          this.onDisconnected(new Error('peer ' + this._state))
          break
      }
    }

    this._peer.onsignalingstatechange = () => {
      if (!this._peer) {
        this._log.warn(`attempting to call 'onsignalingstatechange' for nonexistent peer`)
        return
      }

      const state = this._peer.signalingState
      this._log.info(`signaling state changed`, { state })

      // The closed signaling state has been deprecated in favor of the closed iceConnectionState.
      // We are watching for it here to add a bit of backward compatibility.
      if (state == 'closed') {
        this.onDisconnected(new Error('signaling state changed to closed'))
      }
    }

    let negotiating = false
    this._peer.onnegotiationneeded = async () => {
      if (!this._peer) {
        this._log.warn(`attempting to call 'onsignalingstatechange' for nonexistent peer`)
        return
      }

      const state = this._peer.signalingState
      this._log.warn(`negotiation is needed`, { state })

      if (negotiating) {
        this._log.info(`negotiation already in progress; skipping...`)
        return
      }

      negotiating = true

      try {
        // If the connection hasn't yet achieved the "stable" state,
        // return to the caller. Another negotiationneeded event
        // will be fired when the state stabilizes.

        if (state != 'stable') {
          this._log.info(`connection isn't stable yet; postponing...`)
          return
        }

        const offer = await this._peer.createOffer()
        await this._peer.setLocalDescription(offer)

        if (offer) {
          this.emit('negotiation', offer)
        } else {
          this._log.warn(`negotiatoion offer is empty`)
        }
      } catch (error: any) {
        this._log.error(`on negotiation needed failed`, { error })
      } finally {
        negotiating = false
      }
    }

    this._peer.ontrack = this.onTrack.bind(this)
    this._peer.ondatachannel = this.onDataChannel.bind(this)
  }

  public async setOffer(sdp: string) {
    if (!this._peer) {
      throw new Error('attempting to set offer for nonexistent peer')
    }

    await this._peer.setRemoteDescription({ type: 'offer', sdp })

    if (this._candidates.length > 0) {
      let candidates = 0
      for (const candidate of this._candidates) {
        try {
          await this._peer.addIceCandidate(candidate)
          candidates++
        } catch (error: any) {
          this._log.warn(`unable to add remote ICE candidate`, { error })
        }
      }

      this._log.debug(`added ${candidates} remote ICE candidates`, { candidates: this._candidates })
      this._candidates = []
    }

    const answer = await this._peer.createAnswer()

    // add stereo=1 to answer sdp to enable stereo audio for chromium
    answer.sdp = answer.sdp?.replace(/(stereo=1;)?useinbandfec=1/, 'useinbandfec=1;stereo=1')

    await this._peer.setLocalDescription(answer)

    if (answer) {
      this.emit('negotiation', answer)
    } else {
      this._log.warn(`negiotation answer is empty`)
    }
  }

  public async setAnswer(sdp: string) {
    if (!this._peer) {
      throw new Error('attempting to set answer for nonexistent peer')
    }

    await this._peer.setRemoteDescription({ type: 'answer', sdp })
  }

  public async close() {
    if (!this._peer) {
      throw new Error('attempting to close nonexistent peer')
    }

    // create and emit video snap before closing connection
    try {
      const imageSrc = await videoSnap(this.video)
      this.emit('fallback', imageSrc)
    } catch (error: any) {
      this._log.warn(`unable to generate video snap`, { error })
    }

    this.onDisconnected(new Error('connection closed'))
  }

  public addTrack(track: MediaStreamTrack, ...streams: MediaStream[]): RTCRtpSender {
    if (!this._peer) {
      throw new Error('attempting to add track for nonexistent peer')
    }

    // @ts-ignore
    const isChromium = !!window.chrome

    // TOOD: Ugly workaround, find real cause of this issue.
    if (isChromium) {
      return this._peer.addTrack(track, ...streams)
    } else {
      return this._peer.addTransceiver(track, { direction: 'sendonly', streams }).sender
    }
  }

  public removeTrack(sender: RTCRtpSender) {
    if (!this._peer) {
      throw new Error('attempting to add track for nonexistent peer')
    }

    this._peer.removeTrack(sender)
  }

  public disconnect() {
    if (typeof this._channel !== 'undefined') {
      // unmount all events
      this._channel.onerror = () => {}
      this._channel.onmessage = () => {}
      this._channel.onopen = () => {}
      this._channel.onclose = () => {}

      try {
        this._channel.close()
      } catch {}

      this._channel = undefined
    }

    if (typeof this._peer != 'undefined') {
      // unmount all events
      this._peer.onicecandidate = () => {}
      this._peer.onicecandidateerror = () => {}
      this._peer.onconnectionstatechange = () => {}
      this._peer.oniceconnectionstatechange = () => {}
      this._peer.onsignalingstatechange = () => {}
      this._peer.onnegotiationneeded = () => {}
      this._peer.ontrack = () => {}
      this._peer.ondatachannel = () => {}

      try {
        this._peer.close()
      } catch {}

      this._peer = undefined
    }

    if (this._statsStop && typeof this._statsStop === 'function') {
      this._statsStop()
      this._statsStop = undefined
    }

    this._track = undefined
    this._state = 'disconnected'
    this._connected = false
    this._candidates = []
  }

  public send(event: 'wheel', data: { delta_x: number; delta_y: number; control_key?: boolean }): void
  public send(event: 'mousemove', data: { x: number; y: number }): void
  public send(event: 'mousedown' | 'mouseup' | 'keydown' | 'keyup', data: { key: number }): void
  public send(event: 'ping', data: number): void
  public send(
    event: 'touchbegin' | 'touchupdate' | 'touchend',
    data: { touch_id: number; x: number; y: number; pressure: number },
  ): void
  public send(event: string, data: any): void {
    if (typeof this._channel === 'undefined' || this._channel.readyState !== 'open') {
      this._log.warn(`attempting to send data, but data-channel is not open`, { event })
      return
    }

    let buffer: ArrayBuffer
    let payload: DataView
    switch (event) {
      case 'mousemove':
        buffer = new ArrayBuffer(7)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.MOVE)
        payload.setUint16(1, 4)
        payload.setUint16(3, data.x)
        payload.setUint16(5, data.y)
        break
      case 'wheel':
        buffer = new ArrayBuffer(8)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.SCROLL)
        payload.setUint16(1, 5)
        payload.setInt16(3, data.delta_x)
        payload.setInt16(5, data.delta_y)
        payload.setUint8(7, data.control_key ? 1 : 0)
        break
      case 'keydown':
        buffer = new ArrayBuffer(7)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.KEY_DOWN)
        payload.setUint16(1, 4)
        payload.setUint32(3, data.key)
        break
      case 'keyup':
        buffer = new ArrayBuffer(7)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.KEY_UP)
        payload.setUint16(1, 4)
        payload.setUint32(3, data.key)
        break
      case 'mousedown':
        buffer = new ArrayBuffer(7)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.BTN_DOWN)
        payload.setUint16(1, 4)
        payload.setUint32(3, data.key)
        break
      case 'mouseup':
        buffer = new ArrayBuffer(7)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.BTN_UP)
        payload.setUint16(1, 4)
        payload.setUint32(3, data.key)
        break
      case 'ping':
        buffer = new ArrayBuffer(11)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.PING)
        payload.setUint16(1, 8)
        payload.setUint32(3, Math.trunc(data / maxUint32))
        payload.setUint32(7, data % maxUint32)
        break
      case 'touchbegin':
      case 'touchupdate':
      case 'touchend':
        buffer = new ArrayBuffer(16)
        payload = new DataView(buffer)
        if (event === 'touchbegin') {
          payload.setUint8(0, OPCODE.TOUCH_BEGIN)
        } else if (event === 'touchupdate') {
          payload.setUint8(0, OPCODE.TOUCH_UPDATE)
        } else if (event === 'touchend') {
          payload.setUint8(0, OPCODE.TOUCH_END)
        }
        payload.setUint16(1, 13)
        payload.setUint32(3, data.touch_id)
        payload.setInt32(7, data.x)
        payload.setInt32(11, data.y)
        payload.setUint8(15, data.pressure)
        break
      default:
        this._log.warn(`unknown data event`, { event })
        return
    }

    this._channel.send(buffer)
  }

  private onTrack(event: RTCTrackEvent) {
    this._log.debug(`received track from peer`, { label: event.track.label })

    const stream = event.streams[0]
    if (!stream) {
      this._log.warn(`no stream provided for track`, { label: event.track.label })
      return
    }

    if (event.track.kind === 'video') {
      this._track = event.track
    }

    this.emit('track', event)
  }

  private onDataChannel(event: RTCDataChannelEvent) {
    this._log.debug(`received data channel from peer`, { label: event.channel.label })

    this._channel = event.channel
    this._channel.binaryType = 'arraybuffer'
    this._channel.onerror = this.onDisconnected.bind(this, new Error('peer data channel error'))
    this._channel.onmessage = this.onData.bind(this)
    this._channel.onopen = this.onConnected.bind(this)
    this._channel.onclose = this.onDisconnected.bind(this, new Error('peer data channel closed'))
  }

  private onData(e: MessageEvent) {
    const payload = new DataView(e.data)
    const event = payload.getUint8(0)
    const length = payload.getUint16(1)

    switch (event) {
      case 1:
        this.emit('cursor-position', {
          x: payload.getUint16(3),
          y: payload.getUint16(5),
        })
        break
      case 2:
        const data = e.data.slice(11, length - 11)

        // TODO: get string from server
        const blob = new Blob([data], { type: 'image/png' })
        const reader = new FileReader()
        reader.onload = (e) => {
          this.emit('cursor-image', {
            width: payload.getUint16(3),
            height: payload.getUint16(5),
            x: payload.getUint16(7),
            y: payload.getUint16(9),
            uri: String(e.target!.result),
          })
        }
        reader.readAsDataURL(blob)

        break
      case 3:
        const nowTs = Date.now()

        const [clientTs1, clientTs2] = [payload.getUint32(3), payload.getUint32(7)]
        const clientTs = clientTs1 * maxUint32 + clientTs2
        const [serverTs1, serverTs2] = [payload.getUint32(11), payload.getUint32(15)]
        const serverTs = serverTs1 * maxUint32 + serverTs2

        this._requestLatency = serverTs - clientTs
        this._responseLatency = nowTs - serverTs

        break
      default:
        this._log.warn(`unhandled webrtc event`, { event, payload })
    }
  }

  private onConnected() {
    if (!this.connected || this._connected) {
      return
    }

    this._log.info(`connected`)
    this.emit('connected')

    this._statsStop = this.statsEmitter()
    this._connected = true
  }

  private onDisconnected(error?: Error) {
    const wasConnected = this._connected
    this.disconnect()

    if (wasConnected) {
      this._log.info(`disconnected`, { error })
      this.emit('disconnected', error)
    }
  }

  private statsEmitter(ms: number = 2000) {
    let bytesReceived: number
    let timestamp: number
    let framesDecoded: number
    let packetsLost: number
    let packetsReceived: number

    const timer = window.setInterval(async () => {
      if (!this._peer) return

      let stats: RTCStatsReport | undefined = undefined
      if (this._peer.getStats.length === 0) {
        stats = await this._peer.getStats()
      } else {
        // callback browsers support
        await new Promise((res) => {
          //@ts-ignore
          this._peer.getStats((stats) => res(stats))
        })
      }

      if (!stats) return

      let report: any = null
      stats.forEach(function (stat) {
        if (stat.type === 'inbound-rtp' && stat.kind === 'video') {
          report = { ...stat }
        }
      })

      if (report === null) return

      if (timestamp) {
        const bytesDiff = (report.bytesReceived - bytesReceived) * 8
        const tsDiff = report.timestamp - timestamp
        const framesDecodedDiff = report.framesDecoded - framesDecoded
        const packetsLostDiff = report.packetsLost - packetsLost
        const packetsReceivedDiff = report.packetsReceived - packetsReceived

        this.emit('stats', {
          // Firefox does not emit any event when starting paused
          // because there is no video report found in stats.
          paused: this.paused,
          bitrate: (bytesDiff / tsDiff) * 1000,
          packetLoss: (packetsLostDiff / (packetsLostDiff + packetsReceivedDiff)) * 100,
          fps: Number(report.framesPerSecond || framesDecodedDiff / (tsDiff / 1000)),
          width: report.frameWidth || NaN,
          height: report.frameHeight || NaN,
          muted: this._track?.muted,
          // latency from ping/pong messages
          latency: this._requestLatency + this._responseLatency,
          requestLatency: this._requestLatency,
          responseLatency: this._responseLatency,
        })
      }

      bytesReceived = report.bytesReceived
      timestamp = report.timestamp
      framesDecoded = report.framesDecoded
      packetsLost = report.packetsLost
      packetsReceived = report.packetsReceived

      this.send('ping', Date.now())
    }, ms)

    return function () {
      window.clearInterval(timer)
    }
  }
}
