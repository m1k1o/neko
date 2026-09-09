import { FileListItem, ScreenResolution } from './types'

export type WebSocketPayloads =
  | SignalProvidePayload
  | SignalOfferPayload
  | SignalAnswerPayload
  | SignalCandidatePayload
  | SignalRequestPayload
  | ChatPayload
  | ChatSendPayload
  | EmojiSendPayload
  | ScreenResolutionPayload
  | KeyboardMapPayload
  | KeyboardModifiersPayload
  | ClipboardSetPayload
  | BroadcastStatusPayload

/*
  SYSTEM MESSAGES/PAYLOADS
*/
export interface SystemInitPayload {
  session_id: string
  control_host: ControlHostPayload
  screen_size: ScreenResolution
  sessions: Record<string, SessionDataPayload>
  settings: SettingsPayload
  touch_events: boolean
  screencast_enabled: boolean
  webrtc: {
    videos: string[]
  }
}

export interface ControlHostPayload {
  id: string
  has_host: boolean
  host_id?: string
}

export interface SessionDataPayload {
  id: string
  profile: {
    name: string
    is_admin: boolean
  }
  state: {
    is_connected: boolean
  }
}

export interface SettingsPayload {
  implicit_hosting: boolean
  locked_logins: boolean
  locked_controls: boolean
  control_protection: boolean
  heartbeat_interval: number
  plugins?: Record<string, unknown>
}

export interface SystemAdminPayload {
  broadcast_status: BroadcastStatusPayload
}

export interface SystemSettingsPayload extends SettingsPayload {
  id: string
}

// system/disconnect
// system/error
export interface SystemMessagePayload {
  title: string
  message: string
}

/*
  SIGNAL MESSAGES/PAYLOADS
*/
export interface SignalProvidePayload {
  sdp: string
  iceservers: RTCIceServer[]
}

export interface SignalRequestPayload {
  /** Browser decoder capabilities, ordered by browser preference. */
  video_codecs?: string[]
  video?: {
    auto?: boolean
    disabled?: boolean
    selector?: Record<string, unknown>
  }
  audio?: {
    disabled?: boolean
  }
}

export interface SignalOfferPayload {
  sdp: string
}

export interface SignalAnswerPayload {
  sdp: string
}

export type SignalCandidatePayload = RTCIceCandidateInit

/*
  SESSION PAYLOADS
*/
export interface SessionIdPayload {
  id: string
}

export interface SessionProfilePayload {
  id: string
  name: string
  is_admin: boolean
}

export interface SessionStatePayload {
  id: string
  is_connected: boolean
  is_watching?: boolean
}

export interface SessionCursorsPayload {
  id: string
  cursors: Array<{ x: number; y: number }>
}

export interface ClipboardSetPayload {
  text: string
}

export interface KeyboardMapPayload {
  layout: string
  variant?: string
}

export interface KeyboardModifiersPayload {
  shift?: boolean
  capslock?: boolean
  control?: boolean
  alt?: boolean
  numlock?: boolean
  meta?: boolean
  super?: boolean
  altgr?: boolean
}

/*
  CHAT PAYLOADS
*/
export interface ChatSendPayload {
  content: string
}
export interface ChatPayload {
  id: string
  content: string
}

export interface EmotePayload {
  id: string
  emote: string
}

export interface EmojiSendPayload {
  emote: string
}

/*
  FILE TRANSFER PAYLOADS
*/
export interface FileTransferUpdatePayload {
  root_dir: string
  enabled: boolean
  user_download: boolean
  user_upload: boolean
  user_delete: boolean
  files: FileListItem[]
}

/*
  SCREEN PAYLOADS
*/
export interface ScreenResolutionPayload extends ScreenResolution {
  id?: string
}

/*
  BROADCAST PAYLOADS
*/
export interface BroadcastStatusPayload {
  url: string
  is_active: boolean
}

export type AdminLockResource = 'login' | 'control' | 'file_transfer'
