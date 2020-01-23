import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import { BaseClient, BaseEvents } from './base'
import { Member } from './types'
import { EVENT } from './events'
import { accessor } from '~/store'

import {
  DisconnectPayload,
  IdentityPayload,
  MemberListPayload,
  MemberDisconnectPayload,
  MemberPayload,
  ControlPayload,
  ControlTargetPayload,
  ChatPayload,
  EmotePayload,
  AdminPayload,
  AdminTargetPayload,
} from './messages'

interface NekoEvents extends BaseEvents {}

export class NekoClient extends BaseClient implements EventEmitter<NekoEvents> {
  private $vue!: Vue
  private $accessor!: typeof accessor

  init(vue: Vue) {
    this.$vue = vue
    this.$accessor = vue.$accessor
  }

  connect(password: string, username: string) {
    const url =
      process.env.NODE_ENV === 'development'
        ? `ws://${process.env.VUE_APP_SERVER}/`
        : `${/https/gi.test(location.protocol) ? 'wss' : 'ws'}://${location.host}/`

    super.connect(url, password, username)
  }

  private get id() {
    return this.$accessor.user.id
  }

  /////////////////////////////
  // Internal Events
  /////////////////////////////
  protected [EVENT.CONNECTING]() {
    this.$accessor.setConnnecting()
  }

  protected [EVENT.CONNECTED]() {
    this.$accessor.setConnected(true)
    this.$accessor.setConnected(true)

    this.$vue.$notify({
      group: 'neko',
      type: 'success',
      title: 'Successfully connected',
      duration: 5000,
      speed: 1000,
    })

    this.$accessor.chat.newMessage({
      id: this.id,
      content: 'connected',
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.DISCONNECTED](reason?: Error) {
    this.$accessor.setConnected(false)

    this.$accessor.remote.clear()
    this.$accessor.user.clear()
    this.$accessor.video.clear()
    this.$accessor.chat.clear()

    this.$vue.$notify({
      group: 'neko',
      type: 'error',
      title: `Disconnected:`,
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

  protected [EVENT.DATA](data: any) {}

  /////////////////////////////
  // System Events
  /////////////////////////////
  protected [EVENT.SYSTEM.DISCONNECT]({ message }: DisconnectPayload) {
    this.onDisconnected(new Error(message))
    this.$vue.$swal({
      title: 'Disconnected!',
      text: message,
      icon: 'error',
      confirmButtonText: 'ok',
    })
  }

  /////////////////////////////
  // Identity Events
  /////////////////////////////
  protected [EVENT.IDENTITY.PROVIDE]({ id }: IdentityPayload) {
    this.$accessor.user.setMember(id)
  }

  /////////////////////////////
  // Member Events
  /////////////////////////////
  protected [EVENT.MEMBER.LIST]({ members }: MemberListPayload) {
    this.$accessor.user.setMembers(members)
  }

  protected [EVENT.MEMBER.CONNECTED](member: MemberPayload) {
    this.$accessor.user.addMember(member)

    if (member.id !== this.id) {
      this.$accessor.chat.newMessage({
        id: member.id,
        content: 'connected',
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
      content: 'disconnected',
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
    const member = this.member(id)
    if (!member) {
      return
    }

    if (this.id === id) {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: `You have the controls`,
        duration: 5000,
        speed: 1000,
      })
    }

    this.$accessor.chat.newMessage({
      id: member.id,
      content: 'took the controls',
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.CONTROL.RELEASE]({ id }: ControlPayload) {
    this.$accessor.remote.clear()
    const member = this.member(id)
    if (!member) {
      return
    }

    if (this.id === id) {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: `You released the controls`,
        duration: 5000,
        speed: 1000,
      })
    }

    this.$accessor.chat.newMessage({
      id: member.id,
      content: 'released the controls',
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
      title: `${member.username} has the controls`,
      text: 'But I let them know you wanted it',
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
      title: `${member.username} is requesting the controls`,
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
    this.$accessor.chat.newMessage({
      id,
      content: `gave the controls to ${member.id == this.id ? 'you' : member.username}`,
      type: 'event',
      created: new Date(),
    })
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
      content: `banned ${member.id == this.id ? 'you' : member.username}`,
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
      content: `kicked ${member.id == this.id ? 'you' : member.username}`,
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
      content: `muted ${member.id == this.id ? 'you' : member.username}`,
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
      content: `unmuted ${member.username}`,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.LOCK]({ id }: AdminPayload) {
    this.$accessor.setLocked(true)
    this.$accessor.chat.newMessage({
      id,
      content: `locked the room`,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.UNLOCK]({ id }: AdminPayload) {
    this.$accessor.setLocked(false)
    this.$accessor.chat.newMessage({
      id,
      content: `unlocked the room`,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.CONTROL]({ id, target }: AdminTargetPayload) {
    this.$accessor.remote.setHost(id)

    if (!target) {
      this.$accessor.chat.newMessage({
        id,
        content: `force took the controls`,
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
      content: `took the controls from ${member.id == this.id ? 'you' : member.username}`,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.RELEASE]({ id, target }: AdminTargetPayload) {
    this.$accessor.remote.clear()
    if (!target) {
      this.$accessor.chat.newMessage({
        id,
        content: `force released the controls`,
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
      content: `released the controls from ${member.id == this.id ? 'you' : member.username}`,
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

    this.$accessor.chat.newMessage({
      id,
      content: `gave the controls to ${member.id == this.id ? 'you' : member.username}`,
      type: 'event',
      created: new Date(),
    })
  }

  // Utilities
  protected member(id: string): Member | undefined {
    return this.$accessor.user.members[id]
  }
}
