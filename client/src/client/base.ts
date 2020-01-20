import EventEmitter from 'eventemitter3'

import { OPCODE } from './data'

import { EVENT, WebSocketEvents } from './events'

import {
  WebSocketMessages,
  WebSocketPayloads,
  IdentityPayload,
  SignalPayload,
  MemberListPayload,
  MemberPayload,
  ControlPayload,
} from './messages'

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
  protected _timeout?: number
  protected _username?: string
  protected _state: RTCIceConnectionState = 'disconnected'

  get socketOpen() {
    return typeof this._ws !== 'undefined' && this._ws.readyState === WebSocket.OPEN
  }

  get peerConnected() {
    return typeof this._peer !== 'undefined' && this._state === 'connected'
  }

  get connected() {
    return this.peerConnected && this.socketOpen
  }

  public connect(url: string, password: string, username: string) {
    if (this.socketOpen) {
      this.emit('warn', `attempting to create websocket while connection open`)
      return
    }

    if (username === '') {
      throw new Error('Must add a username') // TODO: Better handleing
    }
    this._username = username

    this._ws = new WebSocket(`${url}ws?password=${password}`)
    this.emit('debug', `connecting to ${this._ws.url}`)
    this._ws.onmessage = this.onMessage.bind(this)
    this._ws.onerror = event => this.onError.bind(this)
    this._ws.onclose = event => this.onDisconnected.bind(this, new Error('websocket closed'))
    this._timeout = setTimeout(this.onTimeout.bind(this), 5000)
    this[EVENT.CONNECTING]()
  }

  public sendData(event: 'wheel' | 'mousemove', data: { x: number; y: number }): void
  public sendData(event: 'mousedown' | 'mouseup' | 'keydown' | 'keyup', data: { key: number }): void
  public sendData(event: string, data: any) {
    if (!this.connected) {
      this.emit('warn', `attemping to send data while dissconneted`)
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
      this.emit('warn', `attemping to send message while dissconneted`)
      return
    }
    this.emit('debug', `sending event '${event}' ${payload ? `with payload: ` : ''}`, payload)
    this._ws!.send(JSON.stringify({ event, ...payload }))
  }

  public createPeer() {
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

    this._peer = new RTCPeerConnection({
      iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
    })

    this._peer.onicecandidate = event => {
      if (event.candidate === null && this._peer!.localDescription) {
        this.emit('debug', `sending event '${EVENT.SIGNAL.PROVIDE}' with payload`, this._peer!.localDescription.sdp)
        this._ws!.send(
          JSON.stringify({
            event: EVENT.SIGNAL.PROVIDE,
            sdp: this._peer!.localDescription.sdp,
          }),
        )
      }
    }

    this._peer.oniceconnectionstatechange = event => {
      this._state = this._peer!.iceConnectionState
      this.emit('debug', `peer connection state chagned: ${this._state}`)

      switch (this._state) {
        case 'connected':
          this.onConnected()
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

    this._peer
      .createOffer()
      .then(d => this._peer!.setLocalDescription(d))
      .catch(err => this.emit('error', err))
  }

  private setRemoteDescription(payload: SignalPayload) {
    if (this.peerConnected) {
      this.emit('warn', `received ${event} with no peer!`)
      return
    }
    this._peer!.setRemoteDescription({ type: 'answer', sdp: payload.sdp })
  }

  private onMessage(e: MessageEvent) {
    const { event, ...payload } = JSON.parse(e.data) as WebSocketMessages

    this.emit('debug', `received websocket event ${event} ${payload ? `with payload: ` : ''}`, payload)

    switch (event) {
      case EVENT.IDENTITY.PROVIDE:
        this[EVENT.IDENTITY.PROVIDE](payload as IdentityPayload)
        this.createPeer()
        break
      case EVENT.SIGNAL.ANSWER:
        this.setRemoteDescription(payload as SignalPayload)
        break
      default:
        // @ts-ignore
        if (typeof this[event] === 'function') {
          // @ts-ignore
          this[event](payload)
        } else {
          this[EVENT.MESSAGE](event, payload)
        }
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

    this.emit('debug', `sending event '${EVENT.IDENTITY.DETAILS}' with payload`, { username: this._username })
    this._ws!.send(
      JSON.stringify({
        event: EVENT.IDENTITY.DETAILS,
        username: this._username,
      }),
    )

    this.emit('debug', `connected`)
    this[EVENT.CONNECTED]()
  }

  private onTimeout() {
    this.emit('debug', `connection timedout`)
    if (this._timeout) {
      clearTimeout(this._timeout)
    }
    this.onDisconnected(new Error('connection timeout'))
  }

  private onDisconnected(reason?: Error) {
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

    this.emit('debug', `disconnected:`, reason)
    this[EVENT.DISCONNECTED](reason)
  }

  [EVENT.MESSAGE](event: string, payload: any) {
    this.emit('warn', `unhandled websocket event '${event}':`, payload)
  }

  abstract [EVENT.CONNECTING](): void
  abstract [EVENT.CONNECTED](): void
  abstract [EVENT.DISCONNECTED](reason?: Error): void
  abstract [EVENT.TRACK](event: RTCTrackEvent): void
  abstract [EVENT.DATA](data: any): void
  abstract [EVENT.IDENTITY.PROVIDE](payload: IdentityPayload): void
}
