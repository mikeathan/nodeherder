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
/** Gap between recordings that is worth stating as skipped; below it the line would be noise. */
export const POSITIONS_GAP_NOTICE_MS = 60_000;
/**
 * Maximum labelled movement marks rendered for one recording's scrubber. Every mark is a real
 * focus stop, so the cap is also a tab-order bound: a dense session must not trap a keyboard
 * user in a hundred stops. Sampled marks state how many they stand for (see movementChapters).
 */
export const MAX_TRANSPORT_MARKS = 40;
/**
 * Maximum recorded-position marks placed inside one recording on the range timeline. The strip is
 * a few hundred pixels wide on its own; a filter with many recordings is further reduced by the
 * global budget in `positions-strip`.
 */
export const MAX_RECORDING_MARKS = 60;
/** Server-side bound for `Session.Samples` (`MaxSamplesPerSession`). */
export const MAX_SESSION_SAMPLES = 240;

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
  /** Recorded positions inside this recording, placed as a percentage of it. */
  activity: RecordingMovement[];
  /** True sample count, which may exceed the rendered blocks (see MAX_RECORDING_MARKS). */
  sampleCount: number;
};

/** One recorded position inside a recording, positioned relative to that recording. */
export type RecordingMovement = { at: number; pct: number };

/** A recording placed on the unified playback timeline. */
export type PlaybackRecording = { id: string; from: number; to: number; samples: number[] };

/** Where a recording sits among the filtered ones, and which recording is adjacent to it. */
export type RecordingNeighbours = {
  /** Zero-based position among the chronological recordings; -1 when the id is unknown. */
  index: number;
  /** Number of filtered recordings. */
  count: number;
  /** Earlier recording id, or null at the start. */
  previous: string | null;
  /** Later recording id, or null at the end. */
  next: string | null;
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
 * the real start/end for labels. Sessions outside the window are dropped. Each recorded position
 * inside the drawn span is placed as a percentage of that span; more than MAX_RECORDING_MARKS are
 * sampled so a dense day cannot create thousands of nodes, while `sampleCount` keeps the true
 * total the payload reported.
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
    // Validated once: the count and the marks must describe the same samples.
    const recorded = sanitiseSamples(session.samples);
    segments.push({
      id: session.id,
      from: session.from,
      to: session.to,
      leftPct: clampPct(((start - from) / span) * 100),
      widthPct: clampPct(((end - start) / span) * 100),
      points: session.points,
      slots: session.slots,
      capped: session.capped,
      activity: activityMarks(recorded, start, end),
      sampleCount: recorded.length,
    });
  }
  return segments;
}

/**
 * Places a recording's recorded positions inside the span it is actually drawn over. Times outside
 * that span are dropped rather than clamped to an edge, so a mark always names a time the drawn
 * segment holds, and their percentage is relative to the drawn span. Dense spans are reduced to at
 * most MAX_RECORDING_MARKS evenly spread marks.
 */
function activityMarks(
  recorded: readonly number[],
  start: number,
  end: number
): RecordingMovement[] {
  const span = Math.max(1, end - start);
  const inside = recorded.filter((at) => at >= start && at <= end);
  const step = Math.ceil(inside.length / MAX_RECORDING_MARKS) || 1;
  return inside
    .filter((_, index) => index % step === 0)
    .map((at) => ({ at, pct: clampPct(((at - start) / span) * 100) }));
}

/**
 * Validates the additive `samples` field at the browser ingress. The field is optional (an older
 * hub omits it) and untrusted: only finite, ordered, duplicated-free numbers are kept, bounded to
 * the server's own per-session cap. Absent or invalid data becomes no blocks rather than
 * corrupting playback state.
 */
export function sanitiseSamples(value: unknown): number[] {
  if (!Array.isArray(value)) return [];
  const kept = new Set<number>();
  for (const entry of value) {
    if (typeof entry !== 'number' || !Number.isFinite(entry) || entry < 0) continue;
    kept.add(entry);
    if (kept.size >= MAX_SESSION_SAMPLES) break;
  }
  return [...kept].sort((a, b) => a - b);
}

