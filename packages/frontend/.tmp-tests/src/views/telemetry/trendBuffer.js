function compareByTimeAsc(a, b) {
    return new Date(a.time).getTime() - new Date(b.time).getTime();
}
export function normalizeTrendPoints(points, maxBuffer) {
    const dedupedByTime = new Map();
    for (const point of points) {
        if (!point.time)
            continue;
        dedupedByTime.set(point.time, { time: point.time, value: point.value });
    }
    return [...dedupedByTime.values()]
        .sort(compareByTimeAsc)
        .slice(-maxBuffer);
}
export function appendTrendPoint(points, point, maxBuffer) {
    if (!point.time) {
        return normalizeTrendPoints(points, maxBuffer);
    }
    const existing = points.find((item) => item.time === point.time);
    if (existing && existing.value === point.value) {
        return points;
    }
    return normalizeTrendPoints([...points, point], maxBuffer);
}
