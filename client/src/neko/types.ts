export interface Member {
  id: string
  username: string
  admin: boolean
  muted: boolean
  connected?: boolean
  ignored?: boolean
}
