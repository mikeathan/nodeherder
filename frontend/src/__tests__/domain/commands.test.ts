import { commandKey, confirmedBy, isExpired, PendingCommand } from '@/domain/commands';

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
    expect(isExpired(p, 4999, 5000)).toBe(false);
    expect(isExpired(p, 5000, 5000)).toBe(true);
  });
});
