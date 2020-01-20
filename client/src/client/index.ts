import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import { BaseClient, BaseEvents } from './base'

import { EVENT } from './events'
import { accessor } from '~/store'
import { IdentityPayload, MemberListPayload, MemberDisconnectPayload, MemberPayload, ControlPayload } from './messages'

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

  /////////////////////////////
  // Internal Events
  /////////////////////////////
  [EVENT.CONNECTING]() {
    this.$accessor.setConnnecting(true)
  }

  [EVENT.CONNECTED]() {
    this.$accessor.setConnected(true)
    this.$accessor.setConnnecting(false)
    this.$accessor.video.clearStream()
    this.$accessor.remote.clearHost()
    this.$vue.$notify({
      group: 'neko',
      type: 'success',
      title: 'Successfully connected',
      duration: 5000,
      speed: 1000,
    })
  }

  [EVENT.DISCONNECTED](reason?: Error) {
    this.$accessor.setConnected(false)
    this.$accessor.setConnnecting(false)
    this.$accessor.video.clearStream()
    this.$accessor.user.clearMembers()
    this.$vue.$notify({
      group: 'neko',
      type: 'error',
      title: `Disconnected`,
      text: reason ? reason.message : undefined,
      duration: 5000,
      speed: 1000,
    })
  }

  [EVENT.TRACK](event: RTCTrackEvent) {
    if (event.track.kind === 'audio') {
      return
    }
    this.$accessor.video.addStream(event.streams[0])
    this.$accessor.video.setStream(0)
  }

  [EVENT.DATA](data: any) {}

  /////////////////////////////
  // Identity Events
  /////////////////////////////
  [EVENT.IDENTITY.PROVIDE]({ id }: IdentityPayload) {
    this.$accessor.user.setMember(id)
  }

  /////////////////////////////
  // Member Events
  /////////////////////////////
  [EVENT.MEMBER.LIST]({ members }: MemberListPayload) {
    this.$accessor.user.setMembers(members)
  }

  [EVENT.MEMBER.CONNECTED](member: MemberPayload) {
    this.$accessor.user.addMember(member)

    if (member.id !== this.$accessor.user.id) {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: `${member.username} connected`,
        duration: 5000,
        speed: 1000,
      })
    }
  }

  [EVENT.MEMBER.DISCONNECTED]({ id }: MemberDisconnectPayload) {
    this.$vue.$notify({
      group: 'neko',
      type: 'info',
      title: `${this.$accessor.user.members[id].username} disconnected`,
      duration: 5000,
      speed: 1000,
    })
    this.$accessor.user.delMember(id)
  }

  /////////////////////////////
  // Control Events
  /////////////////////////////
  [EVENT.CONTROL.LOCKED]({ id }: ControlPayload) {
    this.$accessor.remote.setHost(id)
    if (this.$accessor.user.id === id) {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: `You have the controls`,
        duration: 5000,
        speed: 1000,
      })
    } else {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: `${this.$accessor.user.members[id].username} took the controls`,
        duration: 5000,
        speed: 1000,
      })
    }
  }

  [EVENT.CONTROL.RELEASE]({ id }: ControlPayload) {
    this.$accessor.remote.clearHost()
    if (this.$accessor.user.id === id) {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: `You released the controls`,
        duration: 5000,
        speed: 1000,
      })
    } else {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: `The controls released from ${this.$accessor.user.members[id].username}`,
        duration: 5000,
        speed: 1000,
      })
    }
  }

  [EVENT.CONTROL.REQUEST]({ id }: ControlPayload) {
    this.$vue.$notify({
      group: 'neko',
      type: 'info',
      title: `${this.$accessor.user.members[id].username} has the controls`,
      text: 'But I let them know you wanted it',
      duration: 5000,
      speed: 1000,
    })
  }

  [EVENT.CONTROL.REQUESTING]({ id }: ControlPayload) {
    this.$vue.$notify({
      group: 'neko',
      type: 'info',
      title: `${this.$accessor.user.members[id].username} is requesting the controls`,
      duration: 5000,
      speed: 1000,
    })
  }
}
