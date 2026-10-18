import { energyBuckets } from '@/domain/history/energy';
test('counter deltas require known endpoints and never distribute across buckets', () => {
  expect(
    energyBuckets(
      [
        { timestamp: 0, value: 10 },
        { timestamp: 50, value: 11 },
        { timestamp: 100, value: 14 },
      ],
      0,
      100,
      2
    )
  ).toEqual([
    { timestamp: 0, value: 1, partial: false },
    { timestamp: 50, value: 3, partial: false },
  ]);
  expect(
    energyBuckets(
      [
        { timestamp: 0, value: 10 },
        { timestamp: 100, value: 14 },
      ],
      0,
      100,
      2
    ).every((b) => b.value === null && b.partial)
  ).toBe(true);
});
test('reset, gap, missing baseline and measured zero remain distinct', () => {
  const bucket = (values: (number | null)[]) =>
    energyBuckets(
      values.map((value, i) => ({ timestamp: i * 50, value })),
      0,
      100,
      1
    )[0];
  expect(bucket([10, 1, 2])).toMatchObject({ value: 1, partial: true });
  expect(bucket([10, null, 12])).toMatchObject({ value: null, partial: true });
  expect(bucket([10, 10, 10])).toMatchObject({ value: 0, partial: false });
  expect(energyBuckets([], 0, 100, 1)[0]).toMatchObject({ value: null, partial: true });
});

test('consumption summaries explicitly distinguish unavailable and observed zero', () => {
  const { energySummary } = require('@/domain/history/energy');
  expect(energySummary([{ timestamp: 0, value: null, partial: true }])).toEqual({
    observed: null,
    complete: 0,
    count: 1,
  });
  expect(energySummary([{ timestamp: 0, value: 0, partial: false }])).toEqual({ observed: 0, complete: 1, count: 1 });
});

test('only cumulative energy counters use consumption bars', () => {
  const { isEnergyCounter } = require('@/domain/history/energy');
  expect(isEnergyCounter({ name: 'energy', unit: 'kWh' })).toBe(true);
  expect(isEnergyCounter({ name: 'power', unit: 'W' })).toBe(false);
  expect(isEnergyCounter({ name: 'energy_today', unit: 'kWh' })).toBe(false);
});
