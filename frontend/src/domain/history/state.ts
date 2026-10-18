import type { Expose } from '@/types/device';
import { binaryStateOf, formatExposeValue } from '@/domain/exposes';
import type { HistoryPoint, HistorySeries } from './types';

export type StateSegment = {
  from: number;
  to: number;
  value: HistoryPoint['value'];
  label: string;
  active: boolean | null;
};
export function binaryState(expose: Expose, value: unknown): boolean | null {
  return binaryStateOf(expose, value);
}
export function stateSegments(series: HistorySeries, expose: Expose): StateSegment[] {
  const segments: StateSegment[] = [];
  let from = series.from;
  let value: HistoryPoint['value'] = null;
  for (const point of series.data) {
    if (point.timestamp > from) appendSegment(segments, segment(from, point.timestamp, value, expose));
    from = point.timestamp;
    value = point.value;
  }
  if (from < series.to) appendSegment(segments, segment(from, series.to, series.partial ? null : value, expose));
  return segments;
}
// Adjacent equal state reports share geometry, while the source event list stays intact.
function appendSegment(segments: StateSegment[], next: StateSegment) {
  const previous = segments[segments.length - 1];
  if (previous && previous.value === next.value) previous.to = next.to;
  else segments.push(next);
}

function segment(from: number, to: number, value: HistoryPoint['value'], expose: Expose): StateSegment {
  const active = binaryState(expose, value);
  return { from, to, value, active, label: active === null ? 'Unknown' : formatExposeValue(expose, value) };
}
export function stateStats(segments: StateSegment[]) {
  const known = segments.filter((s) => s.active !== null);
  const duration = known.reduce((n, s) => n + s.to - s.from, 0);
  const active = known.filter((s) => s.active).reduce((n, s) => n + s.to - s.from, 0);
  let transitions = 0;
  let previous: boolean | null = null;
  for (const s of segments) {
    if (previous !== null && s.active !== null && previous !== s.active) transitions++;
    previous = s.active;
  }
  return { duration, active, transitions, percent: duration ? (active / duration) * 100 : null };
}

/** Dense timelines use a labelled active-duration view, never inflated event widths. */
export function stateDensity(segments: StateSegment[], from: number, to: number, count = 64) {
  const width = (to - from) / count;
  const activeSegments = segments.filter((s) => s.active === true);
  const knownSegments = segments.filter((s) => s.active !== null);
  return Array.from({ length: count }, (_, i) => {
    const start = from + i * width;
    const end = start + width;
    const active = overlapDuration(activeSegments, start, end);
    const known = overlapDuration(knownSegments, start, end);
    return { from: start, to: end, percent: known ? (active / known) * 100 : null };
  });
}
function overlapDuration(segments: StateSegment[], from: number, to: number) {
  return segments.reduce((sum, s) => sum + Math.max(0, Math.min(to, s.to) - Math.max(from, s.from)), 0);
}
