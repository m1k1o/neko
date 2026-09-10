import EventEmitter from 'eventemitter3'
import { EVENT, WebSocketEvents } from './events'
import { ConnectionStateMachine } from '~/sdk/connection-state'
import { MediaInput } from '~/sdk/media-protocol'
import { MediaSession } from '~/sdk/media-session'
import { SignalingMessage, SignalingTransport } from '~/sdk/signaling'

import {
  WebSocketPayloads,
  SignalProvidePayload,
  SignalCandidatePayload,
  SignalOfferPayload,
  SignalRequestPayload,
} from './messages'

export interface BaseEvents {
  info: (...message: any[]) => void
  warn: (...message: any[]) => void
  debug: (...message: any[]) => void
  error: (error: Error) => void
}

export abstract class BaseClient extends EventEmitter<BaseEvents> {
  protected readonly signaling: SignalingTransport
  private readonly mediaSession: MediaSession
  protected _ws_heartbeat?: number
  protected _control_heartbeat?: number
  protected _peer?: RTCPeerConnection
  protected _timeout?: number
  protected _state: RTCIceConnectionState = 'disconnected'
  protected _id = ''
  protected _candidates: RTCIceCandidateInit[] = []
  protected _controlEpoch = 0
  private readonly connectionMachine = new ConnectionStateMachine()
  private negotiationQueue: Promise<void> = Promise.resolve()
  private remoteDescriptionSet = false

