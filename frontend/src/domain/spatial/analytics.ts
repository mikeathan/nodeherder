/*
 * Session-only spatial analytics: bounded frame sampling, heat cells, recent
 * trails and hourly activity. Nothing here is persisted or sent to the device.
 */
import type { SpatialTarget } from './index';

/** One sampled instant. Positions are copied so later updates cannot mutate history. */
export type SpatialFrame = { at: number; targets: SpatialTarget[] };
/** Aggregated target-seconds for one heat cell. */
export type HeatCell = { x: number; y: number; count: number; intensity: number };
/** Milliseconds per second for sampling and display conversions. */
export const MILLISECONDS_PER_SECOND = 1000;
/** Session-only retention window in milliseconds. */
export const SPATIAL_SESSION_MS = 60 * 60 * MILLISECONDS_PER_SECOND;
/** Hard cap on retained one-second frames. */
export const SPATIAL_FRAME_LIMIT = 3600;
/** Sampling cadence for session observations. */
export const SPATIAL_SAMPLE_INTERVAL_MS = 1000;
/** Milliseconds per minute for observation-window arithmetic. */
export const MINUTE_MS = 60_000;
/** Selectable observation windows in minutes. */
export const OBSERVATION_WINDOW_MINUTES = [5, 15, 60] as const;
/** Default observation window in minutes. */
export const DEFAULT_OBSERVATION_MINUTES = 15;
/** Maximum targets sampled per frame. */
export const MAX_SPATIAL_TARGETS = 3;
/** Heat grid cell size in millimetres. */
export const HEAT_CELL_MM = 250;
/** Recent-trail window in milliseconds. */
export const TRAIL_WINDOW_MS = 10_000;
/** A sample gap that splits a trail into separate segments. */
export const TRAIL_GAP_MS = 1_500;

/** Adds one sampled frame with copied positions; bounds history to one hour / 3600 frames. */
export function appendFrame(
  frames: readonly SpatialFrame[],
  targets: readonly SpatialTarget[],
  now: number
): SpatialFrame[] {
  return [
    ...frames.filter((frame) => frame.at > now - SPATIAL_SESSION_MS),
    { at: now, targets: targets.slice(0, MAX_SPATIAL_TARGETS).map((target) => ({ ...target })) },
  ].slice(-SPATIAL_FRAME_LIMIT);
}

/** Aggregates target-seconds per 250 mm cell over an inclusive requested window; no empty cells are invented. */
export function heatCells(frames: readonly SpatialFrame[], since: number): HeatCell[] {
  const cells = new Map<string, { x: number; y: number; count: number }>();
  for (const frame of frames) {
    if (frame.at < since) continue;
    for (const target of frame.targets) {
      const x = Math.floor(target.x / HEAT_CELL_MM) * HEAT_CELL_MM;
      const y = Math.floor(target.y / HEAT_CELL_MM) * HEAT_CELL_MM;
      const key = `${x}:${y}`;
      const cell = cells.get(key) ?? { x, y, count: 0 };
      cell.count += 1;
      cells.set(key, cell);
    }
  }
  const maximum = Math.max(1, ...Array.from(cells.values(), (cell) => cell.count));
  return Array.from(cells.values(), (cell) => ({ ...cell, intensity: cell.count / maximum }));
}

/** Up to ten seconds of sampled trails, separated when a target disappears or samples have a gap. */
export function targetTrails(
  frames: readonly SpatialFrame[],
  now: number
): { id: number; targets: SpatialTarget[] }[] {
  const trails: { id: number; targets: SpatialTarget[] }[] = [];
  const recent = frames.filter((frame) => frame.at >= now - TRAIL_WINDOW_MS);
  for (let id = 1; id <= MAX_SPATIAL_TARGETS; id++) {
    let points: SpatialTarget[] = [];
    let previous = 0;
    for (const frame of recent) {
      const target = frame.targets.find((candidate) => candidate.id === id);
      if (!target || frame.at - previous > TRAIL_GAP_MS) {
        if (points.length > 1) trails.push({ id, targets: points });
        points = [];
      }
      if (target) points.push(target);
      previous = frame.at;
    }
    if (points.length > 1) trails.push({ id, targets: points });
  }
  return trails;
}

/** Hour buckets preserve local timezone labels and count only frames with observed targets. */
export function activityByHour(frames: readonly SpatialFrame[], since: number): { label: string; count: number }[] {
  const buckets = new Map<number, number>();
  for (const frame of frames) {
    if (frame.at < since || !frame.targets.length) continue;
    const hour = new Date(frame.at);
    hour.setMinutes(0, 0, 0);
    const key = hour.getTime();
    buckets.set(key, (buckets.get(key) ?? 0) + frame.targets.length);
  }
  return Array.from(buckets, ([at, count]) => ({
    label: new Date(at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', timeZoneName: 'short' }),
    count,
  }));
}
