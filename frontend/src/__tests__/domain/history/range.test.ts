import { historyRange, validHistoryRange } from '@/domain/history/range';
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
