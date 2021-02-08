import EventEmitter from 'eventemitter3'
import { Logger } from '../utils/logger'

export const OPCODE = {
  MOVE: 0x01,
  SCROLL: 0x02,
  KEY_DOWN: 0x03,
  KEY_UP: 0x04,
} as const

export interface WebRTCStats {
  bitrate: number
  packetLoss: number
  fps: number
  width: number
  height: number
}

export interface NekoWebRTCEvents {
  connecting: () => void
  connected: () => void
  disconnected: (error?: Error) => void
  track: (event: RTCTrackEvent) => void
  candidate: (candidate: RTCIceCandidateInit) => void
  stats: (stats: WebRTCStats) => void
}

export class NekoWebRTC extends EventEmitter<NekoWebRTCEvents> {
  private _peer?: RTCPeerConnection
  private _channel?: RTCDataChannel
  private _state: RTCIceConnectionState = 'disconnected'
  private _log: Logger
  private _statsStop?: () => void

  constructor() {
    super()

    this._log = new Logger('webrtc')
  }

  get supported() {
    return typeof RTCPeerConnection !== 'undefined' && typeof RTCPeerConnection.prototype.addTransceiver !== 'undefined'
  }

  get connected() {
    return (
      typeof this._peer !== 'undefined' &&
      ['connected', 'checking', 'completed'].includes(this._state) &&
      typeof this._channel !== 'undefined' &&
      this._channel.readyState == 'open'
    )
  }

  candidates: RTCIceCandidateInit[] = []
  public async setCandidate(candidate: RTCIceCandidateInit) {
    if (!this._peer) {
      this.candidates.push(candidate)
      return
    }

    this._peer.addIceCandidate(candidate)
    this._log.debug(`adding remote ICE candidate`, candidate)
  }

  public async connect(sdp: string, lite: boolean, servers: string[]): Promise<string> {
    this._log.info(`connecting`)

    if (!this.supported) {
      throw new Error('browser does not support webrtc')
    }

    if (this.connected) {
      throw new Error('attempting to create peer while connected')
    }

    this.emit('connecting')

    this._peer = new RTCPeerConnection()
    if (lite !== true) {
      this._peer = new RTCPeerConnection({
        iceServers: [{ urls: servers }],
      })
    }

    this._peer.onconnectionstatechange = (event) => {
      this._log.debug(`peer connection state changed`, this._peer ? this._peer.connectionState : undefined)
    }

    this._peer.onsignalingstatechange = (event) => {
      this._log.debug(`peer signaling state changed`, this._peer ? this._peer.signalingState : undefined)
    }

    this._peer.onicecandidate = (event: RTCPeerConnectionIceEvent) => {
      if (!event.candidate) {
        this._log.debug(`sent all remote ICE candidates`)
        return
      }

      const init = event.candidate.toJSON()
      this.emit('candidate', init)
      this._log.debug(`sending remote ICE candidate`, init)
    }

    this._peer.oniceconnectionstatechange = (event) => {
      this._state = this._peer!.iceConnectionState
      this._log.debug(`peer ice connection state changed: ${this._peer!.iceConnectionState}`)

      switch (this._state) {
        case 'disconnected':
          this.onDisconnected(new Error('peer disconnected'))
          break
        case 'failed':
          this.onDisconnected(new Error('peer failed'))
          break
        case 'closed':
          this.onDisconnected(new Error('peer closed'))
          break
      }
    }

    this._peer.ontrack = this.onTrack.bind(this)
    this._peer.addTransceiver('audio', { direction: 'recvonly' })
    this._peer.addTransceiver('video', { direction: 'recvonly' })

    this._channel = this._peer.createDataChannel('data')
    this._channel.onerror = this.onDisconnected.bind(this, new Error('peer data channel error'))
    this._channel.onmessage = this.onData.bind(this)
    this._channel.onopen = this.onConnected.bind(this)
    this._channel.onclose = this.onDisconnected.bind(this, new Error('peer data channel closed'))

    this._peer.setRemoteDescription({ type: 'offer', sdp })

    if (this.candidates.length > 0) {
      for (const candidate of this.candidates) {
        this._peer.addIceCandidate(candidate)
      }

      this._log.debug(`added ${this.candidates.length} remote ICE candidates`, this.candidates)
      this.candidates = []
    }

    const answer = await this._peer.createAnswer()
    this._peer!.setLocalDescription(answer)

    if (!answer.sdp) {
      throw new Error('sdp answer is empty')
    }

    return answer.sdp
  }

