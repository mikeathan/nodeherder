import {
  ACTIVITY_BUCKETS,
  buildZoneActivity,
  occupancyShare,
  zoneActivityRequest,
  zoneHistoryEntities,
} from '@/domain/spatial/activity';
import type { HistorySeries } from '@/domain/history/types';
import type { Device } from '@/types/device';
import type { SpatialZone } from '@/types/spatial';

const zone = { id: 'desk', name: 'Desk', kind: 'detection', shape: 'rectangle', points: [] } as unknown as SpatialZone;
const derived = (exposes: Record<string, unknown>) =>
  ({ id: 'presence-0xabc', exposes: exposes as Device['exposes'] }) as Device;
const entity = (name: string, zoneId: string) => ({
  name,
  attributes: { zoneId },
  type: 'binary',
  access_mode: 'read',
  data: false,
});

test('resolves only mapped zone entities from the derived device', () => {
  const device = derived({
    zone_a_occupied: entity('zone_a_occupied', 'desk'),
    zone_a_count: { ...entity('zone_a_count', 'desk'), type: 'numeric' },
    occupied: { ...entity('occupied', ''), attributes: {} },
  });
  const entities = zoneHistoryEntities(device, [zone]);
  expect(entities.get('desk')).toEqual({ occupied: 'zone_a_occupied', count: 'zone_a_count' });
  expect(zoneActivityRequest(device, [zone], 0, 1000)?.exposes.sort()).toEqual(['zone_a_count', 'zone_a_occupied']);
  expect(zoneActivityRequest(derived({}), [zone], 0, 1000)).toBeNull();
});

test('share covers only known time and never counts unknown as empty', () => {
  expect(occupancyShare([{ at: 0, value: true }, { at: 1000, value: false }], 0, 2000)).toEqual({
    share: 50,
    trueMs: 1000,
    knownMs: 2000,
  });
  // First sample arrives at 1000: the earlier half is unknown, not empty.
  expect(occupancyShare([{ at: 1000, value: true }], 0, 2000)).toEqual({
    share: 100,
    trueMs: 1000,
    knownMs: 1000,
  });
  expect(occupancyShare([], 0, 2000).share).toBeNull();
});

test('builds transitions, peak count, last occupancy and a partial flag', () => {
  const occupied: HistorySeries = {
    name: 'zone_a_occupied', type: 'binary', unit: '', from: 1000, to: 3000, partial: false,
    data: [
      { timestamp: 1000, value: true },
      { timestamp: 2000, value: false },
      { timestamp: 2500, value: true },
    ],
  };
  const count: HistorySeries = {
    name: 'zone_a_count', type: 'numeric', unit: '', from: 0, to: 3000, partial: false,
    data: [
      { timestamp: 1000, value: 1 },
      { timestamp: 2000, value: 3 },
    ],
  };
  const activity = buildZoneActivity([occupied, count], [zone], new Map([['desk', { occupied: occupied.name, count: count.name }]]), 1000, 3000, 4);
  expect(activity[0]).toMatchObject({
    zoneId: 'desk',
    transitions: 2,
    peakCount: 3,
    lastOccupiedAt: 2500,
    partial: false,
  });
  expect(activity[0].sharePct).toBeCloseTo(75, 6);
  expect(activity[0].buckets).toHaveLength(4);
});

test('a missing or late first sample marks the zone partial', () => {
  const late: HistorySeries = {
    name: 'zone_a_occupied', type: 'binary', unit: '', from: 1000, to: 2000, partial: false,
    data: [{ timestamp: 1500, value: true }],
  };
  const lateActivity = buildZoneActivity([late], [zone], new Map([['desk', { occupied: late.name }]]), 1000, 2000, 2);
  expect(lateActivity[0].partial).toBe(true);
  expect(lateActivity[0].sharePct).toBe(100);
  const none = buildZoneActivity([], [zone], new Map([['desk', {}]]), 0, 1000, ACTIVITY_BUCKETS);
  expect(none[0].sharePct).toBeNull();
  expect(none[0].partial).toBe(true);
});
