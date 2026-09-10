import { encodeMediaInput, MediaInput, MediaInputEvent } from './media-protocol'

export type MediaInputData = Record<string, number | boolean | undefined>

export interface MediaSessionOptions {
  onData?: (event: MessageEvent) => void
  onError?: (error: Error | Event) => void
  onTrack?: (event: RTCTrackEvent) => void
  onDataChannelClosed?: () => void
}

/**
 * Owns browser media primitives that are independent from signaling or UI:
 * the negotiated data channel, microphone sender, and remote media callbacks.
 */
export class MediaSession {
  private readonly options: MediaSessionOptions
  private peer?: RTCPeerConnection
  private channel?: RTCDataChannel
  private micStream?: MediaStream
  private micSender?: RTCRtpSender

  constructor(options: MediaSessionOptions = {}) {
    this.options = options
  }

  get dataChannelOpen() {
    return this.channel?.readyState === 'open'
  }

  get microphoneActive() {
    return this.micStream !== undefined
  }

  attachPeer(peer: RTCPeerConnection) {
    this.peer = peer
  }

  attachDataChannel(channel: RTCDataChannel) {
    if (this.channel && this.channel !== channel) {
      this.detachDataChannel(this.channel)
    }

    this.channel = channel
    channel.binaryType = 'arraybuffer'
    channel.onerror = (event) => this.options.onError?.(event)
    channel.onmessage = (event) => this.options.onData?.(event)
    channel.onclose = () => {
      if (this.channel === channel) {
        this.options.onDataChannelClosed?.()
      }
    }
  }

  sendData(event: MediaInputEvent, data: MediaInputData) {
    if (!this.channel || this.channel.readyState !== 'open') {
      return false
    }

    this.channel.send(encodeMediaInput({ event, ...data } as MediaInput))
    return true
  }

  async enableMicrophone() {
    if (!this.peer) {
      throw new Error('cannot enable microphone without a media peer')
    }
    if (this.microphoneActive) {
      return
    }

    this.micStream = await navigator.mediaDevices.getUserMedia({ audio: true })
    const audioTrack = this.micStream.getAudioTracks()[0]
    if (!audioTrack) {
      this.micStream.getTracks().forEach((track) => track.stop())
      this.micStream = undefined
      throw new Error('microphone stream has no audio track')
    }

    try {
      this.micSender = this.peer.addTrack(audioTrack, this.micStream)
    } catch (error) {
      this.micStream.getTracks().forEach((track) => track.stop())
      this.micStream = undefined
      throw error
    }
  }

  disableMicrophone() {
    if (this.micSender && this.peer) {
      try {
        this.peer.removeTrack(this.micSender)
      } catch (error) {
        this.options.onError?.(error instanceof Error ? error : new Error(String(error)))
      }
      this.micSender = undefined
    }

    if (this.micStream) {
      this.micStream.getTracks().forEach((track) => track.stop())
      this.micStream = undefined
    }
  }

  close() {
    this.detachDataChannel(this.channel)
    this.channel = undefined
    this.disableMicrophone()
    this.peer = undefined
  }

  private detachDataChannel(channel?: RTCDataChannel) {
    if (!channel) {
      return
    }

    channel.onmessage = null
    channel.onerror = null
    channel.onclose = null
    try {
      channel.close()
    } catch (error) {
      this.options.onError?.(error instanceof Error ? error : new Error(String(error)))
    }
  }
}
