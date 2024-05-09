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

  get default(): DefaultApi {
    return new Api.DefaultApi(this.config)
  }

  get sessions(): SessionsApi {
    return new Api.SessionsApi(this.config)
  }

  get room(): RoomApi {
    return new Api.RoomApi(this.config)
  }

  get members(): MembersApi {
    return new Api.MembersApi(this.config)
  }
}

export type DefaultApi = Api.DefaultApi
export type SessionsApi = Api.SessionsApi
export type RoomApi = Api.RoomApi
export type MembersApi = Api.MembersApi
