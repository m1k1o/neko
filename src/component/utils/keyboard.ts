//
// TODO: add support for other keyboards
//
import Keyboard from './keyboards/guacamole'

// conditional import at build time:
// __KEYBOARD__ is replaced by the value of the env variable KEYBOARD

export interface KeyboardInterface {
  onkeydown?: (keysym: number) => boolean
  onkeyup?: (keysym: number) => void
  release: (keysym: number) => void
  listenTo: (element: Element | Document) => void
  removeListener: () => void
}

export function NewKeyboard(element?: Element): KeyboardInterface {
  return Keyboard(element)
}
