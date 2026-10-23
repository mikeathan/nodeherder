import { HISTORY_PRESETS, historyRange, validHistoryRange, clampHistoryRange, panHistoryRange, zoomHistoryRange } from '@/domain/history/range';
test('preset 3d selects a rolling 72 hour window', () => {
  expect(HISTORY_PRESETS.find((p) => p.value === '72')?.label).toBe('3d');
  expect(historyRange('72', 100000000)).toEqual({ from: 100000000 - 72 * 3600000, to: 100000000 });
});
test('rolling windows retain elapsed time and invalid/custom windows are explicit', () => {
  expect(historyRange('24', 100000000)).toEqual({ from: 13600000, to: 100000000 });
  expect(validHistoryRange(historyRange('custom', 0, '', ''))).toBe(false);
  expect(validHistoryRange(historyRange('custom', 0, '2026-10-17T10:00Z', '2026-10-17T14:00Z'))).toBe(true);
  expect(validHistoryRange({ from: 0, to: 32 * 86400000 })).toBe(false);
  expect(validHistoryRange({ from: 100, to: 0 })).toBe(false);
});
test('calendar Today starts at local midnight rather than subtracting 24 hours', () => {
  const now = new Date(2026, 9, 18, 12, 30).getTime();
  const start = new Date(historyRange('today', now).from);
  expect(start.getHours()).toBe(0);
  expect(start.getDate()).toBe(18);
});
test('Yesterday uses consecutive local midnights', () => {
  const now = new Date(2026, 2, 30, 12).getTime();
  expect(historyRange('yesterday', now)).toEqual({ from: new Date(2026, 2, 29).getTime(), to: new Date(2026, 2, 30).getTime() });
});
test('zoom and pan clamp identically and narrow Today never escapes its query', () => {
  const bounds = { from: 0, to: 3600000 };
  expect(clampHistoryRange({ from: -100, to: 100 }, bounds)).toEqual({ from: 0, to: 300000 });
  expect(zoomHistoryRange(bounds, bounds, 0.5)).toEqual({ from: 900000, to: 2700000 });
  expect(panHistoryRange({ from: 0, to: 1000000 }, bounds, -1)).toEqual({ from: 0, to: 1000000 });
  const short = { from: 0, to: 60000 };
  expect(zoomHistoryRange(short, short, 0.5)).toEqual(short);
});
