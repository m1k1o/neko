export type StreamSelectorType = 'exact' | 'nearest' | 'lower' | 'higher'

export interface StreamSelector {
  type: StreamSelectorType
  id?: string
  bitrate?: number
}

export interface PeerRequest {
  video?: PeerVideoRequest
  audio?: PeerAudioRequest
}

export interface PeerVideo {
  disabled: boolean
  id: string
  auto: boolean
}

export interface PeerVideoRequest {
  disabled?: boolean
  selector?: StreamSelector
  auto?: boolean
}

export interface PeerAudio {
  disabled: boolean
}

export interface PeerAudioRequest {
  disabled?: boolean
}

export interface WebRTCStats {
  paused: boolean
  bitrate: number
  packetLoss: number
  fps: number
  width: number
  height: number
  muted?: boolean
  latency: number
  requestLatency: number
  responseLatency: number
}

export interface CursorPosition {
  x: number
  y: number
}

export interface CursorImage {
  width: number
  height: number
  x: number
  y: number
  uri: string
}
