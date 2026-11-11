import {
  derivedPresenceDeviceId,
  derivedPresenceDisplay,
  derivedPresenceValues,
  derivedZoneSignals,
} from '@/domain/spatial/derived';
import { device, expose } from '../../helpers/devices';

test('derived device identifiers mirror the hub convention', () => {
  expect(derivedPresenceDeviceId('0xabc')).toBe('presence-0xabc');
  expect(derivedPresenceDeviceId('')).toBe('presence-');
});

test('zone signals resolve by zone attribute, not by guessing names', () => {
  const derived = device('presence-0xabc', {}, [
    expose('zone_abc_occupied', { data: true, attributes: { zoneId: 'desk', zoneName: 'Desk' } as never }),
    expose('zone_abc_count', { data: 2, attributes: { zoneId: 'desk' } as never }),
    expose('zone_abc_dwell_seconds', { data: 42, attributes: { zoneId: 'desk' } as never }),
    expose('zone_xyz_occupied', { data: false, attributes: { zoneId: 'door' } as never }),
    expose('unrelated', { data: 7, attributes: null }),
    expose('zone_nogeo_occupied', { data: true }),
  ]);
  const zones = derivedZoneSignals(derived);
  expect(zones.get('desk')).toEqual({ occupied: true, count: 2, dwell: 42 });
  expect(zones.get('door')).toEqual({ occupied: false, count: null, dwell: null });
  expect(zones.size).toBe(2);
});

test('unusable values stay null instead of becoming measured zero', () => {
  const derived = device('presence-0xabc', {}, [
    expose('zone_abc_occupied', { data: 'ON', attributes: { zoneId: 'desk' } as never }),
    expose('zone_abc_count', { data: Number.NaN, attributes: { zoneId: 'desk' } as never }),
  ]);
  expect(derivedZoneSignals(derived).get('desk')).toEqual({ occupied: null, count: null, dwell: null });
  expect(derivedZoneSignals(undefined).size).toBe(0);
});

test('missing derived occupancy is not a measured clear', () => {
  expect(derivedPresenceDisplay({ occupied: null, targetCount: null, lastZone: null, emptySeconds: null })).toBe(
    'not-derived'
  );
  expect(derivedPresenceDisplay({ occupied: true, targetCount: 1, lastZone: 'Desk', emptySeconds: 0 })).toBe(
    'occupied'
  );
  expect(derivedPresenceDisplay({ occupied: false, targetCount: 0, lastZone: null, emptySeconds: 12 })).toBe(
    'clear'
  );
});

test('global values use the exported entity names and blank text stays unknown', () => {
  const derived = device('presence-0xabc', {}, [
    expose('occupied', { data: true }),
    expose('target_count', { data: 1 }),
    expose('last_zone', { data: 'Desk' }),
    expose('empty_seconds', { data: 0 }),
  ]);
  expect(derivedPresenceValues(derived)).toEqual({ occupied: true, targetCount: 1, lastZone: 'Desk', emptySeconds: 0 });
  expect(derivedPresenceValues(device('presence-empty', {}, [expose('last_zone', { data: '  ' })]))).toEqual({
    occupied: null,
    targetCount: null,
    lastZone: null,
    emptySeconds: null,
  });
});
