import { ICEServer } from '../internal/webrtc'

export interface Message {
  event: string | undefined
  payload: any
}

/////////////////////////////
// System
/////////////////////////////

export interface SystemInit {
  event: string | undefined
  session_id: string
  control_host: ControlHost
  screen_size: ScreenSize
  sessions: Record<string, SessionData>
  implicit_hosting: boolean
}

export interface SystemAdmin {
  event: string | undefined
  screen_sizes_list: ScreenSize[]
  broadcast_status: BroadcastStatus
}

export interface SystemDisconnect {
  event: string | undefined
  message: string
}

/////////////////////////////
// Signal
/////////////////////////////

export interface SignalProvide {
  event: string | undefined
  sdp: string
  iceservers: ICEServer[]
  video: string
  videos: string[]
}

export interface SignalCandidate extends RTCIceCandidateInit {
  event: string | undefined
}

export interface SignalAnswer {
  event: string | undefined
  sdp: string
}

export interface SignalVideo {
  event: string | undefined
  video: string
}

/////////////////////////////
// Session
/////////////////////////////

export interface SessionID {
  event: string | undefined
  id: string
}

export interface MemberProfile {
  event: string | undefined
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
  event: string | undefined
  id: string
  is_connected: boolean
  is_watching: boolean
}

export interface SessionData {
  event: string | undefined
  id: string
  profile: MemberProfile
  is_connected: boolean
  is_watching: boolean
}

/////////////////////////////
// Control
/////////////////////////////

export interface ControlHost {
  event: string | undefined
  has_host: boolean
  host_id: string | undefined
}

// TODO: New.
export interface ControlMove {
  event: string | undefined
  x: number
  y: number
}

// TODO: New.
export interface ControlScroll {
  event: string | undefined
  x: number
  y: number
}

// TODO: New.
export interface ControlKey {
  event: string | undefined
  key: number
}

/////////////////////////////
// Screen
/////////////////////////////

export interface ScreenSize {
  event: string | undefined
  width: number
  height: number
  rate: number
}

/////////////////////////////
// Clipboard
/////////////////////////////

export interface ClipboardData {
  event: string | undefined
  text: string
}

/////////////////////////////
// Keyboard
/////////////////////////////

export interface KeyboardModifiers {
  event: string | undefined
  caps_lock: boolean
  num_lock: boolean
  scroll_lock: boolean
}

export interface KeyboardMap {
  event: string | undefined
  layout: string
  variant: string
}

/////////////////////////////
// Broadcast
/////////////////////////////

export interface BroadcastStatus {
  event: string | undefined
  is_active: boolean
  url: string | undefined
}

/////////////////////////////
// Send
/////////////////////////////

export interface SendMessage {
  event: string | undefined
  sender: string
  subject: string
  body: string
}
