import EventEmitter from 'eventemitter3'
import { OPCODE } from './data'
import { EVENT, WebSocketEvents } from './events'

import { WebSocketMessages, WebSocketPayloads, SignalProvidePayload } from './messages'

export interface BaseEvents {
  info: (...message: any[]) => void
  warn: (...message: any[]) => void
  debug: (...message: any[]) => void
  error: (error: Error) => void
}

export abstract class BaseClient extends EventEmitter<BaseEvents> {
  protected _ws?: WebSocket
  protected _peer?: RTCPeerConnection
  protected _channel?: RTCDataChannel
  protected _timeout?: NodeJS.Timeout
  protected _displayname?: string
  protected _state: RTCIceConnectionState = 'disconnected'
  protected _id = ''

  get id() {
    return this._id
  }

  get supported() {
    return typeof RTCPeerConnection !== 'undefined' && typeof RTCPeerConnection.prototype.addTransceiver !== 'undefined'
  }

  get socketOpen() {
    return typeof this._ws !== 'undefined' && this._ws.readyState === WebSocket.OPEN
  }

  get peerConnected() {
    return typeof this._peer !== 'undefined' && ['connected', 'checking', 'completed'].includes(this._state)
  }

  get connected() {
    return this.peerConnected && this.socketOpen
  }

  public connect(url: string, password: string, displayname: string) {
    if (this.socketOpen) {
      this.emit('warn', `attempting to create websocket while connection open`)
      return
    }

    if (!this.supported) {
      this.onDisconnected(new Error('browser does not support webrtc (RTCPeerConnection missing)'))
      return
    }

    if (displayname === '') {
      throw new Error('Must add a displayname') // TODO: Better handling
    }

    this._displayname = displayname
    this[EVENT.CONNECTING]()

    try {
      this._ws = new WebSocket(`${url}ws?password=${password}`)
      this.emit('debug', `connecting to ${this._ws.url}`)
      this._ws.onmessage = this.onMessage.bind(this)
      this._ws.onerror = event => this.onError.bind(this)
      this._ws.onclose = event => this.onDisconnected.bind(this, new Error('websocket closed'))
      this._timeout = setTimeout(this.onTimeout.bind(this), 15000)
    } catch (err) {
      this.onDisconnected(err)
    }
  }

  protected disconnect() {
    if (this._timeout) {
      clearTimeout(this._timeout)
    }

    if (this.socketOpen) {
      try {
        this._ws!.close()
      } catch (err) {}
      this._ws = undefined
    }

    if (this.peerConnected) {
      try {
        this._peer!.close()
      } catch (err) {}
      this._peer = undefined
    }

    this._state = 'disconnected'
    this._displayname = undefined
    this._id = ''
  }

  public sendData(event: 'wheel' | 'mousemove', data: { x: number; y: number }): void
  public sendData(event: 'mousedown' | 'mouseup' | 'keydown' | 'keyup', data: { key: number }): void
  public sendData(event: string, data: any) {
    if (!this.connected) {
      this.emit('warn', `attempting to send data while disconnected`)
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
        buffer = new ArrayBuffer(5)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.KEY_DOWN)
        payload.setUint16(1, 1, true)
        payload.setUint16(3, data.key, true)
        break
      case 'keyup':
      case 'mouseup':
        buffer = new ArrayBuffer(5)
        payload = new DataView(buffer)
        payload.setUint8(0, OPCODE.KEY_UP)
        payload.setUint16(1, 1, true)
        payload.setUint16(3, data.key, true)
        break
      default:
        this.emit('warn', `unknown data event: ${event}`)
    }

