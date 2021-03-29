import GuacamoleKeyboard from './guacamole-keyboard.js'

export interface GuacamoleKeyboardInterface {
  /**
   * Fired whenever the user presses a key with the element associated
   * with this Guacamole.Keyboard in focus.
   *
   * @event
   * @param {Number} keysym The keysym of the key being pressed.
   * @return {Boolean} true if the key event should be allowed through to the
   *                   browser, false otherwise.
   */
  onkeydown?: (keysym: number) => boolean

  /**
   * Fired whenever the user releases a key with the element associated
   * with this Guacamole.Keyboard in focus.
   *
   * @event
   * @param {Number} keysym The keysym of the key being released.
   */
  onkeyup?: (keysym: number) => void

  /**
   * Marks a key as pressed, firing the keydown event if registered. Key
   * repeat for the pressed key will start after a delay if that key is
   * not a modifier. The return value of this function depends on the
   * return value of the keydown event handler, if any.
   *
   * @param {Number} keysym The keysym of the key to press.
   * @return {Boolean} true if event should NOT be canceled, false otherwise.
   */
  press: (keysym: number) => boolean

  /**
   * Marks a key as released, firing the keyup event if registered.
   *
   * @param {Number} keysym The keysym of the key to release.
   */
  release: (keysym: number) => void

  /**
   * Presses and releases the keys necessary to type the given string of
   * text.
   *
   * @param {String} str
   *     The string to type.
   */
  type: (str: string) => void

  /**
   * Resets the state of this keyboard, releasing all keys, and firing keyup
   * events for each released key.
   */
  reset: () => void

  /**
   * Attaches event listeners to the given Element, automatically translating
   * received key, input, and composition events into simple keydown/keyup
   * events signalled through this Guacamole.Keyboard's onkeydown and
   * onkeyup handlers.
   *
   * @param {Element|Document} element
   *     The Element to attach event listeners to for the sake of handling
   *     key or input events.
   */
  listenTo: (element: Element | Document) => void
}

export default function (element?: Element): GuacamoleKeyboardInterface {
  const Keyboard = {}

  GuacamoleKeyboard.bind(Keyboard, element)()

  return Keyboard as GuacamoleKeyboardInterface
}
