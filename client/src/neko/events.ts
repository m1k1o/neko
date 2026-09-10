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
    ADMIN: 'system/admin',
    SETTINGS: 'system/settings',
    HEARTBEAT: 'system/heartbeat',
    DISCONNECT: 'system/disconnect',
    ERROR: 'system/error',
  },
  CLIENT: {
    HEARTBEAT: 'client/heartbeat',
  },
  SIGNAL: {
    REQUEST: 'signal/request',
    RESTART: 'signal/restart',
    OFFER: 'signal/offer',
    ANSWER: 'signal/answer',
    PROVIDE: 'signal/provide',
    CANDIDATE: 'signal/candidate',
    CLOSE: 'signal/close',
  },
  SESSION: {
    CREATED: 'session/created',
    DELETED: 'session/deleted',
    PROFILE: 'session/profile',
    STATE: 'session/state',
    CURSORS: 'session/cursors',
  },
  CONTROL: {
    HOST: 'control/host',
    RELEASE: 'control/release',
    REQUEST: 'control/request',
    RENEW: 'control/renew',
  },
  CHAT: {
    INIT: 'chat/init',
    MESSAGE: 'chat/message',
    EMOTE: 'chat/emote',
  },
  FILETRANSFER: {
    UPDATE: 'filetransfer/update',
  },
  OPENINAPP: {
    INIT: 'openinapp/init',
    OPENLINK: 'openinapp/openlink',
  },
  SCREEN: {
    UPDATED: 'screen/updated',
    SET: 'screen/set',
  },
  CLIPBOARD: {
    UPDATED: 'clipboard/updated',
    SET: 'clipboard/set',
  },
  BROADCAST: {
    STATUS: 'broadcast/status',
  },
  KEYBOARD: {
    MAP: 'keyboard/map',
    MODIFIERS: 'keyboard/modifiers',
  },
} as const

export type Events = typeof EVENT

export type WebSocketEvents =
  | SystemEvents
  | ClientEvents
  | ControlEvents
  | SessionEvents
  | SignalEvents
  | ChatEvents
  | FileTransferEvents
  | OpenInAppEvents
  | ScreenEvents
  | ClipboardEvents
  | KeyboardEvents
  | BroadcastEvents

export type ControlEvents =
  | typeof EVENT.CONTROL.HOST
  | typeof EVENT.CONTROL.RELEASE
  | typeof EVENT.CONTROL.REQUEST
  | typeof EVENT.CONTROL.RENEW

export type SystemEvents =
  | typeof EVENT.SYSTEM.INIT
  | typeof EVENT.SYSTEM.ADMIN
  | typeof EVENT.SYSTEM.SETTINGS
  | typeof EVENT.SYSTEM.HEARTBEAT
  | typeof EVENT.SYSTEM.DISCONNECT
  | typeof EVENT.SYSTEM.ERROR
export type ClientEvents = typeof EVENT.CLIENT.HEARTBEAT

export type SessionEvents =
  | typeof EVENT.SESSION.CREATED
  | typeof EVENT.SESSION.DELETED
  | typeof EVENT.SESSION.PROFILE
  | typeof EVENT.SESSION.STATE
  | typeof EVENT.SESSION.CURSORS

export type SignalEvents =
  | typeof EVENT.SIGNAL.REQUEST
  | typeof EVENT.SIGNAL.RESTART
  | typeof EVENT.SIGNAL.OFFER
  | typeof EVENT.SIGNAL.ANSWER
  | typeof EVENT.SIGNAL.PROVIDE
  | typeof EVENT.SIGNAL.CANDIDATE
  | typeof EVENT.SIGNAL.CLOSE

export type ChatEvents = typeof EVENT.CHAT.INIT | typeof EVENT.CHAT.MESSAGE | typeof EVENT.CHAT.EMOTE

export type FileTransferEvents = typeof EVENT.FILETRANSFER.UPDATE

export type OpenInAppEvents = typeof EVENT.OPENINAPP.INIT | typeof EVENT.OPENINAPP.OPENLINK

export type ScreenEvents = typeof EVENT.SCREEN.UPDATED | typeof EVENT.SCREEN.SET

export type ClipboardEvents = typeof EVENT.CLIPBOARD.UPDATED | typeof EVENT.CLIPBOARD.SET

export type KeyboardEvents = typeof EVENT.KEYBOARD.MAP | typeof EVENT.KEYBOARD.MODIFIERS

export type BroadcastEvents = typeof EVENT.BROADCAST.STATUS
