import type { SpatialPoint, SpatialSession } from '@/services/spatial-history.service';
import type { SpatialDisplayTrack } from '@/domain/spatial/tracks';
import { clock } from '@/domain/spatial/playback';

/**
 * The recorded positions as canvas tracks at a playhead.
 *
 * A slot whose last sample is older than `staleAfterMs` is presented as last-known rather than
 * live, and a slot inside an idle span is presented as idle. Both states are derived here, not in
 * the component, so the canvas, the transport and the tests cannot disagree about them.
 *
 * @param points Recorded positions, ordered by `at` (the API guarantee).
 * @param at Playhead in epoch milliseconds.
 * @param inIdleGap Whether the playhead currently sits inside a span with no recorded sample.
 * @param staleAfterMs Age after which a position is last-known.
 * @returns One track per slot, or none before the first sample.
 */
export function playbackTracks(
  points: readonly SpatialPoint[],
  at: number,
  inIdleGap: boolean,
  staleAfterMs: number
): SpatialDisplayTrack[] {
  if (!at) return [];
  const latest = new Map<number, SpatialPoint>();
  for (const point of points) {
    if (point.at > at) break;
    latest.set(point.slot, point);
  }
  return [...latest.entries()].map(([slot, point]): SpatialDisplayTrack => {
    const stale = at - point.at > staleAfterMs;
    const idle = !stale && inIdleGap;
    return {
      id: slot,
      label: stale ? `Slot ${slot} · last seen ${clock(point.at)}` : `Slot ${slot}`,
      slot,
      x: point.x,
      y: point.y,
      rawX: point.x,
      rawY: point.y,
      vx: 0,
      vy: 0,
      at: point.at,
      colorIndex: (slot - 1) % 4,
      state: stale ? 'held' : idle ? 'idle' : 'live',
    };
  });
}

/** A recorded path for one slot, as the canvas expects it. */
export type PlaybackTrail = {
  id: number;
  targets: { id: number; x: number; y: number; distance: number }[];
};

/**
 * Converts one slot's ordered samples into a canvas polyline. The canvas owns rendering; this
 * owns the shape, so trails and recent trails cannot drift apart.
 */
export function playbackTrail(slot: number, points: readonly SpatialPoint[]): PlaybackTrail {
  return {
    id: slot,
    targets: points.map((point) => ({
      id: point.slot,
      x: point.x,
      y: point.y,
      distance: 0,
    })),
  };
}

/**
 * Connectors between consecutive samples that are far apart in time or space.
 *
 * A position that appears somewhere else must explain where it came from, otherwise a large jump
 * looks like the target teleporting. Only the last two samples per slot are considered, so the
 * cost stays O(slots) regardless of recording length.
 *
 * @param points Recorded positions, ordered by `at`.
 * @param at Playhead in epoch milliseconds; samples after it are ignored.
 * @param idleGapMs A pause longer than this draws a connector.
 * @param hopMm A distance jump larger than this draws a connector.
 * @returns Previous/current pairs, in slot order.
 */
export function playbackHopLinks(
  points: readonly SpatialPoint[],
  at: number,
  idleGapMs: number,
  hopMm: number
): { from: SpatialPoint; to: SpatialPoint }[] {
  if (!at) return [];
  const pairs = new Map<number, { current?: SpatialPoint; previous?: SpatialPoint }>();
  for (const point of points) {
    if (point.at > at) break;
    const entry = pairs.get(point.slot) ?? {};
    entry.previous = entry.current;
    entry.current = point;
    pairs.set(point.slot, entry);
  }
  const links: { from: SpatialPoint; to: SpatialPoint }[] = [];
  for (const { current, previous } of pairs.values()) {
    if (!current || !previous) continue;
    const jumped = Math.hypot(current.x - previous.x, current.y - previous.y) > hopMm;
    if (current.at - previous.at > idleGapMs || jumped) links.push({ from: previous, to: current });
  }
  return links;
}

/**
 * The timestamp of the newest sample at or before the playhead, or 0 when none precedes it.
 *
 * Used as the canvas pulse key: a new value means a new sample became current, which is the only
 * moment a movement cue should fire.
 */
export function latestSampleAt(points: readonly SpatialPoint[], at: number): number {
  let latest = 0;
  for (const point of points) {
    if (point.at > at) break;
    latest = point.at;
  }
  return latest;
}

/**
 * The span of the preset window that actually holds recordings, or the window itself when the
 * filter is empty.
 *
 * The range timeline fits the recorded envelope instead of the whole preset: a four-minute
 * recording inside a 24 h window would otherwise be a sub-pixel sliver. Sessions come back
 * ascending, so the first start and last end bound every recording.
 *
 * @param sessions Filtered recordings, ascending by `from`.
 * @param from Filter start, epoch milliseconds.
 * @param to Filter end, epoch milliseconds.
 */
export function recordedEnvelope(
  sessions: readonly SpatialSession[],
  from: number,
  to: number
): { from: number; to: number } {
  return {
    from: sessions[0]?.from ?? from,
    to: sessions.at(-1)?.to ?? to,
  };
}
