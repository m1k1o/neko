export default interface State {
  connection: Connection
  video: Video
  control: Control
  screen: Screen
  member_id: string | null
  members: Record<string, Member>
}

/////////////////////////////
// Connection
/////////////////////////////
export interface Connection {
  authenticated: boolean
  websocket: 'unavailable' | 'disconnected' | 'connecting' | 'connected'
  webrtc: WebRTC
  type: 'webrtc' | 'fallback' | 'none'
}

export interface WebRTC {
  status: 'unavailable' | 'disconnected' | 'connecting' | 'connected'
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
}

/////////////////////////////
// Video
/////////////////////////////
export interface Video {
  playable: boolean
  playing: boolean
  volume: number
  muted: boolean
  fullscreen: boolean
}

/////////////////////////////
// Control
/////////////////////////////
export interface Control {
  scroll: Scroll
  cursor: Cursor
  clipboard: Clipboard | null
  host_id: string | null
  implicit_hosting: boolean
}

export interface Scroll {
  inverse: boolean
  sensitivity: number
}

export interface Cursor {
  image: CursorImage | null
  position: CursorPosition | null
}

export interface CursorImage {
  uri: string
  width: number
  height: number
  x: number
  y: number
}

export interface CursorPosition {
  x: number
  y: number
}

export interface Clipboard {
  text: string
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
// Member
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

export interface MemberState {
  is_connected: boolean
  is_watching: boolean
}

export interface Member {
  id: string
  profile: MemberProfile
  state: MemberState
}
