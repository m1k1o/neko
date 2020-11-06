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

export class NekoMessages {
  /////////////////////////////
  // System Events
  /////////////////////////////
  public [EVENT.SYSTEM.DISCONNECT]({ message }: DisconnectPayload) {
    console.log('EVENT.SYSTEM.DISCONNECT')
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
    //this.$accessor.user.setMembers(members)
  }

  public [EVENT.MEMBER.CONNECTED](member: MemberPayload) {
    console.log('EVENT.MEMBER.CONNECTED')
    //this.$accessor.user.addMember(member)
  }

  public [EVENT.MEMBER.DISCONNECTED]({ id }: MemberDisconnectPayload) {
    console.log('EVENT.MEMBER.DISCONNECTED')
    //const member = this.member(id)
    //if (!member) {
    //  return
    //}
    //
    //this.$accessor.user.delMember(id)
  }

  /////////////////////////////
  // Control Events
  /////////////////////////////
  public [EVENT.CONTROL.LOCKED]({ id }: ControlPayload) {
    console.log('EVENT.CONTROL.LOCKED')
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
    //const member = this.member(target)
    //if (!member) {
    //  return
    //}
    //
    //this.$accessor.remote.setHost(member)
    //this.$accessor.remote.changeKeyboard()
  }

  public [EVENT.CONTROL.CLIPBOARD]({ text }: ControlClipboardPayload) {
    console.log('EVENT.CONTROL.CLIPBOARD')
    //this.$accessor.remote.setClipboard(text)
  }

  /////////////////////////////
  // Screen Events
  /////////////////////////////
  public [EVENT.SCREEN.CONFIGURATIONS]({ configurations }: ScreenConfigurationsPayload) {
    console.log('EVENT.SCREEN.CONFIGURATIONS')
    //this.$accessor.video.setConfigurations(configurations)
  }

  public [EVENT.SCREEN.RESOLUTION]({ id, width, height, rate }: ScreenResolutionPayload) {
    console.log('EVENT.SCREEN.RESOLUTION')
    //this.$accessor.video.setResolution({ width, height, rate })
  }

  /////////////////////////////
  // Broadcast Events
  /////////////////////////////
  public [EVENT.BROADCAST.STATUS](payload: BroadcastStatusPayload) {
    console.log('EVENT.BROADCAST.STATUS')
    //this.$accessor.settings.broadcastStatus(payload)
  }

  /////////////////////////////
  // Admin Events
  /////////////////////////////
  public [EVENT.ADMIN.BAN]({ id, target }: AdminTargetPayload) {
    console.log('EVENT.ADMIN.BAN')
    // TODO
  }

  public [EVENT.ADMIN.KICK]({ id, target }: AdminTargetPayload) {
    console.log('EVENT.ADMIN.KICK')
    // TODO
  }

  public [EVENT.ADMIN.MUTE]({ id, target }: AdminTargetPayload) {
    console.log('EVENT.ADMIN.MUTE')
    //if (!target) {
    //  return
    //}
    //
    //this.$accessor.user.setMuted({ id: target, muted: true })
  }

  public [EVENT.ADMIN.UNMUTE]({ id, target }: AdminTargetPayload) {
    console.log('EVENT.ADMIN.UNMUTE')
    //if (!target) {
    //  return
    //}
    //
    //this.$accessor.user.setMuted({ id: target, muted: false })
  }

  public [EVENT.ADMIN.LOCK]({ id }: AdminPayload) {
    console.log('EVENT.ADMIN.LOCK')
    //this.$accessor.setLocked(true)
  }

  public [EVENT.ADMIN.UNLOCK]({ id }: AdminPayload) {
    console.log('EVENT.ADMIN.UNLOCK')
    //this.$accessor.setLocked(false)
  }

  public [EVENT.ADMIN.CONTROL]({ id, target }: AdminTargetPayload) {
    console.log('EVENT.ADMIN.CONTROL')
    //this.$accessor.remote.setHost(id)
    //this.$accessor.remote.changeKeyboard()
  }

  public [EVENT.ADMIN.RELEASE]({ id, target }: AdminTargetPayload) {
    console.log('EVENT.ADMIN.RELEASE')
    //this.$accessor.remote.reset()
  }

  public [EVENT.ADMIN.GIVE]({ id, target }: AdminTargetPayload) {
    console.log('EVENT.ADMIN.GIVE')
    //if (!target) {
    //  return
    //}
    //
    //const member = this.member(target)
    //if (member) {
    //  this.$accessor.remote.setHost(member)
    //  this.$accessor.remote.changeKeyboard()
    //}
  }

  // Utilities
  //public member(id: string): Member | undefined {
  //  return this.$accessor.user.members[id]
  //}
}
