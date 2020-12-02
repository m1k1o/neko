export interface Message {
  event: string | undefined
  payload: any
}

/////////////////////////////
// System
/////////////////////////////

export interface SystemInit {
  event: string | undefined
  member_id: string
  control_host: ControlHost
  screen_size: ScreenSize
  members: Record<string, MemberData>
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
  lite: boolean
  ice: string[]
}

export interface SignalAnswer {
  event: string | undefined
  sdp: string
}

/////////////////////////////
// Member
/////////////////////////////

// TODO: New.
export interface MemberID {
  event: string | undefined
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
  can_access_clipboard: boolean
}

export interface MemberData {
  event: string | undefined
  id: string
  profile: MemberProfile
  is_connected: boolean
  is_receiving: boolean
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

export interface KeyboardLayout {
  event: string | undefined
  layout: string
}

/////////////////////////////
// Broadcast
/////////////////////////////

export interface BroadcastStatus {
  event: string | undefined
  is_active: boolean
  url: string | undefined
}
