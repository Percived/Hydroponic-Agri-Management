export interface TrendPoint {
  time: string
  value: number
}

function compareByTimeAsc(a: TrendPoint, b: TrendPoint): number {
  return new Date(a.time).getTime() - new Date(b.time).getTime()
}

export function normalizeTrendPoints(points: TrendPoint[], maxBuffer: number): TrendPoint[] {
  const dedupedByTime = new Map<string, TrendPoint>()

  for (const point of points) {
    if (!point.time) continue
    dedupedByTime.set(point.time, { time: point.time, value: point.value })
  }

  return [...dedupedByTime.values()]
    .sort(compareByTimeAsc)
    .slice(-maxBuffer)
}

export function appendTrendPoint(points: TrendPoint[], point: TrendPoint, maxBuffer: number): TrendPoint[] {
  if (!point.time) {
    return normalizeTrendPoints(points, maxBuffer)
  }

  const existing = points.find((item) => item.time === point.time)
  if (existing && existing.value === point.value) {
    return points
  }

  return normalizeTrendPoints([...points, point], maxBuffer)
}
