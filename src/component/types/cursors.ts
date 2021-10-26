export type CursorDrawFunction = (ctx: CanvasRenderingContext2D, x: number, y: number, id: string) => void

export interface Dimension {
  width: number
  height: number
}
