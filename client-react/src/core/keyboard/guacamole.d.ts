// Vendored Apache Guacamole keyboard (guacamole.js), also shipped by the legacy client.
export default class GuacamoleKeyboard {
  constructor(element?: Element)
  // return true to let the browser handle the key, false to prevent default
  onkeydown?: (keysym: number) => boolean
  onkeyup?: (keysym: number) => void
  press(keysym: number): boolean
  release(keysym: number): void
  reset(): void
  listenTo(element: Element | Document): void
}
