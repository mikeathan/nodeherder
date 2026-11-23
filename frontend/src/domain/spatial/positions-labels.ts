import type { SpatialSession } from '@/services/spatial-history.service';
import { formatMillimetres } from '@/utils/units';
import {
  clock as clockLabel,
  durationLabel,
  type PlaybackGap,
  type RecordingNeighbours,
  type SessionSegment,
} from '@/domain/spatial/playback';

/**
 * User-facing labels for the recorded-positions surfaces.
 *
 * The transport, the range timeline, the recordings rail and the sheet all describe the same
 * few things (a recording, a moment, a gap, a sample count). Each used to build those strings
 * itself, so the same recording was worded differently in the rail and in the timeline. This
 * module owns the wording; components only place it.
 *
 * All timestamps are epoch milliseconds in the viewer's local zone, as `clock` renders them.
 */

/** "1 recording" / "2 recordings"; the one pluraliser for this family. */
export function countLabel(count: number, singular: string, plural = `${singular}s`): string {
  return `${count} ${count === 1 ? singular : plural}`;
}

/**
 * Whether a recording spans enough time to be scrubbed.
 *
 * A single stored sample has `to === from`, so its duration reads "0 s", both axis ends are the
 * same instant and the mark row is empty. Rendering that as a dead control looks like a fault.
 */
export function hasPlayableSpan(session: SpatialSession): boolean {
  return session.to - session.from >= 1_000;
}

/**
 * The precision a time label needs for a given window.
 *
 * A day-wide window hides the date; a window under an hour needs seconds, otherwise four
 * recordings a few minutes apart all print the same range.
 */
export function timeLabelOptions(spanMs: number): { date: boolean; seconds: boolean } {
  return { date: spanMs > 86_400_000, seconds: spanMs < 3_600_000 };
}

/** "02:05 PM – 02:07 PM" for two timestamps. */
export function clockRange(
  from: number,
  to: number,
  options: { date?: boolean; seconds?: boolean } = {}
): string {
  return `${clockLabel(from, options)} – ${clockLabel(to, options)}`;
}

/**
 * "Recording 2 of 4 · 11/21, 02:05 PM – 11/21, 02:07 PM".
 *
 * Reached through the transport, so it also states where in the filter the operator is: paging
 * without knowing the position is disorienting.
 */
export function recordingPositionLabel(
  session: SpatialSession,
  neighbours: RecordingNeighbours
): string {
  const position =
    neighbours.index >= 0
      ? `Recording ${neighbours.index + 1} of ${neighbours.count}`
      : 'Recording';
  return `${position} · ${clockRange(session.from, session.to, { date: true, seconds: false })}`;
}

/** "Skipped 4 min with no recording, 02:05 PM – 02:09 PM." — for a crossed gap. */
export function skippedGapLabel(gap: PlaybackGap): string {
  return `Skipped ${durationLabel(gap.to - gap.from)} with no recording, ${clockRange(gap.from, gap.to)}.`;
}

/** "Next recording is after a 4 min gap with no recording." — for a pending gap. */
export function pendingGapLabel(gap: PlaybackGap): string {
  return `Next recording is after a ${durationLabel(gap.to - gap.from)} gap with no recording.`;
}

/**
 * The idle line, bucketed so a live region does not re-announce every tick.
 *
 * The playhead advances every 200 ms, so an unrounded duration inside an `aria-live` region
 * re-announces the same sentence five times a second and makes the screen reader unusable.
 * Even whole seconds still re-announce a long idle tail every second, so the duration is bucketed
 * to ten seconds under a minute and to minutes beyond it. Returns '' while the bucket is zero.
 */
export function idleRecordingLabel(idleForMs: number, lastSampleAt: number): string {
  const bucketMs = idleForMs < 60_000 ? 10_000 : 60_000;
  const seconds = Math.floor(idleForMs / bucketMs) * (bucketMs / 1_000);
  if (seconds <= 0) return '';
  return `No movement recorded for ${durationLabel(seconds * 1_000)} · last sample ${clockLabel(lastSampleAt)}`;
}

/**
 * The one status line under the transport.
 *
 * Order matters: a skipped gap and an idle stretch describe where the playhead is right now;
 * the sample count and the next-recording hint are context.
 */
export function transportStatusLabel(input: {
  skippedGap: PlaybackGap | null;
  pendingGap: PlaybackGap | null;
  session: SpatialSession;
  idleForMs: number;
  lastSampleAt: number;
  awaitingCrossing: boolean;
  reachedEnd: boolean;
  movementCount: number;
  renderedMovements: number;
}): string {
  if (input.skippedGap) return skippedGapLabel(input.skippedGap);
  if (!hasPlayableSpan(input.session)) {
    return `One recorded sample at ${clockLabel(input.session.from)}. Nothing to play or scrub.`;
  }
  const idle = idleRecordingLabel(input.idleForMs, input.lastSampleAt);
  if (idle) return idle;
  if (input.awaitingCrossing) return 'Loading the next recording…';
  if (input.reachedEnd) return 'End of the filtered recordings.';
  if (input.pendingGap) return pendingGapLabel(input.pendingGap);
  return movementCountLabel(input.movementCount, input.renderedMovements);
}

