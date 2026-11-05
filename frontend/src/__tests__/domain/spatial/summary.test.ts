import { spatialHeadline, spatialSummary } from '@/domain/spatial/summary';

test('precedence never claims a position while offline, unmapped or silent', () => {
  expect(spatialSummary({ online: false, mapped: true, tracks: [{ state: 'live' }] }).tone).toBe('offline');
  expect(spatialSummary({ online: true, mapped: false, tracks: [] }).tone).toBe('unmapped');
  expect(spatialSummary({ online: true, mapped: true, tracks: [] })).toEqual({ tone: 'clear', label: 'Clear' });
  expect(spatialSummary({ online: true, mapped: true, tracks: [{ state: 'idle' }] }).tone).toBe('idle');
});

test('counts tracks and names the occupied zone', () => {
  expect(spatialSummary({ online: true, mapped: true, tracks: [{ state: 'live' }] })).toEqual({
    tone: 'tracked',
    label: '1 track',
  });
  expect(
    spatialSummary({ online: true, mapped: true, tracks: [{ state: 'live' }, { state: 'held' }], occupiedZones: ['Desk'] })
  ).toEqual({ tone: 'tracked', label: '2 tracks · Desk' });
});

test('the header headline stays honest about what a frame can prove', () => {
  expect(spatialHeadline({ online: true, configured: false, activeSlots: 0 })).toEqual({
    tone: 'muted',
    label: 'Not set up',
  });
  expect(spatialHeadline({ online: false, configured: true, activeSlots: 2 })).toEqual({
    tone: 'danger',
    label: 'Offline',
  });
  // Offline wins over a missing mapping: the device is unreachable, not merely unconfigured.
  expect(spatialHeadline({ online: false, configured: false, activeSlots: 0 }).label).toBe('Offline');
  expect(spatialHeadline({ online: true, configured: true, activeSlots: 0 })).toEqual({ tone: 'ok', label: 'Clear' });
  expect(spatialHeadline({ online: true, configured: true, activeSlots: 1 }).label).toBe('1 target');
  expect(spatialHeadline({ online: true, configured: true, activeSlots: 2 }).label).toBe('2 targets');
});
