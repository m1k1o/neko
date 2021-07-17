import { ICEServer } from '../internal/webrtc'

export interface Message {
  event?: string
  payload: any
}

/////////////////////////////
// System
/////////////////////////////

export interface SystemWebRTC {
  event?: string
  videos: string[]
}

export interface SystemInit {
  event?: string
  session_id: string
  control_host: ControlHost
  screen_size: ScreenSize
  sessions: Record<string, SessionData>
  implicit_hosting: boolean
  screencast_enabled: boolean
  webrtc: SystemWebRTC
}

export interface SystemAdmin {
  event?: string
  screen_sizes_list: ScreenSize[]
  broadcast_status: BroadcastStatus
}

export interface SystemDisconnect {
  event?: string
  message: string
}

/////////////////////////////
// Signal
/////////////////////////////

export interface SignalProvide {
  event?: string
  sdp: string
  iceservers: ICEServer[]
  video: string
}

export interface SignalCandidate extends RTCIceCandidateInit {
  event?: string
}

export interface SignalAnswer {
  event?: string
  sdp: string
}

export interface SignalVideo {
  event?: string
  video: string
}

/////////////////////////////
// Session
/////////////////////////////

export interface SessionID {
  event?: string
  id: string
}

export interface MemberProfile {
  event?: string
  id: string
  name: string
  is_admin: boolean
  can_login: boolean
  can_connect: boolean
  can_watch: boolean
  can_host: boolean
  can_access_clipboard: boolean
}

export interface SessionState {
  event?: string
  id: string
  is_connected: boolean
  is_watching: boolean
}

export interface SessionData {
  event?: string
  id: string
  profile: MemberProfile
  is_connected: boolean
  is_watching: boolean
}

/////////////////////////////
// Control
/////////////////////////////

export interface ControlHost {
  event?: string
  has_host: boolean
  host_id: string | undefined
}

// TODO: New.
export interface ControlMove {
  event?: string
  x: number
  y: number
}

// TODO: New.
export interface ControlScroll {
  event?: string
  x: number
  y: number
}

// TODO: New.
export interface ControlKey {
  event?: string
  key: number
}

/////////////////////////////
// Screen
/////////////////////////////

export interface ScreenSize {
  event?: string
  width: number
  height: number
  rate: number
}

/////////////////////////////
// Clipboard
/////////////////////////////

export interface ClipboardData {
  event?: string
  text: string
}

/////////////////////////////
// Keyboard
/////////////////////////////

export interface KeyboardModifiers {
  event?: string
  caps_lock: boolean
  num_lock: boolean
  scroll_lock: boolean
}

export interface KeyboardMap {
  event?: string
  layout: string
  variant: string
}

/////////////////////////////
// Broadcast
/////////////////////////////

export interface BroadcastStatus {
  event?: string
  is_active: boolean
  url: string | undefined
}

/////////////////////////////
// Send
/////////////////////////////

export interface SendMessage {
  event?: string
  sender: string
  subject: string
  body: string
}