/** The movement summary for one recording, stating when the marks were sampled. */
export function movementCountLabel(count: number, rendered = count): string {
  if (!count) return 'No movement recorded in this recording.';
  if (rendered < count) return `${count} movements in this recording, ${rendered} shown`;
  return `${countLabel(count, 'movement')} in this recording`;
}

/** The range-timeline head: what the strip is drawn over. */
export function rangeHeadlineLabel(count: number, from: number, to: number): string {
  const options = timeLabelOptions(to - from);
  return `${countLabel(count, 'recording')} between ${clockLabel(from, options)} and ${clockLabel(to, options)}`;
}

/** The range-timeline accessible name for one recording's segment. */
export function segmentLabel(segment: SessionSegment, showDate: boolean): string {
  const slots = segment.slots.length ? segment.slots.join(', ') : 'none';
  const capped = segment.capped ? ', capped' : '';
  const activity = segment.sampleCount
    ? `, ${countLabel(segment.sampleCount, 'recorded position')}`
    : ', nothing recorded';
  return `${clockRange(segment.from, segment.to, { date: showDate })}, ${durationLabel(
    segment.to - segment.from
  )}, ${countLabel(segment.points, 'point')}, slot ${slots}${capped}${activity}`;
}

/**
 * The accessible name for one recorded-position mark on the range timeline.
 *
 * The summary carries timestamps, not per-sample slots, so the mark names its time and the
 * recording it belongs to; it must not imply a slot the wire did not provide.
 */
export function movementMarkLabel(
  segment: SessionSegment,
  at: number,
  showDate: boolean
): string {
  return `Movement at ${clockLabel(at, { date: showDate })} in ${clockRange(
    segment.from,
    segment.to,
    { date: showDate }
  )}`;
}

/**
 * The accessible name for one movement mark on the recording scrubber.
 *
 * The scrubber's points do carry a slot, so the mark may name it; the range timeline's summary
 * does not, which is why `movementMarkLabel` must not.
 */
export function movementChapterLabel(at: number, slot: number): string {
  return `Movement at ${clockLabel(at)} on slot ${slot}`;
}

/** The scrubber's value text: the playhead against the recording's end. */
export function playbackValueLabel(at: number, to: number): string {
  return `${clockLabel(at)} of ${clockLabel(to)}`;
}

/** "350 mm of 1.2 km": distance covered against the recording's total. */
export function distanceProgressLabel(coveredMm: number, totalMm: number): string {
  return `${formatMillimetres(coveredMm)} of ${formatMillimetres(totalMm)}`;
}

/**
 * The control a recording keeps when it has no rendered movement marks (older hub, or a tight
 * global budget). It selects the recording and plays from its start; it never implies a measured
 * position.
 */
export function fallbackMarkLabel(segment: SessionSegment, showDate: boolean): string {
  return `Play from ${clockLabel(segment.from, { date: showDate })} · ${segmentLabel(
    segment,
    showDate
  )}`;
}

/** The rail row's time range and metadata: "02:05 – 02:07 PM" and "2 min · slot 1". */
export function railTimeLabel(session: SpatialSession, spanMs: number): string {
  return clockRange(session.from, session.to, timeLabelOptions(spanMs));
}

/**
 * The rail row's second line.
 *
 * A single sample is not "0 s · 1 point"; it is one sample from one slot. Zero-second durations
 * are an artefact of sampling, not a length worth reading.
 */
export function railMetaLabel(session: SpatialSession): string {
  const slotLabel = session.slots.length === 1 ? `slot ${session.slots.join(', ')}` : `slots ${
    session.slots.length ? session.slots.join(', ') : 'none'
  }`;
  const duration = session.to - session.from;
  const length =
    duration >= 1_000 ? durationLabel(duration) : countLabel(session.points, 'sample');
  return `${length} · ${slotLabel}`;
}

/**
 * Empty copy, scoped to what the surface is.
 *
 * The range timeline and the recordings list had drifted to different sentences for the same
 * state ("No recordings" vs "No recorded positions"). Both are owned here now, so the wording is
 * a decision rather than an accident: the timeline speaks of the range it draws, the list of the
 * recordings it lists.
 */
export const RANGE_EMPTY_COPY =
  'No recordings in this range. Enable “Record history” on the device and let a target move.';
export const RECORDINGS_EMPTY_COPY =
  'No recorded positions in this range. Enable “Record history” on the device and let a target move.';
export const RECORDINGS_LOADING_COPY = 'Loading recordings…';

/**
 * What an empty-looking recording means, given what its own summary claimed.
 *
 * The summary counts points inside the filter and the detail returns them, so the two can
 * legitimately disagree (a cap, or a chunk pruned between the requests). Saying only "no
 * positions" makes the operator think the recording is broken.
 */
export function emptyRecordingLabel(listedPoints: number): string {
  if (listedPoints <= 0) return 'No recorded positions in this session.';
  return `This recording lists ${countLabel(listedPoints, 'position')} in this range, but none could be loaded.`;
}
