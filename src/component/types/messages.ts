import type { ICEServer } from '../internal/webrtc'
import type { Settings, ScreenSize } from './state'
import type { PeerRequest, PeerVideo, PeerAudio } from './webrtc'

/////////////////////////////
// System
/////////////////////////////

export interface SystemSettingsUpdate extends Settings {
  id: string
}

export interface SystemWebRTC {
  videos: string[]
}

export interface SystemInit {
  session_id: string
  control_host: ControlHost
  screen_size: ScreenSize
  sessions: Record<string, SessionData>
  settings: Settings
  touch_events: boolean
  screencast_enabled: boolean
  webrtc: SystemWebRTC
}

export interface SystemAdmin {
  screen_sizes_list: ScreenSize[]
  broadcast_status: BroadcastStatus
}

export type SystemLogs = SystemLog[]

export interface SystemLog {
  level: 'debug' | 'info' | 'warn' | 'error'
  fields: Record<string, string>
  message: string
}

export interface SystemDisconnect {
  message: string
}

/////////////////////////////
// Signal
/////////////////////////////

export type SignalRequest = PeerRequest

export interface SignalProvide {
  sdp: string
  iceservers: ICEServer[]
  video: PeerVideo
  audio: PeerAudio
}

export type SignalCandidate = RTCIceCandidateInit

export interface SignalDescription {
  sdp: string
}

export type SignalVideo = PeerVideo

export type SignalAudio = PeerAudio

/////////////////////////////
// Session
/////////////////////////////

export interface SessionID {
  id: string
}

export interface MemberProfile {
  id: string
  name: string
  is_admin: boolean
  can_login: boolean
  can_connect: boolean
  can_watch: boolean
  can_host: boolean
  can_share_media: boolean
  can_access_clipboard: boolean
  sends_inactive_cursor: boolean
  can_see_inactive_cursors: boolean
}

export interface SessionState {
  id: string
  is_connected: boolean
  is_watching: boolean
}

export interface SessionData {
  id: string
  profile: MemberProfile
  is_connected: boolean
  is_watching: boolean
}

export interface SessionCursor {
  id: string
  x: number
  y: number
}

/////////////////////////////
// Control
/////////////////////////////

export interface ControlHost {
  id: string
  has_host: boolean
  host_id: string | undefined
}

export interface ControlScroll {
  delta_x: number
  delta_y: number
  control_key: boolean
}

export interface ControlPos {
  x: number
  y: number
}

export interface ControlButton extends Partial<ControlPos> {
  code: number
}

export interface ControlKey extends Partial<ControlPos> {
  keysym: number
}

export interface ControlTouch extends Partial<ControlPos> {
  touch_id: number
  pressure: number
}

/////////////////////////////
// Screen
/////////////////////////////

export interface ScreenSizeUpdate extends ScreenSize {
  id: string
}

/////////////////////////////
// Clipboard
/////////////////////////////

export interface ClipboardData {
  text: string
}

/////////////////////////////
// Keyboard
/////////////////////////////

export interface KeyboardModifiers {
  caps_lock: boolean
  num_lock: boolean
  scroll_lock: boolean
}

export interface KeyboardMap {
  layout: string
  variant: string
}

/////////////////////////////
// Broadcast
/////////////////////////////

export interface BroadcastStatus {
  is_active: boolean
  url: string | undefined
}

/////////////////////////////
// Send
/////////////////////////////

export interface SendMessage {
  sender: string
  subject: string
  body: string
}
