// Apache Guacamole keyboard, the same file as upstream master's client/src/utils/guacamole-keyboard.js
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
