import type { HistoryPoint } from './types';
import { RENDER_LIMIT } from './types';

export function numericStats(points: HistoryPoint[]) {
  const values = points.map((p) => p.value).filter((v): v is number => typeof v === 'number' && Number.isFinite(v));
  if (!values.length) return null;
  const sum = values.reduce((a, b) => a + b, 0);
  return {
    min: Math.min(...values),
    max: Math.max(...values),
    avg: sum / values.length,
    last: values[values.length - 1],
  };
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
function reduceRun(points: HistoryPoint[], limit: number): HistoryPoint[] {
  if (points.length <= limit) return points;
  const result = [points[0]];
  const bucketCount = Math.floor((limit - 2) / 2);
  const width = Math.ceil((points.length - 2) / bucketCount);
  for (let start = 1; start < points.length - 1; start += width) {
    result.push(...bucketExtrema(points.slice(start, Math.min(points.length - 1, start + width))));
  }
  result.push(points[points.length - 1]);
  return [...new Set(result)].sort((a, b) => a.timestamp - b.timestamp);
}
function bucketExtrema(points: HistoryPoint[]): HistoryPoint[] {
  const low = points.reduce((a, b) => (Number(a.value) < Number(b.value) ? a : b));
  const high = points.reduce((a, b) => (Number(a.value) > Number(b.value) ? a : b));
  return [low, high];
}
