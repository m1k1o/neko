import EventEmitter from 'eventemitter3'

import { NekoWebSocket } from './websocket'
import { NekoWebRTC, WebRTCStats } from './webrtc'

export interface NekoConnectionEvents {
  connecting: () => void
  connected: () => void
  disconnected: (error?: Error) => void
}

export class NekoConnection extends EventEmitter<NekoConnectionEvents> {
  staysConnected = false

  websocket = new NekoWebSocket()
  webrtc = new NekoWebRTC()

  constructor() {
    super()

    //
    // websocket events
    //

    this.websocket.on('message', async (event: string, payload: any) => {
      
    })
    this.websocket.on('connecting', () => {
      
    })
    this.websocket.on('connected', () => {
      
    })
    this.websocket.on('disconnected', () => {
      
    })

    //
    // webrtc events
    //

    this.webrtc.on('track', (event: RTCTrackEvent) => {
      
    })
    this.webrtc.on('candidate', (candidate: RTCIceCandidateInit) => {
      
    })
    this.webrtc.on('stats', (stats: WebRTCStats) => {
      
    })
    this.webrtc.on('connecting', () => {
      
    })
    this.webrtc.on('connected', () => {
      
    })
    this.webrtc.on('disconnected', () => {

    })

  }

  public async connect(): Promise<void> {
    this.staysConnected = true
  }

  public async disconnect(): Promise<void> {
    this.staysConnected = false
  }
}
