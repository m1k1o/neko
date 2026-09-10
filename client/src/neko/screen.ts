import { ScreenConfigurations, ScreenResolution } from './types'

export interface ScreenRect {
  left: number
  top: number
  width: number
  height: number
}

export interface ScreenPoint {
  x: number
  y: number
}

export function normalizeScreenConfigurations(configurations: ScreenConfigurations): ScreenResolution[] {
  return configurations
    .filter(({ width, height, rate }) => width >= 600 && height >= 300 && (rate === 30 || rate === 60))
    .map(({ width, height, rate }) => ({ width, height, rate }))
    .sort((a, b) => {
      if (b.width === a.width && b.height === a.height) {
        return b.rate - a.rate
      }
      if (b.width === a.width) {
        return b.height - a.height
      }
      return b.width - a.width
    })
}

/** Convert a CSS pointer position into the current logical desktop space. */
export function mapPointerToScreen(
  clientX: number,
  clientY: number,
  rect: ScreenRect,
  resolution: Pick<ScreenResolution, 'width' | 'height'>,
): ScreenPoint | undefined {
  const { width, height } = resolution
  if (rect.width <= 0 || rect.height <= 0 || width <= 0 || height <= 0) {
    return undefined
  }

  return {
    x: Math.min(Math.max(Math.round((width / rect.width) * (clientX - rect.left)), 0), width - 1),
    y: Math.min(Math.max(Math.round((height / rect.height) * (clientY - rect.top)), 0), height - 1),
  }
}
