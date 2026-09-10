import { mapPointerToScreen, ScreenPoint, ScreenRect } from '../neko/screen'
import { ScreenResolution } from '../neko/types'

export interface ControlInputOptions {
  scroll: number
  invertScroll: boolean
  lineHeight?: number
}

export interface NormalizedWheel {
  x: number
  y: number
  controlKey: boolean
}

/**
 * Framework-independent input normalization for the remote desktop.
 * Permission checks and transport calls stay in the UI adapter; this class
 * only turns browser coordinates and wheel events into desktop-space values.
 */
export class ControlInputController {
  private options: Required<ControlInputOptions>

  constructor(options: ControlInputOptions) {
    this.options = {
      lineHeight: 19,
      ...options,
    }
  }

  update(options: ControlInputOptions) {
    this.options = {
      lineHeight: this.options.lineHeight,
      ...options,
    }
  }

  pointer(
    event: Pick<MouseEvent, 'clientX' | 'clientY'>,
    rect: ScreenRect,
    resolution: Pick<ScreenResolution, 'width' | 'height'>,
  ): ScreenPoint | undefined {
    return mapPointerToScreen(event.clientX, event.clientY, rect, resolution)
  }

  wheel(event: Pick<WheelEvent, 'deltaX' | 'deltaY' | 'deltaMode' | 'ctrlKey'>): NormalizedWheel {
    let x = event.deltaX
    let y = event.deltaY

    if (event.deltaMode !== 0) {
      x *= this.options.lineHeight
      y *= this.options.lineHeight
    }

    if (this.options.invertScroll) {
      x *= -1
      y *= -1
    }

    return {
      x: clamp(x, -this.options.scroll, this.options.scroll),
      y: clamp(y, -this.options.scroll, this.options.scroll),
      controlKey: event.ctrlKey,
    }
  }
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max)
}
