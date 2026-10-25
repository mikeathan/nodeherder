import {
  formatHistoryAxisLabel,
  formatHistoryClock,
  formatHistoryCoverage,
  formatHistoryMinutes,
  formatHistoryMoment,
  formatHistoryPreciseDuration,
  formatHistorySharePercent,
  formatHistorySpan,
} from '@/domain/history/format';

test('minute formatting rounds, never shows zero and splits at the hour', () => {
  expect(formatHistoryMinutes(0)).toBe('1 min');
  expect(formatHistoryMinutes(1_000)).toBe('1 min');
  expect(formatHistoryMinutes(4 * 60_000)).toBe('4 min');
  expect(formatHistoryMinutes(59.4 * 60_000)).toBe('59 min');
  expect(formatHistoryMinutes(60 * 60_000)).toBe('1 h');
  expect(formatHistoryMinutes(65 * 60_000)).toBe('1 h 5 min');
  expect(formatHistoryMinutes(23 * 60 * 60_000 + 55 * 60_000)).toBe('23 h 55 min');
});

test('precise formatting keeps seconds for very short intervals only', () => {
  expect(formatHistoryPreciseDuration(400)).toBe('under 1 s');
  expect(formatHistoryPreciseDuration(35_000)).toBe('35 s');
  expect(formatHistoryPreciseDuration(60_000)).toBe('1 min');
  expect(formatHistoryPreciseDuration(90_000)).toBe('2 min');
});

test('span formatting treats a positive sub-minute span as "under 1 min"', () => {
  expect(formatHistorySpan(0)).toBe('under 1 min');
  expect(formatHistorySpan(45_000)).toBe('under 1 min');
  expect(formatHistorySpan(5 * 60_000)).toBe('5 min');
  expect(formatHistorySpan(90 * 60_000)).toBe('1 h 30 min');
});

test('coverage distinguishes tiny, zero and missing coverage', () => {
  expect(formatHistoryCoverage(0, 0)).toBe('Unknown');
  expect(formatHistoryCoverage(0, 1000)).toBe('0%');
  expect(formatHistoryCoverage(1, 1000)).toBe('<1%');
  expect(formatHistoryCoverage(131, 1000)).toBe('13%');
  expect(formatHistoryCoverage(1000, 1000)).toBe('100%');
});

test('share percentages round and keep an unknown bin distinct', () => {
  expect(formatHistorySharePercent(null)).toBeNull();
  expect(formatHistorySharePercent(0)).toBe('0%');
  expect(formatHistorySharePercent(41.6)).toBe('42%');
  expect(formatHistorySharePercent(100)).toBe('100%');
});

test('moments, clock labels and axis labels follow the range', () => {
  const stamp = new Date(2026, 9, 24, 18, 20, 0).getTime();
  expect(formatHistoryMoment(stamp)).toContain('18:20');
  expect(formatHistoryClock(stamp)).toMatch(/18:20/);
  expect(formatHistoryAxisLabel(stamp, 3_600_000, false)).toMatch(/18:20/);
  expect(formatHistoryAxisLabel(stamp, 24 * 3_600_000, true)).not.toMatch(/18:20/);
  expect(formatHistoryAxisLabel(stamp, 3 * 24 * 3_600_000, false)).not.toMatch(/18:20/);
});
