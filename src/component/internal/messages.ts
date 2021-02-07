import Vue from 'vue'
import * as EVENT from '../types/events'
import * as message from '../types/messages'

import EventEmitter from 'eventemitter3'
import { Logger } from '../utils/logger'
import { NekoWebSocket } from './websocket'
import NekoState from '../types/state'

export interface NekoEvents {
  // connection events
  ['connection.websocket']: (state: 'connected' | 'connecting' | 'disconnected') => void
  ['connection.webrtc']: (state: 'connected' | 'connecting' | 'disconnected') => void
  ['connection.disconnect']: (message: string) => void

  // drag and drop events
  ['upload.drop.started']: () => void
  ['upload.drop.progress']: (progressEvent: ProgressEvent) => void
  ['upload.drop.finished']: (error: Error | null) => void

  // upload dialog events
  ['upload.dialog.requested']: () => void
  ['upload.dialog.overlay']: (id: string) => void
  ['upload.dialog.closed']: () => void

  // custom messages events
  ['receive.unicast']: (sender: string, subject: string, body: any) => void
  ['receive.broadcast']: (sender: string, subject: string, body: any) => void

  // member events
  ['member.created']: (id: string) => void
  ['member.deleted']: (id: string) => void
  ['member.updated']: (id: string) => void

  // room events
  ['room.control.host']: (hasHost: boolean, hostID: string | undefined) => void
  ['room.screen.updated']: (width: number, height: number, rate: number) => void
  ['room.clipboard.updated']: (text: string) => void
  ['room.broadcast.status']: (isActive: boolean, url: string | undefined) => void
}

export class NekoMessages extends EventEmitter<NekoEvents> {
  private state: NekoState
  private _log: Logger

  constructor(websocket: NekoWebSocket, state: NekoState) {
    super()

    this._log = new Logger('messages')
    this.state = state
    websocket.on('message', async (event: string, payload: any) => {
      // @ts-ignore
      if (typeof this[event] === 'function') {
        // @ts-ignore
        this[event](payload)
      } else {
        this._log.warn(`unhandled websocket event '${event}':`, payload)
      }
    })
  }

  /////////////////////////////
  // System Events
  /////////////////////////////

  protected [EVENT.SYSTEM_INIT](conf: message.SystemInit) {
    this._log.debug('EVENT.SYSTEM_INIT')
    Vue.set(this.state, 'member_id', conf.member_id)
    Vue.set(this.state.control, 'implicit_hosting', conf.implicit_hosting)

    for (const id in conf.members) {
      this[EVENT.MEMBER_CREATED](conf.members[id])
    }

    this[EVENT.SCREEN_UPDATED](conf.screen_size)
    this[EVENT.CONTROL_HOST](conf.control_host)
    if (conf.cursor_image) {
      this[EVENT.CURSOR_IMAGE](conf.cursor_image)
    }
  }

  protected [EVENT.SYSTEM_ADMIN]({ screen_sizes_list, broadcast_status }: message.SystemAdmin) {
    this._log.debug('EVENT.SYSTEM_ADMIN')

    const list = screen_sizes_list.sort((a, b) => {
      if (b.width === a.width && b.height == a.height) {
        return b.rate - a.rate
      } else if (b.width === a.width) {
        return b.height - a.height
      }
      return b.width - a.width
    })

    Vue.set(this.state.screen, 'configurations', list)

    this[EVENT.BORADCAST_STATUS](broadcast_status)
  }

  protected [EVENT.SYSTEM_DISCONNECT]({ message }: message.SystemDisconnect) {
    this._log.debug('EVENT.SYSTEM_DISCONNECT')
    Vue.set(this.state.connection, 'authenticated', false)
    this.emit('connection.disconnect', message)
  }

  protected [EVENT.CURSOR_IMAGE]({ uri, width, height, x, y }: message.CursorImage) {
    this._log.debug('EVENT.CURSOR_IMAGE')
    Vue.set(this.state.control, 'cursor', { uri, width, height, x, y })
  }

  /////////////////////////////
  // Signal Events
  /////////////////////////////

  protected [EVENT.SIGNAL_PROVIDE]({ event, video, videos }: message.SignalProvide) {
    this._log.debug('EVENT.SIGNAL_PROVIDE')
    Vue.set(this.state.connection.webrtc, 'video', video)
    Vue.set(this.state.connection.webrtc, 'videos', videos)
    // TODO: Handle.
  }

