import type { CursorImage } from './webrtc'

export type CursorDrawFunction = (
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  cursorElement: HTMLImageElement,
  cursorImage: CursorImage,
  sessionId: string,
) => void

export type InactiveCursorDrawFunction = (
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  sessionId: string,
) => void

export interface Dimension {
  width: number
  height: number
}

export interface KeyboardModifiers {
  shift?: boolean
  capslock?: boolean
  control?: boolean
  alt?: boolean
  numlock?: boolean
  meta?: boolean
  super?: boolean
  altgr?: boolean
}
