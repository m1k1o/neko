import EventEmitter from 'eventemitter3'
import { EVENT, WebSocketEvents } from './events'
import { ConnectionStateMachine } from '~/sdk/connection-state'
import { encodeMediaInput, MediaInput } from '~/sdk/media-protocol'
import { SignalingMessage, SignalingTransport } from '~/sdk/signaling'

import {
  WebSocketMessages,
  WebSocketPayloads,
  SignalProvidePayload,
  SignalCandidatePayload,
  SignalOfferPayload,
  SignalAnswerMessage,
} from './messages'

export interface BaseEvents {
  info: (...message: any[]) => void
  warn: (...message: any[]) => void
  debug: (...message: any[]) => void
  error: (error: Error) => void
}

export abstract class BaseClient extends EventEmitter<BaseEvents> {
  protected readonly signaling: SignalingTransport
  protected _ws_heartbeat?: number
  protected _peer?: RTCPeerConnection
  protected _channel?: RTCDataChannel
  protected _timeout?: number
  protected _displayname?: string
  protected _state: RTCIceConnectionState = 'disconnected'
  protected _id = ''
  protected _candidates: RTCIceCandidate[] = []
  protected _micStream?: MediaStream
  protected _micSender?: RTCRtpSender
  protected _micActive = false
  private readonly connectionMachine = new ConnectionStateMachine()
  private negotiationQueue: Promise<void> = Promise.resolve()
  private remoteDescriptionSet = false

  constructor() {
    super()
    this.signaling = new SignalingTransport({
      onMessage: (message) => this.onMessage(message),
      onError: (error) => this.onError(error),
      onClose: () => this.onDisconnected(new Error('websocket closed')),
    })
  }

  get id() {
    return this._id
  }

  get supported() {
    return typeof RTCPeerConnection !== 'undefined' && typeof RTCPeerConnection.prototype.addTransceiver !== 'undefined'
  }

  get socketOpen() {
    return this.signaling.open
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

    this._displayname = displayname
    this.transitionConnection('connect')
    const signalingURL = `${url}?password=${encodeURIComponent(password)}&username=${encodeURIComponent(displayname)}`
    this.emit('debug', `connecting to ${signalingURL}`)
    this.signaling.connect(signalingURL)
    this._timeout = window.setTimeout(this.onTimeout.bind(this), 15000)
  }

  protected disconnect() {
    if (this._timeout) {
      clearTimeout(this._timeout)
      this._timeout = undefined
    }

    if (this._ws_heartbeat) {
      clearInterval(this._ws_heartbeat)
      this._ws_heartbeat = undefined
    }

    this.signaling.close()
    this.remoteDescriptionSet = false

    if (this._channel) {
      // reset all events
      this._channel.onmessage = () => {}
      this._channel.onerror = () => {}
      this._channel.onclose = () => {}

      try {
        this._channel.close()
      } catch (err) {}

      this._channel = undefined
    }

    if (this._peer) {
      // reset all events
      this._peer.onconnectionstatechange = () => {}
      this._peer.onsignalingstatechange = () => {}
      this._peer.oniceconnectionstatechange = () => {}
      this._peer.ontrack = () => {}

      try {
        this._peer.close()
      } catch (err) {}

      this._peer = undefined
    }

    this.disableMicrophone()

    this._state = 'disconnected'
    this._displayname = undefined
    this._id = ''
  }

  get microphoneActive() {
    return this._micActive
  }

  public async enableMicrophone(): Promise<void> {
    if (!this._peer) {
      this.emit('warn', 'attempting to enable microphone with no peer connection')
      return
    }

    if (this._micActive) {
      this.emit('debug', 'microphone already active')
      return
    }

    try {
      this._micStream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const audioTrack = this._micStream.getAudioTracks()[0]
      this._micSender = this._peer.addTrack(audioTrack, this._micStream)
      this._micActive = true
      this.emit('info', `microphone enabled: ${audioTrack.label}`)
    } catch (err: any) {
      this.emit('error', err)
      throw err
    }
  }

  public disableMicrophone(): void {
    if (this._micSender && this._peer) {
      try {
        this._peer.removeTrack(this._micSender)
      } catch (err) {
        this.emit('warn', 'failed to remove mic track from peer', err)
      }
      this._micSender = undefined
    }

    if (this._micStream) {
      this._micStream.getTracks().forEach((t) => t.stop())
      this._micStream = undefined
    }

    this._micActive = false
    this.emit('info', 'microphone disabled')
  }

  public sendData(event: MediaInput['event'], data: Omit<MediaInput, 'event'>) {
    if (!this.connected) {
      this.emit('warn', `attempting to send data while disconnected`)
      return
    }

    if (!this._channel || this._channel.readyState !== 'open') {
      this.emit('warn', `attempting to send data while data channel is not open`)
      return
    }

    try {
      this._channel.send(encodeMediaInput({ event, ...data } as MediaInput))
    } catch (error: any) {
      this.emit('error', error instanceof Error ? error : new Error(String(error)))
    }
  }

