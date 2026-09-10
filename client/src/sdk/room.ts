export interface RoomHttpClient {
  get<T>(url: string): Promise<{ data: T }>
  post<T = unknown>(url: string, data?: unknown): Promise<{ data: T }>
}

export interface ControlStatus {
  has_host: boolean
  host_id?: string
  epoch: number
}

/** REST room operations that do not depend on Vue, Vuex, or UI services. */
export class RoomClient {
  constructor(private readonly http: RoomHttpClient, private readonly apiURL: string) {}

  async controlStatus() {
    const response = await this.http.get<ControlStatus>(`${this.apiURL}/room/control`)
    return response.data
  }

  async requestControl() {
    await this.http.post(`${this.apiURL}/room/control/request`)
  }

  async releaseControl() {
    await this.http.post(`${this.apiURL}/room/control/release`)
  }

  async takeControl() {
    await this.http.post(`${this.apiURL}/room/control/take`)
  }

  async giveControl(sessionID: string) {
    await this.http.post(`${this.apiURL}/room/control/give/${encodeURIComponent(sessionID)}`)
  }

  async resetControl() {
    await this.http.post(`${this.apiURL}/room/control/reset`)
  }
}