/**
 * Chronological order of the filtered recordings. The API returns ascending `from`, but
 * playback ordering is a domain rule and must not depend on transport order.
 */
export function recordingOrder(sessions: readonly SpatialSession[]): string[] {
  return [...sessions].sort((a, b) => a.from - b.from).map((session) => session.id);
}

/**
 * Where `id` sits among the filtered recordings and which recordings are adjacent. `previous`
 * is null at the start and `next` null at the end, so the controls can be disabled honestly:
 * Previous and Next always move exactly one recording, which keeps repeated presses predictable.
 */
export function recordingNeighbours(
  sessions: readonly SpatialSession[],
  id: string
): RecordingNeighbours {
  const order = recordingOrder(sessions);
  const index = order.indexOf(id);
  if (index < 0) {
    return { index: -1, count: order.length, previous: null, next: null };
  }
  return {
    index,
    count: order.length,
    previous: order[index - 1] ?? null,
    next: order[index + 1] ?? null,
  };
}

/**
 * The filtered recordings as one chronological timeline, including the gaps between them.
 * Playback uses this to cross a boundary explicitly: a gap is reported rather than silently
 * accelerated over, because no movement was recorded there.
 */
export function playbackRecordings(sessions: readonly SpatialSession[]): PlaybackRecording[] {
  return [...sessions]
    .sort((a, b) => a.from - b.from)
    .map((session) => ({
      id: session.id,
      from: session.from,
      to: session.to,
      samples: sanitiseSamples(session.samples),
    }));
}

/** Explicit skipped interval between two consecutive recordings. */
export type PlaybackGap = { from: number; to: number };

/**
 * Gap between `currentId` and the recording that follows it, or null when none follows. A gap
 * shorter than POSITIONS_GAP_NOTICE_MS is crossed silently: the player should not stop the
 * story to announce ten seconds.
 */
export function gapAfter(
  recordings: readonly PlaybackRecording[],
  currentId: string
): PlaybackGap | null {
  const index = recordings.findIndex((recording) => recording.id === currentId);
  const current = recordings[index];
  const next = recordings[index + 1];
  if (!current || !next) return null;
  return noticeGap(current.to, next.from);
}

/**
 * Gap that precedes `currentId`, reported when a continuous advance has just crossed it. Null
 * for the first recording or when the interruption was too short to mention.
 */
export function gapBefore(
  recordings: readonly PlaybackRecording[],
  currentId: string
): PlaybackGap | null {
  const index = recordings.findIndex((recording) => recording.id === currentId);
  const previous = recordings[index - 1];
  const current = recordings[index];
  if (!previous || !current) return null;
  return noticeGap(previous.to, current.from);
}

function noticeGap(from: number, to: number): PlaybackGap | null {
  return to - from > POSITIONS_GAP_NOTICE_MS ? { from, to } : null;
}

/** A movement mark on the recording scrubber, carrying its sampled position. */
export type MovementChapter = { at: number; slot: number; pct: number };

/**
 * Movement episodes for the scrubber, each placed as a percentage of the recording. Dense
 * recordings are sampled deterministically and `total` reports the true count, so the labels
 * can say "24 movements, 12 shown" instead of pretending the sampled set is everything.
 */
export function movementChapters(
  points: readonly SpatialPoint[],
  from: number,
  to: number,
  gapMs = POSITIONS_IDLE_GAP_MS
): { chapters: MovementChapter[]; total: number } {
  const span = Math.max(1, to - from);
  const episodes = movementEpisodes(points, gapMs);
  const step = Math.ceil(episodes.length / MAX_TRANSPORT_MARKS) || 1;
  const chapters = episodes
    .filter((_, index) => index % step === 0)
    .map((episode) => ({
      at: episode.at,
      slot: episode.slot,
      pct: clampPct(((episode.at - from) / span) * 100),
    }));
  return { chapters, total: episodes.length };
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
