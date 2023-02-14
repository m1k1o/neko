import * as webrtcTypes from './webrtc'
import * as reconnectorTypes from './reconnector'

export default interface State {
  authenticated: boolean
  connection: Connection
  video: Video
  control: Control
  screen: Screen
  session_id: string | null
  sessions: Record<string, Session>
  settings: Settings
  cursors: Cursors
  mobile_keyboard_open: boolean
}

/////////////////////////////
// Connection
/////////////////////////////

export interface Connection {
  url: string
  token?: string
  status: 'disconnected' | 'connecting' | 'connected'
  websocket: WebSocket
  webrtc: WebRTC
  screencast: boolean
  type: 'webrtc' | 'fallback' | 'none'
}

export interface WebSocket {
  connected: boolean
  config: ReconnectorConfig
}

export interface WebRTC {
  connected: boolean
  stable: boolean
  config: ReconnectorConfig
  stats: WebRTCStats | null
  video: string | null
  bitrate: number | null
  videos: string[]
}

export interface ReconnectorConfig extends reconnectorTypes.ReconnectorConfig {}

export interface WebRTCStats extends webrtcTypes.WebRTCStats {}

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
  locked: boolean
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
  sync: boolean
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
  can_share_media: boolean
  can_access_clipboard: boolean
  sends_inactive_cursor: boolean
  can_see_inactive_cursors: boolean
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

/////////////////////////////
// Settings
/////////////////////////////

export interface Settings {
  private_mode: boolean
  implicit_hosting: boolean
  inactive_cursors: boolean
  merciful_reconnect: boolean
}

/////////////////////////////
// Cursors
/////////////////////////////

type Cursors = SessionCursors[]

export interface SessionCursors {
  id: string
  cursors: Cursor[]
}

export interface Cursor {
  x: number
  y: number
}
