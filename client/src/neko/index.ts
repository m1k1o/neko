import EventEmitter from 'eventemitter3'
import { BaseClient, BaseEvents } from './base'
import { EVENT } from './events'
import { AuthClient } from '~/sdk/auth'
import { RoomClient } from '~/sdk/room'
import { NetworkQualityMonitor } from '~/sdk/network-monitor'
import { set } from '~/utils/localstorage'
import { NekoClientRuntime } from './runtime'

import {
  SystemMessagePayload,
  ChatPayload,
  ChatInitPayload,
  EmotePayload,
  ScreenResolutionPayload,
  BroadcastStatusPayload,
  SystemInitPayload,
  SystemAdminPayload,
  SystemSettingsPayload,
  SessionDataPayload,
  SessionIdPayload,
  SessionProfilePayload,
  SessionStatePayload,
  FileTransferUpdatePayload,
} from './messages'

interface NekoEvents extends BaseEvents {}

export class NekoClient extends BaseClient implements EventEmitter<NekoEvents> {
  private runtime!: NekoClientRuntime
  private auth!: AuthClient
  private roomClient!: RoomClient
  private url!: string
  private apiURL = ''
  private networkMonitor?: NetworkQualityMonitor

  init(runtime: NekoClientRuntime) {
    const url =
      process.env.NODE_ENV === 'development'
        ? `ws://${location.host.split(':')[0]}:${process.env.VUE_APP_SERVER_PORT}/api/ws`
        : location.protocol.replace(/^http/, 'ws') +
          '//' +
          location.host +
          location.pathname.replace(/\/$/, '') +
          '/api/ws'

    this.initWithURL(runtime, url)
  }

  initWithURL(runtime: NekoClientRuntime, url: string) {
    this.runtime = runtime
    this.url = url
    const httpURL = url.replace(/^ws/, 'http')
    this.apiURL = httpURL.replace(/\/api\/ws$/, '/api')
    // Keep static assets (emoji, keyboard layouts) on the application root.
    this.runtime.http.defaults.baseURL = httpURL.replace(/\/api\/ws$/, '')
    this.runtime.http.defaults.withCredentials = true
    this.auth = new AuthClient(this.runtime.http, this.apiURL)
    this.roomClient = new RoomClient(this.runtime.http, this.apiURL)
  }

  get room() {
    return this.roomClient
  }

  private get state() {
    return this.runtime.state
  }

  private get ui() {
    return this.runtime.ui
  }

  private cleanup() {
    this.stopNetworkMonitor()
    this.state.connection.setConnected(false)
    this.state.remote.reset()
    this.state.user.reset()
    this.state.video.reset()
    this.state.chat.reset()
  }

  async login(password: string, displayname: string) {
    if (this.state.connection.connecting) {
      return
    }

    this.state.connection.setConnecting()
    try {
      const token = await this.auth.login(displayname, password)
      this.connect(this.url, token)
    } catch (error) {
      const reason = this.toError(error)
      this.auth.clear()
      this.state.connection.setConnected(false)
      this.state.connection.setError(reason.message)
    }
  }

  async logout() {
    this.disconnect()
    this.cleanup()
    try {
      await this.auth.logout()
    } catch (error) {
      // A closed session is already safe to discard locally.
    }
    this.ui.alert({
      title: this.ui.translate('connection.logged_out'),
      icon: 'info',
      confirmButtonText: this.ui.translate('connection.button_confirm'),
    })
  }

  private toError(error: unknown): Error {
    const response = (error as { response?: { data?: { message?: string } } })?.response
    const message = response?.data?.message
    if (message) {
      return new Error(message)
    }
    if (error instanceof Error) {
      return error
    }
    return new Error('login request failed')
  }