  public sendMessage(event: WebSocketEvents, payload?: WebSocketPayloads) {
    if (!this.connected) {
      this.emit('warn', `attempting to send message while disconnected`)
      return
    }
    this.emit('debug', `sending event '${event}' ${payload ? `with payload: ` : ''}`, payload)
    if (!this.signaling.send({ event, ...payload } as SignalingMessage)) {
      this.emit('warn', `unable to send websocket event '${event}' while signaling socket is not open`)
    }
  }

  public async createPeer(lite: boolean, servers: RTCIceServer[]) {
    this.emit('debug', `creating peer`)
    if (!this.socketOpen) {
      this.emit('warn', `attempting to create peer with no websocket: `, `state: ${this.signaling.state}`)
      return
    }

    if (this.peerConnected) {
      this.emit('warn', `attempting to create peer while connected`)
      return
    }

    if (lite !== true) {
      this._peer = new RTCPeerConnection({
        iceServers: servers,
      })
    } else {
      this._peer = new RTCPeerConnection()
    }

    this._peer.onconnectionstatechange = () => {
      this.emit('debug', `peer connection state changed`, this._peer ? this._peer.connectionState : undefined)
    }

    this._peer.onsignalingstatechange = () => {
      this.emit('debug', `peer signaling state changed`, this._peer ? this._peer.signalingState : undefined)
    }

    this._peer.oniceconnectionstatechange = () => {
      this._state = this._peer!.iceConnectionState

      this.emit('debug', `peer ice connection state changed: ${this._peer!.iceConnectionState}`)

      switch (this._state) {
        case 'checking':
          if (this._timeout) {
            clearTimeout(this._timeout)
            this._timeout = undefined
          }
          break
        case 'connected':
          this.onConnected()
          break
        case 'disconnected':
          this.transitionConnection('reconnect')
          break
        // https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Signaling_and_video_calling#ice_connection_state
        // We don't watch the disconnected signaling state here as it can indicate temporary issues and may
        // go back to a connected state after some time. Watching it would close the video call on any temporary
        // network issue.
        case 'failed':
          this.onDisconnected(new Error('peer failed'))
          break
        case 'closed':
          this.onDisconnected(new Error('peer closed'))
          break
      }
    }

    this._peer.ontrack = this.onTrack.bind(this)

    this._peer.onicecandidate = (event: RTCPeerConnectionIceEvent) => {
      if (!event.candidate) {
        this.emit('debug', `sent all local ICE candidates`)
        return
      }

      const init = event.candidate.toJSON()
      this.emit('debug', `sending local ICE candidate`, init)

      if (!this.signaling.send({ event: EVENT.SIGNAL.CANDIDATE, data: JSON.stringify(init) })) {
        this.emit('warn', 'unable to send local ICE candidate while signaling socket is not open')
      }
    }

    this._peer.ondatachannel = (event: RTCDataChannelEvent) => {
      this.emit('debug', `received data channel '${event.channel.label}'`)
      this.attachDataChannel(event.channel)
    }

    this._peer.onnegotiationneeded = () => {
      this.queueNegotiation()
    }

    // Keep a client-created channel as a compatibility fallback for legacy
    // servers. Modern servers negotiate their data channel in the offer and
    // replace this channel through ondatachannel above.
    this.attachDataChannel(this._peer.createDataChannel('data'))
  }

  private attachDataChannel(channel: RTCDataChannel) {
    if (this._channel && this._channel !== channel) {
      const previous = this._channel
      previous.onopen = null
      previous.onmessage = null
      previous.onerror = null
      previous.onclose = null
      try {
        previous.close()
      } catch (error) {
        this.emit('debug', 'failed to close compatibility data channel', error)
      }
    }

    this._channel = channel
    channel.binaryType = 'arraybuffer'
    channel.onerror = this.onError.bind(this)
    channel.onmessage = this.onData.bind(this)
    channel.onclose = () => {
      if (this._channel === channel) {
        this.onDisconnected(new Error('peer data channel closed'))
      }
    }
  }

  private queueNegotiation() {
    this.negotiationQueue = this.negotiationQueue
      .then(async () => {
        if (
          !this._peer ||
          !this.remoteDescriptionSet ||
          !this.signaling.open ||
          this._peer.signalingState !== 'stable'
        ) {
          return
        }

        this.emit('debug', 'creating renegotiation offer')
        const offer = await this._peer.createOffer()
        await this._peer.setLocalDescription(offer)
        const description = this._peer.localDescription
        if (!description?.sdp) {
          throw new Error('renegotiation produced an empty SDP offer')
        }

        if (!this.signaling.send({ event: EVENT.SIGNAL.OFFER, sdp: description.sdp })) {
          throw new Error('unable to send renegotiation offer while signaling socket is not open')
        }
      })
      .catch((error) => this.onError(error instanceof Error ? error : new Error(String(error))))
  }

