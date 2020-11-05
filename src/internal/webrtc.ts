import EventEmitter from 'eventemitter3'
import { Logger } from '../utils/logger'

export const OPCODE = {
  MOVE: 0x01,
  SCROLL: 0x02,
  KEY_DOWN: 0x03,
  KEY_UP: 0x04,
} as const

export interface NekoWebRTCEvents {
  connecting: () => void
  connected: () => void
  disconnected: (error?: Error) => void
  track: (event: RTCTrackEvent) => void
}

export abstract class NekoWebRTC extends EventEmitter<NekoWebRTCEvents> {
  private _peer?: RTCPeerConnection
  private _channel?: RTCDataChannel
  private _state: RTCIceConnectionState = 'disconnected'
  private _log: Logger

  constructor() {
    super()
  
    this._log = new Logger('webrtc')
  }

  get supported() {
    return typeof RTCPeerConnection !== 'undefined' && typeof RTCPeerConnection.prototype.addTransceiver !== 'undefined'
  }

  get connected() {
    return typeof this._peer !== 'undefined' && ['connected', 'checking', 'completed'].includes(this._state)
  }

  public async connect(sdp: string, lite: boolean, servers: string[]): Promise<string> {
    this._log.debug(`creating peer`)
  
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

    this._peer.oniceconnectionstatechange = (event) => {
      this._state = this._peer!.iceConnectionState
      this._log.debug(`peer ice connection state changed: ${this._peer!.iceConnectionState}`)

      switch (this._state) {
        case 'checking':
          break
        case 'connected':
          this.onConnected()
          break
        case 'failed':
          this.onDisconnected(new Error('peer failed'))
          break
        case 'disconnected':
          this.onDisconnected(new Error('peer disconnected'))
          break
      }
    }

    this._peer.ontrack = this.onTrack.bind(this)
    this._peer.addTransceiver('audio', { direction: 'recvonly' })
    this._peer.addTransceiver('video', { direction: 'recvonly' })

    this._channel = this._peer.createDataChannel('data')
    this._channel.onerror = this.onError.bind(this)
    this._channel.onmessage = this.onData.bind(this)
    this._channel.onclose = this.onDisconnected.bind(this, new Error('peer data channel closed'))

    this._peer.setRemoteDescription({ type: 'offer', sdp })

    let answer = await this._peer.createAnswer()
    this._peer!.setLocalDescription(answer)

    if (!answer.sdp) {
      throw new Error('sdp answer is empty')
    }

    return answer.sdp
  }

  public disconnect() {
    if (this.connected) {
      try {
        this._peer!.close()
      } catch (err) {}

      this._peer = undefined
      this._channel = undefined
    }

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
    }

    // @ts-ignore
    if (typeof buffer !== 'undefined') {
      this._channel!.send(buffer)
    }
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

  private onError(event: Event) {
    this._log.error((event as ErrorEvent).error)
  }

  private onConnected() {
    if (!this.connected) {
      this._log.warn(`onConnected called while being disconnected`)
      return
    }

    this._log.debug(`connected`)
    this.emit('connected')
  }

  private onDisconnected(reason?: Error) {
    this.disconnect()

    this._log.debug(`disconnected:`, reason)
    this.emit('disconnected', reason)
  }
}
