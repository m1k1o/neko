/* eslint-disable */

export interface Point {
  x: number
  y: number
}

// movement: percent is 0-1
export function getMovementXYatPercent(points: Point[], percent: number): Point {
  if (points.length == 0) { console.error('no points specified'); return { x:0, y:0 } }
  if (points.length == 1) return points[0]
  if (points.length == 2) return getLineXYatPercent(points[0], points[1], percent)
  if (points.length == 3) return getQuadraticBezierXYatPercent(points[0], points[1], points[2], percent)
  if (points.length == 4) return getCubicBezierXYatPercent(points[0], points[1], points[2], points[3], percent)
  console.error('max 4 points supported'); return points[4]
}

// line: percent is 0-1
export function getLineXYatPercent(startPt: Point, endPt: Point, percent: number) : Point {
  return {
    x: startPt.x + (endPt.x - startPt.x) * percent,
    y: startPt.y + (endPt.y - startPt.y) * percent,
  };
}

// quadratic bezier: percent is 0-1
export function getQuadraticBezierXYatPercent(startPt: Point, controlPt: Point, endPt: Point, percent: number): Point {
  return {
    x: Math.pow(1 - percent, 2) * startPt.x + 2 * (1 - percent) * percent * controlPt.x + Math.pow(percent, 2) * endPt.x,
    y: Math.pow(1 - percent, 2) * startPt.y + 2 * (1 - percent) * percent * controlPt.y + Math.pow(percent, 2) * endPt.y,
  }
}

// cubic bezier percent is 0-1
export function getCubicBezierXYatPercent(startPt: Point, controlPt1: Point, controlPt2: Point, endPt: Point, percent: number): Point {
  return {
    x: cubicN(percent, startPt.x, controlPt1.x, controlPt2.x, endPt.x),
    y: cubicN(percent, startPt.y, controlPt1.y, controlPt2.y, endPt.y),
  }
}

// cubic helper formula at percent distance
function cubicN(pct: number, a: number, b: number, c: number, d: number): number {
  var t2 = pct * pct;
  var t3 = t2 * pct;
  return a + (-a * 3 + pct * (3 * a - a * pct)) * pct + (3 * b + pct * (-6 * b + b * 3 * pct)) * pct + (c * 3 - c * 3 * pct) * t2 + d * t3;
}
