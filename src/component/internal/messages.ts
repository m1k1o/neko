import Vue from 'vue'
import * as EVENT from '../types/events'
import * as message from '../types/messages'

import EventEmitter from 'eventemitter3'
import { Logger } from '../utils/logger'
import { NekoConnection } from './connection'
import NekoState from '../types/state'

export interface NekoEvents {
  // connection events
  ['connection.status']: (status: 'connected' | 'connecting' | 'disconnected') => void
  ['connection.type']: (status: 'fallback' | 'webrtc' | 'none') => void
  ['connection.webrtc.sdp']: (type: 'local' | 'remote', data: string) => void
  ['connection.webrtc.sdp.candidate']: (type: 'local' | 'remote', data: RTCIceCandidateInit) => void
  ['connection.closed']: (error?: Error) => void

  // drag and drop events
  ['upload.drop.started']: () => void
  ['upload.drop.progress']: (progressEvent: ProgressEvent) => void
  ['upload.drop.finished']: (error?: Error) => void

  // upload dialog events
  ['upload.dialog.requested']: () => void
  ['upload.dialog.overlay']: (id: string) => void
  ['upload.dialog.closed']: () => void

  // custom messages events
  ['receive.unicast']: (sender: string, subject: string, body: any) => void
  ['receive.broadcast']: (sender: string, subject: string, body: any) => void

  // session events
  ['session.created']: (id: string) => void
  ['session.deleted']: (id: string) => void
  ['session.updated']: (id: string) => void

  // room events
  ['room.control.host']: (hasHost: boolean, hostID?: string) => void
  ['room.screen.updated']: (width: number, height: number, rate: number) => void
  ['room.clipboard.updated']: (text: string) => void
  ['room.broadcast.status']: (isActive: boolean, url?: string) => void
}

export class NekoMessages extends EventEmitter<NekoEvents> {
  private _connection: NekoConnection
  private _state: NekoState
  private _localLog: Logger
  private _remoteLog: Logger

  constructor(connection: NekoConnection, state: NekoState) {
    super()

    this._connection = connection
    this._state = state
    this._localLog = new Logger('messages')
    this._remoteLog = connection.getLogger('messages')

    this._connection.websocket.on('message', async (event: string, payload: any) => {
      // @ts-ignore
      if (typeof this[event] === 'function') {
        try {
          // @ts-ignore
          this[event](payload)
        } catch (error: any) {
          this._remoteLog.error(`error while processing websocket event`, { event, error })
        }
      } else {
        this._remoteLog.warn(`unhandled websocket event`, { event, payload })
      }
    })

    this._connection.webrtc.on('candidate', (candidate: RTCIceCandidateInit) => {
      this._connection.websocket.send(EVENT.SIGNAL_CANDIDATE, candidate)
      this.emit('connection.webrtc.sdp.candidate', 'local', candidate)
    })
  }

  /////////////////////////////
  // System Events
  /////////////////////////////

  protected [EVENT.SYSTEM_INIT](conf: message.SystemInit) {
    this._localLog.debug(`EVENT.SYSTEM_INIT`)
    Vue.set(this._state, 'session_id', conf.session_id)
    Vue.set(this._state.control, 'implicit_hosting', conf.implicit_hosting)
    Vue.set(this._state.connection, 'screencast', conf.screencast_enabled)
    Vue.set(this._state.connection.webrtc, 'videos', conf.webrtc.videos)

    for (const id in conf.sessions) {
      this[EVENT.SESSION_CREATED](conf.sessions[id])
    }

    this[EVENT.SCREEN_UPDATED](conf.screen_size)
    this[EVENT.CONTROL_HOST](conf.control_host)
  }

  protected [EVENT.SYSTEM_ADMIN]({ screen_sizes_list, broadcast_status }: message.SystemAdmin) {
    this._localLog.debug(`EVENT.SYSTEM_ADMIN`)

    const list = screen_sizes_list.sort((a, b) => {
      if (b.width === a.width && b.height == a.height) {
        return b.rate - a.rate
      } else if (b.width === a.width) {
        return b.height - a.height
      }
      return b.width - a.width
    })

    Vue.set(this._state.screen, 'configurations', list)

    this[EVENT.BORADCAST_STATUS](broadcast_status)
  }

  protected [EVENT.SYSTEM_DISCONNECT]({ message }: message.SystemDisconnect) {
    this._localLog.debug(`EVENT.SYSTEM_DISCONNECT`)
    this._connection.close(new Error(message))
  }

  /////////////////////////////
  // Signal Events
  /////////////////////////////

  protected async [EVENT.SIGNAL_PROVIDE]({ sdp: remoteSdp, video, iceservers }: message.SignalProvide) {
    this._localLog.debug(`EVENT.SIGNAL_PROVIDE`)
    this.emit('connection.webrtc.sdp', 'remote', remoteSdp)

    const localSdp = await this._connection.webrtc.connect(remoteSdp, iceservers)
    this._connection.websocket.send(EVENT.SIGNAL_ANSWER, {
      sdp: localSdp,
    })

    this.emit('connection.webrtc.sdp', 'local', localSdp)
    Vue.set(this._state.connection.webrtc, 'video', video)
  }

