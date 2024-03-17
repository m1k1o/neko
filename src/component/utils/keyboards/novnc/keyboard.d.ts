declare export default class Keyboard {
  constructor (element?: Element)

  _target: Element | Document | null
  _keyDownList: { [key: string]: number }
  _altGrArmed: boolean
  _eventHandlers: {
    keyup: (event: KeyboardEvent) => void
    keydown: (event: KeyboardEvent) => void
    blur: () => void
  }

  _sendKeyEvent(keysym: number, code: string, down: boolean): void
  _getKeyCode(e: KeyboardEvent): string
  _handleKeyDown(e: KeyboardEvent): void
  _handleKeyUp(e: KeyboardEvent): void
  _handleAltGrTimeout(): void
  _allKeysUp(): void

  onkeyevent: (keysym: number | null, code: string, down: boolean) => boolean
  grab: () => void
  ungrab: () => void
}
