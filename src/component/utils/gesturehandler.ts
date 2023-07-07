// https://github.com/novnc/noVNC/blob/ca6527c1bf7131adccfdcc5028964a1e67f9018c/core/input/gesturehandler.js#L246
import gh from './gesturehandler.js'

const g = gh as GestureHandlerConstructor
export default g

interface GestureHandlerConstructor {
  new (): GestureHandler
}

export interface GestureHandler {
  attach(element: Element): void
  detach(): void
}
