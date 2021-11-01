export type CursorDrawFunction = (ctx: CanvasRenderingContext2D, x: number, y: number, cursorTag: string) => void

export interface Dimension {
  width: number
  height: number
}
