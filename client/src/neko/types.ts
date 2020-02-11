export interface Member {
  id: string
  username: string
  admin: boolean
  muted: boolean
  connected?: boolean
  ignored?: boolean
}

export interface ScreenConfigurations {
  [index: number]: ScreenConfiguration
}

export interface ScreenConfiguration {
  width: string
  height: string
  rates: { [index: number]: number }
}
