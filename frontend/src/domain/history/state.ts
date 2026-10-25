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
export type StateHistoryBin = {
  from: number;
  to: number;
  /** Exact active duration represented by the bin, in milliseconds. */
  active: number;
  /** Exact known duration represented by the bin, in milliseconds. */
  known: number;
  transitions: number;
  /** Active percent of known duration; null when the entire bin is unknown. */
  percent: number | null;
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
export type StateTransition = { from: number; before: boolean; after: boolean; segment: StateSegment };

/**
 * A transition is a real change between two consecutive known intervals.
 * Unknown time breaks the chain, so a gap never invents a state change.
 * Iterating avoids building an array when a caller only counts transitions.
 */
function eachTransition(segments: StateSegment[], visit: (transition: StateTransition) => void) {
  let previousActive: boolean | null = null;
  for (const segment of segments) {
    if (segment.active === null) {
      previousActive = null;
      continue;
    }
    if (previousActive !== null && previousActive !== segment.active) {
      visit({ from: segment.from, before: previousActive, after: segment.active, segment });
    }
    previousActive = segment.active;
  }
}

export function stateTransitions(segments: StateSegment[]): StateTransition[] {
  const transitions: StateTransition[] = [];
  eachTransition(segments, (transition) => transitions.push(transition));
  return transitions;
}

export function stateStats(segments: StateSegment[]) {
  const known = segments.filter((s) => s.active !== null);
  const duration = known.reduce((n, s) => n + s.to - s.from, 0);
  const active = known.filter((s) => s.active).reduce((n, s) => n + s.to - s.from, 0);
  let transitions = 0;
  eachTransition(segments, () => transitions++);
  return { duration, active, transitions, percent: duration ? (active / duration) * 100 : null };
}

/** Clip source intervals to a display range without mutating them or inventing state. */
export function clipStateSegments(segments: StateSegment[], from: number, to: number): StateSegment[] {
  if (!(to > from)) return [];
  return segments
    .filter((segment) => segment.to > from && segment.from < to)
    .map((segment) => ({
      ...segment,
      from: Math.max(segment.from, from),
      to: Math.min(segment.to, to),
    }));
}

export type RecentActivityOptions = {
  /** How many latest transitions form the cluster. */
  transitions?: number;
  /** Context added on each side of the cluster. */
  padding?: number;
  /** Smallest focused window, when the loaded range allows it. */
  minimum?: number;
};
export type RecentActivityWindow = {
  from: number;
  to: number;
  clusterFrom: number;
  clusterTo: number;
  transitionCount: number;
};

/**
 * Find a readable window around the latest transitions without changing the
 * series or its statistics. Returns null when there are fewer than two
 * transitions or the padded focus window is the full loaded range.
 */
export function recentActivityWindow(
  segments: StateSegment[],
  from: number,
  to: number,
  options: RecentActivityOptions = {}
): RecentActivityWindow | null {
  if (!(to > from)) return null;
  const transitionCount = Math.max(1, Math.floor(options.transitions ?? 10));
  const padding = Math.max(0, options.padding ?? 60000);
  const minimum = Math.max(0, Math.min(options.minimum ?? 8 * 60000, to - from));
  const transitions: StateSegment[] = stateTransitions(segments).map((transition) => transition.segment);
  if (transitions.length < 2) return null;
  const recent = transitions.slice(-transitionCount);
  const first = recent[0].from;
  const last = recent[recent.length - 1].from;
  let start = Math.max(from, first - padding);
  let end = Math.min(to, last + padding);
  if (end <= start) return null;
  if (end - start < minimum) {
    end = Math.min(to, Math.max(end, start + minimum));
    if (end - start < minimum) start = Math.max(from, end - minimum);
  }
  return end - start >= to - from
    ? null
    : { from: start, to: end, clusterFrom: first, clusterTo: last, transitionCount: recent.length };
}

/**
 * Aggregate sorted, non-overlapping state intervals into equal time bins.
 * Work is O(segments + bins) because each interval can intersect only the
 * bins it crosses. Unknown time contributes to neither known nor active time.
 */
export function stateHistoryBins(
  segments: StateSegment[],
  from: number,
  to: number,
  count: number
): StateHistoryBin[] {
  const columns = Math.floor(count);
  if (!Number.isFinite(from) || !Number.isFinite(to) || !Number.isFinite(columns) || to <= from || columns <= 0) return [];
  const binWidth = (to - from) / columns;
  const bins: StateHistoryBin[] = Array.from({ length: columns }, (_, i) => ({
    from: from + i * binWidth,
    to: i === columns - 1 ? to : from + (i + 1) * binWidth,
    active: 0,
    known: 0,
    transitions: 0,
    percent: null,
  }));

  for (let i = 1; i < segments.length; i++) {
    const previous = segments[i - 1];
    const next = segments[i];
    if (previous.active === null || next.active === null || previous.active === next.active) continue;
    if (next.from < from || next.from >= to) continue;
    const index = Math.min(columns - 1, Math.floor((next.from - from) / binWidth));
    bins[index].transitions++;
  }

  let firstIntersecting = 0;
  for (const bin of bins) {
    while (firstIntersecting < segments.length && segments[firstIntersecting].to <= bin.from) firstIntersecting++;
    for (let i = firstIntersecting; i < segments.length && segments[i].from < bin.to; i++) {
      const segment = segments[i];
      const overlap = Math.max(0, Math.min(bin.to, segment.to) - Math.max(bin.from, segment.from));
      if (!overlap || segment.active === null) continue;
      bin.known += overlap;
      if (segment.active) bin.active += overlap;
    }
    bin.percent = bin.known ? (bin.active / bin.known) * 100 : null;
  }
  return bins;
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
