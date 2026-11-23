import { describe, expect, test } from '@jest/globals';
import {
  clock,
  durationLabel,
  gapAfter,
  idleSpans,
  movementChapters,
  movementEpisodes,
  MAX_RECORDING_MARKS,
  nextMovement,
  playedDistanceMm,
  playbackRecordings,
  recordingNeighbours,
  recordingOrder,
  sanitiseSamples,
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

  test('accepts only finite, non-negative, de-duplicated movement timestamps', () => {
    expect(sanitiseSamples(undefined)).toEqual([]);
    expect(sanitiseSamples('nope')).toEqual([]);
    expect(sanitiseSamples([5, 0, -2, 'x', 5, Number.NaN, Number.POSITIVE_INFINITY, 1])).toEqual([
      0, 1, 5,
    ]);
  });

  test('places a movement inside its own recording and drops out-of-range marks', () => {
    const [segment] = sessionSegments([
      session({ id: 'a', from: 0, to: 100, samples: [0, 25, 75, 500] }),
    ], 0, 1_000);
    expect(segment.activity.map(({ at, pct }) => ({ at, pct }))).toEqual([
      { at: 0, pct: 0 },
      { at: 25, pct: 25 },
      { at: 75, pct: 75 },
    ]);
    expect(segment.sampleCount).toBe(4);
  });

  test('reduces dense activity but keeps the true sample count', () => {
    const samples = Array.from({ length: 100 }, (_, index) => index + 1);
    const [segment] = sessionSegments([session({ id: 'a', from: 1, to: 100, samples })], 1, 100);
    expect(segment.activity.length).toBeLessThanOrEqual(MAX_RECORDING_MARKS);
    expect(segment.activity.length).toBeGreaterThan(0);
    expect(segment.sampleCount).toBe(100);
  });

  test('orders recordings chronologically regardless of transport order', () => {
    const sessions = [
      session({ id: 'late', from: 900, to: 950 }),
      session({ id: 'early', from: 0, to: 50 }),
      session({ id: 'middle', from: 500, to: 550 }),
    ];
    expect(recordingOrder(sessions)).toEqual(['early', 'middle', 'late']);
    expect(playbackRecordings(sessions).map((recording) => recording.id)).toEqual([
      'early',
      'middle',
      'late',
    ]);
  });

  test('resolves neighbours and boundary flags', () => {
    const sessions = [
      session({ id: 'a', from: 0, to: 50 }),
      session({ id: 'b', from: 500, to: 550 }),
      session({ id: 'c', from: 900, to: 950 }),
    ];
    expect(recordingNeighbours(sessions, 'a')).toEqual({
      index: 0,
      count: 3,
      previous: null,
      next: 'b',
    });
    expect(recordingNeighbours(sessions, 'b')).toEqual({
      index: 1,
      count: 3,
      previous: 'a',
      next: 'c',
    });
    expect(recordingNeighbours(sessions, 'c')).toMatchObject({ index: 2, next: null });
    expect(recordingNeighbours(sessions, 'missing')).toMatchObject({ index: -1, count: 3 });
  });

  test('reports only gaps long enough to be worth announcing', () => {
    const recordings = playbackRecordings([
      session({ id: 'a', from: 0, to: 10_000 }),
      session({ id: 'b', from: 10_030_000, to: 10_040_000 }),
      session({ id: 'c', from: 10_050_000, to: 10_060_000 }),
    ]);
    expect(gapAfter(recordings, 'a')).toEqual({ from: 10_000, to: 10_030_000 });
    expect(gapAfter(recordings, 'b')).toBeNull();
    expect(gapAfter(recordings, 'c')).toBeNull();
  });

  test('labels the scrubber chapters and states how many were sampled', () => {
    // Each sample here follows a pause longer than the 10 s episode gap, so all three count.
    const points = [point(0, 1), point(30_000, 1), point(90_000, 1)];
    const { chapters, total } = movementChapters(points, 0, 90_000);
    expect(total).toBe(3);
    expect(chapters.map((chapter) => chapter.at)).toEqual([0, 30_000, 90_000]);
    expect(chapters[1].pct).toBeCloseTo(100 / 3, 5);
  });
});

test('places marks relative to the span the segment is drawn over, not the whole recording', () => {
  const [segment] = sessionSegments(
    [session({ id: 'a', from: 0, to: 1_000, samples: [0, 250, 500, 750, 1_000] })],
    500,
    1_000
  );
  // The segment is clipped to [500, 1000]; its marks must be placed across that drawn span, or a
  // mark would point at a moment the visible segment does not represent.
  expect(segment.activity.map(({ at, pct }) => ({ at, pct }))).toEqual([
    { at: 500, pct: 0 },
    { at: 750, pct: 50 },
    { at: 1_000, pct: 100 },
  ]);
});
