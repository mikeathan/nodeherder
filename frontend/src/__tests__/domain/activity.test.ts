import { describe, expect, test } from '@jest/globals';
import { recordUpdate, snapshotDevices } from '@/domain/activity';
import { device, expose } from '../helpers/devices';

describe('activity feed (AC-27)', () => {
  const devices = [device('d1', {}, [expose('power', { data: 10 }), expose('state', { data: 'ON' })])];
  test('records changed values with from/to, ignores repeats and unknown exposes', () => {
    const snap = snapshotDevices(devices);
    let entries = recordUpdate([], snap, { id: 'd1', last_seen: '', availability: undefined as never, data: { power: 12, state: 'ON', mystery: 1 } }, 1000);
    expect(entries).toHaveLength(1);
    expect(entries[0].changes).toEqual([{ expose: 'power', from: 10, to: 12 }]);
    entries = recordUpdate(entries, snap, { id: 'd1', last_seen: '', availability: undefined as never, data: { power: 12 } }, 2000);
    expect(entries).toHaveLength(1);
  });
  test('is bounded, newest first', () => {
    const snap = snapshotDevices(devices);
    let entries: ReturnType<typeof recordUpdate> = [];
    for (let i = 0; i < 10; i++) entries = recordUpdate(entries, snap, { id: 'd1', last_seen: '', availability: undefined as never, data: { power: i + 100 } }, i, 3);
    expect(entries.map((e) => e.changes[0].to)).toEqual([109, 108, 107]);
  });
});


import { filterActivityEntries } from '@/domain/activity';

describe('activity device filter', () => {
  const entries = [{ deviceId: 'lamp', id: 3 }, { deviceId: 'sensor', id: 2 }, { deviceId: 'lamp', id: 1 }, { deviceId: 'removed-id', id: 0 }];
  const names: Record<string, string> = { lamp: 'Living room light', sensor: 'Attic temperature sensor' };
  const name = (id: string) => names[id];
  test('matches partial names without case or surrounding whitespace, preserving order', () => {
    const result = filterActivityEntries(entries, '  LIVING ROOM  ', name);
    expect(result.map((e) => e.id)).toEqual([3, 1]);
    expect(entries).toHaveLength(4);
    expect(filterActivityEntries(entries, '', name)).toBe(entries);
    expect(filterActivityEntries(entries, '   ', name)).toBe(entries);
  });
  test('handles removed devices, no matches and literal punctuation', () => {
    expect(filterActivityEntries(entries, 'removed', name).map((e) => e.id)).toEqual([0]);
    expect(filterActivityEntries(entries, 'missing', name)).toEqual([]);
    expect(filterActivityEntries(entries, '[', name)).toEqual([]);
    expect(filterActivityEntries([], 'lamp', name)).toEqual([]);
  });
  test('finds matches beyond the visible window without truncating or mutating the retained feed', () => {
    const retained = Array.from({ length: 100 }, (_, i) => ({ id: i, deviceId: i < 70 ? 'sensor' : 'lamp' }));
    expect(filterActivityEntries(retained, 'light', name).slice(0, 60)).toHaveLength(30);
    expect(retained).toHaveLength(100);
  });
});
