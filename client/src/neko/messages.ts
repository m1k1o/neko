import {
  EVENT,
  WebSocketEvents,
  SystemEvents,
  ControlEvents,
  MemberEvents,
  SignalEvents,
  ChatEvents,
  ScreenEvents,
  AdminEvents,
  FileTransferEvents,
} from './events'
import { FileListItem, Member, ScreenConfigurations, ScreenResolution } from './types'

export type WebSocketMessages =
  | WebSocketMessage
  | SignalProvideMessage
  | SignalOfferMessage
  | SignalAnswerMessage
  | SignalCandidateMessage
  | MemberListMessage
  | MemberConnectMessage
  | MemberDisconnectMessage
  | ControlMessage
  | ScreenResolutionMessage
  | ScreenConfigurationsMessage
  | ChatMessage

export type WebSocketPayloads =
  | SignalProvidePayload
  | SignalOfferPayload
  | SignalAnswerPayload
  | SignalCandidatePayload
  | MemberListPayload
  | Member
  | ControlPayload
  | ControlClipboardPayload
  | ControlKeyboardPayload
  | ChatPayload
  | ChatSendPayload
  | EmojiSendPayload
  | ScreenResolutionPayload
  | ScreenConfigurationsPayload
  | AdminPayload
  | AdminLockPayload
  | BroadcastStatusPayload
  | BroadcastCreatePayload

export interface WebSocketMessage {
  event: WebSocketEvents | string
}

/*
  SYSTEM MESSAGES/PAYLOADS
*/
// system/init
export interface SystemInit extends WebSocketMessage, SystemInitPayload {
  event: typeof EVENT.SYSTEM.INIT
}
export interface SystemInitPayload {
  implicit_hosting: boolean
  locks: Record<string, string>
  file_transfer: boolean
  heartbeat_interval: number
}

// system/disconnect
// system/error
export interface SystemMessage extends WebSocketMessage, SystemMessagePayload {
  event: typeof EVENT.SYSTEM.DISCONNECT | typeof EVENT.SYSTEM.ERROR
}
export interface SystemMessagePayload {
  title: string
  message: string
}

/*
  SIGNAL MESSAGES/PAYLOADS
*/
// signal/provide
export interface SignalProvideMessage extends WebSocketMessage, SignalProvidePayload {
  event: typeof EVENT.SIGNAL.PROVIDE
}
export interface SignalProvidePayload {
  id: string
  lite: boolean
  ice: RTCIceServer[]
  sdp: string
}

// signal/offer
export interface SignalOfferMessage extends WebSocketMessage, SignalOfferPayload {
  event: typeof EVENT.SIGNAL.OFFER
}
export interface SignalOfferPayload {
  sdp: string
}

// signal/answer
export interface SignalAnswerMessage extends WebSocketMessage, SignalAnswerPayload {
  event: typeof EVENT.SIGNAL.ANSWER
}
export interface SignalAnswerPayload {
  sdp: string
  displayname: string
}

// signal/candidate
export interface SignalCandidateMessage extends WebSocketMessage, SignalCandidatePayload {
  event: typeof EVENT.SIGNAL.CANDIDATE
}
export interface SignalCandidatePayload {
  data: string
}

/*
  MEMBER MESSAGES/PAYLOADS
*/
// member/list
export interface MemberListMessage extends WebSocketMessage, MemberListPayload {
  event: typeof EVENT.MEMBER.LIST
}
export interface MemberListPayload {
  members: Member[]
}

// member/connected
export interface MemberConnectMessage extends WebSocketMessage, MemberPayload {
  event: typeof EVENT.MEMBER.CONNECTED
}
export type MemberPayload = Member

// member/disconnected
export interface MemberDisconnectMessage extends WebSocketMessage, MemberPayload {
  event: typeof EVENT.MEMBER.DISCONNECTED
}
export interface MemberDisconnectPayload {
  id: string
}

/*
  CONTROL MESSAGES/PAYLOADS
*/
// control/locked & control/release & control/request
export interface ControlMessage extends WebSocketMessage, ControlPayload {
  event: ControlEvents
}
export interface ControlPayload {
  id: string
}

export interface ControlTargetPayload {
  id: string
  target: string
}

export interface ControlClipboardPayload {
  text: string
}

export interface ControlKeyboardPayload {
  layout?: string
  capsLock?: boolean
  numLock?: boolean
  scrollLock?: boolean
}

/*
  CHAT PAYLOADS
*/
// chat/message
export interface ChatMessage extends WebSocketMessage, ChatPayload {
  event: typeof EVENT.CHAT.MESSAGE
}

export interface ChatSendPayload {
  content: string
}
export interface ChatPayload {
  id: string
  content: string
}

// chat/emoji
export interface ChatEmoteMessage extends WebSocketMessage, EmotePayload {
  event: typeof EVENT.CHAT.EMOTE
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
export interface FileTransferListMessage extends WebSocketMessage, FileTransferListPayload {
  event: FileTransferEvents
}

export interface FileTransferListPayload {
  cwd: string
  files: FileListItem[]
}

/*
  SCREEN PAYLOADS
*/
export interface ScreenResolutionMessage extends WebSocketMessage, ScreenResolutionPayload {
  event: ScreenEvents
}

export interface ScreenResolutionPayload extends ScreenResolution {
  id?: string
}

export interface ScreenConfigurationsMessage extends WebSocketMessage, ScreenConfigurationsPayload {
  event: ScreenEvents
}

export interface ScreenConfigurationsPayload {
  configurations: ScreenConfigurations
}

/*
  BROADCAST PAYLOADS
*/
export interface BroadcastCreatePayload {
  url: string
}

export interface BroadcastStatusPayload {
  url: string
  isActive: boolean
}

/*
  ADMIN PAYLOADS
*/
export interface AdminMessage extends WebSocketMessage, AdminPayload {
  event: AdminEvents
}

export interface AdminPayload {
  id: string
}

export interface AdminTargetMessage extends WebSocketMessage, AdminTargetPayload {
  event: AdminEvents
}

export interface AdminTargetPayload {
  id: string
  target?: string
}

export interface AdminLockMessage extends WebSocketMessage, AdminLockPayload {
  event: AdminEvents
  id: string
}

export type AdminLockResource = 'login' | 'control' | 'file_transfer'

export interface AdminLockPayload {
  resource: AdminLockResource
}
