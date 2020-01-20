export const EVENT = {
  // Internal Events
  CONNECTING: 'CONNECTING',
  CONNECTED: 'CONNECTED',
  DISCONNECTED: 'DISCONNECTED',
  TRACK: 'TRACK',
  MESSAGE: 'MESSAGE',
  DATA: 'DATA',

  // Websocket Events
  DISCONNECT: 'disconnect',
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
  },
  CHAT: {
    MESSAGE: 'chat/message',
    EMOJI: 'chat/emoji',
  },
  ADMIN: {
    BAN: 'admin/ban',
    KICK: 'admin/kick',
    LOCK: 'admin/lock',
    MUTE: 'admin/mute',
    UNMUTE: 'admin/unmute',
    FORCE: {
      CONTROL: 'admin/force/control',
      RELEASE: 'admin/force/release',
    },
  },
} as const

export type Events = typeof EVENT
export type WebSocketEvents = SystemEvents | ControlEvents | IdentityEvents | MemberEvents | SignalEvents | ChatEvents
export type SystemEvents = typeof EVENT.DISCONNECT
export type ControlEvents = typeof EVENT.CONTROL.LOCKED | typeof EVENT.CONTROL.RELEASE | typeof EVENT.CONTROL.REQUEST
export type IdentityEvents = typeof EVENT.IDENTITY.PROVIDE | typeof EVENT.IDENTITY.DETAILS
export type MemberEvents = typeof EVENT.MEMBER.LIST | typeof EVENT.MEMBER.CONNECTED | typeof EVENT.MEMBER.DISCONNECTED
export type SignalEvents = typeof EVENT.SIGNAL.ANSWER | typeof EVENT.SIGNAL.PROVIDE
export type ChatEvents = typeof EVENT.CHAT.MESSAGE | typeof EVENT.CHAT.EMOJI
