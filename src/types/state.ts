export interface ScreenSize {
  width: number
  height: number
  rate: number
}

export interface Scroll {
  sensitivity: number
  invert: boolean
}

export default interface State {
  id: string | null
  display_name: string | null
  screen_size: ScreenSize
  scroll: Scroll
  websocket: 'connected' | 'connecting' | 'disconnected'
  webrtc: 'connected' | 'connecting' | 'disconnected'
}