  public async setRemoteOffer(sdp: string) {
    if (!this._peer) {
      this.emit('warn', `attempting to set remote offer while disconnected`)
      return
    }

    try {
      await this._peer.setRemoteDescription({ type: 'offer', sdp })
      this.remoteDescriptionSet = true

      for (const candidate of this._candidates) {
        await this._peer.addIceCandidate(candidate)
      }
      this._candidates = []

      const d = await this._peer.createAnswer()

      // add stereo=1 to answer sdp to enable stereo audio for chromium
      d.sdp = d.sdp?.replace(/(stereo=1;)?useinbandfec=1/, 'useinbandfec=1;stereo=1')

      await this._peer.setLocalDescription(d)
      if (!this._peer.localDescription?.sdp) {
        throw new Error('remote offer produced an empty SDP answer')
      }

      if (
        !this.signaling.send({
          event: EVENT.SIGNAL.ANSWER,
          sdp: this._peer.localDescription.sdp,
          displayname: this._displayname,
        })
      ) {
        throw new Error('unable to send SDP answer while signaling socket is not open')
      }
    } catch (err: any) {
      this.onError(err)
    }
  }

  public async setRemoteAnswer(sdp: string) {
    if (!this._peer) {
      this.emit('warn', `attempting to set remote answer while disconnected`)
      return
    }

    try {
      await this._peer.setRemoteDescription({ type: 'answer', sdp })
      this.remoteDescriptionSet = true
    } catch (error: any) {
      this.onError(error)
    }
  }

  private async onMessage(message: SignalingMessage) {
    const { event, ...payload } = message as WebSocketMessages

    this.emit('debug', `received websocket event ${event} ${payload ? `with payload: ` : ''}`, payload)

    if (event === EVENT.SIGNAL.PROVIDE) {
      const { sdp, lite, ice, id } = payload as SignalProvidePayload
      this._id = id
      await this.createPeer(lite, ice)
      await this.setRemoteOffer(sdp)
      return
    }

    if (event === EVENT.SIGNAL.OFFER) {
      const { sdp } = payload as SignalOfferPayload
      await this.setRemoteOffer(sdp)
      return
    }

    if (event === EVENT.SIGNAL.ANSWER) {
      const { sdp } = payload as SignalAnswerMessage
      await this.setRemoteAnswer(sdp)
      return
    }

    if (event === EVENT.SIGNAL.CANDIDATE) {
      const { data } = payload as SignalCandidatePayload
      const candidate: RTCIceCandidate = JSON.parse(data)
      if (this._peer) {
        try {
          await this._peer.addIceCandidate(candidate)
        } catch (error: any) {
          this.onError(error)
        }
      } else {
        this._candidates.push(candidate)
      }
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

  private onError(error: Error | Event) {
    if (error instanceof Error) {
      this.emit('error', error)
      return
    }

    const eventError = (error as ErrorEvent).error
    this.emit('error', eventError instanceof Error ? eventError : new Error('WebRTC or signaling error'))
  }

  private onConnected() {
    if (this._timeout) {
      clearTimeout(this._timeout)
      this._timeout = undefined
    }

    if (!this.connected) {
      this.emit('warn', `onConnected called while being disconnected`)
      return
    }

    const transition = this.connectionMachine.transition('connected')
    if (!transition.changed) {
      return
    }

    this.emit('debug', `connected`)
    this[EVENT.CONNECTED]()
  }

  private onTimeout() {
    this.emit('debug', `connection timeout`)
    if (this._timeout) {
      clearTimeout(this._timeout)
      this._timeout = undefined
    }
    this.onDisconnected(new Error('connection timeout'))
  }

  protected onDisconnected(reason?: Error) {
    const transition = this.connectionMachine.transition('disconnect')
    this.disconnect()
    this.emit('debug', `disconnected:`, reason)
    if (transition.changed) {
      this[EVENT.DISCONNECTED](reason)
    }
  }

  private transitionConnection(event: 'connect' | 'reconnect') {
    const transition = this.connectionMachine.transition(event)
    if (!transition.changed) return

    if (event === 'connect') {
      this[EVENT.CONNECTING]()
    } else {
      this[EVENT.RECONNECTING]()
    }
  }

  protected [EVENT.MESSAGE](event: string, payload: any) {
    this.emit('warn', `unhandled websocket event '${event}':`, payload)
  }

  protected abstract [EVENT.RECONNECTING](): void
  protected abstract [EVENT.CONNECTING](): void
  protected abstract [EVENT.CONNECTED](): void
  protected abstract [EVENT.DISCONNECTED](reason?: Error): void
  protected abstract [EVENT.TRACK](event: RTCTrackEvent): void
  protected abstract [EVENT.DATA](data: any): void
}
