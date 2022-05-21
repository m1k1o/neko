import * as Api from '../api'

export class NekoApi {
  public readonly config = new Api.Configuration({
    basePath: location.href.replace(/\/+$/, ''),
    baseOptions: { withCredentials: true },
  })

  public setUrl(url: string) {
    this.config.basePath = url.replace(/\/+$/, '')
  }

  public setToken(token: string) {
    this.config.accessToken = token
  }

  get url(): string {
    return this.config.basePath || location.href.replace(/\/+$/, '')
  }

  get session(): SessionApi {
    return new Api.SessionApi(this.config)
  }

  get room(): RoomApi {
    return new Api.RoomApi(this.config)
  }

  get members(): MembersApi {
    return new Api.MembersApi(this.config)
  }
}

export type SessionApi = Api.SessionApi
export type RoomApi = Api.RoomApi
export type MembersApi = Api.MembersApi
