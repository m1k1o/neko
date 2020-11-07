import Vue from 'vue'
import { Member } from '../types/structs'
import { EVENT } from '../types/events'
import {
  DisconnectPayload,
  SignalProvidePayload,
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

export class NekoMessages {
  _eventEmmiter: EventEmitter

  constructor(eventEmitter: EventEmitter) {
    this._eventEmmiter = eventEmitter
  }

  private emit(event: string, ...payload: any) {
    this._eventEmmiter.emit(event, ...payload)
  }

  /////////////////////////////
  // System Events
  /////////////////////////////
  public [EVENT.SYSTEM.DISCONNECT]({ message }: DisconnectPayload) {
    console.log('EVENT.SYSTEM.DISCONNECT')
    this.emit('disconnect', message)
    //this.onDisconnected(new Error(message))
    //this.$vue.$swal({
    //  title: this.$vue.$t('connection.disconnected'),
    //  text: message,
    //  icon: 'error',
    //  confirmButtonText: this.$vue.$t('connection.button_confirm') as string,
    //})
  }

  /////////////////////////////
  // Member Events
  /////////////////////////////
  public [EVENT.MEMBER.LIST]({ members }: MemberListPayload) {
    console.log('EVENT.MEMBER.LIST')
    this.emit('member.list', members)
    //this.$accessor.user.setMembers(members)
  }

  public [EVENT.MEMBER.CONNECTED](member: MemberPayload) {
    console.log('EVENT.MEMBER.CONNECTED')
    this.emit('member.connected', member.id)
    //this.$accessor.user.addMember(member)
  }

  public [EVENT.MEMBER.DISCONNECTED]({ id }: MemberDisconnectPayload) {
    console.log('EVENT.MEMBER.DISCONNECTED')
    this.emit('member.disconnected', id)
    //this.$accessor.user.delMember(id)
  }

  /////////////////////////////
  // Control Events
  /////////////////////////////
  public [EVENT.CONTROL.LOCKED]({ id }: ControlPayload) {
    console.log('EVENT.CONTROL.LOCKED')
    this.emit('host.change', id)
    //this.$accessor.remote.setHost(id)
    //this.$accessor.remote.changeKeyboard()
    //
    //const member = this.member(id)
    //if (!member) {
    //  return
    //}
    //
    //if (this.id === id) {
    //  this.$vue.$notify({
    //    group: 'neko',
    //    type: 'info',
    //    title: this.$vue.$t('notifications.controls_taken', { name: this.$vue.$t('you') }) as string,
    //    duration: 5000,
    //    speed: 1000,
    //  })
    //}
  }

  public [EVENT.CONTROL.RELEASE]({ id }: ControlPayload) {
    console.log('EVENT.CONTROL.RELEASE')
    this.emit('host.change', null)
    //this.$accessor.remote.reset()
    //const member = this.member(id)
    //if (!member) {
    //  return
    //}
    //
    //if (this.id === id) {
    //  this.$vue.$notify({
    //    group: 'neko',
    //    type: 'info',
    //    title: this.$vue.$t('notifications.controls_released', { name: this.$vue.$t('you') }) as string,
    //    duration: 5000,
    //    speed: 1000,
    //  })
    //}
  }

  public [EVENT.CONTROL.REQUEST]({ id }: ControlPayload) {
    console.log('EVENT.CONTROL.REQUEST')
    this.emit('control.request', id)
    //const member = this.member(id)
    //if (!member) {
    //  return
    //}
    //
    //this.$vue.$notify({
    //  group: 'neko',
    //  type: 'info',
    //  title: this.$vue.$t('notifications.controls_has', { name: member.displayname }) as string,
    //  text: this.$vue.$t('notifications.controls_has_alt') as string,
    //  duration: 5000,
    //  speed: 1000,
    //})
  }

  public [EVENT.CONTROL.REQUESTING]({ id }: ControlPayload) {
    console.log('EVENT.CONTROL.REQUESTING')
    this.emit('control.requesting', id)
    //const member = this.member(id)
    //if (!member || member.ignored) {
    //  return
    //}
    //
    //this.$vue.$notify({
    //  group: 'neko',
    //  type: 'info',
    //  title: this.$vue.$t('notifications.controls_requesting', { name: member.displayname }) as string,
    //  duration: 5000,
    //  speed: 1000,
    //})
  }

  public [EVENT.CONTROL.GIVE]({ id, target }: ControlTargetPayload) {
    console.log('EVENT.CONTROL.GIVE')
    this.emit('host.change', target)
    //this.$accessor.remote.setHost(target)
    //this.$accessor.remote.changeKeyboard()
  }

  public [EVENT.CONTROL.CLIPBOARD]({ text }: ControlClipboardPayload) {
    console.log('EVENT.CONTROL.CLIPBOARD')
    this.emit('clipboard.update', text)
    //this.$accessor.remote.setClipboard(text)
  }

  /////////////////////////////
  // Screen Events
  /////////////////////////////
  public [EVENT.SCREEN.CONFIGURATIONS]({ configurations }: ScreenConfigurationsPayload) {
    console.log('EVENT.SCREEN.CONFIGURATIONS')
    this.emit('screen.configuration', configurations)
    //this.$accessor.video.setConfigurations(configurations)
  }

  public [EVENT.SCREEN.RESOLUTION]({ id, width, height, rate }: ScreenResolutionPayload) {
    console.log('EVENT.SCREEN.RESOLUTION')
    this.emit('screen.size', width, height, rate)
    //this.$accessor.video.setResolution({ width, height, rate })
  }

  /////////////////////////////
  // Broadcast Events
  /////////////////////////////
  public [EVENT.BROADCAST.STATUS](payload: BroadcastStatusPayload) {
    console.log('EVENT.BROADCAST.STATUS')
    this.emit('broadcast.status', payload)
    //this.$accessor.settings.broadcastStatus(payload)
  }

  /////////////////////////////
  // Admin Events
  /////////////////////////////
  public [EVENT.ADMIN.BAN]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.BAN')
    this.emit('member.ban', id, target)
    // TODO
  }

  public [EVENT.ADMIN.KICK]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.KICK')
    this.emit('member.kick', id, target)
    // TODO
  }

  public [EVENT.ADMIN.MUTE]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.MUTE')
    this.emit('member.muted', id, target)
    //this.$accessor.user.setMuted({ id: target, muted: true })
  }

  public [EVENT.ADMIN.UNMUTE]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.UNMUTE')
    this.emit('member.unmuted', id, target)
    //this.$accessor.user.setMuted({ id: target, muted: false })
  }

  public [EVENT.ADMIN.LOCK]({ id }: AdminPayload) {
    console.log('EVENT.ADMIN.LOCK')
    this.emit('room.locked', id)
    //this.$accessor.setLocked(true)
  }

  public [EVENT.ADMIN.UNLOCK]({ id }: AdminPayload) {
    console.log('EVENT.ADMIN.UNLOCK')
    this.emit('room.unlocked', id)
    //this.$accessor.setLocked(false)
  }

  public [EVENT.ADMIN.CONTROL]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.CONTROL')
    this.emit('host.change', id)
    //this.$accessor.remote.setHost(id)
    //this.$accessor.remote.changeKeyboard()
  }

  public [EVENT.ADMIN.RELEASE]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.RELEASE')
    this.emit('host.change', null)
    //this.$accessor.remote.reset()
  }

  public [EVENT.ADMIN.GIVE]({ id, target }: AdminTargetPayload) {
    if (!target) return

    console.log('EVENT.ADMIN.GIVE')
    this.emit('host.change', target)
    //this.$accessor.remote.setHost(target)
    //this.$accessor.remote.changeKeyboard()
  }

  // Utilities
  //public member(id: string): Member | undefined {
  //  return this.$accessor.user.members[id]
  //}
}
