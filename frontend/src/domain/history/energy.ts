import type { HistoryPoint, HistorySeries } from './types';
export function isEnergyCounter(series: HistorySeries): boolean {
  return series.name === 'energy' && /^(kwh|wh)$/i.test(series.unit);
}

export type EnergyBucket = { timestamp: number; value: number | null; partial: boolean };

/** Only deltas whose endpoints belong to the same bucket are assigned to it. */
export function energyBuckets(points: HistoryPoint[], from: number, to: number, count = 12): EnergyBucket[] {
  const width = (to - from) / count;
  return Array.from({ length: count }, (_, i) => {
    const start = from + i * width;
    const end = start + width;
    const samples = points.filter((p) => p.timestamp >= start && p.timestamp <= end);
    return energyBucket(samples, start, end);
  });
}
function energyBucket(points: HistoryPoint[], start: number, end: number): EnergyBucket {
  let value = 0;
  let valid = 0;
  let broken = false;
  for (let i = 1; i < points.length; i++) {
    const a = points[i - 1].value;
    const b = points[i].value;
    if (typeof a !== 'number' || typeof b !== 'number' || b < a) {
      broken = true;
      continue;
    }
    value += b - a;
    valid++;
  }
  const boundary = points[0]?.timestamp === start && points[points.length - 1]?.timestamp === end;
  return { timestamp: start, value: valid ? value : null, partial: broken || !boundary };
}

export function energySummary(buckets: EnergyBucket[]) {
  const known = buckets.filter((bucket) => bucket.value !== null);
  return {
    observed: known.length ? known.reduce((sum, bucket) => sum + (bucket.value ?? 0), 0) : null,
    complete: buckets.filter((bucket) => !bucket.partial).length,
    count: buckets.length,
  };
}
