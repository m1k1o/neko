import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import { BaseClient, BaseEvents } from './base'
import { Member } from './types'
import { EVENT } from './events'
import { accessor } from '~/store'

import {
  SystemMessagePayload,
  MemberListPayload,
  MemberDisconnectPayload,
  MemberPayload,
  ControlPayload,
  ControlTargetPayload,
  ChatPayload,
  EmotePayload,
  ControlClipboardPayload,
  ScreenConfigurationsPayload,
  ScreenResolutionPayload,
  BroadcastStatusPayload,
  AdminTargetPayload,
  AdminLockMessage,
  SystemInitPayload,
  AdminLockResource,
  FileTransferListPayload,
} from './messages'

interface NekoEvents extends BaseEvents {}

export class NekoClient extends BaseClient implements EventEmitter<NekoEvents> {
  private $vue!: Vue
  private $accessor!: typeof accessor
  private url!: string

  init(vue: Vue) {
    const url =
      process.env.NODE_ENV === 'development'
        ? `ws://${location.host.split(':')[0]}:${process.env.VUE_APP_SERVER_PORT}/ws`
        : location.protocol.replace(/^http/, 'ws') + '//' + location.host + location.pathname.replace(/\/$/, '') + '/ws'

    this.initWithURL(vue, url)
  }

  initWithURL(vue: Vue, url: string) {
    this.$vue = vue
    this.$accessor = vue.$accessor
    this.url = url
    // convert ws url to http url
    this.$vue.$http.defaults.baseURL = url.replace(/^ws/, 'http').replace(/\/ws$/, '')
  }

  private cleanup() {
    this.$accessor.setConnected(false)
    this.$accessor.remote.reset()
    this.$accessor.user.reset()
    this.$accessor.video.reset()
    this.$accessor.chat.reset()
  }

  login(password: string, displayname: string) {
    this.connect(this.url, password, displayname)
  }

  logout() {
    this.disconnect()
    this.cleanup()
    this.$vue.$swal({
      title: this.$vue.$t('connection.logged_out'),
      icon: 'info',
      confirmButtonText: this.$vue.$t('connection.button_confirm') as string,
    })
  }

