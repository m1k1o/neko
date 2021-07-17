export default interface State {
  authenticated: boolean
  connection: Connection
  video: Video
  control: Control
  screen: Screen
  session_id: string | null
  sessions: Record<string, Session>
}

/////////////////////////////
// Connection
/////////////////////////////

export interface Connection {
  status: 'disconnected' | 'connecting' | 'connected'
  webrtc: WebRTC
  screencast: boolean
  type: 'webrtc' | 'screencast' | 'none'
}

export interface WebRTC {
  stats: WebRTCStats | null
  video: string | null
  videos: string[]
  auto: boolean
}

export interface WebRTCStats {
  bitrate: number
  packetLoss: number
  fps: number
  width: number
  height: number
  muted?: boolean
}

/////////////////////////////
// Video
/////////////////////////////

export interface Video {
  playable: boolean
  playing: boolean
  volume: number
  muted: boolean
}

/////////////////////////////
// Control
/////////////////////////////

export interface Control {
  scroll: Scroll
  clipboard: Clipboard | null
  keyboard: Keyboard
  host_id: string | null
  implicit_hosting: boolean
}

export interface Scroll {
  inverse: boolean
  sensitivity: number
}

export interface Clipboard {
  text: string
}

export interface Keyboard {
  layout: string
  variant: string
}

/////////////////////////////
// Screen
/////////////////////////////

export interface Screen {
  size: ScreenSize
  configurations: ScreenSize[]
}

export interface ScreenSize {
  width: number
  height: number
  rate: number
}

/////////////////////////////
// Session
/////////////////////////////

export interface MemberProfile {
  name: string
  is_admin: boolean
  can_login: boolean
  can_connect: boolean
  can_watch: boolean
  can_host: boolean
  can_access_clipboard: boolean
}

export interface SessionState {
  is_connected: boolean
  is_watching: boolean
}

export interface Session {
  id: string
  profile: MemberProfile
  state: SessionState
}
