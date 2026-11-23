import type { SessionSegment } from '@/domain/spatial/playback';

/**
 * Geometry for the range timeline: where each recording is drawn.
 *
 * This used to live in `PositionsCoverageBar.vue`, so the component owned layout maths as well as
 * presentation. It is a pure function of the segments and the measured track width, which makes it
 * testable without mounting a component.
 */

/**
 * Target minimum drawn width for a recording, so its movement marks stay readable and hittable.
 * The row is widened to this where the track allows and rescaled when it does not: keeping every
 * recording visible wins over the minimum, so a very dense row can draw narrower segments.
 */
export const MIN_SEGMENT_PX = 28;

/** Global budget for focusable movement marks across the whole timeline. */
export const MAX_TIMELINE_MARKS = 120;

/**
 * The per-recording share of the global mark budget.
 *
 * A per-recording cap alone lets a 500-recording filter expose thousands of focus stops ahead of
 * the transport. Splitting one budget keeps the tab order bounded; a zero share means only the
 * selected/pending recordings keep a control.
 */
export function marksPerSegment(
  segmentCount: number,
  budget = MAX_TIMELINE_MARKS
): number {
  return Math.max(0, Math.floor(budget / Math.max(1, segmentCount)));
}

/** One recording as the timeline draws it. */
export type RenderedSegment = {
  segment: SessionSegment;
  /** Left offset and width as a percentage of the whole track. */
  leftPct: number;
  widthPct: number;
};

/**
 * Places every recording on the track at a readable minimum width.
 *
 * @param segments Recordings already placed proportionally by `sessionSegments`.
 * @param trackWidth Measured track width in pixels; 0 before the first measurement, in which
 * case nothing is widened and the row keeps its proportional layout.
 */
export function renderSegments(
  segments: readonly SessionSegment[],
  trackWidth: number
): RenderedSegment[] {
  /*
   * Before the first measurement (and while the timeline is hidden) there is no pixel width to
   * widen into. Returning the proportional layout keeps the drawn order stable; the previous code
   * divided by a 1 px floor, which expanded every recording to full width on first paint.
   */
  if (trackWidth <= 0) {
    return segments.map((segment) => ({
      segment,
      leftPct: segment.leftPct,
      widthPct: segment.widthPct,
    }));
  }
  const minPct = (MIN_SEGMENT_PX / trackWidth) * 100;
  /*
   * Keep each recording at its proportional left where it fits, but never let it overlap the one
   * before it. Widening a 2 s recording to the readable minimum while leaving every left edge
   * proportional made adjacent short recordings sit on top of each other, so the later one was
   * unclickable. A recording is pushed right only when its widened predecessor reaches it.
   */
  let cursor = 0;
  const placed = segments.map((segment) => {
    const left = Math.max(segment.leftPct, cursor);
    const width = Math.max(minPct, segment.widthPct);
    cursor = left + width;
    return { segment, left, width };
  });
  // Keep the row inside the track: clipping the last recordings would hide them entirely.
  const scale = cursor > 100 ? 100 / cursor : 1;
  return placed.map((entry) => ({
    segment: entry.segment,
    leftPct: entry.left * scale,
    widthPct: entry.width * scale,
  }));
}

/**
 * Evenly reduces ordered marks to at most `limit`, preserving order.
 *
 * A single dense recording is bounded by `MAX_RECORDING_MARKS`, but a filter can hold hundreds
 * of recordings; the timeline splits one global budget across them so its tab order cannot grow
 * without bound. Both endpoints survive whenever `limit` allows two picks; a limit of one can only
 * keep the first mark.
 */
export function sampleMarks<T extends { at: number }>(marks: readonly T[], limit: number): T[] {
  if (limit <= 0) return [];
  if (marks.length <= limit) return [...marks];
  if (limit === 1) return [marks[0]];
  // Spread the picks over the closed range so both endpoints survive, not just the opening.
  const step = (marks.length - 1) / (limit - 1);
  return Array.from({ length: limit }, (_, index) => marks[Math.round(index * step)]);
}

/**
 * The recorded position nearest a click, as an epoch-millisecond timestamp.
 *
 * Clicking anywhere in a recording plays from the moment closest to the pointer, so reaching a
 * specific instant takes one gesture instead of hunting a 2 px mark.
 *
 * @param segment The recording the pointer landed in.
 * @param ratio Pointer position across that recording, 0 at its start and 1 at its end.
 */
export function nearestPositionAt(segment: SessionSegment, ratio: number): number {
  const positions = segment.activity;
  if (!positions.length) return segment.from;
  const clamped = Math.min(1, Math.max(0, ratio));
  const at = segment.from + clamped * (segment.to - segment.from);
  let best = positions[0];
  for (const position of positions) {
    if (Math.abs(position.at - at) < Math.abs(best.at - at)) best = position;
  }
  return best.at;
}

/**
 * Placement of a timestamp inside a window, as a percentage, clamped to the window.
 *
 * The transport's scrubber marks and the range timeline both place things this way, and each
 * used to carry its own copy of the clamp. A zero-length window (one sample) yields 0 rather than
 * a division by zero.
 *
 * @param at Timestamp to place, epoch milliseconds.
 * @param from Window start; a window shorter than 1 ms is treated as 1 ms.
 * @param to Window end.
 */
export function positionPct(at: number, from: number, to: number): number {
  const span = Math.max(1, to - from);
  return Math.min(100, Math.max(0, ((at - from) / span) * 100));
}

/** A span (gap or idle stretch) placed on a track, as CSS offsets. */
export function positionedSpan(
  span: { from: number; to: number },
  window: { from: number; to: number }
): { left: string; width: string } {
  const left = positionPct(span.from, window.from, window.to);
  const right = positionPct(span.to, window.from, window.to);
  return { left: `${left}%`, width: `${Math.max(0, right - left)}%` };
}

/** Total recorded positions across recordings; the range head states it once. */
export function totalSampleCount(segments: readonly SessionSegment[]): number {
  return segments.reduce((total, segment) => total + segment.sampleCount, 0);
}
