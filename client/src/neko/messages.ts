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
} from './events'
import { Member, ScreenConfigurations, ScreenResolution } from './types'

export type WebSocketMessages =
  | WebSocketMessage
  | SignalProvideMessage
  | SignalAnswerMessage
  | MemberListMessage
  | MemberConnectMessage
  | MemberDisconnectMessage
  | ControlMessage
  | ScreenResolutionMessage
  | ScreenConfigurationsMessage
  | ChatMessage

export type WebSocketPayloads =
  | SignalProvidePayload
  | SignalAnswerPayload
  | MemberListPayload
  | Member
  | ControlPayload
  | ControlClipboardPayload
  | ChatPayload
  | ChatSendPayload
  | EmojiSendPayload
  | ScreenResolutionPayload
  | ScreenConfigurationsPayload
  | AdminPayload

export interface WebSocketMessage {
  event: WebSocketEvents | string
}

/*
  SYSTEM MESSAGES/PAYLOADS
*/
// system/disconnect
export interface DisconnectMessage extends WebSocketMessage, DisconnectPayload {
  event: typeof EVENT.SYSTEM.DISCONNECT
}
export interface DisconnectPayload {
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
  ice: string[]
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
