export const EVENT = {
  SYSTEM: {
    DISCONNECT: 'system/disconnect',
  },
  SIGNAL: {
    ANSWER: 'signal/answer',
    PROVIDE: 'signal/provide',
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
    CLIPBOARD: 'control/clipboard',
    KEYBOARD: 'control/keyboard',
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
    MUTE: 'admin/mute',
    UNMUTE: 'admin/unmute',
    LOCK: 'admin/lock',
    UNLOCK: 'admin/unlock',
    CONTROL: 'admin/control',
    RELEASE: 'admin/release',
    GIVE: 'admin/give',
  },
} as const

export type Events = typeof EVENT

export type WebSocketEvents =
  | SystemEvents
  | SignalEvents
  | MemberEvents
  | ControlEvents
  | ScreenEvents
  | BroadcastEvents
  | AdminEvents

export type SystemEvents = typeof EVENT.SYSTEM.DISCONNECT

export type SignalEvents = typeof EVENT.SIGNAL.ANSWER | typeof EVENT.SIGNAL.PROVIDE

export type MemberEvents = typeof EVENT.MEMBER.LIST | typeof EVENT.MEMBER.CONNECTED | typeof EVENT.MEMBER.DISCONNECTED

export type ControlEvents =
  | typeof EVENT.CONTROL.LOCKED
  | typeof EVENT.CONTROL.RELEASE
  | typeof EVENT.CONTROL.REQUEST
  | typeof EVENT.CONTROL.GIVE
  | typeof EVENT.CONTROL.CLIPBOARD
  | typeof EVENT.CONTROL.KEYBOARD

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
