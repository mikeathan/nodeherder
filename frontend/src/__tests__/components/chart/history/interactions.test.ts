import { keyboardHistoryRange, pinchHistoryRange } from '@/components/chart/history/interactions';
const bounds = { from: 0, to: 3600000 };
test('all input paths obey the same min/max and keyboard pan policy', () => {
  expect(keyboardHistoryRange('+', bounds, bounds)).toEqual({ from: 900000, to: 2700000 });
  expect(pinchHistoryRange(bounds, bounds, 0.5, 0.5)).toEqual(keyboardHistoryRange('+', bounds, bounds));
  expect(keyboardHistoryRange('ArrowLeft', bounds, bounds)).toEqual(bounds);
  expect(keyboardHistoryRange('0', { from: 100, to: 300100 }, bounds)).toEqual(bounds);
  expect(keyboardHistoryRange('Tab', bounds, bounds)).toBeNull();
});
