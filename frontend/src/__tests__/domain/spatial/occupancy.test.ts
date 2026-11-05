import { createZoneOccupancyTracker, SPATIAL_ZONE_EXIT_MS, zonePreviewText } from '@/domain/spatial/occupancy';
import { rectanglePoints } from '@/domain/spatial/workspace';
import type { SpatialZone } from '@/types/spatial';

const desk: SpatialZone = {
  id: 'desk',
  name: 'Desk',
  kind: 'detection',
  shape: 'rectangle',
  points: rectanglePoints(-500, 1000, 1000, 1000),
};
const mask: SpatialZone = {
  id: 'fan',
  name: 'Fan',
  kind: 'exclusion',
  shape: 'rectangle',
  points: rectanglePoints(-500, 1000, 400, 400),
};

test('requires two inside samples to report occupancy', () => {
  const tracker = createZoneOccupancyTracker();
  const inside = [{ x: 0, y: 1500 }];
  expect(tracker.update([desk], inside, 0).get('desk')?.occupied).toBe(false);
  expect(tracker.update([desk], inside, 250).get('desk')?.occupied).toBe(true);
});

test('keeps occupancy through the exit grace period and then clears', () => {
  const tracker = createZoneOccupancyTracker();
  tracker.update([desk], [{ x: 0, y: 1500 }], 0);
  tracker.update([desk], [{ x: 0, y: 1500 }], 250);
  expect(tracker.update([desk], [{ x: 0, y: 5000 }], 500).get('desk')?.occupied).toBe(true);
  expect(tracker.update([desk], [{ x: 0, y: 5000 }], 500 + SPATIAL_ZONE_EXIT_MS - 100).get('desk')?.occupied).toBe(true);
  expect(tracker.update([desk], [{ x: 0, y: 5000 }], 500 + SPATIAL_ZONE_EXIT_MS).get('desk')?.occupied).toBe(false);
});

test('honours exclusion masks and inclusive boundaries', () => {
  const tracker = createZoneOccupancyTracker();
  const masked = [{ x: -400, y: 1100 }];
  tracker.update([desk, mask], masked, 0);
  expect(tracker.update([desk, mask], masked, 250).get('desk')?.occupied).toBe(false);
  expect(tracker.update([desk, mask], masked, 250).get('fan')?.occupied).toBe(true);

  const corner = [{ x: -500, y: 1000 }];
  tracker.update([desk], corner, 500);
  expect(tracker.update([desk], corner, 750).get('desk')?.occupied).toBe(true);
});

test('forgets zones removed from the workspace', () => {
  const tracker = createZoneOccupancyTracker();
  tracker.update([desk], [{ x: 0, y: 1500 }], 0);
  const afterRemoval = tracker.update([], [{ x: 0, y: 1500 }], 250);
  expect(afterRemoval.size).toBe(0);
});

test('preview text states its source and never reports missing data as clear', () => {
  expect(zonePreviewText(undefined, false, false)).toContain('offline');
  expect(zonePreviewText(undefined, true, false)).toContain('waiting for position data');
  expect(zonePreviewText({ occupied: false }, true, false)).toContain('waiting for position data');
  expect(zonePreviewText({ occupied: true }, true, true)).toBe('Target inside');
  expect(zonePreviewText({ occupied: false }, true, true)).toBe('No target inside');
});
