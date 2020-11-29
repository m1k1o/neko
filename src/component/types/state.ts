export default interface State {
  connection: Connection
  video: Video
  control: Control
  screen: Screen
  member: Member
  members: Member[]
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
  clipboard: Clipboard
  host: Member | null
}

export interface Scroll {
  inverse: boolean
  sensitivity: number
}

export interface Clipboard {
  data: string | null
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
export interface Member {
  id: string | null
  name: string | null
  is_admin: boolean
  is_watching: boolean
  is_controlling: boolean
  can_watch: boolean
  can_control: boolean
  clipboard_access: boolean
}
