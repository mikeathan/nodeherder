import { describe, expect, test } from '@jest/globals';
import {
  latestSampleAt,
  playbackHopLinks,
  playbackTracks,
  playbackTrail,
  recordedEnvelope,
} from '@/domain/spatial/positions-playback';
import type { SpatialPoint, SpatialSession } from '@/services/spatial-history.service';

function point(at: number, slot: number, x: number, y: number): SpatialPoint {
  return { at, slot, x, y };
}

describe('positions playback derivation', () => {
  test('renders one track per slot at the playhead, ignoring future samples', () => {
    const points = [point(1_000, 1, 0, 0), point(2_000, 1, 100, 0), point(3_000, 2, 0, 100)];
    const tracks = playbackTracks(points, 2_500, false, 30_000);
    expect(tracks.map((track) => track.slot)).toEqual([1]);
    expect(tracks[0]).toMatchObject({ x: 100, state: 'live' });
  });

  test('marks a stale sample as held and an idle one as idle', () => {
    const points = [point(1_000, 1, 0, 0)];
    expect(playbackTracks(points, 1_000, false, 500)[0].state).toBe('live');
    expect(playbackTracks(points, 5_000, false, 1_000)[0].state).toBe('held');
    expect(playbackTracks(points, 1_000, true, 30_000)[0].state).toBe('idle');
    expect(playbackTracks(points, 0, false, 30_000)).toEqual([]);
  });

  test('converts a slot history to a polyline without inventing distance', () => {
    const trail = playbackTrail(2, [point(1, 2, 10, 20), point(2, 2, 30, 40)]);
    expect(trail.id).toBe(2);
    expect(trail.targets).toEqual([
      { id: 2, x: 10, y: 20, distance: 0 },
      { id: 2, x: 30, y: 40, distance: 0 },
    ]);
  });

  test('connects a pause or a spatial hop but not ordinary movement', () => {
    const near = [point(0, 1, 0, 0), point(1_000, 1, 100, 0)];
    expect(playbackHopLinks(near, 2_000, 10_000, 1_000)).toEqual([]);
    const paused = [point(0, 1, 0, 0), point(20_000, 1, 100, 0)];
    expect(playbackHopLinks(paused, 30_000, 10_000, 1_000)).toHaveLength(1);
    const jumped = [point(0, 1, 0, 0), point(1_000, 1, 5_000, 0)];
    expect(playbackHopLinks(jumped, 2_000, 10_000, 1_000)).toHaveLength(1);
  });

  test('reports the newest sample at or before the playhead', () => {
    const points = [point(1_000, 1, 0, 0), point(2_000, 1, 0, 0), point(3_000, 1, 0, 0)];
    expect(latestSampleAt(points, 0)).toBe(0);
    expect(latestSampleAt(points, 2_500)).toBe(2_000);
    expect(latestSampleAt(points, 9_000)).toBe(3_000);
  });

  test('fits the recorded envelope, falling back to the filter window when empty', () => {
    const sessions: SpatialSession[] = [
      { id: 'a', from: 10, to: 20, points: 1, slots: [1], capped: false },
      { id: 'b', from: 30, to: 40, points: 1, slots: [1], capped: false },
    ];
    expect(recordedEnvelope(sessions, 0, 100)).toEqual({ from: 10, to: 40 });
    expect(recordedEnvelope([], 0, 100)).toEqual({ from: 0, to: 100 });
  });
});
