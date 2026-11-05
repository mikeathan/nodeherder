import { zonePreviewRows } from '@/domain/spatial/view';
import { derivedZoneSignals } from '@/domain/spatial/derived';
import { derivedPresenceDeviceId } from '@/domain/spatial/derived';
import { rectanglePoints } from '@/domain/spatial/workspace';
import { device, expose } from '../../helpers/devices';
import type { SpatialZone } from '@/types/spatial';

const desk: SpatialZone = {
  id: 'desk',
  name: 'Desk',
  kind: 'detection',
  shape: 'rectangle',
  points: rectanglePoints(0, 1000, 1000, 1000),
};

function derived() {
  return device(derivedPresenceDeviceId('0xabc'), {}, [
    expose('zone_abc_occupied', { data: true, attributes: { zoneId: 'desk', zoneName: 'Desk' } as never }),
    expose('zone_abc_count', { data: 2, attributes: { zoneId: 'desk' } as never }),
  ]);
}

test('a hub-derived value wins and says so', () => {
  const rows = zonePreviewRows(device('0xabc'), [desk], new Map(), true, true, derivedZoneSignals(derived()));
  expect(rows[0].preview).toBe('Derived by the hub: target inside · 2 targets');
});

test('without derived signals the geometric preview remains the fallback', () => {
  const rows = zonePreviewRows(device('0xabc'), [desk], new Map([['desk', { occupied: true }]]), true, true);
  expect(rows[0].preview).toBe('Target inside');
});

test('missing data and offline are never reported as a measured clear', () => {
  expect(zonePreviewRows(device('0xabc'), [desk], new Map(), true, false)[0].preview).toContain('waiting for position data');
  expect(zonePreviewRows(device('0xabc'), [desk], new Map(), false, false)[0].preview).toContain('offline');
});

test('an offline derived device falls back instead of showing frozen values', () => {
  const rows = zonePreviewRows(device('0xabc'), [desk], new Map([['desk', { occupied: false }]]), true, true, derivedZoneSignals(derived()), false);
  expect(rows[0].preview).toBe('No target inside');
});