  public disconnect() {
    try {
      this._peer!.close()
    } catch (err) {}

    this._peer = undefined
    this._channel = undefined
    this._state = 'disconnected'
  }

  public send(event: 'wheel' | 'mousemove', data: { x: number; y: number }): void
  public send(event: 'mousedown' | 'mouseup' | 'keydown' | 'keyup', data: { key: number }): void
  public send(event: string, data: any): void {
    if (!this.connected) {
      this._log.warn(`attempting to send data while disconnected`)
      return
    }

    let buffer: ArrayBuffer
    let payload: DataView
    switch (event) {
      case 'mousemove':
        buffer = new ArrayBuffer(7)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.MOVE)
        payload.setUint16(1, 4, true)
        payload.setUint16(3, data.x, true)
        payload.setUint16(5, data.y, true)
        break
      case 'wheel':
        buffer = new ArrayBuffer(7)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.SCROLL)
        payload.setUint16(1, 4, true)
        payload.setInt16(3, data.x, true)
        payload.setInt16(5, data.y, true)
        break
      case 'keydown':
      case 'mousedown':
        buffer = new ArrayBuffer(11)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.KEY_DOWN)
        payload.setUint16(1, 8, true)
        payload.setBigUint64(3, BigInt(data.key), true)
        break
      case 'keyup':
      case 'mouseup':
        buffer = new ArrayBuffer(11)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.KEY_UP)
        payload.setUint16(1, 8, true)
        payload.setBigUint64(3, BigInt(data.key), true)
        break
      default:
        this._log.warn(`unknown data event: ${event}`)
        return
    }

    this._channel!.send(buffer)
  }

  // not-implemented
  private onData(e: MessageEvent) {}

  private onTrack(event: RTCTrackEvent) {
    this._log.debug(`received ${event.track.kind} track from peer: ${event.track.id}`, event)
    const stream = event.streams[0]

    if (!stream) {
      this._log.warn(`no stream provided for track ${event.track.id}(${event.track.label})`)
      return
    }

    this.emit('track', event)
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

  private onDisconnected(reason?: Error) {
    this.disconnect()

    this._log.info(`disconnected:`, reason?.message)
    this.emit('disconnected', reason)

    if (this._statsStop && typeof this._statsStop === 'function') {
      this._statsStop()
      this._statsStop = undefined
    }
  }

  private statsEmitter(ms: number = 2000) {
    let bytesReceived: number
    let timestamp: number
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
          report = stat
        }
      })

      if (report === null) return

      if (timestamp) {
        const bytesDiff = (report.bytesReceived - bytesReceived) * 8
        const tsDiff = report.timestamp - timestamp
        const packetsLostDiff = report.packetsLost - packetsLost
        const packetsReceivedDiff = report.packetsReceived - packetsReceived

        this.emit('stats', {
          bitrate: (bytesDiff / tsDiff) * 1000,
          packetLoss: (packetsLostDiff / (packetsLostDiff + packetsReceivedDiff)) * 100,
          fps: report.framesPerSecond,
          width: report.frameWidth,
          height: report.frameHeight,
        })
      }

      bytesReceived = report.bytesReceived
      timestamp = report.timestamp
      packetsLost = report.packetsLost
      packetsReceived = report.packetsReceived
    }, ms)

    return function () {
      clearInterval(timer)
    }
  }
}
