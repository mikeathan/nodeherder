import { dailyHistoryStats } from './stats';
import type { HistorySeries } from './types';
export type HistoryMode = 'history' | 'statistics';
export function historyExportUnavailable(series?: HistorySeries): string {
  if (!series?.data.length || !series.data.some((p) => p.value !== null)) return 'No recorded samples to export.';
  return series.partial ? 'Export is unavailable for partial history. Narrow the period for complete results.' : '';
}
const cell = (value: unknown) => {
  const text = value === null || value === undefined ? '' : String(value);
  // Treat external enum values as text in spreadsheet applications as well as CSV readers.
  const safe = typeof value === 'string' && /^[=+\-@\t\r]/.test(text) ? `'${text}` : text;
  return /[",\r\n]/.test(safe) ? `"${safe.replace(/"/g, '""')}"` : safe;
};
export function historyCSV(series: HistorySeries, mode: HistoryMode): string {
  const error = historyExportUnavailable(series);
  if (error) throw new Error(error);
  const rows: unknown[][] = mode === 'statistics'
    ? [['bucket_start', 'bucket_end', 'min', 'average', 'max', 'last', 'samples', 'unit'],
      ...dailyHistoryStats(series).map((b) => [new Date(b.start).toISOString(), new Date(b.end).toISOString(),
        b.min, b.avg, b.max, b.last, b.count, series.unit])]
    : [['timestamp', 'value', 'unit'], ...series.data.map((p) => [new Date(p.timestamp).toISOString(), p.value, series.unit])];
  return rows.map((row) => row.map(cell).join(',')).join('\r\n') + '\r\n';
}
