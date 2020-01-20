import { WebSocketEvents, EVENT } from './events'
import { Member } from './types'

export type WebSocketMessages =
  | WebSocketMessage
  | IdentityMessage
  | SignalMessage
  | MemberListMessage
  | MembeConnectMessage
  | MembeDisconnectMessage
  | ControlMessage
  | ChatMessage

export type WebSocketPayloads =
  | IdentityPayload
  | SignalPayload
  | MemberListPayload
  | Member
  | ControlPayload
  | ChatPayload

export interface WebSocketMessage {
  event: WebSocketEvents | string
}

/*
  IDENTITY MESSAGES/PAYLOADS
*/
// identity/provide
export interface IdentityMessage extends WebSocketMessage, IdentityPayload {
  event: typeof EVENT.IDENTITY.PROVIDE
}
export interface IdentityPayload {
  id: string
}

/*
  SIGNAL MESSAGES/PAYLOADS
*/
// signal/answer
export interface SignalMessage extends WebSocketMessage, SignalPayload {
  event: typeof EVENT.SIGNAL.ANSWER
}
export interface SignalPayload {
  sdp: string
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
export interface MembeConnectMessage extends WebSocketMessage, MemberPayload {
  event: typeof EVENT.MEMBER.CONNECTED
}
export type MemberPayload = Member

// member/disconnected
export interface MembeDisconnectMessage extends WebSocketMessage, MemberPayload {
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
  event: typeof EVENT.CONTROL.LOCKED | typeof EVENT.CONTROL.RELEASE | typeof EVENT.CONTROL.REQUEST
}
export interface ControlPayload {
  id: string
}

/*
  CHAT PAYLOADS
*/
// chat/send & chat/receive
export interface ChatMessage extends WebSocketMessage, ChatPayload {
  event: typeof EVENT.CHAT.SEND | typeof EVENT.CHAT.RECEIVE
}

export interface ChatPayload {
  id: string
  content: string
}
