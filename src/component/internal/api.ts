import * as Api from '../api'

export class NekoApi {
  private _config = new Api.Configuration({
    basePath: location.href.replace(/\/+$/, ''),
    baseOptions: { withCredentials: true },
  })

  public setUrl(url: string) {
    this._config.basePath = url.replace(/\/+$/, '')
  }

  public setToken(token: string) {
    this._config.accessToken = token
  }

  get url(): string {
    return this._config.basePath || location.href.replace(/\/+$/, '')
  }

  get session(): SessionApi {
    return new Api.SessionApi(this._config)
  }

  get room(): RoomApi {
    return new Api.RoomApi(this._config)
  }

  get members(): MembersApi {
    return new Api.MembersApi(this._config)
  }
}

export type SessionApi = Api.SessionApi
export type RoomApi = Api.RoomApi
export type MembersApi = Api.MembersApi
