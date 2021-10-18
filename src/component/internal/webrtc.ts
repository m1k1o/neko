import EventEmitter from 'eventemitter3'
import { WebRTCStats, CursorPosition, CursorImage } from '../types/webrtc'
import { Logger } from '../utils/logger'

export const OPCODE = {
  MOVE: 0x01,
  SCROLL: 0x02,
  KEY_DOWN: 0x03,
  KEY_UP: 0x04,
  BTN_DOWN: 0x05,
  BTN_UP: 0x06,
} as const

export interface ICEServer {
  urls: string
  username: string
  credential: string
}

export interface NekoWebRTCEvents {
  connected: () => void
  disconnected: (error?: Error) => void
  track: (event: RTCTrackEvent) => void
  candidate: (candidate: RTCIceCandidateInit) => void
  stats: (stats: WebRTCStats) => void
  ['cursor-position']: (data: CursorPosition) => void
  ['cursor-image']: (data: CursorImage) => void
}

export class NekoWebRTC extends EventEmitter<NekoWebRTCEvents> {
  private _peer?: RTCPeerConnection
  private _channel?: RTCDataChannel
  private _track?: MediaStreamTrack
  private _state: RTCIceConnectionState = 'disconnected'
  private _candidates: RTCIceCandidateInit[] = []
  private _statsStop?: () => void

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

    this._peer.addIceCandidate(candidate)
    this._log.debug(`adding remote ICE candidate`, { candidate })
  }

  public async connect(sdp: string, iceServers: ICEServer[]): Promise<string> {
    if (!this.supported) {
      throw new Error('browser does not support webrtc')
    }

    if (this.connected) {
      throw new Error('attempting to create peer while connected')
    }

    this._log.info(`connecting`)

    this._peer = new RTCPeerConnection({ iceServers })

    if (iceServers.length == 0) {
      this._log.warn(`iceservers are empty`)
    }

    this._peer.onicecandidate = (event: RTCPeerConnectionIceEvent) => {
      if (!event.candidate) {
        this._log.debug(`sent all remote ICE candidates`)
        return
      }

      const init = event.candidate.toJSON()
      this.emit('candidate', init)
      this._log.debug(`sending remote ICE candidate`, { init })
    }

    this._peer.onicecandidateerror = (event: Event) => {
      const e = event as RTCPeerConnectionIceErrorEvent
      const fields = { error: e.errorText, code: e.errorCode, port: e.port, url: e.url }
      this._log.warn(`ICE candidate error`, fields)
    }

    this._peer.onconnectionstatechange = () => {
      const state = this._peer!.connectionState
      this._log.info(`peer connection state changed`, { state })

      switch (state) {
        // Chrome sends failed state change only for connectionState and not iceConnectionState, and firefox
        // does not support connectionState at all.
        case 'closed':
        case 'failed':
          this.onDisconnected(new Error('peer ' + state))
          break
      }
    }

    this._peer.oniceconnectionstatechange = () => {
      this._state = this._peer!.iceConnectionState
      this._log.info(`peer ice connection state changed`, { state: this._state })

      switch (this._state) {
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
      const state = this._peer!.iceConnectionState
      this._log.info(`peer signaling state changed`, { state })

      switch (state) {
        // The closed signaling state has been deprecated in favor of the closed iceConnectionState.
        // We are watching for it here to add a bit of backward compatibility.
        case 'closed':
        case 'failed':
          this.onDisconnected(new Error('peer ' + state))
          break
      }
    }

    this._peer.onnegotiationneeded = () => {
      this._log.warn(`negotiation is needed`)
    }

    this._peer.ontrack = this.onTrack.bind(this)
    this._peer.ondatachannel = this.onDataChannel.bind(this)
    this._peer.addTransceiver('audio', { direction: 'recvonly' })
    this._peer.addTransceiver('video', { direction: 'recvonly' })

    return await this.offer(sdp)
  }

  public async offer(sdp: string) {
    if (!this._peer) {
      throw new Error('attempting to set offer for nonexistent peer')
    }

    this._peer.setRemoteDescription({ type: 'offer', sdp })

    if (this._candidates.length > 0) {
      for (const candidate of this._candidates) {
        this._peer.addIceCandidate(candidate)
      }

      this._log.debug(`added ${this._candidates.length} remote ICE candidates`, { candidates: this._candidates })
      this._candidates = []
    }

    const answer = await this._peer.createAnswer()
    this._peer!.setLocalDescription(answer)

    if (!answer.sdp) {
      throw new Error('sdp answer is empty')
    }

    return answer.sdp
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

    this._track = undefined
    this._state = 'disconnected'
    this._candidates = []
  }

  public send(event: 'wheel' | 'mousemove', data: { x: number; y: number }): void
  public send(event: 'mousedown' | 'mouseup' | 'keydown' | 'keyup', data: { key: number }): void
  public send(event: string, data: any): void {
    if (!this.connected) {
      this._log.warn(`attempting to send data while disconnected`, { event })
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
        buffer = new ArrayBuffer(7)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.SCROLL)
        payload.setUint16(1, 4)
        payload.setInt16(3, data.x)
        payload.setInt16(5, data.y)
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
      default:
        this._log.warn(`unknown data event`, { event })
        return
    }

    this._channel!.send(buffer)
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
      default:
        this._log.warn(`unhandled webrtc event`, { event, payload })
    }
  }

  private onConnected() {
    if (!this.connected) {
      this._log.warn(`onConnected called while being disconnected`)
      return
    }

    this._log.info(`connected`)
    this.emit('connected')

    this._statsStop = this.statsEmitter()
  }

  private onDisconnected(error?: Error) {
    this.disconnect()

    this._log.info(`disconnected`, { error })
    this.emit('disconnected', error)

    if (this._statsStop && typeof this._statsStop === 'function') {
      this._statsStop()
      this._statsStop = undefined
    }
  }

  private statsEmitter(ms: number = 2000) {
    let bytesReceived: number
    let timestamp: number
    let framesDecoded: number
    let packetsLost: number
    let packetsReceived: number

    const timer = setInterval(async () => {
      if (!this._peer) return

      let stats: RTCStatsReport | undefined = undefined
      if (this._peer.getStats.length === 0) {
        stats = await this._peer.getStats()
      } else {
        // callback browsers support
        await new Promise((res, rej) => {
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
          bitrate: (bytesDiff / tsDiff) * 1000,
          packetLoss: (packetsLostDiff / (packetsLostDiff + packetsReceivedDiff)) * 100,
          fps: Number(report.framesPerSecond || framesDecodedDiff / (tsDiff / 1000)),
          width: report.frameWidth || NaN,
          height: report.frameHeight || NaN,
          muted: this._track?.muted,
        })
      }

      bytesReceived = report.bytesReceived
      timestamp = report.timestamp
      framesDecoded = report.framesDecoded
      packetsLost = report.packetsLost
      packetsReceived = report.packetsReceived
    }, ms)

    return function () {
      clearInterval(timer)
    }
  }
}
