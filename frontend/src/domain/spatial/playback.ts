import type { SpatialPoint, SpatialSession } from '@/services/spatial-history.service';

/**
 * Silence after which a recorded sample is treated as a new movement episode rather than a
 * continuation. The recorder stores a point only on ≥300 mm movement, so elapsed time — not
 * distance — distinguishes a pause from travel.
 */
export const POSITIONS_IDLE_GAP_MS = 10_000;
/** Trailing path window drawn brighter than the older trail. */
export const POSITIONS_RECENT_TRAIL_MS = 15_000;
/** Sample-to-sample jump that still deserves a connector when the gap was short. */
export const POSITIONS_HOP_MM = 1_000;
/** Age after which a rendered position is presented as last-known rather than current. */
export const POSITIONS_LAST_KNOWN_MS = 30_000;

/** One session placed inside the preset window, expressed as percentages of that window. */
export type SessionSegment = {
  id: string;
  from: number;
  to: number;
  leftPct: number;
  widthPct: number;
  points: number;
  slots: number[];
  capped: boolean;
};

/** First sample after a pause, offered as a "next movement" destination. */
export type MovementEpisode = { at: number; slot: number; gapMs: number };

/** Interval with no recorded sample. It is not evidence that a target was still. */
export type IdleSpan = { from: number; to: number };

/** Per-slot path points up to a time, plus the subset inside the recent window. */
export type TrailGroup = { slot: number; points: SpatialPoint[]; recent: SpatialPoint[] };

function clampPct(value: number) {
  return Math.min(100, Math.max(0, value));
}

/**
 * Places sessions inside `[from, to]` proportionally, clipping the drawn part while keeping
 * the real start/end for labels. Sessions outside the window are dropped.
 */
export function sessionSegments(
  sessions: readonly SpatialSession[],
  from: number,
  to: number
): SessionSegment[] {
  const span = Math.max(1, to - from);
  const segments: SessionSegment[] = [];
  for (const session of sessions) {
    const start = Math.max(session.from, from);
    const end = Math.min(session.to, to);
    if (end < start) continue;
    segments.push({
      id: session.id,
      from: session.from,
      to: session.to,
      leftPct: clampPct(((start - from) / span) * 100),
      widthPct: clampPct(((end - start) / span) * 100),
      points: session.points,
      slots: session.slots,
      capped: session.capped,
    });
  }
  return segments;
}

/**
 * Finds samples that start a movement episode: the first sample, then each sample that
 * follows a pause longer than `gapMs`. `points` must be ordered by `at` (the API guarantee).
 */
export function movementEpisodes(
  points: readonly SpatialPoint[],
  gapMs = POSITIONS_IDLE_GAP_MS
): MovementEpisode[] {
  const episodes: MovementEpisode[] = [];
  let previousAt: number | null = null;
  for (const point of points) {
    if (previousAt === null) {
      episodes.push({ at: point.at, slot: point.slot, gapMs: 0 });
    } else {
      const gap = point.at - previousAt;
      if (gap > gapMs) episodes.push({ at: point.at, slot: point.slot, gapMs: gap });
    }
    previousAt = point.at;
  }
  return episodes;
}

/** Returns the first episode strictly after `after`, or null when playback is at the end. */
export function nextMovement(
  episodes: readonly MovementEpisode[],
  after: number
): MovementEpisode | null {
  return episodes.find((episode) => episode.at > after) ?? null;
}

/** Intervals without a recorded sample, including the head and tail of the session. */
export function idleSpans(
  points: readonly SpatialPoint[],
  from: number,
  to: number,
  gapMs = POSITIONS_IDLE_GAP_MS
): IdleSpan[] {
  const spans: IdleSpan[] = [];
  let previous = from;
  for (const point of points) {
    if (point.at - previous > gapMs) spans.push({ from: previous, to: point.at });
    previous = Math.max(previous, point.at);
  }
  if (to - previous > gapMs) spans.push({ from: previous, to });
  return spans;
}

/**
 * Groups points up to `at` by slot. `recent` holds the points inside the trailing window plus
 * the one before it, so the highlighted line connects to the older path instead of starting
 * mid-air. Slots with no sample in the window keep an empty `recent` list.
 */
export function trailGroups(
  points: readonly SpatialPoint[],
  at: number,
  windowMs = POSITIONS_RECENT_TRAIL_MS
): TrailGroup[] {
  const bySlot = new Map<number, SpatialPoint[]>();
  for (const point of points) {
    if (point.at > at) break;
    const list = bySlot.get(point.slot) ?? [];
    list.push(point);
    bySlot.set(point.slot, list);
  }
  const since = at - windowMs;
  return [...bySlot.entries()].map(([slot, slotPoints]) => {
    const firstRecent = slotPoints.findIndex((point) => point.at >= since);
    if (firstRecent < 0) return { slot, points: slotPoints, recent: [] };
    const start = firstRecent > 0 ? firstRecent - 1 : 0;
    return { slot, points: slotPoints, recent: slotPoints.slice(start) };
  });
}

/**
 * Distance travelled per slot up to `at`, in millimetres. Summing per slot keeps interleaved
 * multi-target samples from being joined into one imaginary path.
 */
export function playedDistanceMm(points: readonly SpatialPoint[], at: number): number {
  let total = 0;
  const last = new Map<number, SpatialPoint>();
  for (const point of points) {
    if (point.at > at) break;
    const previous = last.get(point.slot);
    if (previous) total += Math.hypot(point.x - previous.x, point.y - previous.y);
    last.set(point.slot, point);
  }
  return Math.round(total);
}

/** Local-time label for a sample; `date` adds the day, `seconds` adds seconds. */
export function clock(
  at: number,
  options: { date?: boolean; seconds?: boolean } = {}
): string {
  if (!at) return '—';
  const { date = false, seconds = true } = options;
  return new Date(at).toLocaleString(undefined, {
    ...(date ? { month: '2-digit', day: '2-digit' } : {}),
    hour: '2-digit',
    minute: '2-digit',
    ...(seconds ? { second: '2-digit' } : {}),
  });
}

/** Compact duration: seconds below a minute, then minutes, then hours and minutes. */
export function durationLabel(ms: number): string {
  if (!Number.isFinite(ms) || ms <= 0) return '0 s';
  if (ms < 60_000) return `${Math.round(ms / 1000)} s`;
  if (ms < 3_600_000) return `${Math.round(ms / 60_000)} min`;
  const hours = Math.floor(ms / 3_600_000);
  const minutes = Math.round((ms % 3_600_000) / 60_000);
  return minutes ? `${hours} h ${minutes} min` : `${hours} h`;
}
