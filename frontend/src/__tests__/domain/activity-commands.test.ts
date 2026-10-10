import { describe, expect, test } from '@jest/globals';
import { recordUpdate, snapshotDevices } from '@/domain/activity';
import { commandKey, confirmedBy, isExpired, PendingCommand } from '@/domain/commands';
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

describe('device commands (AC-04)', () => {
  const p: PendingCommand = { key: commandKey('d1', 'state'), deviceId: 'd1', expose: 'state', value: 'ON', sentAt: 0 };
  test('confirmed only by an update carrying the requested value for that device', () => {
    expect(confirmedBy([p], { id: 'd1', last_seen: '', availability: undefined as never, data: { state: 'ON' } })).toEqual([p.key]);
    expect(confirmedBy([p], { id: 'd1', last_seen: '', availability: undefined as never, data: { state: 'OFF' } })).toEqual([]);
    expect(confirmedBy([p], { id: 'd2', last_seen: '', availability: undefined as never, data: { state: 'ON' } })).toEqual([]);
    expect(confirmedBy([p], { id: 'd1', last_seen: '', availability: undefined as never, data: { power: 1 } })).toEqual([]);
  });
  test('numbers and strings compare loosely', () => {
    const n: PendingCommand = { ...p, key: 'k', expose: 'brightness', value: 128 };
    expect(confirmedBy([n], { id: 'd1', last_seen: '', availability: undefined as never, data: { brightness: '128' } })).toEqual(['k']);
  });
  test('expiry', () => {
    expect(isExpired(p, 4999)).toBe(false);
    expect(isExpired(p, 5000)).toBe(true);
  });
});
