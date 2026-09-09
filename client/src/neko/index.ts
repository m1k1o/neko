import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import { BaseClient, BaseEvents } from './base'
import { EVENT } from './events'
import { accessor } from '~/store'
import { NetworkQuality } from '~/store/connection'
import { set } from '~/utils/localstorage'

import {
  SystemMessagePayload,
  ChatPayload,
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
  private $vue!: Vue
  private $accessor!: typeof accessor
  private url!: string
  private apiURL = ''
  private token = ''
  private networkMonitor?: number
  private previousNetworkCounters?: { packetsReceived: number; packetsLost: number }

  init(vue: Vue) {
    const url =
      process.env.NODE_ENV === 'development'
        ? `ws://${location.host.split(':')[0]}:${process.env.VUE_APP_SERVER_PORT}/api/ws`
        : location.protocol.replace(/^http/, 'ws') +
          '//' +
          location.host +
          location.pathname.replace(/\/$/, '') +
          '/api/ws'

    this.initWithURL(vue, url)
  }

  initWithURL(vue: Vue, url: string) {
    this.$vue = vue
    this.$accessor = vue.$accessor
    this.url = url
    const httpURL = url.replace(/^ws/, 'http')
    this.apiURL = httpURL.replace(/\/api\/ws$/, '/api')
    // Keep static assets (emoji, keyboard layouts) on the application root.
    this.$vue.$http.defaults.baseURL = httpURL.replace(/\/api\/ws$/, '')
    this.$vue.$http.defaults.withCredentials = true
  }

  private cleanup() {
    this.stopNetworkMonitor()
    this.$accessor.connection.setConnected(false)
    this.$accessor.remote.reset()
    this.$accessor.user.reset()
    this.$accessor.video.reset()
    this.$accessor.chat.reset()
  }

  async login(password: string, displayname: string) {
    if (this.$accessor.connection.connecting) {
      return
    }

    this.$accessor.connection.setConnecting()
    try {
      const response = await this.$vue.$http.post<{ token?: string }>(`${this.apiURL}/login`, {
        username: displayname,
        password,
      })
      this.token = response.data.token || ''
      this.setAuthToken()
      this.connect(this.url, this.token)
    } catch (error) {
      const reason = this.toError(error)
      this.token = ''
      this.setAuthToken()
      this.$accessor.connection.setConnected(false)
      this.$accessor.connection.setError(reason.message)
    }
  }

  async logout() {
    this.disconnect()
    this.cleanup()
    try {
      await this.$vue.$http.post(`${this.apiURL}/logout`)
    } catch (error) {
      // A closed session is already safe to discard locally.
    }
    this.token = ''
    this.setAuthToken()
    this.$vue.$swal({
      title: this.$vue.$t('connection.logged_out'),
      icon: 'info',
      confirmButtonText: this.$vue.$t('connection.button_confirm') as string,
    })
  }

  private setAuthToken() {
    const headers = this.$vue.$http.defaults.headers.common
    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`
    } else {
      delete headers.Authorization
    }
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
    this.$accessor.connection.setState('reconnecting')
    this.$vue.$notify({
      group: 'neko',
      type: 'warning',
      title: this.$vue.$t('connection.reconnecting') as string,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.CONNECTING]() {
    this.$accessor.connection.setConnecting()
  }

  protected [EVENT.CONNECTED]() {
    this.$accessor.user.setMember(this.id)
    this.$accessor.connection.setConnected(true)
    set('displayname', this.$accessor.displayname)
    set('password', this.$accessor.password)
    this.startNetworkMonitor()

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
    if (!this.$accessor.connection.connected && reason) {
      this.$accessor.connection.setError(reason.message)
    }
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

  private startNetworkMonitor() {
    this.stopNetworkMonitor()
    this.previousNetworkCounters = undefined
    this.updateNetworkQuality()
    this.networkMonitor = window.setInterval(() => this.updateNetworkQuality(), 5000)
  }

  private stopNetworkMonitor() {
    if (this.networkMonitor) {
      window.clearInterval(this.networkMonitor)
      this.networkMonitor = undefined
    }
    this.previousNetworkCounters = undefined
  }

  private async updateNetworkQuality() {
    if (!this._peer || !this.$accessor.connection.connected) {
      return
    }

    try {
      const stats = await this._peer.getStats()
      let packetsReceived = 0
      let packetsLost = 0
      let rtt: number | null = null

      stats.forEach((stat: any) => {
        if (stat.type === 'inbound-rtp' && (stat.kind === 'video' || stat.mediaType === 'video')) {
          packetsReceived += Number(stat.packetsReceived || 0)
          packetsLost += Number(stat.packetsLost || 0)
        }

        if (
          stat.type === 'candidate-pair' &&
          (stat.state === 'succeeded' || stat.nominated === true) &&
          typeof stat.currentRoundTripTime === 'number'
        ) {
          rtt = stat.currentRoundTripTime * 1000
        }
      })

      const previous = this.previousNetworkCounters
      this.previousNetworkCounters = { packetsReceived, packetsLost }

      const receivedDelta = previous ? Math.max(0, packetsReceived - previous.packetsReceived) : packetsReceived
      const lostDelta = previous ? Math.max(0, packetsLost - previous.packetsLost) : packetsLost
      const totalPackets = receivedDelta + lostDelta
      const packetLoss = totalPackets > 0 ? lostDelta / totalPackets : 0
      const quality = this.classifyNetworkQuality(rtt, packetLoss, totalPackets > 0 || rtt !== null)

      this.$accessor.connection.setNetworkQuality({ quality, rtt: rtt === null ? null : Math.round(rtt) })
    } catch (error) {
      // getStats is best effort; a temporary failure must not affect the media session.
    }
  }

  private classifyNetworkQuality(rtt: number | null, packetLoss: number, hasStats: boolean): NetworkQuality {
    if (!hasStats) {
      return 'unknown'
    }

    if ((rtt !== null && rtt > 350) || packetLoss > 0.08) {
      return 'poor'
    }

    if ((rtt !== null && rtt > 180) || packetLoss > 0.03) {
      return 'fair'
    }

    return 'good'
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
  protected [EVENT.SYSTEM.INIT]({ session_id, control_host, sessions, settings }: SystemInitPayload) {
    this._id = session_id
    this.$accessor.remote.setHost(control_host.has_host ? control_host.host_id || '' : '')
    this.$accessor.remote.setImplicitHosting(settings.implicit_hosting)
    this.$accessor.remote.setLocked(settings.locked_controls)
    this.setLockState('login', settings.locked_logins)
    this.setLockState('control', settings.locked_controls)
    this.setLockState('file_transfer', settings.plugins?.['filetransfer.enabled'] === false)
    this.$accessor.user.setMembers(
      Object.values(sessions).map((session) => ({
        id: session.id,
        displayname: session.profile.name,
        admin: session.profile.is_admin,
        muted: false,
        connected: session.state.is_connected,
      })),
    )

    if (settings.heartbeat_interval > 0) {
      if (this._ws_heartbeat) clearInterval(this._ws_heartbeat)
      this._ws_heartbeat = window.setInterval(
        () => this.sendMessage(EVENT.CLIENT.HEARTBEAT),
        settings.heartbeat_interval * 1000,
      )
    }
  }

  protected [EVENT.SYSTEM.ADMIN]({ broadcast_status }: SystemAdminPayload) {
    this.$accessor.settings.broadcastStatus({
      url: broadcast_status.url,
      isActive: broadcast_status.is_active,
    })
  }

  protected [EVENT.SYSTEM.SETTINGS](settings: SystemSettingsPayload) {
    this.$accessor.remote.setImplicitHosting(settings.implicit_hosting)
    this.$accessor.remote.setLocked(settings.locked_controls)
    this.setLockState('login', settings.locked_logins)
    this.setLockState('control', settings.locked_controls)
    this.setLockState('file_transfer', settings.plugins?.['filetransfer.enabled'] === false)
  }

  private setLockState(resource: 'login' | 'control' | 'file_transfer', locked: boolean) {
    if (locked) {
      this.$accessor.setLocked(resource)
    } else {
      this.$accessor.setUnlocked(resource)
    }
  }

  protected [EVENT.CONTROL.HOST]({ has_host, host_id }: { has_host: boolean; host_id?: string }) {
    this.$accessor.remote.setHost(has_host ? host_id || '' : '')
  }

  protected [EVENT.SYSTEM.DISCONNECT]({ message }: SystemMessagePayload) {
    if (message == 'kicked') {
      this.$accessor.logout()
      message = this.$vue.$t('connection.kicked') as string
    }

    if (!this.$accessor.connection.connected && message) {
      this.$accessor.connection.setError(message)
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
    if (!this.$accessor.connection.connected && message) {
      this.$accessor.connection.setError(message)
    }

    this.$vue.$swal({
      title,
      text: message,
      icon: 'error',
      confirmButtonText: this.$vue.$t('connection.button_confirm') as string,
    })
  }

  /////////////////////////////
  // Session Events
  /////////////////////////////
  protected [EVENT.SESSION.CREATED](session: SessionDataPayload) {
    this.$accessor.user.addMember({
      id: session.id,
      displayname: session.profile.name,
      admin: session.profile.is_admin,
      muted: false,
      connected: session.state.is_connected,
    })
  }

  protected [EVENT.SESSION.DELETED]({ id }: SessionIdPayload) {
    this.$accessor.user.delMember(id)
  }

  protected [EVENT.SESSION.PROFILE]({ id, name, is_admin }: SessionProfilePayload) {
    const member = this.member(id)
    if (!member) return
    this.$accessor.user.addMember({ ...member, displayname: name, admin: is_admin })
  }

  protected [EVENT.SESSION.STATE]({ id, is_connected }: SessionStatePayload) {
    const member = this.member(id)
    if (!member) return
    if (is_connected) {
      this.$accessor.user.addMember({ ...member, connected: true })
    } else {
      this.$accessor.user.delMember(id)
    }
  }

  protected [EVENT.SESSION.CURSORS]() {}

  /////////////////////////////
  // Control Events
  /////////////////////////////
  protected [EVENT.CONTROL.RELEASE]({ id }: SessionIdPayload) {
    if (id === this.id) this.$accessor.remote.reset()
  }

  protected [EVENT.CONTROL.REQUEST]({ id }: SessionIdPayload) {
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

  protected [EVENT.CLIPBOARD.UPDATED]({ text }: { text: string }) {
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
  protected [EVENT.FILETRANSFER.UPDATE]({
    root_dir,
    enabled,
    user_download,
    user_upload,
    user_delete,
    files,
  }: FileTransferUpdatePayload) {
    this.$accessor.files.setCwd(root_dir)
    this.$accessor.files.setLoading(false)
    this.$accessor.files.setFileList(files)
    this.$accessor.files.setUserDownload(user_download)
    this.$accessor.files.setUserUpload(user_upload)
    this.$accessor.files.setUserDelete(user_delete)
    this.$accessor.remote.setFileTransfer(enabled)
  }

  /////////////////////////////
  // Open in App Events
  /////////////////////////////
  protected [EVENT.OPENINAPP.INIT]({ enabled }: { enabled: boolean }) {
    this.$accessor.openinapp.setEnabled(enabled)
  }

  /////////////////////////////
  // Screen Events
  /////////////////////////////
  protected [EVENT.SCREEN.UPDATED]({ id, width, height, rate }: ScreenResolutionPayload) {
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
    this.$accessor.settings.broadcastStatus({ url: payload.url, isActive: payload.is_active })
  }

  // Utilities
  protected member(id: string) {
    return this.$accessor.user.members[id]
  }
}
