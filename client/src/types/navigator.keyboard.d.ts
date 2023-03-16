// navigator.keyboard.d.ts

// Type declarations for Keyboard API
// https://developer.mozilla.org/en-US/docs/Web/API/Keyboard_API
interface Keyboard {
  lock(keyCodes?: array<string>): Promise<void>
  unlock(): void
}

interface NavigatorKeyboard {
  // Only available in a secure context.
  readonly keyboard?: Keyboard
}

interface Navigator extends NavigatorKeyboard {}
