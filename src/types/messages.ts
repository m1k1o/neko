import {
  EVENT,
  WebSocketEvents,
  SystemEvents,
  SignalEvents,
  MemberEvents,
  ControlEvents,
  ScreenEvents,
  BroadcastEvents,
  AdminEvents,
} from './events'

import {
  Member,
  ScreenConfigurations,
  ScreenResolution
} from './structs'

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

export type WebSocketPayloads =
  | SignalProvidePayload
  | SignalAnswerPayload
  | MemberListPayload
  | Member
  | ControlPayload
  | ControlClipboardPayload
  | ControlKeyboardPayload
  | ScreenResolutionPayload
  | ScreenConfigurationsPayload
  | AdminPayload
  | BroadcastStatusPayload
  | BroadcastCreatePayload

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

export interface ControlKeyboardPayload {
  layout?: string
  capsLock?: boolean
  numLock?: boolean
  scrollLock?: boolean
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
  url:   string
}

export interface BroadcastStatusPayload {
  url:      string
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
