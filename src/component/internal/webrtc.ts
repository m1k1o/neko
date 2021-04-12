import EventEmitter from 'eventemitter3'
import { Logger } from '../utils/logger'

export const OPCODE = {
  MOVE: 0x01,
  SCROLL: 0x02,
  KEY_DOWN: 0x03,
  KEY_UP: 0x04,
  BTN_DOWN: 0x05,
  BTN_UP: 0x06,
} as const

export interface WebRTCStats {
  bitrate: number
  packetLoss: number
  fps: number
  width: number
  height: number
}

export interface ICEServer {
  urls: string
  username: string
  credential: string
}

export interface NekoWebRTCEvents {
  connecting: () => void
  connected: () => void
  disconnected: (error?: Error) => void
  track: (event: RTCTrackEvent) => void
  candidate: (candidate: RTCIceCandidateInit) => void
  stats: (stats: WebRTCStats) => void
  ['cursor-position']: (data: { x: number; y: number }) => void
  ['cursor-image']: (data: { width: number; height: number; x: number; y: number; uri: string }) => void
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

  public async connect(sdp: string, iceServers: ICEServer[]): Promise<string> {
    this._log.info(`connecting`)

    if (!this.supported) {
      throw new Error('browser does not support webrtc')
    }

    if (this.connected) {
      throw new Error('attempting to create peer while connected')
    }

    this.emit('connecting')

    this._peer = new RTCPeerConnection({ iceServers })

    if (iceServers.length == 0) {
      this._log.warn(`iceservers are empty`)
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
    this._peer.ondatachannel = this.onDataChannel.bind(this)
    this._peer.addTransceiver('audio', { direction: 'recvonly' })
    this._peer.addTransceiver('video', { direction: 'recvonly' })
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
        this._log.warn(`unknown data event: ${event}`)
        return
    }

    this._channel!.send(buffer)
  }

  private onTrack(event: RTCTrackEvent) {
    this._log.debug(`received ${event.track.kind} track from peer: ${event.track.id}`, event)
    const stream = event.streams[0]

    if (!stream) {
      this._log.warn(`no stream provided for track ${event.track.id}(${event.track.label})`)
      return
    }

    this.emit('track', event)
  }

  private onDataChannel(event: RTCDataChannelEvent) {
    this._log.debug(`received data channel from peer: ${event.channel.label}`, event)

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
        this._log.warn(`unhandled webrtc event '${event}'.`, payload)
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
