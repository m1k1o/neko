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
  available_screen_sizes: ScreenSize[]
  scroll: Scroll
  is_controlling: boolean
  websocket: 'connected' | 'connecting' | 'disconnected'
  webrtc: 'connected' | 'connecting' | 'disconnected'
}
