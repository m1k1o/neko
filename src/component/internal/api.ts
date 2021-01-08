import * as Api from '../api'

export class NekoApi {
  api_configuration = new Api.Configuration()

  public connect(url: string, id: string, secret: string) {
    this.api_configuration = new Api.Configuration({
      basePath: url,
      baseOptions: {
        auth: {
          username: id,
          password: secret,
        },
      },
    })
  }

  public disconnect() {
    this.api_configuration = new Api.Configuration()
  }

  get room(): RoomApi {
    return new Api.RoomApi(this.api_configuration)
  }

  get members(): MembersApi {
    return new Api.MembersApi(this.api_configuration)
  }
}

export type RoomApi = Api.RoomApi
export type MembersApi = Api.MembersApi
