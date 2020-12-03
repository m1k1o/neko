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
  websocket: 'disconnected' | 'connecting' | 'connected'
  webrtc: 'disconnected' | 'connecting' | 'connected'
  type: 'webrtc' | 'fallback' | 'none'
  can_watch: boolean
  can_control: boolean
  clipboard_access: boolean
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
  clipboard: Clipboard | null
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
  is_receiving: boolean
}

export interface Member {
  id: string
  profile: MemberProfile
  state: MemberState
}
