import { describe, expect, test } from '@jest/globals';
import {
  PRESENCE_COPY,
  presenceDistanceLabel,
  presenceEmptyLabel,
  presenceOverviewModel,
  presenceStateLabel,
  presenceTargetLabel,
  presenceZoneChips,
  presenceZoneLabel,
  presenceZoneOccupancyLabel,
  serverPresenceLabel,
} from '@/domain/spatial/presence';
import { device, expose } from '../../helpers/devices';

function spatialDevice(id: string, name: string) {
  return device(id, { friendly_name: name, provider: 'mqtt' as never }, [
    expose('target_x', { unit: 'mm' }),
    expose('target_y', { unit: 'mm' }),
    expose('target_active', { type: 'binary' as never }),
  ]);
}

describe('presence labels', () => {
  test('state labels distinguish the compact and card wording', () => {
    expect(presenceStateLabel('not-derived')).toBe('Not derived');
    expect(presenceStateLabel('not-derived', 'card')).toBe('Presence not derived yet');
    expect(presenceStateLabel('occupied')).toBe('Occupied');
    expect(presenceStateLabel('clear')).toBe('Clear');
  });

  test('detail labels are optional and pluralised', () => {
    expect(presenceTargetLabel(1)).toBe('1 target');
    expect(presenceTargetLabel(2)).toBe('2 targets');
    expect(presenceTargetLabel(null)).toBeNull();
    expect(presenceZoneLabel('Desk')).toBe('last zone Desk');
    expect(presenceZoneLabel(null)).toBeNull();
    expect(presenceEmptyLabel(12)).toBe('empty 12 s');
    expect(presenceEmptyLabel(null)).toBeNull();
  });

  test('distance labels use the shared formatter and never invent a position', () => {
    expect(presenceDistanceLabel(1500)).toBe('1.5 m from sensor');
    expect(presenceDistanceLabel(2000)).toBe('2 m from sensor');
    expect(presenceDistanceLabel(null)).toBe(PRESENCE_COPY.tile.noPosition);
    expect(presenceDistanceLabel(Number.NaN)).toBe(PRESENCE_COPY.tile.noPosition);
  });

  test('zone occupancy words stay distinct from the global state', () => {
    expect(presenceZoneOccupancyLabel(null)).toBe('unknown');
    expect(presenceZoneOccupancyLabel(true)).toBe('occupied');
    expect(presenceZoneOccupancyLabel(false)).toBe('clear');
  });
});

describe('presence zone chips', () => {
  test('resolve names and labels from derived zone entities', () => {
    const derived = device('presence-0xabc', {}, [
      expose('zone_desk_occupied', { data: true, attributes: { zoneId: 'desk', zoneName: 'Desk' } as never }),
      expose('zone_door_occupied', { data: 'ON', attributes: { zoneId: 'door', zoneName: 'Door' } as never }),
      expose('unrelated', { data: 7 }),
    ]);
    expect(presenceZoneChips(derived)).toEqual([
      { id: 'desk', name: 'Desk', occupied: true, label: 'occupied' },
      { id: 'door', name: 'Door', occupied: null, label: 'unknown' },
    ]);
    expect(presenceZoneChips(undefined)).toEqual([]);
  });
});

describe('presence overview model', () => {
  const first = spatialDevice('a', 'First sensor');
  const second = spatialDevice('b', 'Second sensor');
  const status = () => 'online';

  test('one sensor is the card itself', () => {
    const model = presenceOverviewModel([first], false, status);
    expect(model.sensors).toEqual([first]);
    expect(model).toMatchObject({
      title: 'First sensor',
      subtitle: 'online',
      showHeader: false,
      showNames: true,
      isSingle: true,
    });
  });

  test('one sensor keeps a header while editing for the remove control', () => {
    const model = presenceOverviewModel([first], true, status);
    expect(model).toMatchObject({ title: 'First sensor', showHeader: true, showNames: false, isSingle: true });
  });

  test('multiple sensors use one section header and named rows', () => {
    const model = presenceOverviewModel([first, second], false, status);
    expect(model).toMatchObject({
      title: PRESENCE_COPY.overview.title,
      subtitle: '2 sensors',
      showHeader: true,
      showNames: true,
      isSingle: false,
    });
    expect(model.sensors).toEqual([first, second]);
  });

  test('no sensors keep the section header and empty list', () => {
    const model = presenceOverviewModel([], false, status);
    expect(model).toMatchObject({ sensors: [], title: 'Presence', subtitle: '0 sensors', showHeader: true });
  });
});

describe('server presence label', () => {
  const values = { occupied: true, targetCount: 1, lastZone: 'Desk', emptySeconds: 0 };

  test('missing derived data is never called clear', () => {
    expect(serverPresenceLabel({ hasDerivedDevice: false, online: true, values })).toBe(
      'Server presence: not derived yet.'
    );
    expect(
      serverPresenceLabel({
        hasDerivedDevice: true,
        online: true,
        values: { occupied: null, targetCount: null, lastZone: null, emptySeconds: null },
      })
    ).toBe('Server presence: not derived yet.');
  });

  test('offline and known states use the same wording as before', () => {
    expect(serverPresenceLabel({ hasDerivedDevice: true, online: false, values })).toBe(
      'Server presence: derived device offline.'
    );
    expect(serverPresenceLabel({ hasDerivedDevice: true, online: true, values })).toBe(
      'Server presence: occupied · 1 target · last zone Desk · empty 0s'
    );
    expect(
      serverPresenceLabel({
        hasDerivedDevice: true,
        online: true,
        values: { occupied: false, targetCount: 0, lastZone: null, emptySeconds: 12 },
      })
    ).toBe('Server presence: clear · 0 targets · last zone none yet · empty 12s');
  });
});
