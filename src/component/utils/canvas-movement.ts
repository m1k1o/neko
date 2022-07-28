/* eslint-disable */

export interface Point {
  x: number
  y: number
}

// movement: percent is 0-1
export function getMovementXYatPercent(p: Point[], percent: number): Point {
  const len = p.length
  if (len == 0) {
    console.error('getMovementXYatPercent: no points specified');
    return { x:0, y:0 }
  }

  if (len == 1) return p[0]
  if (len == 2) return getLineXYatPercent(p[0], p[1], percent)
  if (len == 3) return getQuadraticBezierXYatPercent(p[0], p[1], p[2], percent)
  if (len == 4) return getCubicBezierXYatPercent(p[0], p[1], p[2], p[3], percent)

  // TODO: Support more than 4 points
  if (len-1 % 3 == 0) return getCubicBezierXYatPercent(p[0], p[(len-1)/3], p[((len-1)/3)*2], p[len-1], percent)
  else if (len-1 % 2 == 0) return getQuadraticBezierXYatPercent(p[0], p[(len-1)/2], p[len-1], percent)
  else return getLineXYatPercent(p[0], p[len-1], percent)
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
