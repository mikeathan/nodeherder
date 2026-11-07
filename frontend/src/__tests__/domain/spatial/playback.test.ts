import { describe, expect, test } from '@jest/globals';
import {
  clock,
  durationLabel,
  idleSpans,
  movementEpisodes,
  nextMovement,
  playedDistanceMm,
  sessionSegments,
  trailGroups,
} from '@/domain/spatial/playback';
import type { SpatialPoint, SpatialSession } from '@/services/spatial-history.service';

function session(overrides: Partial<SpatialSession>): SpatialSession {
  return { id: 's', from: 0, to: 1, points: 1, slots: [1], capped: false, ...overrides };
}

function point(at: number, slot: number, x = 0, y = 0): SpatialPoint {
  return { at, slot, x, y };
}

describe('positions playback helpers', () => {
  test('places sessions proportionally and clips them to the window', () => {
    const segments = sessionSegments(
      [
        session({ id: 'a', from: 0, to: 100 }),
        session({ id: 'b', from: 500, to: 600 }),
        session({ id: 'c', from: 2_000, to: 3_000 }),
        session({ id: 'd', from: -200, to: 50 }),
      ],
      0,
      1_000
    );
    expect(segments.map(({ id, leftPct, widthPct }) => ({ id, leftPct, widthPct }))).toEqual([
      { id: 'a', leftPct: 0, widthPct: 10 },
      { id: 'b', leftPct: 50, widthPct: 10 },
      { id: 'd', leftPct: 0, widthPct: 5 },
    ]);
  });

  test('marks the first sample and each sample after a pause as a movement episode', () => {
    const episodes = movementEpisodes([
      point(0, 1),
      point(2_000, 1),
      point(12_000, 1),
      point(30_000, 1),
    ]);
    expect(episodes.map((episode) => episode.at)).toEqual([0, 30_000]);
    expect(episodes[1].gapMs).toBe(18_000);
  });

  test('next movement only looks forward and stops at the end', () => {
    const episodes = movementEpisodes([point(0, 1), point(30_000, 1)]);
    expect(nextMovement(episodes, -1)?.at).toBe(0);
    expect(nextMovement(episodes, 0)?.at).toBe(30_000);
    expect(nextMovement(episodes, 30_000)).toBeNull();
  });

  test('reports gaps, including the head and tail of a session', () => {
    expect(idleSpans([point(15_000, 1), point(30_000, 1)], 0, 45_000)).toEqual([
      { from: 0, to: 15_000 },
      { from: 15_000, to: 30_000 },
      { from: 30_000, to: 45_000 },
    ]);
  });

  test('keeps a connector point before the recent trail window', () => {
    const [group] = trailGroups([point(0, 1), point(5_000, 1), point(20_000, 1)], 20_000, 15_000);
    expect(group.points.map((entry) => entry.at)).toEqual([0, 5_000, 20_000]);
    expect(group.recent.map((entry) => entry.at)).toEqual([0, 5_000, 20_000]);
  });

  test('leaves a slot without recent samples empty instead of inventing a line', () => {
    const [group] = trailGroups([point(0, 1)], 20_000, 15_000);
    expect(group.recent).toEqual([]);
  });

  test('sums distance per slot up to the playhead only', () => {
    const points = [point(0, 1, 0, 0), point(1_000, 2, 1_000, 0), point(2_000, 1, 300, 400)];
    expect(playedDistanceMm(points, -1)).toBe(0);
    expect(playedDistanceMm(points, 1_000)).toBe(0);
    expect(playedDistanceMm(points, 2_000)).toBe(500);
  });

  test('formats clock values and durations for the transport labels', () => {
    expect(clock(0)).toBe('—');
    expect(clock(1_700_000_000_000)).toMatch(/\d{2}:\d{2}:\d{2}/);
    expect(durationLabel(0)).toBe('0 s');
    expect(durationLabel(45_000)).toBe('45 s');
    expect(durationLabel(240_000)).toBe('4 min');
    expect(durationLabel(3_600_000)).toBe('1 h');
    expect(durationLabel(3_660_000)).toBe('1 h 1 min');
  });
});
