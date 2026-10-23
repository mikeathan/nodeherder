import { numericStats } from './numeric';
import type { HistoryPoint, HistorySeries } from './types';
import type { Expose } from '@/types/device';
import { displayNumber } from '@/domain/controls';

/** Convert every raw sample before aggregation, including nonlinear mired → kelvin. */
export function statisticsDisplaySeries(series: HistorySeries, expose: Expose): HistorySeries {
  return { ...series, unit: expose.name === 'brightness' ? '%' : expose.name === 'color_temp' ? 'K' : series.unit,
    data: series.data.map((p) => ({ ...p, value: typeof p.value === 'number' ? displayNumber(expose, p.value) : p.value })) };
}

export type DailyHistoryBucket = {
  start: number; end: number; min: number | null; avg: number | null;
  max: number | null; last: number | null; count: number; partial: boolean;
};
export function coveredHistoryWindow(series: HistorySeries) {
  return series.data.length ? { from: series.data[0].timestamp, to: series.data[series.data.length - 1].timestamp } : null;
}

/** Samples belong to [local midnight, next local midnight); the query itself includes its end. */
export function dailyHistoryStats(series: HistorySeries): DailyHistoryBucket[] {
  if (series.type !== 'numeric') return [];
  const day = new Date(series.from);
  day.setHours(0, 0, 0, 0);
  const result: DailyHistoryBucket[] = [];
  const covered = coveredHistoryWindow(series);
  let index = 0;
  // Inclusive query endpoints still own their calendar day, even for a null report.
  const hasEndReport = series.data.at(-1)?.timestamp === series.to;
  while (day.getTime() < series.to || (day.getTime() === series.to && hasEndReport)) {
    const start = day.getTime();
    day.setDate(day.getDate() + 1);
    day.setHours(0, 0, 0, 0);
    const end = day.getTime();
    const points: HistoryPoint[] = [];
    while (index < series.data.length && series.data[index].timestamp < end) {
      const point = series.data[index++];
      if (point.timestamp >= Math.max(start, series.from) && point.timestamp <= series.to) points.push(point);
    }
    const stats = numericStats(points);
    const count = points.filter((p) => typeof p.value === 'number' && Number.isFinite(p.value)).length;
    result.push({ start, end, min: stats?.min ?? null, avg: stats?.avg ?? null, max: stats?.max ?? null,
      last: stats?.last ?? null, count,
      partial: start < series.from || end > series.to || (series.partial && (!covered || end > covered.to)) });
  }
  return result;
}
