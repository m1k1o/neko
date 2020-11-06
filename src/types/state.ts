interface ScreenSize {
  width: number
  height: number
  rate: number
}

export default interface State {
  id: string | null
  display_name: string | null
  screen_size: ScreenSize
}
