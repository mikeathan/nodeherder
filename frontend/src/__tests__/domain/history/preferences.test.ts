import { readHistoryPreference, writeHistoryPreference } from '@/domain/history/preferences';
const storage = (value: unknown) => ({ getItem: () => JSON.stringify(value), setItem: jest.fn() });
test('valid custom bounds restore exactly while malformed bounds keep only a valid expose', () => {
  const value = { version: 1, preset: 'custom', expose: 'temperature', custom: { from: 0, to: 10000 } };
  expect(readHistoryPreference(storage(value), 'device', ['humidity', 'temperature'])).toEqual(value);
  for (const custom of [undefined, { from: 'bad', to: 1 }, { from: 2, to: 1 }, { from: 9e15, to: 9e15 + 60000 }, { from: 0, to: 32 * 86400000 }]) {
    expect(readHistoryPreference(storage({ ...value, custom }), 'd', ['temperature'])).toEqual({ version: 1, preset: '24', expose: 'temperature' });
  }
});
test('unknown versions/exposes and unavailable storage are harmless', () => {
  expect(readHistoryPreference(storage({ version: 2, expose: 'gone' }), 'd', ['temperature']).expose).toBe('temperature');
  expect(readHistoryPreference(storage({ version: 1, preset: '6', expose: 'gone' }), 'd', ['temperature']).expose).toBe('temperature');
  expect(readHistoryPreference(storage({ version: 1, preset: '6', expose: 'gone' }), 'd', ['temperature']).preset).toBe('24');
  const broken = { getItem: () => { throw new Error('blocked'); }, setItem: () => { throw new Error('full'); } };
  expect(readHistoryPreference(broken, 'd', []).preset).toBe('24');
  expect(() => writeHistoryPreference(broken, 'd', { version: 1, preset: '24', expose: '' })).not.toThrow();
});