  protected async [EVENT.SIGNAL_RESTART]({ sdp }: message.SignalAnswer) {
    this._localLog.debug(`EVENT.SIGNAL_RESTART`)
    this.emit('connection.webrtc.sdp', 'remote', sdp)

    const localSdp = await this._connection.webrtc.offer(sdp)
    this._connection.websocket.send(EVENT.SIGNAL_ANSWER, {
      sdp: localSdp,
    })

    this.emit('connection.webrtc.sdp', 'local', localSdp)
  }

  protected [EVENT.SIGNAL_CANDIDATE](candidate: message.SignalCandidate) {
    this._localLog.debug(`EVENT.SIGNAL_CANDIDATE`)
    this._connection.webrtc.setCandidate(candidate)
    this.emit('connection.webrtc.sdp.candidate', 'remote', candidate)
  }

  protected [EVENT.SIGNAL_VIDEO]({ video }: message.SignalVideo) {
    this._localLog.debug(`EVENT.SIGNAL_VIDEO`, { video })
    Vue.set(this._state.connection.webrtc, 'video', video)
  }

  /////////////////////////////
  // Session Events
  /////////////////////////////

  protected [EVENT.SESSION_CREATED]({ id, ...data }: message.SessionData) {
    this._localLog.debug(`EVENT.SESSION_CREATED`, { id })
    Vue.set(this._state.sessions, id, data)
    this.emit('session.created', id)
  }

  protected [EVENT.SESSION_DELETED]({ id }: message.SessionID) {
    this._localLog.debug(`EVENT.SESSION_DELETED`, { id })
    Vue.delete(this._state.sessions, id)
    this.emit('session.deleted', id)
  }

  protected [EVENT.SESSION_PROFILE]({ id, ...profile }: message.MemberProfile) {
    this._localLog.debug(`EVENT.SESSION_PROFILE`, { id })
    Vue.set(this._state.sessions[id], 'profile', profile)
    this.emit('session.updated', id)
  }

  protected [EVENT.SESSION_STATE]({ id, ...state }: message.SessionState) {
    this._localLog.debug(`EVENT.SESSION_STATE`, { id })
    Vue.set(this._state.sessions[id], 'state', state)
    this.emit('session.updated', id)
  }

  /////////////////////////////
  // Control Events
  /////////////////////////////

  protected [EVENT.CONTROL_HOST]({ has_host, host_id }: message.ControlHost) {
    this._localLog.debug(`EVENT.CONTROL_HOST`)

    if (has_host && host_id) {
      Vue.set(this._state.control, 'host_id', host_id)
    } else {
      Vue.set(this._state.control, 'host_id', null)
    }

    this.emit('room.control.host', has_host, host_id)
  }

  /////////////////////////////
  // Screen Events
  /////////////////////////////

  protected [EVENT.SCREEN_UPDATED]({ width, height, rate }: message.ScreenSize) {
    this._localLog.debug(`EVENT.SCREEN_UPDATED`)
    Vue.set(this._state.screen, 'size', { width, height, rate })
    this.emit('room.screen.updated', width, height, rate)
  }

  /////////////////////////////
  // Clipboard Events
  /////////////////////////////

  protected [EVENT.CLIPBOARD_UPDATED]({ text }: message.ClipboardData) {
    this._localLog.debug(`EVENT.CLIPBOARD_UPDATED`)
    Vue.set(this._state.control, 'clipboard', { text })
    this.emit('room.clipboard.updated', text)
  }

  /////////////////////////////
  // Broadcast Events
  /////////////////////////////

  protected [EVENT.BORADCAST_STATUS]({ url, is_active }: message.BroadcastStatus) {
    this._localLog.debug(`EVENT.BORADCAST_STATUS`)
    // TODO: Handle.
    this.emit('room.broadcast.status', is_active, url)
  }

  /////////////////////////////
  // Send Events
  /////////////////////////////

  protected [EVENT.SEND_UNICAST]({ sender, subject, body }: message.SendMessage) {
    this._localLog.debug(`EVENT.SEND_UNICAST`)
    this.emit('receive.unicast', sender, subject, body)
  }

  protected [EVENT.SEND_BROADCAST]({ sender, subject, body }: message.SendMessage) {
    this._localLog.debug(`EVENT.BORADCAST_STATUS`)
    this.emit('receive.broadcast', sender, subject, body)
  }

  /////////////////////////////
  // FileChooserDialog Events
  /////////////////////////////

  protected [EVENT.FILE_CHOOSER_DIALOG_OPENED]({ id }: message.SessionID) {
    this._localLog.debug(`EVENT.FILE_CHOOSER_DIALOG_OPENED`)

    if (id == this._state.session_id) {
      this.emit('upload.dialog.requested')
    } else {
      this.emit('upload.dialog.overlay', id)
    }
  }

  protected [EVENT.FILE_CHOOSER_DIALOG_CLOSED]({ id }: message.SessionID) {
    this._localLog.debug(`EVENT.FILE_CHOOSER_DIALOG_CLOSED`)
    this.emit('upload.dialog.closed')
  }
}