    // @ts-ignore
    if (typeof buffer !== 'undefined') {
      this._channel!.send(buffer)
    }
  }

  public sendMessage(event: WebSocketEvents, payload?: WebSocketPayloads) {
    if (!this.connected) {
      this.emit('warn', `attempting to send message while disconnected`)
      return
    }
    this.emit('debug', `sending event '${event}' ${payload ? `with payload: ` : ''}`, payload)
    this._ws!.send(JSON.stringify({ event, ...payload }))
  }

  public createPeer(sdp: string, lite: boolean, servers: string[]) {
    this.emit('debug', `creating peer`)
    if (!this.socketOpen) {
      this.emit(
        'warn',
        `attempting to create peer with no websocket: `,
        this._ws ? `state: ${this._ws.readyState}` : 'no socket',
      )
      return
    }

    if (this.peerConnected) {
      this.emit('warn', `attempting to create peer while connected`)
      return
    }

    this._peer = new RTCPeerConnection()
    if (lite !== true) {
      this._peer = new RTCPeerConnection({
        iceServers: [{ urls: servers }],
      })
    }

    this._peer.onconnectionstatechange = event => {
      this.emit('debug', `peer connection state changed`, this._peer ? this._peer.connectionState : undefined)
    }

    this._peer.onsignalingstatechange = event => {
      this.emit('debug', `peer signaling state changed`, this._peer ? this._peer.signalingState : undefined)
    }

    this._peer.oniceconnectionstatechange = event => {
      this._state = this._peer!.iceConnectionState

      this.emit('debug', `peer ice connection state changed: ${this._peer!.iceConnectionState}`)

      switch (this._state) {
        case 'checking':
          if (this._timeout) {
            clearTimeout(this._timeout)
          }
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
    this._peer
      .createAnswer()
      .then(d => {
        this._peer!.setLocalDescription(d)
        this._ws!.send(
          JSON.stringify({
            event: EVENT.SIGNAL.ANSWER,
            sdp: d.sdp,
            displayname: this._displayname,
          }),
        )
      })
      .catch(err => this.emit('error', err))
  }

  private onMessage(e: MessageEvent) {
    const { event, ...payload } = JSON.parse(e.data) as WebSocketMessages

    this.emit('debug', `received websocket event ${event} ${payload ? `with payload: ` : ''}`, payload)

    if (event === EVENT.SIGNAL.PROVIDE) {
      const { sdp, lite, ice, id } = payload as SignalProvidePayload
      this._id = id
      this.createPeer(sdp, lite, ice)
      return
    }

    // @ts-ignore
    if (typeof this[event] === 'function') {
      // @ts-ignore
      this[event](payload)
    } else {
      this[EVENT.MESSAGE](event, payload)
    }
  }

  private onData(e: MessageEvent) {
    this[EVENT.DATA](e.data)
  }

  private onTrack(event: RTCTrackEvent) {
    this.emit('debug', `received ${event.track.kind} track from peer: ${event.track.id}`, event)
    const stream = event.streams[0]
    if (!stream) {
      this.emit('warn', `no stream provided for track ${event.track.id}(${event.track.label})`)
      return
    }
    this[EVENT.TRACK](event)
  }

  private onError(event: Event) {
    this.emit('error', (event as ErrorEvent).error)
  }

  private onConnected() {
    if (this._timeout) {
      clearTimeout(this._timeout)
    }

    if (!this.connected) {
      this.emit('warn', `onConnected called while being disconnected`)
      return
    }

    this.emit('debug', `connected`)
    this[EVENT.CONNECTED]()
  }

  private onTimeout() {
    this.emit('debug', `connection timeout`)
    if (this._timeout) {
      clearTimeout(this._timeout)
    }
    this.onDisconnected(new Error('connection timeout'))
  }

  protected onDisconnected(reason?: Error) {
    this.disconnect()
    this.emit('debug', `disconnected:`, reason)
    this[EVENT.DISCONNECTED](reason)
  }

  protected [EVENT.MESSAGE](event: string, payload: any) {
    this.emit('warn', `unhandled websocket event '${event}':`, payload)
  }

  protected abstract [EVENT.CONNECTING](): void
  protected abstract [EVENT.CONNECTED](): void
  protected abstract [EVENT.DISCONNECTED](reason?: Error): void
  protected abstract [EVENT.TRACK](event: RTCTrackEvent): void
  protected abstract [EVENT.DATA](data: any): void
}
