import type { HistoryPoint } from './types';
import { RENDER_LIMIT } from './types';

/**
 * Min, max, mean and last value of a series, or null when it holds no finite number.
 *
 * Computed in a single pass with running values rather than `Math.min(...values)`. The spread form
 * is not merely slower, it is unsafe: passing more than roughly 120,000 arguments exceeds the
 * engine's call stack and throws RangeError. The HTTP contract caps a series at HISTORY_LIMIT
 * today, so that crash was not reachable from the network, but it made correctness a property of
 * the caller rather than of this function — and an exported helper should not have a hidden
 * upper bound on its input.
 *
 * A single pass also avoids materialising an intermediate array of the values, which for a
 * 10,000-point series halves the allocation this function performs.
 */
export function numericStats(points: HistoryPoint[]) {
  let min = Number.POSITIVE_INFINITY;
  let max = Number.NEGATIVE_INFINITY;
  let sum = 0;
  let count = 0;
  let last = 0;

  for (const point of points) {
    const value = point.value;
    // Null is a gap, not a zero: counting it would drag the mean down and report a minimum that
    // never happened. Non-finite numbers are excluded for the same reason.
    if (typeof value !== 'number' || !Number.isFinite(value)) continue;
    if (value < min) min = value;
    if (value > max) max = value;
    sum += value;
    count++;
    last = value;
  }

  if (count === 0) return null;
  return { min, max, avg: sum / count, last };
}

/** Decimate independent continuous runs: a missing interval can never become a line. */
export function decimate(points: HistoryPoint[], limit = RENDER_LIMIT): HistoryPoint[] {
  if (points.length <= limit) return points;
  const runs = continuousRuns(points);
  const gaps = runs.filter((run) => run[0].value === null).length;
  const numericRuns = runs.length - gaps;
  if (!numericRuns) return runs.flat();
  const budget = Math.floor((limit - gaps) / numericRuns);
  // Too many disjoint runs cannot be drawn faithfully in the rendering budget.
  if (budget < 4) return [];
  return runs.flatMap((run) => (run[0].value === null ? run : reduceRun(run, budget)));
}

function continuousRuns(points: HistoryPoint[]): HistoryPoint[][] {
  const runs: HistoryPoint[][] = [];
  for (const point of points) appendToRun(runs, point);
  return runs;
}

function appendToRun(runs: HistoryPoint[][], point: HistoryPoint) {
  const last = runs[runs.length - 1];
  if (!last || (last[0].value === null) !== (point.value === null)) {
    runs.push([point]);
    return;
  }
  if (point.value !== null) last.push(point);
}

/**
 * Reduce one continuous run to at most `limit` points, keeping its shape.
 *
 * The run is split into buckets and each bucket contributes its lowest and highest point, so a
 * spike survives compression even when it is one sample wide. Endpoints are always kept, because
 * a chart whose line does not start or end where the data does is misleading.
 *
 * Buckets are walked with indices into the original slice rather than by calling `slice` per
 * bucket. The previous version allocated a fresh array per bucket — for a 10,000-point run that
 * was hundreds of short-lived arrays on every render, which is exactly the allocation pattern that
 * makes a chart stutter when the clock ticks.
 */
function reduceRun(points: HistoryPoint[], limit: number): HistoryPoint[] {
  if (points.length <= limit) return points;

  const lastIndex = points.length - 1;
  const result: HistoryPoint[] = [points[0]];
  const bucketCount = Math.max(1, Math.floor((limit - 2) / 2));
  const width = Math.ceil((lastIndex - 1) / bucketCount);

  for (let start = 1; start < lastIndex; start += width) {
    const end = Math.min(lastIndex, start + width);
    const { low, high } = bucketExtrema(points, start, end);
    result.push(low, high);
  }
  result.push(points[lastIndex]);

  // Buckets can produce the same point twice (a flat run has low === high), and the boundaries can
  // repeat an endpoint. Sorting by timestamp also puts an out-of-order run back in order.
  result.sort((a, b) => a.timestamp - b.timestamp);
  return dedupeByTimestamp(result);
}

/** Lowest and highest point of `points[start:end]`, compared by numeric value. */
function bucketExtrema(points: HistoryPoint[], start: number, end: number) {
  let low = points[start];
  let high = points[start];
  for (let index = start + 1; index < end; index++) {
    const point = points[index];
    // A null inside a numeric run cannot occur (runs are split on null), so Number() is safe here.
    if (Number(point.value) < Number(low.value)) low = point;
    if (Number(point.value) > Number(high.value)) high = point;
  }
  return { low, high };
}

/**
 * Remove points sharing a timestamp, keeping the first.
 *
 * Deduplicating on the sorted array is cheaper than the previous `[...new Set(result)]`, which
 * relied on identity and did not remove distinct objects for the same timestamp.
 */
function dedupeByTimestamp(points: HistoryPoint[]): HistoryPoint[] {
  const result: HistoryPoint[] = [];
  for (const point of points) {
    if (result.length === 0 || result[result.length - 1].timestamp !== point.timestamp) {
      result.push(point);
    }
  }
  return result;
}
