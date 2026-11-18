import { decodeHistory } from '@/contracts/history';
import { numericStats, decimate } from '@/domain/history/numeric';
import { RENDER_LIMIT } from '@/domain/history/types';

test('statistics distinguish no values, zero and sparse samples', () => {
  expect(numericStats([])).toBeNull();
  expect(numericStats([{ timestamp: 0, value: null }])).toBeNull();
  expect(
    numericStats([
      { timestamp: 0, value: 0 },
      { timestamp: 1, value: 4 },
    ])
  ).toEqual({ min: 0, max: 4, avg: 2, last: 4 });
});

test('decimation preserves endpoints, gaps and a brief spike within render budget', () => {
  const data = Array.from({ length: 10000 }, (_, timestamp) => ({ timestamp, value: 0 as number | null }));
  data[5001].value = 800;
  data[7002].value = null;
  const result = decimate(data);
  expect(result.length).toBeLessThanOrEqual(1000);
  expect(result[0]).toBe(data[0]);
  expect(result[result.length - 1]).toBe(data[9999]);
  expect(result).toContain(data[5001]);
  expect(result).toContain(data[7002]);
  expect(decimate([{ timestamp: 1, value: 0 }])).toHaveLength(1);
  expect(decimate(data.map((p) => ({ ...p, value: null }))).length).toBeLessThanOrEqual(1000);
});

test('separate null runs stay separate after reducing long histories', () => {
  const source = Array.from({ length: 10000 }, (_, timestamp) => ({
    timestamp,
    value: timestamp % 200 === 100 ? null : timestamp,
  }));
  const result = decimate(source);
  expect(result.filter((p) => p.value === null)).toHaveLength(50);
  expect(result.length).toBeLessThanOrEqual(1000);
  expect(decimate(source.map((p) => ({ ...p, value: p.timestamp % 2 ? null : 1 })))).toEqual([]);
});

/*
 * Large-input behaviour at the real sample limit.
 *
 * These assertions replace a wall-clock timing check that failed intermittently. A performance
 * budget measured with `performance.now()` cannot be deterministic: it depends on JIT state, GC,
 * CPU frequency scaling and whatever else the machine is doing, so the same code passes and fails
 * on consecutive runs. That made the whole suite untrustworthy — a genuinely slow change would
 * have been indistinguishable from noise, and a red build meant nothing.
 *
 * What is actually worth asserting is the property that makes it fast: bounded output, preserved
 * structure, and no quadratic behaviour. Bounds are exact and repeatable; milliseconds are not.
 * Timing remains available as an explicit benchmark (numeric.bench.ts), run on purpose.
 */
describe('at the 10,000-point sample limit', () => {
  /** A full-size series with a smooth signal, as a real temperature history would look. */
  const series = Array.from({ length: 10000 }, (_, timestamp) => ({
    timestamp,
    value: Math.sin(timestamp / 100),
  }));

  test('statistics are correct, not approximate', () => {
    const stats = numericStats(series);
    expect(stats).not.toBeNull();
    // Derived from the fixture rather than written as literals: sin never lands exactly on ±1 at
    // an integer sample, so asserting 1 would be asserting something untrue about the data.
    const values = series.map((point) => point.value);
    expect(stats?.min).toBe(Math.min(...values));
    expect(stats?.max).toBe(Math.max(...values));
    expect(stats?.last).toBe(Math.sin(9999 / 100));
  });

  test('a gap is excluded from the statistics rather than counted as zero', () => {
    // The bug this guards: treating null as 0 would report a minimum of 0 and a mean pulled down,
    // which on a temperature chart reads as a reading that never happened.
    const withGap = [
      { timestamp: 0, value: 10 },
      { timestamp: 1, value: null },
      { timestamp: 2, value: 20 },
    ];
    expect(numericStats(withGap)).toEqual({ min: 10, max: 20, avg: 15, last: 20 });
  });

  test('statistics survive a series far beyond the response limit', () => {
    // Guards a real crash: the previous implementation used Math.min(...values), and spreading
    // more than roughly 120,000 arguments exceeds the call stack. The HTTP contract caps a series
    // at HISTORY_LIMIT today, but numericStats is exported and took any array, so the limit was a
    // property of the caller rather than of the function.
    const huge = Array.from({ length: 300000 }, (_, timestamp) => ({ timestamp, value: timestamp }));
    expect(() => numericStats(huge)).not.toThrow();
    expect(numericStats(huge)).toEqual({ min: 0, max: 299999, avg: 149999.5, last: 299999 });
  });

  test('reduction stays within the render budget', () => {
    // The property the timing check was standing in for: whatever the input size, the chart
    // renders a bounded number of points.
    expect(decimate(series).length).toBeLessThanOrEqual(RENDER_LIMIT);
  });

  test('reduction is linear, so a doubled input does not quadruple the work', () => {
    // A quadratic implementation would still pass a bounded-output check while being unusable on
    // a slow device. This compares two sizes and allows generous headroom: the point is to catch
    // an order-of-magnitude regression, not to measure milliseconds.
    const doubled = [...series, ...series.map((p) => ({ ...p, timestamp: p.timestamp + 10000 }))];
    const small = timeReduction(series);
    const large = timeReduction(doubled);
    expect(large).toBeLessThan(Math.max(small * 8, 200));
  });

  test('reduction keeps the first and last point, and every timestamp ordered', () => {
    const result = decimate(series);
    expect(result[0].timestamp).toBe(0);
    expect(result[result.length - 1].timestamp).toBe(9999);
    for (let index = 1; index < result.length; index++) {
      expect(result[index].timestamp).toBeGreaterThan(result[index - 1].timestamp);
    }
  });

  test('a full decode of the largest allowed response is bounded', () => {
    const request = { deviceId: 'benchmark', exposes: ['temperature'], from: 0, to: 10000 };
    const raw = [
      {
        expose: 'temperature',
        values: [
          {
            device_id: 'benchmark',
            unit: '°C',
            value: {
              name: 'temperature',
              type: 'numeric',
              from: 0,
              to: 10000,
              data: series,
              partial: false,
            },
          },
        ],
      },
    ];

    const normalized = decodeHistory(raw, request)[0].data;
    expect(normalized).toHaveLength(10000);
    expect(decimate(normalized).length).toBeLessThanOrEqual(RENDER_LIMIT);
    expect(numericStats(normalized)).not.toBeNull();
  });
});

/** Best of repeated reductions, so a single GC pause cannot dominate the comparison. */
function timeReduction(points: { timestamp: number; value: number | null }[]): number {
  let best = Number.POSITIVE_INFINITY;
  for (let run = 0; run < 5; run++) {
    const start = performance.now();
    decimate(points);
    best = Math.min(best, performance.now() - start);
  }
  return best;
}
