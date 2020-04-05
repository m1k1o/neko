export interface Member {
  id: string
  displayname: string
  admin: boolean
  muted: boolean
  connected?: boolean
  ignored?: boolean
}

export interface ScreenConfigurations {
  [index: string]: ScreenConfiguration
}

export interface ScreenConfiguration {
  width: number
  height: number
  rates: { [index: string]: number }
}

export interface ScreenResolution {
  width: number
  height: number
  rate: number
}
