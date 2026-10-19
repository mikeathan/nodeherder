import { eventCounts } from '@/domain/history/events';
import type { HistorySeries } from '@/domain/history/types';
test('counts repeated equal reports and includes the right endpoint once', () => {
  const source: HistorySeries = {
    name: 'action',
    type: 'enum',
    unit: '',
    from: 0,
    to: 100,
    partial: false,
    data: [
      { timestamp: 20, value: 'press' },
      { timestamp: 21, value: 'press' },
      { timestamp: 100, value: 'press' },
    ],
  };
  expect(eventCounts(source, 2).data.map((p) => p.value)).toEqual([2, 1]);
});

test('truncated event history never labels the unscanned tail as zero events', () => {
  const source: HistorySeries = {
    name: 'action',
    type: 'enum',
    unit: '',
    from: 0,
    to: 100,
    partial: true,
    data: [
      { timestamp: 10, value: 'press' },
      { timestamp: 60, value: 'press' },
    ],
  };
  expect(eventCounts(source, 4).data.map((point) => point.value)).toEqual([1, 0, null, null]);
});

test('an empty truncated response has no known event-count buckets', () => {
  const source: HistorySeries = { name: 'action', type: 'enum', unit: '', from: 0, to: 100, partial: true, data: [] };
  const result = eventCounts(source);
  expect(result.data).toHaveLength(24);
  expect(result.data.every((point) => point.value === null)).toBe(true);
});