  protected [EVENT.SIGNAL_CANDIDATE]({ event, ...candidate }: message.SignalCandidate) {
    this._log.debug('EVENT.SIGNAL_CANDIDATE')
    // TODO: Handle.
  }

  protected [EVENT.SIGNAL_VIDEO]({ event, video }: message.SignalVideo) {
    this._log.debug('EVENT.SIGNAL_VIDEO')
    Vue.set(this.state.connection.webrtc, 'video', video)
  }

  /////////////////////////////
  // Member Events
  /////////////////////////////

  protected [EVENT.MEMBER_CREATED]({ id, ...data }: message.MemberData) {
    this._log.debug('EVENT.MEMBER_CREATED', id)
    Vue.set(this.state.members, id, data)
    this.emit('member.created', id)
  }

  protected [EVENT.MEMBER_DELETED]({ id }: message.MemberID) {
    this._log.debug('EVENT.MEMBER_DELETED', id)
    Vue.delete(this.state.members, id)
    this.emit('member.deleted', id)
  }

  protected [EVENT.MEMBER_PROFILE]({ id, ...profile }: message.MemberProfile) {
    this._log.debug('EVENT.MEMBER_PROFILE', id)
    Vue.set(this.state.members[id], 'profile', profile)
    this.emit('member.updated', id)
  }

  protected [EVENT.MEMBER_STATE]({ id, ...state }: message.MemberState) {
    this._log.debug('EVENT.MEMBER_STATE', id)
    Vue.set(this.state.members[id], 'state', state)
    this.emit('member.updated', id)
  }

  /////////////////////////////
  // Control Events
  /////////////////////////////

  protected [EVENT.CONTROL_HOST]({ has_host, host_id }: message.ControlHost) {
    this._log.debug('EVENT.CONTROL_HOST')

    if (has_host && host_id) {
      Vue.set(this.state.control, 'host_id', host_id)
    } else {
      Vue.set(this.state.control, 'host_id', null)
    }

    this.emit('room.control.host', has_host, host_id)
  }

  /////////////////////////////
  // Screen Events
  /////////////////////////////

  protected [EVENT.SCREEN_UPDATED]({ width, height, rate }: message.ScreenSize) {
    this._log.debug('EVENT.SCREEN_UPDATED')
    Vue.set(this.state.screen, 'size', { width, height, rate })
    this.emit('room.screen.updated', width, height, rate)
  }
  /////////////////////////////
  // Clipboard Events
  /////////////////////////////

  protected [EVENT.CLIPBOARD_UPDATED]({ text }: message.ClipboardData) {
    this._log.debug('EVENT.CLIPBOARD_UPDATED')
    Vue.set(this.state.control, 'clipboard', { text })
    this.emit('room.clipboard.updated', text)
  }

  /////////////////////////////
  // Broadcast Events
  /////////////////////////////

  protected [EVENT.BORADCAST_STATUS]({ event, url, is_active }: message.BroadcastStatus) {
    this._log.debug('EVENT.BORADCAST_STATUS')
    // TODO: Handle.
    this.emit('room.broadcast.status', is_active, url)
  }

  /////////////////////////////
  // Send Events
  /////////////////////////////

  protected [EVENT.SEND_UNICAST]({ sender, subject, body }: message.SendMessage) {
    this._log.debug('EVENT.SEND_UNICAST')
    this.emit('receive.unicast', sender, subject, body)
  }

  protected [EVENT.SEND_BROADCAST]({ sender, subject, body }: message.SendMessage) {
    this._log.debug('EVENT.BORADCAST_STATUS')
    this.emit('receive.broadcast', sender, subject, body)
  }

  /////////////////////////////
  // FileChooserDialog Events
  /////////////////////////////

  protected [EVENT.FILE_CHOOSER_DIALOG_OPENED]({ id }: message.MemberID) {
    this._log.debug('EVENT.FILE_CHOOSER_DIALOG_OPENED')

    if (id == this.state.member_id) {
      this.emit('upload.dialog.requested')
    } else {
      this.emit('upload.dialog.overlay', id)
    }
  }

  protected [EVENT.FILE_CHOOSER_DIALOG_CLOSED]({ id }: message.MemberID) {
    this._log.debug('EVENT.FILE_CHOOSER_DIALOG_CLOSED')
    this.emit('upload.dialog.closed')
  }
}
