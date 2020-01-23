export const EVENT = {
  // Internal Events
  CONNECTING: 'CONNECTING',
  CONNECTED: 'CONNECTED',
  DISCONNECTED: 'DISCONNECTED',
  TRACK: 'TRACK',
  MESSAGE: 'MESSAGE',
  DATA: 'DATA',

  // Websocket Events
  SYSTEM: {
    DISCONNECT: 'system/disconnect',
  },
  SIGNAL: {
    ANSWER: 'signal/answer',
    PROVIDE: 'signal/provide',
  },
  IDENTITY: {
    PROVIDE: 'identity/provide',
    DETAILS: 'identity/details',
  },
  MEMBER: {
    LIST: 'member/list',
    CONNECTED: 'member/connected',
    DISCONNECTED: 'member/disconnected',
  },
  CONTROL: {
    LOCKED: 'control/locked',
    RELEASE: 'control/release',
    REQUEST: 'control/request',
    REQUESTING: 'control/requesting',
    GIVE: 'control/give',
  },
  CHAT: {
    MESSAGE: 'chat/message',
    EMOTE: 'chat/emote',
  },
  ADMIN: {
    BAN: 'admin/ban',
    KICK: 'admin/kick',
    LOCK: 'admin/lock',
    UNLOCK: 'admin/unlock',
    MUTE: 'admin/mute',
    UNMUTE: 'admin/unmute',
    CONTROL: 'admin/control',
    RELEASE: 'admin/release',
    GIVE: 'admin/give',
  },
} as const

export type Events = typeof EVENT

export type WebSocketEvents =
  | SystemEvents
  | ControlEvents
  | IdentityEvents
  | MemberEvents
  | SignalEvents
  | ChatEvents
  | AdminEvents

export type ControlEvents =
  | typeof EVENT.CONTROL.LOCKED
  | typeof EVENT.CONTROL.RELEASE
  | typeof EVENT.CONTROL.REQUEST
  | typeof EVENT.CONTROL.GIVE

export type SystemEvents = typeof EVENT.SYSTEM.DISCONNECT
export type IdentityEvents = typeof EVENT.IDENTITY.PROVIDE | typeof EVENT.IDENTITY.DETAILS
export type MemberEvents = typeof EVENT.MEMBER.LIST | typeof EVENT.MEMBER.CONNECTED | typeof EVENT.MEMBER.DISCONNECTED
export type SignalEvents = typeof EVENT.SIGNAL.ANSWER | typeof EVENT.SIGNAL.PROVIDE
export type ChatEvents = typeof EVENT.CHAT.MESSAGE | typeof EVENT.CHAT.EMOTE
export type AdminEvents =
  | typeof EVENT.ADMIN.BAN
  | typeof EVENT.ADMIN.KICK
  | typeof EVENT.ADMIN.LOCK
  | typeof EVENT.ADMIN.UNLOCK
  | typeof EVENT.ADMIN.MUTE
  | typeof EVENT.ADMIN.UNMUTE
  | typeof EVENT.ADMIN.CONTROL
  | typeof EVENT.ADMIN.RELEASE
  | typeof EVENT.ADMIN.GIVE
