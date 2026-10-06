// Client state + wire types, mirroring server/pkg/types (master).

// data channel opcodes, big-endian (server/internal/webrtc/payload)
export const OP = {
  // client -> server
  MOVE: 1,
  SCROLL: 2,
  KEY_DOWN: 3,
  KEY_UP: 4,
  BTN_DOWN: 5,
  BTN_UP: 6,
  TOUCH_BEGIN: 8,
  TOUCH_UPDATE: 9,
  TOUCH_END: 10,
  // server -> client
  CURSOR_POSITION: 1,
  CURSOR_IMAGE: 2,
}

export interface MemberProfile {
  name: string
  avatar?: string
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
  connected_since?: string
  not_connected_since?: string
  is_watching: boolean
  watching_since?: string
  not_watching_since?: string
}

export interface Session {
  id: string
  profile: MemberProfile
  state: SessionState
}

export interface Settings {
  private_mode: boolean
  locked_logins: boolean
  locked_controls: boolean
  control_protection: boolean
  implicit_hosting: boolean
  inactive_cursors: boolean
  merciful_reconnect: boolean
  heartbeat_interval?: number
  plugins?: Record<string, any>
}

export interface ScreenSize {
  width: number
  height: number
  rate: number
}

export interface State {
  authenticated: boolean
  connection: {
    url: string
    token?: string
    status: 'disconnected' | 'connecting' | 'connected'
  }
  video: {
    playable: boolean
    playing: boolean
    volume: number
    muted: boolean
    // the browser refused to autoplay with sound, so playback started muted
    mutedByAutoplay: boolean
  }
  control: {
    host_id: string | null
    // local lock: stop sending input without giving up control
    locked: boolean
    clipboard: { text: string } | null
    scroll: { inverse: boolean; sensitivity: number }
    keyboard: { layout: string; variant: string }
    // server supports native touch events (otherwise touch emulates the mouse)
    touch: boolean
  }
  screen: {
    size: ScreenSize
    configurations: ScreenSize[]
  }
  session_id: string | null
  sessions: Record<string, Session>
  settings: Settings
  mobile_keyboard_open: boolean
}

export interface CursorImage {
  width: number
  height: number
  x: number
  y: number
  uri: string
}

// events emitted by NekoClient.events
export interface NekoEvents {
  'connection.status': (status: State['connection']['status']) => void
  'connection.closed': (error?: Error) => void
  'session.created': (id: string) => void
  'session.deleted': (id: string) => void
  'session.updated': (id: string) => void
  'room.control.host': (hasHost: boolean, hostId: string | undefined, by: string) => void
  'room.control.request': (id: string) => void
  'room.screen.updated': (width: number, height: number, rate: number, by: string) => void
  'room.settings.updated': (settings: Settings, by: string) => void
  'room.clipboard.updated': (text: string) => void
  'room.broadcast.status': (active: boolean, url?: string) => void
  'receive.unicast': (sender: string, subject: string, body: any) => void
  'receive.broadcast': (sender: string, subject: string, body: any) => void
  'upload.drop.progress': (p: { loaded: number; total: number }) => void
  'upload.drop.finished': (error?: Error) => void
  'overlay.click': (e: MouseEvent) => void
  // anything the core does not handle itself (plugins: chat/*, filetransfer/*, ...)
  message: (event: string, payload: any) => void
}
