// https://github.com/apache/guacamole-client/blob/1ca1161a68030565a37319ec6275556dfcd1a1af/guacamole-common-js/src/main/webapp/modules/Keyboard.js
import GuacamoleKeyboard from './guacamole/keyboard'
import type { Interface } from './guacamole/keyboard'

export interface GuacamoleKeyboardInterface extends Interface {
  removeListener: () => void
}

export default function (element?: Element): GuacamoleKeyboardInterface {
  const keyboard = {} as GuacamoleKeyboardInterface

  GuacamoleKeyboard.bind(keyboard, element)()

  // add removeListener function
  keyboard.removeListener = function () {
    // Guacamole Keyboard does not provide destroy functions
  }

  return keyboard
}
