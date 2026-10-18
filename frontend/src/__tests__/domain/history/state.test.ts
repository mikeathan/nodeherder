import { binaryState, stateSegments, stateStats } from '@/domain/history/state';
import { expose } from '../../helpers/devices';
import type { HistorySeries } from '@/domain/history/types';
const contact = expose('contact', { type: 'binary', data: true, values: { on: false, off: true } });
const series: HistorySeries = {
  name: 'contact',
  type: 'binary',
  unit: '',
  from: 0,
  to: 100,
  partial: false,
  data: [
    { timestamp: 10, value: true },
    { timestamp: 20, value: true },
    { timestamp: 40, value: false },
    { timestamp: 45, value: true },
  ],
};
test('contacts respect inverted values, unknown coverage and genuine transitions', () => {
  const segments = stateSegments(series, contact);
  expect(segments[0]).toMatchObject({ from: 0, to: 10, active: null, label: 'Unknown' });
  expect(segments.find((s) => s.from === 40)).toMatchObject({ to: 45, active: true, label: 'Open' });
  expect(stateStats(segments)).toMatchObject({ duration: 90, active: 5, transitions: 2 });
  expect(binaryState(contact, 'garbage')).toBeNull();
  expect(stateStats(stateSegments({ ...series, data: [] }, contact)).percent).toBeNull();
});
test('same-time reports have no invented duration and unknown breaks transition inference', () => {
  const s = stateSegments(
    {
      ...series,
      data: [
        { timestamp: 0, value: true },
        { timestamp: 0, value: false },
        { timestamp: 10, value: null },
        { timestamp: 30, value: true },
      ],
    },
    contact
  );
  expect(s).toHaveLength(3);
  expect(stateStats(s).transitions).toBe(0);
});

test('dense history retains known/unknown duration without enlarging short events', () => {
  const repeated = stateSegments(
    {
      ...series,
      data: [
        { timestamp: 0, value: true },
        { timestamp: 20, value: true },
      ],
    },
    contact
  );
  expect(repeated).toHaveLength(1);
  expect(stateSegments({ ...series, partial: true }, contact).slice(-1)[0].active).toBeNull();
});

test('density uses real overlap and leaves entirely unknown buckets unknown', () => {
  const { stateDensity } = require('@/domain/history/state');
  const segments = stateSegments(series, contact);
  const buckets = stateDensity(segments, 0, 100, 10);
  expect(buckets[0].percent).toBeNull();
  expect(buckets[4].percent).toBe(50);
});
