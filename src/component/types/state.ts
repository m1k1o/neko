import type * as webrtcTypes from './webrtc'
import type * as reconnectorTypes from './reconnector'

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
  video: PeerVideo
  audio: PeerAudio
  videos: string[]
}

export interface ReconnectorConfig extends reconnectorTypes.ReconnectorConfig {}

export interface WebRTCStats extends webrtcTypes.WebRTCStats {}

export interface PeerVideo extends webrtcTypes.PeerVideo {}

export interface PeerAudio extends webrtcTypes.PeerAudio {}

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
  touch: Touch
  host_id: string | null
  is_host: boolean
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

export interface Touch {
  enabled: boolean
  supported: boolean
}

/////////////////////////////
// Screen
/////////////////////////////

export interface Screen {
  size: ScreenSize
  configurations: ScreenSize[]
  sync: ScreenSync
}

export interface ScreenSize {
  width: number
  height: number
  rate: number
}

export interface ScreenSync {
  enabled: boolean
  multiplier: number
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
  plugins?: Record<string, any>
}

export interface SessionState {
  is_connected: boolean
  connected_since?: Date
  not_connected_since?: Date
  is_watching: boolean
  watching_since?: Date
  not_watching_since?: Date
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
  locked_logins: boolean
  locked_controls: boolean
  control_protection: boolean
  implicit_hosting: boolean
  inactive_cursors: boolean
  merciful_reconnect: boolean
  plugins?: Record<string, any>
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
