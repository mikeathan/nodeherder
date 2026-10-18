import type { HistoryPoint, HistorySeries } from './types';

/** Count reports, not transitions: identical button presses are separate events. */
export function eventCounts(series: HistorySeries, count = 24): HistorySeries {
  const width = (series.to - series.from) / count;
  const buckets: HistoryPoint[] = Array.from({ length: count }, (_, i) => ({
    timestamp: series.from + i * width,
    value: 0,
  }));
  for (const point of series.data) {
    const index = Math.min(count - 1, Math.floor((point.timestamp - series.from) / width));
    buckets[index].value = Number(buckets[index].value) + 1;
  }
  if (series.partial) {
    const knownUntil = series.data[series.data.length - 1]?.timestamp ?? series.from;
    // A response cap truncates the tail; unscanned buckets are not zero events.
    for (const bucket of buckets) {
      if (bucket.timestamp + width > knownUntil) bucket.value = null;
    }
  }
  return { ...series, name: 'event_count', type: 'numeric', unit: 'events', data: buckets };
}
