import { describe, expect, test } from '@jest/globals';
import {
  MAX_TIMELINE_MARKS,
  marksPerSegment,
  MIN_SEGMENT_PX,
  nearestPositionAt,
  positionedSpan,
  positionPct,
  renderSegments,
  sampleMarks,
  totalSampleCount,
} from '@/domain/spatial/positions-strip';
import type { SessionSegment } from '@/domain/spatial/playback';

function segment(overrides: Partial<SessionSegment>): SessionSegment {
  return {
    id: 's',
    from: 0,
    to: 1_000,
    leftPct: 0,
    widthPct: 10,
    points: 1,
    slots: [1],
    capped: false,
    activity: [],
    sampleCount: 0,
    ...overrides,
  };
}

describe('positions strip geometry', () => {
  test('draws a narrow recording at the readable minimum and keeps the row inside the track', () => {
    const narrow = segment({ id: 'a', widthPct: 0.5 });
    const wide = segment({ id: 'b', widthPct: 50, leftPct: 50 });
    const rendered = renderSegments([narrow, wide], 1_000);
    expect(rendered[0].widthPct).toBeGreaterThan(0.5);
    const sumPct = rendered.reduce((total, entry) => total + entry.widthPct, 0);
    expect(sumPct).toBeLessThanOrEqual(100.000_001);
    // Before measurement there is nothing to widen, so the proportional layout stands.
    const unmeasured = renderSegments([narrow], 0);
    expect(unmeasured[0].widthPct).toBe(0.5);
    expect(unmeasured[0].leftPct).toBe(0);
    expect(MIN_SEGMENT_PX).toBeGreaterThan(0);
  });

  test('widened recordings never overlap, so an adjacent short one stays clickable', () => {
    // Two 2 s recordings in a multi-hour window both widen to the 28 px minimum. Their
    // proportional lefts would put the second underneath the first.
    const first = segment({ id: 'first', leftPct: 10, widthPct: 0.2 });
    const second = segment({ id: 'second', leftPct: 10.3, widthPct: 0.2 });
    const rendered = renderSegments([first, second], 1_000);
    const firstRight = rendered[0].leftPct + rendered[0].widthPct;
    expect(rendered[1].leftPct).toBeGreaterThanOrEqual(firstRight - 1e-9);
    expect(rendered[1].widthPct * 10).toBeGreaterThanOrEqual(MIN_SEGMENT_PX - 1e-9);
  });

  test('bounds the marks a timeline can expose and keeps order and endpoints', () => {
    const marks = Array.from({ length: 10 }, (_, index) => ({ at: index }));
    expect(sampleMarks(marks, 0)).toEqual([]);
    expect(sampleMarks(marks, 20)).toEqual(marks);
    const sampled = sampleMarks(marks, 3);
    expect(sampled).toHaveLength(3);
    // Order preserved and both endpoints survive, so the bound never truncates to the opening.
    expect(sampled.map((mark) => mark.at)).toEqual([0, 5, 9]);
    expect(sampleMarks(marks, 1)).toEqual([marks[0]]);
    // Near the length, the picks must stay strictly ascending with no duplicates.
    const dense = Array.from({ length: 120 }, (_, index) => ({ at: index }));
    const nearlyAll = sampleMarks(dense, 119);
    expect(nearlyAll).toHaveLength(119);
    expect(new Set(nearlyAll.map((mark) => mark.at)).size).toBe(119);
  });

  test('splits one global mark budget across every recording', () => {
    for (const count of [1, 2, 7, 120, 500]) {
      const share = marksPerSegment(count);
      expect(share * count).toBeLessThanOrEqual(MAX_TIMELINE_MARKS);
    }
    expect(marksPerSegment(0)).toBe(MAX_TIMELINE_MARKS);
    expect(marksPerSegment(500)).toBe(0);
  });

  test('places a timestamp or span on a track without dividing by a zero-length window', () => {
    expect(positionPct(500, 0, 1_000)).toBe(50);
    // Outside the window is clamped, not extrapolated off the track.
    expect(positionPct(-500, 0, 1_000)).toBe(0);
    expect(positionPct(9_000, 0, 1_000)).toBe(100);
    // A single-sample window has no span to divide by.
    expect(positionPct(1_000, 1_000, 1_000)).toBe(0);
    expect(positionedSpan({ from: 250, to: 750 }, { from: 0, to: 1_000 })).toEqual({
      left: '25%',
      width: '50%',
    });
  });

  test('totals recorded positions across recordings', () => {
    expect(totalSampleCount([segment({ sampleCount: 2 }), segment({ sampleCount: 3 })])).toBe(5);
    expect(totalSampleCount([])).toBe(0);
  });

  test('resolves the recorded position nearest a click', () => {
    const target = segment({
      from: 0,
      to: 1_000,
      activity: [
        { at: 0, pct: 0 },
        { at: 500, pct: 50 },
        { at: 1_000, pct: 100 },
      ],
    });
    expect(nearestPositionAt(target, 0)).toBe(0);
    expect(nearestPositionAt(target, 0.45)).toBe(500);
    expect(nearestPositionAt(target, 0.9)).toBe(1_000);
    // A click outside the span is clamped, not extrapolated.
    expect(nearestPositionAt(target, 5)).toBe(1_000);
    expect(nearestPositionAt(target, -2)).toBe(0);
    // Nothing recorded: the recording's own start, never an invented moment.
    expect(nearestPositionAt(segment({ activity: [], from: 42 }), 0.5)).toBe(42);
  });
});
