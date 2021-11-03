import { CursorImage } from './webrtc'

export type CursorDrawFunction = (
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  cursorElement: HTMLImageElement,
  cursorImage: CursorImage,
  cursorTag: string,
) => void

export type InactiveCursorDrawFunction = (
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  cursorTag: string,
) => void

export interface Dimension {
  width: number
  height: number
}

export interface KeyboardModifiers {
  capslock: boolean
  numlock: boolean
}