  constructor() {
    super()
    this.mediaSession = new MediaSession({
      onData: (event) => this.onData(event),
      onError: (error) => this.onError(error),
      onDataChannelClosed: () => this.onDisconnected(new Error('peer data channel closed')),
    })
    this.signaling = new SignalingTransport({
      onMessage: (message) => this.onMessage(message),
      onOpen: () => this.onSignalingOpen(),
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

  get connectionState() {
    return this.connectionMachine.state
  }

  get connected() {
    return this.connectionMachine.state === 'connected'
  }

  public connect(url: string, token: string) {
    if (this.socketOpen) {
      this.emit('warn', `attempting to create websocket while connection open`)
      return
    }

    if (!this.supported) {
      this.onDisconnected(new Error('browser does not support webrtc (RTCPeerConnection missing)'))
      return
    }

    this.transitionConnection('connect')
    const signalingURL = new URL(url, window.location.href)
    if (token) {
      signalingURL.searchParams.set('token', token)
    }
    this.emit('debug', `connecting to ${signalingURL.origin}${signalingURL.pathname}`)
    this.signaling.connect(signalingURL.toString())
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
    if (this._control_heartbeat) {
      clearInterval(this._control_heartbeat)
      this._control_heartbeat = undefined
    }

    this.signaling.close()
    this.remoteDescriptionSet = false

    this.mediaSession.close()

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

    this.connectionMachine.reset()
    this._state = 'disconnected'
    this._id = ''
  }

  get microphoneActive() {
    return this.mediaSession.microphoneActive
  }

  public async enableMicrophone(): Promise<void> {
    if (this.mediaSession.microphoneActive) {
      this.emit('debug', 'microphone already active')
      return
    }

    try {
      await this.mediaSession.enableMicrophone()
      this.emit('info', 'microphone enabled')
    } catch (error) {
      const reason = error instanceof Error ? error : new Error(String(error))
      this.emit('error', reason)
      throw reason
    }
  }

  public disableMicrophone(): void {
    this.mediaSession.disableMicrophone()
    this.emit('info', 'microphone disabled')
  }

  public sendData(event: 'mousemove', data: { x: number; y: number; epoch?: number }): void
  public sendData(event: 'wheel', data: { x: number; y: number; controlKey?: boolean; epoch?: number }): void
  public sendData(event: 'mousedown' | 'mouseup' | 'keydown' | 'keyup', data: { key: number; epoch?: number }): void
  public sendData(event: MediaInput['event'], data: Record<string, number | boolean | undefined>) {
    if (!this.connected) {
      this.emit('warn', `attempting to send data while disconnected`)
      return
    }

    if (!this.mediaSession.dataChannelOpen) {
      this.emit('warn', `attempting to send data while data channel is not open`)
      return
    }

    try {
      this.mediaSession.sendData(event, { ...data, epoch: this._controlEpoch })
    } catch (error) {
      this.emit('error', error instanceof Error ? error : new Error(String(error)))
    }
  }

  public setControlEpoch(epoch: number) {
    if (!Number.isSafeInteger(epoch) || epoch < 0) {
      throw new RangeError('control epoch must be a non-negative safe integer')
    }
    this._controlEpoch = epoch
  }

  public sendMessage(event: WebSocketEvents, payload?: WebSocketPayloads) {
    if (!this.connected) {
      this.emit('warn', `attempting to send message while disconnected`)
      return
    }
    this.emit('debug', `sending event '${event}' ${payload ? `with payload: ` : ''}`, payload)
    const message: SignalingMessage = payload ? { event, payload } : { event }
    if (!this.signaling.send(message)) {
      this.emit('warn', `unable to send websocket event '${event}' while signaling socket is not open`)
    }
  }

  private onSignalingOpen() {
    this.emit('debug', 'signaling socket open; requesting media peer')
    const payload: SignalRequestPayload = {
      video_codecs: this.getSupportedVideoCodecs(),
      video: { auto: true },
      audio: {},
    }
    if (!this.signaling.send({ event: EVENT.SIGNAL.REQUEST, payload })) {
      this.emit('warn', 'unable to request media peer while signaling socket is not open')
    }
  }

  private getSupportedVideoCodecs(): string[] | undefined {
    if (typeof RTCRtpReceiver === 'undefined' || typeof RTCRtpReceiver.getCapabilities !== 'function') {
      return undefined
    }

    const capabilities = RTCRtpReceiver.getCapabilities('video')
    if (!capabilities?.codecs) {
      return undefined
    }

    const supported = new Set<string>()
    for (const codec of capabilities.codecs) {
      const [, name] = codec.mimeType.split('/')
      const normalized = name?.toLowerCase()
      if (normalized && ['av1', 'h264', 'h265', 'vp8', 'vp9'].includes(normalized)) {
        supported.add(normalized)
      }
    }

    return supported.size > 0 ? Array.from(supported) : undefined
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

    // Start gathering a small candidate pool before the remote SDP arrives.
    // BUNDLE and RTCP mux match the server offer and avoid negotiating
    // unnecessary transports during the first connection.
    const configuration: RTCConfiguration = {
      bundlePolicy: 'max-bundle',
      iceCandidatePoolSize: 1,
      rtcpMuxPolicy: 'require',
    }
    if (lite !== true) {
      configuration.iceServers = servers
    }
    this._peer = new RTCPeerConnection(configuration)
    this.mediaSession.attachPeer(this._peer)

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

      if (!this.signaling.send({ event: EVENT.SIGNAL.CANDIDATE, payload: init })) {
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
  }

  private attachDataChannel(channel: RTCDataChannel) {
    this.mediaSession.attachDataChannel(channel)
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

        if (!this.signaling.send({ event: EVENT.SIGNAL.OFFER, payload: { sdp: description.sdp } })) {
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
      await this.flushCandidates()

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
          payload: { sdp: this._peer.localDescription.sdp },
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
      await this.flushCandidates()
    } catch (error: any) {
      this.onError(error)
    }
  }

  private async onMessage(message: SignalingMessage) {
    const { event, payload = {} } = message

    this.emit('debug', `received websocket event ${event} ${payload ? `with payload: ` : ''}`, payload)

    if (event === EVENT.SIGNAL.PROVIDE) {
      const { sdp, iceservers } = payload as SignalProvidePayload
      await this.createPeer(iceservers.length === 0, iceservers)
      await this.setRemoteOffer(sdp)
      return
    }

    if (event === EVENT.SIGNAL.OFFER || event === EVENT.SIGNAL.RESTART) {
      const { sdp } = payload as SignalOfferPayload
      await this.setRemoteOffer(sdp)
      return
    }

    if (event === EVENT.SIGNAL.ANSWER) {
      const { sdp } = payload as SignalOfferPayload
      await this.setRemoteAnswer(sdp)
      return
    }

    if (event === EVENT.SIGNAL.CANDIDATE) {
      const candidate = payload as SignalCandidatePayload
      // Candidates can race the async SDP setter. Buffer them until the
      // remote description is installed instead of triggering a failed
      // addIceCandidate call and waiting for a reconnect.
      if (this._peer && this.remoteDescriptionSet) {
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

    if (event === EVENT.SIGNAL.CLOSE) {
      this.onDisconnected(new Error('media peer closed by server'))
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

  private async flushCandidates() {
    if (!this._peer || this._candidates.length === 0) {
      return
    }

    const candidates = this._candidates.splice(0)
    await Promise.all(candidates.map((candidate) => this._peer!.addIceCandidate(candidate)))
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

    if (!this.peerConnected || !this.socketOpen) {
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
