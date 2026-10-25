import {
  binaryState,
  clipStateSegments,
  recentActivityWindow,
  stateHistoryBins,
  stateSegments,
  stateStats,
  stateTransitions,
} from '@/domain/history/state';
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
test('boolean history normalises to live ON/OFF words', () => {
  const lightState = expose('state', { type: 'binary', values: { on: 'ON', off: 'OFF' } });
  const segments = stateSegments(
    {
      ...series,
      name: 'state',
      from: 0,
      to: 100,
      data: [
        { timestamp: 0, value: true },
        { timestamp: 50, value: false },
      ],
    },
    lightState
  );
  expect(segments[0]).toMatchObject({ active: true, label: 'On' });
  expect(segments[1]).toMatchObject({ active: false, label: 'Off' });
  expect(stateStats(segments)).toMatchObject({ transitions: 1, duration: 100 });
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
  const bins = stateHistoryBins(s, 0, 100, 10);
  expect(bins.reduce((known, bin) => known + bin.known, 0)).toBe(stateStats(s).duration);
  expect(bins.reduce((active, bin) => active + bin.active, 0)).toBe(stateStats(s).active);
  expect(bins.reduce((transitions, bin) => transitions + bin.transitions, 0)).toBe(0);
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

test('pixel bins conserve exact active and known duration and count transitions once', () => {
  const segments = stateSegments(series, contact);
  const before = structuredClone(segments);
  const bins = stateHistoryBins(segments, series.from, series.to, 4);
  expect(bins.map(({ from, to }) => [from, to])).toEqual([[0, 25], [25, 50], [50, 75], [75, 100]]);
  expect(bins.map((bin) => bin.percent)).toEqual([0, 20, 0, 0]);
  expect(bins.map((bin) => bin.transitions)).toEqual([0, 2, 0, 0]);
  expect(bins.reduce((sum, bin) => sum + bin.active, 0)).toBe(stateStats(segments).active);
  expect(bins.reduce((sum, bin) => sum + bin.known, 0)).toBe(stateStats(segments).duration);
  expect(segments).toEqual(before);
});

test('pixel bins preserve unknown-only intervals and partial tails as unknown', () => {
  const segments = stateSegments(series, contact);
  const bins = stateHistoryBins(segments, series.from, series.to, 10);
  expect(bins[0]).toMatchObject({ from: 0, to: 10, active: 0, known: 0, percent: null, transitions: 0 });

  const partial = stateSegments({ ...series, partial: true }, contact);
  const partialBins = stateHistoryBins(partial, series.from, series.to, 10);
  expect(partialBins[9]).toMatchObject({ from: 90, to: 100, active: 0, known: 0, percent: null });
});

test('pixel bins handle empty dimensions without inventing data', () => {
  const segments = stateSegments(series, contact);
  expect(stateHistoryBins(segments, 0, 100, 0)).toEqual([]);
  expect(stateHistoryBins(segments, 0, 100, Infinity)).toEqual([]);
  expect(stateHistoryBins(segments, 100, 0, 10)).toEqual([]);
  expect(stateHistoryBins([], 0, 100, 4).every((bin) => bin.percent === null)).toBe(true);
});

test('density uses real overlap and leaves entirely unknown buckets unknown', () => {
  const { stateDensity } = require('@/domain/history/state');
  const segments = stateSegments(series, contact);
  const buckets = stateDensity(segments, 0, 100, 10);
  expect(buckets[0].percent).toBeNull();
  expect(buckets[4].percent).toBe(50);
});

test('clipping keeps only the visible intervals and never invents state', () => {
  const segments = stateSegments(series, contact);
  const before = structuredClone(segments);
  const clipped = clipStateSegments(segments, 15, 42);
  expect(clipped.map(({ from, to }) => [from, to])).toEqual([[15, 40], [40, 42]]);
  expect(clipped[0].label).toBe('Closed');
  expect(clipped[1].active).toBe(true);
  expect(segments).toEqual(before);
  expect(clipStateSegments(segments, 50, 50)).toEqual([]);
});

test('recent activity focuses latest transitions with context and clamps to the range', () => {
  const cluster = stateSegments(
    {
      ...series,
      from: 0,
      to: 1000,
      data: [
        { timestamp: 900, value: true },
        { timestamp: 910, value: false },
        { timestamp: 950, value: true },
        { timestamp: 960, value: false },
        { timestamp: 970, value: true },
      ],
    },
    contact
  );
  expect(recentActivityWindow(cluster, 0, 1000, { padding: 0, minimum: 15 })).toMatchObject({ from: 910, to: 970, clusterFrom: 910, clusterTo: 970, transitionCount: 4 });
  expect(recentActivityWindow(cluster, 0, 1000, { transitions: 2, padding: 0, minimum: 15 })).toMatchObject({ from: 960, to: 975, clusterFrom: 960, clusterTo: 970, transitionCount: 2 });
  expect(recentActivityWindow(cluster, 0, 1000, { padding: 0, minimum: 40 })).toMatchObject({ from: 910, to: 970, clusterFrom: 910, clusterTo: 970, transitionCount: 4 });
  expect(recentActivityWindow(cluster, 945, 1000, { padding: 0, minimum: 5 })).toMatchObject({ from: 945, to: 970, clusterFrom: 910, clusterTo: 970, transitionCount: 4 });
  expect(recentActivityWindow(cluster, 900, 970, { padding: 0, minimum: 15 })).not.toBeNull();
});

test('recent activity declines single transitions, empty ranges and the full range', () => {
  const single = stateSegments(
    { ...series, from: 0, to: 100, data: [{ timestamp: 50, value: false }] },
    contact
  );
  expect(recentActivityWindow(single, 0, 100)).toBeNull();
  expect(recentActivityWindow([], 0, 100)).toBeNull();
  expect(recentActivityWindow(stateSegments(series, contact), 0, 100, { minimum: 100, padding: 100 })).toBeNull();
});

test('recent activity uses the latest ten transitions in a long burst', () => {
  const points = Array.from({ length: 31 }, (_, index) => ({
    timestamp: 100 + index * 10,
    value: index % 2 === 0,
  }));
  const segments = stateSegments({ ...series, from: 0, to: 1000, data: points }, contact);
  const window = recentActivityWindow(segments, 0, 1000, { padding: 0, minimum: 10 });
  expect(window).toMatchObject({ from: 310, to: 400, clusterFrom: 310, clusterTo: 400, transitionCount: 10 });
});

test('transitions are shared with stats and never invented across unknown time', () => {
  const segments = stateSegments(series, contact);
  const transitions = stateTransitions(segments);
  expect(stateStats(segments).transitions).toBe(transitions.length);
  expect(transitions.map((transition) => transition.from)).toEqual([40, 45]);
  expect(transitions[0]).toMatchObject({ before: false, after: true });
  expect(transitions[0].segment.from).toBe(40);

  const gapped = stateSegments(
    {
      ...series,
      data: [
        { timestamp: 0, value: true },
        { timestamp: 10, value: null },
        { timestamp: 20, value: false },
      ],
    },
    contact
  );
  expect(stateTransitions(gapped)).toEqual([]);
  expect(stateStats(gapped).transitions).toBe(0);
});