  /////////////////////////////
  // Internal Events
  /////////////////////////////
  protected [EVENT.RECONNECTING]() {
    this.state.connection.setState('reconnecting')
    this.ui.notify({
      group: 'neko',
      type: 'warning',
      title: this.ui.translate('connection.reconnecting'),
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.CONNECTING]() {
    this.state.connection.setConnecting()
  }

  protected [EVENT.CONNECTED]() {
    this.state.user.setMember(this.id)
    this.state.connection.setConnected(true)
    // Screen metadata moved from the deprecated websocket events to the REST
    // room API. Load it after the session is authenticated so pointer mapping
    // is based on the actual desktop size instead of the 1280x720 defaults.
    void this.state.video.screenGet().catch((error: unknown) => {
      this.ui.log.warn('failed to load the current screen size', error)
    })
    if (this.state.user.admin) {
      void this.state.video.screenConfigurations().catch((error: unknown) => {
        this.ui.log.warn('failed to load screen configurations', error)
      })
    }
    set('displayname', this.state.session.displayname)
    set('password', this.state.session.password)
    this.startNetworkMonitor()

    this.ui.notify({
      group: 'neko',
      clean: true,
    })

    this.ui.notify({
      group: 'neko',
      type: 'success',
      title: this.ui.translate('connection.connected'),
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.DISCONNECTED](reason?: Error) {
    if (!this.state.connection.connected && reason) {
      this.state.connection.setError(reason.message)
    }
    this.cleanup()

    this.ui.notify({
      group: 'neko',
      type: 'error',
      title: this.ui.translate('connection.disconnected'),
      text: reason ? reason.message : undefined,
      duration: 5000,
      speed: 1000,
    })
  }

  private startNetworkMonitor() {
    this.stopNetworkMonitor()
    this.networkMonitor = new NetworkQualityMonitor({
      onSample: ({ quality, rtt }) => this.state.connection.setNetworkQuality({ quality, rtt }),
    })
    if (this._peer) {
      this.networkMonitor.start(this._peer)
    }
  }

  private stopNetworkMonitor() {
    this.networkMonitor?.stop()
    this.networkMonitor = undefined
  }

  protected [EVENT.TRACK](event: RTCTrackEvent) {
    const { track, streams } = event
    if (track.kind === 'audio') {
      return
    }

    this.state.video.addTrack([track, streams[0]])
    this.state.video.setStream(0)
  }

  protected [EVENT.DATA]() {}

  /////////////////////////////
  // System Events
  /////////////////////////////
  protected [EVENT.SYSTEM.INIT]({ session_id, control_host, screen_size, sessions, settings }: SystemInitPayload) {
    // The websocket has authenticated the session at this point. Allow the
    // login surface to leave while WebRTC continues negotiating in parallel.
    this.state.connection.setAuthenticated(true)
    this._id = session_id
    this.state.user.setMember(session_id)
    this.setControlEpoch(control_host.epoch)
    this.state.remote.setEpoch(control_host.epoch)
    this.state.video.setResolution(screen_size)
    this.state.remote.setHost(control_host.has_host ? control_host.host_id || '' : '')
    this.state.remote.setImplicitHosting(settings.implicit_hosting)
    this.state.remote.setLocked(settings.locked_controls)
    this.setLockState('login', settings.locked_logins)
    this.setLockState('control', settings.locked_controls)
    this.setLockState('file_transfer', settings.plugins?.['filetransfer.enabled'] === false)
    this.state.user.setMembers(
      Object.values(sessions).map((session) => ({
        id: session.id,
        displayname: session.profile.name,
        avatar: session.profile.avatar,
        admin: session.profile.is_admin,
        muted: false,
        connected: session.state.is_connected,
      })),
    )

    if (settings.heartbeat_interval > 0) {
      if (this._ws_heartbeat) clearInterval(this._ws_heartbeat)
      this._ws_heartbeat = window.setInterval(() => {
        this.sendMessage(EVENT.CLIENT.HEARTBEAT)
      }, settings.heartbeat_interval * 1000)
    }

    if (this._control_heartbeat) clearInterval(this._control_heartbeat)
    if (settings.control_lease_ttl > 0) {
      const renewInterval = Math.max(1000, Math.floor((settings.control_lease_ttl * 1000) / 3))
      this._control_heartbeat = window.setInterval(() => {
        if (this.state.remote.controlling) {
          this.sendMessage(EVENT.CONTROL.RENEW, { epoch: this._controlEpoch })
        }
      }, renewInterval)
    }
  }

  protected [EVENT.SYSTEM.ADMIN]({ broadcast_status }: SystemAdminPayload) {
    this.state.settings.broadcastStatus({
      url: broadcast_status.url,
      isActive: broadcast_status.is_active,
    })
  }

  protected [EVENT.SYSTEM.SETTINGS](settings: SystemSettingsPayload) {
    this.state.remote.setImplicitHosting(settings.implicit_hosting)
    this.state.remote.setLocked(settings.locked_controls)
    this.setLockState('login', settings.locked_logins)
    this.setLockState('control', settings.locked_controls)
    this.setLockState('file_transfer', settings.plugins?.['filetransfer.enabled'] === false)
  }

  private setLockState(resource: 'login' | 'control' | 'file_transfer', locked: boolean) {
    if (locked) {
      this.state.session.setLocked(resource)
    } else {
      this.state.session.setUnlocked(resource)
    }
  }

  protected [EVENT.CONTROL.HOST]({ has_host, host_id, epoch }: { has_host: boolean; host_id?: string; epoch: number }) {
    this.setControlEpoch(epoch)
    this.state.remote.setEpoch(epoch)
    this.state.remote.setHost(has_host ? host_id || '' : '')
  }

  protected [EVENT.SYSTEM.DISCONNECT]({ message }: SystemMessagePayload) {
    if (message == 'kicked') {
      this.state.session.logout()
      message = this.ui.translate('connection.kicked')
    }

    if (!this.state.connection.connected && message) {
      this.state.connection.setError(message)
    }

    this.onDisconnected(new Error(message))

    this.ui.alert({
      title: this.ui.translate('connection.disconnected'),
      text: message,
      icon: 'error',
      confirmButtonText: this.ui.translate('connection.button_confirm'),
    })
  }

  protected [EVENT.SYSTEM.ERROR]({ title, message }: SystemMessagePayload) {
    if (!this.state.connection.connected && message) {
      this.state.connection.setError(message)
    }

    this.ui.alert({
      title,
      text: message,
      icon: 'error',
      confirmButtonText: this.ui.translate('connection.button_confirm'),
    })
  }

  /////////////////////////////
  // Session Events
  /////////////////////////////
  protected [EVENT.SESSION.CREATED](session: SessionDataPayload) {
    this.state.user.addMember({
      id: session.id,
      displayname: session.profile.name,
      avatar: session.profile.avatar,
      admin: session.profile.is_admin,
      muted: false,
      connected: session.state.is_connected,
    })
  }

  protected [EVENT.SESSION.DELETED]({ id }: SessionIdPayload) {
    this.state.user.delMember(id)
  }

  protected [EVENT.SESSION.PROFILE]({ id, name, is_admin, avatar }: SessionProfilePayload) {
    const member = this.member(id)
    if (!member) return
    this.state.user.addMember({ ...member, displayname: name, avatar, admin: is_admin })
  }

  protected [EVENT.SESSION.STATE]({ id, is_connected }: SessionStatePayload) {
    const member = this.member(id)
    if (!member) return
    if (is_connected) {
      this.state.user.addMember({ ...member, connected: true })
    } else {
      this.state.user.delMember(id)
    }
  }

  protected [EVENT.SESSION.CURSORS]() {}

  /////////////////////////////
  // Control Events
  /////////////////////////////
  protected [EVENT.CONTROL.RELEASE]({ id }: SessionIdPayload) {
    if (id === this.id) this.state.remote.reset()
  }

  protected [EVENT.CONTROL.REQUEST]({ id }: SessionIdPayload) {
    const member = this.member(id)
    if (!member) {
      return
    }

    this.ui.notify({
      group: 'neko',
      type: 'info',
      title: this.ui.translate('notifications.controls_has', { name: member.displayname }),
      text: this.ui.translate('notifications.controls_has_alt'),
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.CLIPBOARD.UPDATED]({ text }: { text: string }) {
    this.state.remote.setClipboard(text)
  }

  /////////////////////////////
  // Chat Events
  /////////////////////////////
  protected [EVENT.CHAT.INIT]({ history }: ChatInitPayload) {
    this.state.chat.restoreHistory(
      (history || []).map((message) => ({
        id: message.id,
        content: this.chatText(message.content),
        name: message.name,
        avatar: message.avatar,
        type: 'text' as const,
        created: message.created ? new Date(message.created) : new Date(),
      })),
    )
  }

  protected [EVENT.CHAT.MESSAGE]({ id, content, created, name, avatar }: ChatPayload) {
    const member = this.member(id)
    if (member && member.ignored) {
      return
    }

    this.state.chat.newMessage({
      id,
      content: this.chatText(content),
      name: name || member?.displayname,
      avatar: avatar || member?.avatar,
      type: 'text',
      created: created ? new Date(created) : new Date(),
    })
  }

  private chatText(content: string | { text: string }) {
    return typeof content === 'string' ? content : content.text
  }

  protected [EVENT.CHAT.EMOTE]({ id, emote }: EmotePayload) {
    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.state.chat.newEmote({ type: emote })
  }

  /////////////////////////////
  // File Transfer Events
  /////////////////////////////
  protected [EVENT.FILETRANSFER.UPDATE]({
    root_dir,
    enabled,
    user_download,
    user_upload,
    user_delete,
    files,
  }: FileTransferUpdatePayload) {
    this.state.files.setCwd(root_dir)
    this.state.files.setLoading(false)
    this.state.files.setFileList(files)
    this.state.files.setUserDownload(user_download)
    this.state.files.setUserUpload(user_upload)
    this.state.files.setUserDelete(user_delete)
    this.state.remote.setFileTransfer(enabled)
  }

  /////////////////////////////
  // Open in App Events
  /////////////////////////////
  protected [EVENT.OPENINAPP.INIT]({ enabled }: { enabled: boolean }) {
    this.state.openinapp.setEnabled(enabled)
  }

  /////////////////////////////
  // Screen Events
  /////////////////////////////
  protected [EVENT.SCREEN.UPDATED]({ id, width, height, rate }: ScreenResolutionPayload) {
    this.state.video.setResolution({ width, height, rate })

    if (!id) {
      return
    }

    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.state.chat.newMessage({
      id,
      content: this.ui.translate('notifications.resolution', {
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
    this.state.settings.broadcastStatus({ url: payload.url, isActive: payload.is_active })
  }

  // Utilities
  protected member(id: string) {
    return this.state.user.members[id]
  }
}
