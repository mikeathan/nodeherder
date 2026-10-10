import { clockTime, isKnownTime, relativeTime } from '@/domain/time';

const now = Date.parse('2026-10-10T12:00:00Z');

describe('relativeTime', () => {
  it.each([
    ['2026-10-10T11:59:30Z', 'just now'],
    ['2026-10-10T11:55:00Z', '5 min ago'],
    ['2026-10-10T09:00:00Z', '3 h ago'],
    ['2026-10-09T11:00:00Z', 'yesterday'],
    ['2026-10-07T12:00:00Z', '3 days ago'],
  ])('%s → %s', (iso, text) => expect(relativeTime(iso, now)).toBe(text));

  it('handles missing, invalid and future values', () => {
    expect(relativeTime(null, now)).toBe('—');
    expect(relativeTime('nope', now)).toBe('—');
    expect(relativeTime('2026-10-10T13:00:00Z', now)).toBe('just now');
  });
});

describe('clockTime', () => {
  it('pads to HH:MM:SS', () => {
    expect(clockTime(new Date(2026, 0, 1, 7, 5, 9).getTime())).toBe('07:05:09');
  });
});

describe('isKnownTime', () => {
  it('accepts timestamps and ISO strings only', () => {
    expect(isKnownTime('2026-10-10T12:00:00Z')).toBe(true);
    expect(isKnownTime(0)).toBe(true);
    expect(isKnownTime('')).toBe(false);
    expect(isKnownTime(null)).toBe(false);
    expect(isKnownTime('soon')).toBe(false);
  });
});