  /////////////////////////////
  // Internal Events
  /////////////////////////////
  protected [EVENT.RECONNECTING]() {
    this.$vue.$notify({
      group: 'neko',
      type: 'warning',
      title: this.$vue.$t('connection.reconnecting') as string,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.CONNECTING]() {
    this.$accessor.setConnnecting()
  }

  protected [EVENT.CONNECTED]() {
    this.$accessor.user.setMember(this.id)
    this.$accessor.setConnected(true)

    this.$vue.$notify({
      group: 'neko',
      clean: true,
    })

    this.$vue.$notify({
      group: 'neko',
      type: 'success',
      title: this.$vue.$t('connection.connected') as string,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.DISCONNECTED](reason?: Error) {
    this.cleanup()

    this.$vue.$notify({
      group: 'neko',
      type: 'error',
      title: this.$vue.$t('connection.disconnected') as string,
      text: reason ? reason.message : undefined,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.TRACK](event: RTCTrackEvent) {
    const { track, streams } = event
    if (track.kind === 'audio') {
      return
    }

    this.$accessor.video.addTrack([track, streams[0]])
    this.$accessor.video.setStream(0)
  }

  protected [EVENT.DATA]() {}

  /////////////////////////////
  // System Events
  /////////////////////////////
  protected [EVENT.SYSTEM.INIT]({ implicit_hosting, locks, file_transfer, heartbeat_interval }: SystemInitPayload) {
    this.$accessor.remote.setImplicitHosting(implicit_hosting)
    this.$accessor.remote.setFileTransfer(file_transfer)

    for (const resource in locks) {
      this[EVENT.ADMIN.LOCK]({
        event: EVENT.ADMIN.LOCK,
        resource: resource as AdminLockResource,
        id: locks[resource],
      })
    }

    if (heartbeat_interval > 0) {
      if (this._ws_heartbeat) clearInterval(this._ws_heartbeat)
      this._ws_heartbeat = window.setInterval(() => this.sendMessage(EVENT.CLIENT.HEARTBEAT), heartbeat_interval * 1000)
    }
  }

  protected [EVENT.SYSTEM.DISCONNECT]({ message }: SystemMessagePayload) {
    if (message == 'kicked') {
      this.$accessor.logout()
      message = this.$vue.$t('connection.kicked') as string
    }

    this.onDisconnected(new Error(message))

    this.$vue.$swal({
      title: this.$vue.$t('connection.disconnected'),
      text: message,
      icon: 'error',
      confirmButtonText: this.$vue.$t('connection.button_confirm') as string,
    })
  }

  protected [EVENT.SYSTEM.ERROR]({ title, message }: SystemMessagePayload) {
    this.$vue.$swal({
      title,
      text: message,
      icon: 'error',
      confirmButtonText: this.$vue.$t('connection.button_confirm') as string,
    })
  }

  /////////////////////////////
  // Member Events
  /////////////////////////////
  protected [EVENT.MEMBER.LIST]({ members }: MemberListPayload) {
    this.$accessor.user.setMembers(members)
    this.$accessor.chat.newMessage({
      id: this.id,
      content: this.$vue.$t('notifications.connected', { name: '' }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.MEMBER.CONNECTED](member: MemberPayload) {
    this.$accessor.user.addMember(member)

    if (member.id !== this.id) {
      this.$accessor.chat.newMessage({
        id: member.id,
        content: this.$vue.$t('notifications.connected', { name: '' }) as string,
        type: 'event',
        created: new Date(),
      })
    }
  }

  protected [EVENT.MEMBER.DISCONNECTED]({ id }: MemberDisconnectPayload) {
    const member = this.member(id)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id: member.id,
      content: this.$vue.$t('notifications.disconnected', { name: '' }) as string,
      type: 'event',
      created: new Date(),
    })

    this.$accessor.user.delMember(id)
  }

  /////////////////////////////
  // Control Events
  /////////////////////////////
  protected [EVENT.CONTROL.LOCKED]({ id }: ControlPayload) {
    this.$accessor.remote.setHost(id)
    this.$accessor.remote.changeKeyboard()

    const member = this.member(id)
    if (!member) {
      return
    }

    if (this.id === id) {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: this.$vue.$t('notifications.controls_taken', {
          name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
        }) as string,
        duration: 5000,
        speed: 1000,
      })
    }

    this.$accessor.chat.newMessage({
      id: member.id,
      content: this.$vue.$t('notifications.controls_taken', { name: '' }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.CONTROL.RELEASE]({ id }: ControlPayload) {
    this.$accessor.remote.reset()
    const member = this.member(id)
    if (!member) {
      return
    }

    if (this.id === id) {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: this.$vue.$t('notifications.controls_released', {
          name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
        }) as string,
        duration: 5000,
        speed: 1000,
      })
    }

    this.$accessor.chat.newMessage({
      id: member.id,
      content: this.$vue.$t('notifications.controls_released', { name: '' }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.CONTROL.REQUEST]({ id }: ControlPayload) {
    const member = this.member(id)
    if (!member) {
      return
    }

    this.$vue.$notify({
      group: 'neko',
      type: 'info',
      title: this.$vue.$t('notifications.controls_has', { name: member.displayname }) as string,
      text: this.$vue.$t('notifications.controls_has_alt') as string,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.CONTROL.REQUESTING]({ id }: ControlPayload) {
    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.$vue.$notify({
      group: 'neko',
      type: 'info',
      title: this.$vue.$t('notifications.controls_requesting', { name: member.displayname }) as string,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.CONTROL.GIVE]({ id, target }: ControlTargetPayload) {
    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.remote.setHost(member)
    this.$accessor.remote.changeKeyboard()

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.controls_given', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.CONTROL.CLIPBOARD]({ text }: ControlClipboardPayload) {
    this.$accessor.remote.setClipboard(text)
  }

  /////////////////////////////
  // Chat Events
  /////////////////////////////
  protected [EVENT.CHAT.MESSAGE]({ id, content }: ChatPayload) {
    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content,
      type: 'text',
      created: new Date(),
    })
  }

  protected [EVENT.CHAT.EMOTE]({ id, emote }: EmotePayload) {
    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.$accessor.chat.newEmote({ type: emote })
  }

  /////////////////////////////
  // File Transfer Events
  /////////////////////////////
  protected [EVENT.FILETRANSFER.LIST]({ cwd, files }: FileTransferListPayload) {
    this.$accessor.files.setCwd(cwd)
    this.$accessor.files.setFileList(files)
  }

  /////////////////////////////
  // Screen Events
  /////////////////////////////
  protected [EVENT.SCREEN.CONFIGURATIONS]({ configurations }: ScreenConfigurationsPayload) {
    this.$accessor.video.setConfigurations(configurations)
  }

  protected [EVENT.SCREEN.RESOLUTION]({ id, width, height, rate }: ScreenResolutionPayload) {
    this.$accessor.video.setResolution({ width, height, rate })

    if (!id) {
      return
    }

    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.resolution', {
        width: width,
        height: height,
        rate: rate,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  /////////////////////////////
  // Broadcast Events
  /////////////////////////////
  protected [EVENT.BROADCAST.STATUS](payload: BroadcastStatusPayload) {
    this.$accessor.settings.broadcastStatus(payload)
  }

  /////////////////////////////
  // Admin Events
  /////////////////////////////
  protected [EVENT.ADMIN.BAN]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.banned', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.KICK]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.kicked', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.MUTE]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    this.$accessor.user.setMuted({ id: target, muted: true })

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.muted', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.UNMUTE]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    this.$accessor.user.setMuted({ id: target, muted: false })

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.unmuted', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.LOCK]({ id, resource }: AdminLockMessage) {
    this.$accessor.setLocked(resource)

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t(`locks.${resource}.notif_locked`) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.UNLOCK]({ id, resource }: AdminLockMessage) {
    this.$accessor.setUnlocked(resource)

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t(`locks.${resource}.notif_unlocked`) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.CONTROL]({ id, target }: AdminTargetPayload) {
    this.$accessor.remote.setHost(id)
    this.$accessor.remote.changeKeyboard()

    if (!target) {
      this.$accessor.chat.newMessage({
        id,
        content: this.$vue.$t('notifications.controls_taken_force') as string,
        type: 'event',
        created: new Date(),
      })
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.controls_taken_steal', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.RELEASE]({ id, target }: AdminTargetPayload) {
    this.$accessor.remote.reset()
    if (!target) {
      this.$accessor.chat.newMessage({
        id,
        content: this.$vue.$t('notifications.controls_released_force') as string,
        type: 'event',
        created: new Date(),
      })
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.controls_released_steal', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.GIVE]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.remote.setHost(member)
    this.$accessor.remote.changeKeyboard()

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.controls_given', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  // Utilities
  protected member(id: string): Member | undefined {
    return this.$accessor.user.members[id]
  }
}
