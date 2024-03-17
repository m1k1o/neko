// https://github.com/novnc/noVNC/blob/ca6527c1bf7131adccfdcc5028964a1e67f9018c/core/input/keyboard.js
import Keyboard from './novnc/keyboard'

export interface NoVncKeyboardInterface extends Keyboard {
  onkeydown?: (keysym: number) => boolean
  onkeyup?: (keysym: number) => void
  release: (keysym: number) => void
  listenTo: (element: Element | Document) => void
  removeListener: () => void
}

export default function (element?: Element): NoVncKeyboardInterface {
  const keyboard = new Keyboard(element) as NoVncKeyboardInterface

  // map on key event to onkeydown and onkeyup
  keyboard.onkeyevent = function (keysym: number | null, code: string, down: boolean) {
    if (keysym === null) return false
    if (down && this.onkeydown) return this.onkeydown(keysym)
    if (!down && this.onkeyup) this.onkeyup(keysym)
    return false
  }

  // add release function
  keyboard.release = function (keysym: number) {
    for (const code in this._keyDownList) {
      if (this._keyDownList[code] === keysym) {
        this._sendKeyEvent(keysym, code, false)
        break
      }
    }
  }

  // add listenTo function
  keyboard.listenTo = function (element: Element | Document) {
    if (element) this._target = element
    this.grab()
  }

  // add removeListener function
  keyboard.removeListener = function () {
    this.ungrab()
  }

  return keyboard
}
