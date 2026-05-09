import test from 'node:test';
import assert from 'node:assert/strict';
import { appendTrendPoint, normalizeTrendPoints } from '../../src/views/telemetry/trendBuffer.js';
test('appendTrendPoint does not append the same SSE point twice', () => {
    const points = [{ time: '2026-05-09T14:18:00.000Z', value: 50 }];
    const next = appendTrendPoint(points, { time: '2026-05-09T14:18:00.000Z', value: 50 }, 360);
    assert.deepEqual(next, points);
});
test('appendTrendPoint keeps points sorted by time when late data arrives', () => {
    const points = [{ time: '2026-05-09T14:20:00.000Z', value: 60 }];
    const next = appendTrendPoint(points, { time: '2026-05-09T14:19:00.000Z', value: 55 }, 360);
    assert.deepEqual(next, [
        { time: '2026-05-09T14:19:00.000Z', value: 55 },
        { time: '2026-05-09T14:20:00.000Z', value: 60 }
    ]);
});
test('normalizeTrendPoints sorts history ascending and removes duplicate timestamps', () => {
    const normalized = normalizeTrendPoints([
        { time: '2026-05-09T14:20:00.000Z', value: 60 },
        { time: '2026-05-09T14:18:00.000Z', value: 50 },
        { time: '2026-05-09T14:20:00.000Z', value: 60 }
    ], 360);
    assert.deepEqual(normalized, [
        { time: '2026-05-09T14:18:00.000Z', value: 50 },
        { time: '2026-05-09T14:20:00.000Z', value: 60 }
    ]);
});
