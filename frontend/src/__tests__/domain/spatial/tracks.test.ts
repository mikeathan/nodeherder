import {
  createTrackAssociator,
  SPATIAL_MERGE_MM,
  SPATIAL_TRACK_RETAIN_MS,
  type SpatialTrackInput,
} from '@/domain/spatial/tracks';

function point(slot: number, x: number, y: number, at = 0): SpatialTrackInput {
  return { slot, x, y, vx: 0, vy: 0, at, state: 'live' };
}

test('keeps a label with the position when slots swap', () => {
  const associator = createTrackAssociator();
  const first = associator.update([point(1, 0, 1000), point(2, 4000, 1000)], 0);
  expect(first.tracks.map((track) => track.label)).toEqual(['Track 1', 'Track 2']);

  const swapped = associator.update([point(1, 4000, 1000, 250), point(2, 0, 1000, 250)], 250);
  expect(swapped.tracks).toHaveLength(2);
  const nearSensor = swapped.tracks.find((track) => track.x === 0);
  expect(nearSensor?.label).toBe('Track 1');
  expect(nearSensor?.slot).toBe(2);
});

test('creates a new track beyond the association gate and keeps the old identity', () => {
  const associator = createTrackAssociator();
  associator.update([point(1, 0, 1000)], 0);
  const moved = associator.update([point(2, 3000, 1000, 250)], 250);
  expect(moved.tracks.map((track) => track.id)).toEqual([2]);

  // Only observed tracks are drawn, but the lost identity is retained for re-acquisition.
  const both = associator.update([point(1, 0, 1000, 500), point(2, 3000, 1000, 500)], 500);
  expect(both.tracks.map((track) => track.id)).toEqual([1, 2]);
});

test('retains identity across a brief loss and forgets it after retention', () => {
  const associator = createTrackAssociator();
  associator.update([point(1, 0, 1000)], 0);
  const reappeared = associator.update([point(1, 0, 1000, 5000)], 5000);
  expect(reappeared.tracks.map((track) => track.id)).toEqual([1]);

  const forgotten = associator.update([point(1, 0, 1000, SPATIAL_TRACK_RETAIN_MS + 6000)], SPATIAL_TRACK_RETAIN_MS + 6000);
  expect(forgotten.tracks.map((track) => track.id)).toEqual([2]);
});

test('merges two slots reporting one target and reverses when they separate', () => {
  const associator = createTrackAssociator();
  const at = (step: number) => step * 250;
  associator.update([point(1, 0, 1000), point(2, SPATIAL_MERGE_MM / 2, 1000)], 0);
  associator.update([point(1, 0, 1000), point(2, SPATIAL_MERGE_MM / 2, 1000)], 250);

  const merged = associator.update([point(1, 0, 1000), point(2, SPATIAL_MERGE_MM / 2, 1000)], 500);
  expect(merged.tracks).toHaveLength(1);
  expect(merged.duplicates).toBe(1);

  const separated = associator.update([point(1, 0, 1000), point(2, 2000, 1000)], at(3));
  expect(separated.tracks).toHaveLength(2);
  expect(separated.duplicates).toBe(0);
});

test('reset releases identities', () => {
  const associator = createTrackAssociator();
  associator.update([point(1, 0, 1000)], 0);
  associator.reset();
  const fresh = associator.update([point(1, 0, 1000)], 250);
  expect(fresh.tracks[0].id).toBe(1);
  expect(fresh.tracks[0].colorIndex).toBe(0);
});
