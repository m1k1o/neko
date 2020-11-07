import { Member, ScreenConfigurations } from '../types/structs'
import { EVENT } from '../types/events'
import {
  DisconnectPayload,
  MemberListPayload,
  MemberDisconnectPayload,
  MemberPayload,
  ControlPayload,
  ControlTargetPayload,
  ControlClipboardPayload,
  ScreenConfigurationsPayload,
  ScreenResolutionPayload,
  BroadcastStatusPayload,
  AdminPayload,
  AdminTargetPayload,
} from '../types/messages'

import EventEmitter from 'eventemitter3'
import { NekoWebSocket } from './websocket'

export interface NekoEvents {
  ['system.websocket']: (state: 'connected' | 'connecting' | 'disconnected') => void
  ['system.webrtc']: (state: 'connected' | 'connecting' | 'disconnected') => void
  ['system.connect']: () => void
  ['system.disconnect']: (message: string) => void
  ['control.host']: (id: string | null) => void
  ['member.list']: (members: Member[]) => void
  ['member.connected']: (id: string) => void
  ['member.disconnected']: (id: string) => void
  ['control.request']: (id: string) => void
  ['control.requesting']: (id: string) => void
  ['clipboard.update']: (text: string) => void
  ['screen.configuration']: (configurations: ScreenConfigurations) => void
  ['screen.size']: (width: number, height: number, rate: number) => void
  ['broadcast.status']: (url: string, isActive: boolean) => void
  ['member.ban']: (id: string, target: string) => void
  ['member.kick']: (id: string, target: string) => void
  ['member.muted']: (id: string, target: string) => void
  ['member.unmuted']: (id: string, target: string) => void
  ['room.locked']: (id: string) => void
  ['room.unlocked']: (id: string) => void
}

export class NekoMessages extends EventEmitter<NekoEvents> {
  constructor(websocket: NekoWebSocket) {
    super()

    websocket.on('message', async (event: string, payload: any) => {
      // @ts-ignore
      if (typeof this[event] === 'function') {
        // @ts-ignore
        this[event](payload)
      } else {
        console.log(`unhandled websocket event '${event}':`, payload)
      }
    })
  }

  /////////////////////////////
  // System Events
  /////////////////////////////
  protected [EVENT.SYSTEM.DISCONNECT]({ message }: DisconnectPayload) {
    console.log('EVENT.SYSTEM.DISCONNECT')
    this.emit('system.disconnect', message)
  }

  /////////////////////////////
  // Member Events
  /////////////////////////////
  protected [EVENT.MEMBER.LIST]({ members }: MemberListPayload) {
    console.log('EVENT.MEMBER.LIST')
    this.emit('member.list', members)
    //user.setMembers(members)
  }

  protected [EVENT.MEMBER.CONNECTED](member: MemberPayload) {
    console.log('EVENT.MEMBER.CONNECTED')
    this.emit('member.connected', member.id)
    //user.addMember(member)
  }

  protected [EVENT.MEMBER.DISCONNECTED]({ id }: MemberDisconnectPayload) {
    console.log('EVENT.MEMBER.DISCONNECTED')
    this.emit('member.disconnected', id)
    //user.delMember(id)
  }

  /////////////////////////////
  // Control Events
  /////////////////////////////
  protected [EVENT.CONTROL.LOCKED]({ id }: ControlPayload) {
    console.log('EVENT.CONTROL.LOCKED')
    this.emit('control.host', id)
    //remote.setHost(id)
    //remote.changeKeyboard()
  }

  protected [EVENT.CONTROL.RELEASE]({ id }: ControlPayload) {
    console.log('EVENT.CONTROL.RELEASE')
    this.emit('control.host', null)
    //remote.reset()
  }

  protected [EVENT.CONTROL.REQUEST]({ id }: ControlPayload) {
    console.log('EVENT.CONTROL.REQUEST')
    this.emit('control.request', id)
  }

  protected [EVENT.CONTROL.REQUESTING]({ id }: ControlPayload) {
    console.log('EVENT.CONTROL.REQUESTING')
    this.emit('control.requesting', id)
  }

  protected [EVENT.CONTROL.GIVE]({ id, target }: ControlTargetPayload) {
    console.log('EVENT.CONTROL.GIVE')
    this.emit('control.host', target)
    //remote.setHost(target)
    //remote.changeKeyboard()
  }

  protected [EVENT.CONTROL.CLIPBOARD]({ text }: ControlClipboardPayload) {
    console.log('EVENT.CONTROL.CLIPBOARD')
    this.emit('clipboard.update', text)
    //remote.setClipboard(text)
  }

  /////////////////////////////
  // Screen Events
  /////////////////////////////
  protected [EVENT.SCREEN.CONFIGURATIONS]({ configurations }: ScreenConfigurationsPayload) {
    console.log('EVENT.SCREEN.CONFIGURATIONS')
    this.emit('screen.configuration', configurations)
    //video.setConfigurations(configurations)
  }

  protected [EVENT.SCREEN.RESOLUTION]({ id, width, height, rate }: ScreenResolutionPayload) {
    console.log('EVENT.SCREEN.RESOLUTION')
    this.emit('screen.size', width, height, rate)
    //video.setResolution({ width, height, rate })
  }

  /////////////////////////////
  // Broadcast Events
  /////////////////////////////
  protected [EVENT.BROADCAST.STATUS](payload: BroadcastStatusPayload) {
    console.log('EVENT.BROADCAST.STATUS')
    this.emit('broadcast.status', payload.url, payload.isActive)
    //settings.broadcastStatus(payload)
  }

  /////////////////////////////
  // Admin Events
  /////////////////////////////
  protected [EVENT.ADMIN.BAN]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.BAN')
    this.emit('member.ban', id, target)
    // TODO
  }

  protected [EVENT.ADMIN.KICK]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.KICK')
    this.emit('member.kick', id, target)
    // TODO
  }

  protected [EVENT.ADMIN.MUTE]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.MUTE')
    this.emit('member.muted', id, target)
    //user.setMuted({ id: target, muted: true })
  }

  protected [EVENT.ADMIN.UNMUTE]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.UNMUTE')
    this.emit('member.unmuted', id, target)
    //user.setMuted({ id: target, muted: false })
  }

  protected [EVENT.ADMIN.LOCK]({ id }: AdminPayload) {
    console.log('EVENT.ADMIN.LOCK')
    this.emit('room.locked', id)
    //setLocked(true)
  }

  protected [EVENT.ADMIN.UNLOCK]({ id }: AdminPayload) {
    console.log('EVENT.ADMIN.UNLOCK')
    this.emit('room.unlocked', id)
    //setLocked(false)
  }

  protected [EVENT.ADMIN.CONTROL]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.CONTROL')
    this.emit('control.host', id)
    //remote.setHost(id)
    //remote.changeKeyboard()
  }

  protected [EVENT.ADMIN.RELEASE]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.RELEASE')
    this.emit('control.host', null)
    //remote.reset()
  }

  protected [EVENT.ADMIN.GIVE]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.GIVE')
    this.emit('control.host', target)
    //remote.setHost(target)
    //remote.changeKeyboard()
  }
}
