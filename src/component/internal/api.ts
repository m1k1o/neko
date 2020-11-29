import * as Api from '../api'

export class NekoApi {
  api_configuration = new Api.Configuration()

  public connect(url: string, id: string, secret: string) {
    this.api_configuration = new Api.Configuration({
      basePath: url,
      headers: {
        Authorization: 'Basic ' + btoa(id + ':' + secret),
      },
    })
  }

  public disconnect() {
    this.api_configuration = new Api.Configuration()
  }

  get admin(): Api.AdminsApi {
    return new Api.AdminsApi(this.api_configuration)
  }

  get user(): Api.UsersApi {
    return new Api.UsersApi(this.api_configuration)
  }

  get host(): Api.HostsApi {
    return new Api.HostsApi(this.api_configuration)
  }
}
