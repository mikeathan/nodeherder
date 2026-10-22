import { dailyHistoryStats, statisticsDisplaySeries } from '@/domain/history/stats';
import type { Expose } from '@/types/device';
import { decimate } from '@/domain/history/numeric';
import type { HistorySeries } from '@/domain/history/types';
const base: HistorySeries = { name: 'temperature', type: 'numeric', unit: '°C', from: 0, to: 1, partial: false, data: [] };
test('local days own midnight once and empty/all-null buckets remain missing', () => {
  const from = new Date(2026, 0, 1, 12).getTime(), midnight = new Date(2026, 0, 2).getTime();
  const series = { ...base, from, to: new Date(2026, 0, 3, 12).getTime(),
    data: [{ timestamp: from, value: 0 }, { timestamp: midnight, value: 10 }, { timestamp: midnight + 1, value: null }] };
  const buckets = dailyHistoryStats(series);
  expect(buckets.map((b) => b.count)).toEqual([1, 1, 0]);
  expect(buckets[0]).toMatchObject({ min: 0, avg: 0, partial: true });
  expect(buckets[1]).toMatchObject({ min: 10, last: 10, partial: false });
  expect(buckets[2]).toMatchObject({ min: null, avg: null, partial: true });
  expect(dailyHistoryStats({ ...series, data: [{ timestamp: from, value: null }] })[0].avg).toBeNull();
});
test('raw sample basis differs from extrema decimation and truncated tails stay empty', () => {
  const from = new Date(2026, 0, 1).getTime();
  const data = Array.from({ length: 3000 }, (_, i) => ({ timestamp: from + i * 1000, value: i === 500 ? 100 : 1 }));
  const series = { ...base, from, to: from + 2 * 86400000, data, partial: true };
  const buckets = dailyHistoryStats(series);
  expect(buckets[0]).toMatchObject({ min: 1, max: 100, count: 3000, partial: true });
  expect(buckets[0].avg).not.toBe(dailyHistoryStats({ ...series, data: decimate(data) })[0].avg);
  expect(buckets[1]).toMatchObject({ avg: null, count: 0, partial: true });
});
test('enumerates every local date over spring DST without a fixed cap', () => {
  // Run this companion test with TZ=Europe/London for the pinned 33-date case.
  const from = Date.parse('2026-03-01T23:30:00Z'), to = from + 31 * 86400000;
  const buckets = dailyHistoryStats({ ...base, from, to });
  const london = Intl.DateTimeFormat().resolvedOptions().timeZone === 'Europe/London';
  if (london) {
    expect(buckets).toHaveLength(33);
    expect(buckets.some((b) => b.end - b.start === 23 * 3600000)).toBe(true);
  }
  expect(buckets.at(-1)?.end).toBeGreaterThan(to);
  expect(buckets[0].partial).toBe(true);
});
test('nonlinear unit conversion precedes statistics and does not mutate stored samples', () => {
  const series = { ...base, name: 'color_temp', unit: 'mired', data: [{ timestamp: 0, value: 500 }, { timestamp: 1, value: 250 }] };
  const displayed = statisticsDisplaySeries(series, { name: 'color_temp', unit: 'mired' } as Expose);
  expect(displayed.unit).toBe('K');
  expect(dailyHistoryStats(displayed)[0]).toMatchObject({ min: 2000, avg: 3000, max: 4000, count: 2 });
  expect(series.data[0].value).toBe(500);
});
