export const EVENT = {
  // Internal Events
  RECONNECTING: 'RECONNECTING',
  CONNECTING: 'CONNECTING',
  CONNECTED: 'CONNECTED',
  DISCONNECTED: 'DISCONNECTED',
  TRACK: 'TRACK',
  MESSAGE: 'MESSAGE',
  DATA: 'DATA',

  // Websocket Events
  SYSTEM: {
    INIT: 'system/init',
    DISCONNECT: 'system/disconnect',
    ERROR: 'system/error',
  },
  CLIENT: {
    HEARTBEAT: 'client/heartbeat',
  },
  SIGNAL: {
    OFFER: 'signal/offer',
    ANSWER: 'signal/answer',
    PROVIDE: 'signal/provide',
    CANDIDATE: 'signal/candidate',
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
    CLIPBOARD: 'control/clipboard',
    GIVE: 'control/give',
    KEYBOARD: 'control/keyboard',
  },
  CHAT: {
    MESSAGE: 'chat/message',
    EMOTE: 'chat/emote',
  },
  FILETRANSFER: {
    LIST: 'filetransfer/list',
    REFRESH: 'filetransfer/refresh',
  },
  SCREEN: {
    CONFIGURATIONS: 'screen/configurations',
    RESOLUTION: 'screen/resolution',
    SET: 'screen/set',
  },
  BROADCAST: {
    STATUS: 'broadcast/status',
    CREATE: 'broadcast/create',
    DESTROY: 'broadcast/destroy',
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
  | ClientEvents
  | ControlEvents
  | MemberEvents
  | SignalEvents
  | ChatEvents
  | FileTransferEvents
  | ScreenEvents
  | BroadcastEvents
  | AdminEvents

export type ControlEvents =
  | typeof EVENT.CONTROL.LOCKED
  | typeof EVENT.CONTROL.RELEASE
  | typeof EVENT.CONTROL.REQUEST
  | typeof EVENT.CONTROL.GIVE
  | typeof EVENT.CONTROL.CLIPBOARD
  | typeof EVENT.CONTROL.KEYBOARD

export type SystemEvents = typeof EVENT.SYSTEM.DISCONNECT
export type ClientEvents = typeof EVENT.CLIENT.HEARTBEAT
export type MemberEvents = typeof EVENT.MEMBER.LIST | typeof EVENT.MEMBER.CONNECTED | typeof EVENT.MEMBER.DISCONNECTED

export type SignalEvents =
  | typeof EVENT.SIGNAL.OFFER
  | typeof EVENT.SIGNAL.ANSWER
  | typeof EVENT.SIGNAL.PROVIDE
  | typeof EVENT.SIGNAL.CANDIDATE

export type ChatEvents = typeof EVENT.CHAT.MESSAGE | typeof EVENT.CHAT.EMOTE

export type FileTransferEvents = typeof EVENT.FILETRANSFER.LIST | typeof EVENT.FILETRANSFER.REFRESH

export type ScreenEvents = typeof EVENT.SCREEN.CONFIGURATIONS | typeof EVENT.SCREEN.RESOLUTION | typeof EVENT.SCREEN.SET

export type BroadcastEvents =
  | typeof EVENT.BROADCAST.STATUS
  | typeof EVENT.BROADCAST.CREATE
  | typeof EVENT.BROADCAST.DESTROY

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
