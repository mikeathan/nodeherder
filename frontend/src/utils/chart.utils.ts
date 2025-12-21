import type {
  RangeBarDataPoint,
  BinaryDataPoint,
  BinaryRange,
  NumericDataPoint,
  NumericStats,
} from '@/types/metrics.type';
import { formatDuration } from './date.utils';
import { getFormattedSensorValueByName } from '@/modules/formatters/sensor-formatter';

/**
 * Normalizes binary events into continuous time ranges.
 * Creates segments from chart start to first event, between events, and last event to chart end.
 */
export function normalizeBinaryEvents(events: BinaryDataPoint[], from: number, to: number): BinaryRange[] {
  if (!events?.length) {
    return [{ value: 'false', start: from, end: to }];
  }

  const ranges: BinaryRange[] = [];

  // First segment: chart start → first event
  ranges.push({
    value: events[0].value,
    start: from,
    end: events[0].timestamp,
  });

  // Middle segments: event transitions
  for (let i = 1; i < events.length; i++) {
    const prevEvent = events[i - 1];
    const currEvent = events[i];

    ranges.push({
      value: prevEvent.value,
      start: prevEvent.timestamp,
      end: currEvent.timestamp,
    });
  }

  // Last segment: last event → chart end
  const lastEvent = events[events.length - 1];
  ranges.push({
    value: lastEvent.value,
    start: lastEvent.timestamp,
    end: to,
  });

  return ranges;
}

/**
 * Merges short-lived state changes (flickers) to reduce visual noise.
 * Combines adjacent segments if they're the same state or if a segment is too short.
 */
export function mergeBinaryFlickers(ranges: BinaryRange[], minDurationMs: number): BinaryRange[] {
  if (ranges.length <= 1) return ranges;

  const merged: BinaryRange[] = [];
  let currentRange = ranges[0];

  for (let i = 1; i < ranges.length; i++) {
    const nextRange = ranges[i];
    const duration = currentRange.end - currentRange.start;

    const isSameState = currentRange.value === nextRange.value;
    const isTooShort = duration < minDurationMs;

    if (isSameState || isTooShort) {
      // Merge: extend current range to include next range
      currentRange = {
        value: nextRange.value,
        start: currentRange.start,
        end: nextRange.end,
      };
    } else {
      // Keep current range and move to next
      merged.push(currentRange);
      currentRange = nextRange;
    }
  }

  // Don't forget the last range
  merged.push(currentRange);
  return merged;
}

export function toBinaryRangeBarData(ranges: BinaryRange[], colorOn: string, colorOff: string): RangeBarDataPoint[] {
  // Convert normalized ranges into Apex-compatible range-bar points.
  return ranges.map((range) => ({
    x: range.value === 'true' ? 'On' : 'Off',
    y: [range.start, range.end] as [number, number],
    fillColor: range.value === 'true' ? colorOn : colorOff,
  }));
}

// Resolve human-friendly label for a binary expose based on its name and value
export function resolveBinaryLabel(exposeName: string, value: string | boolean): string {
  const boolVal = typeof value === 'string' ? value === 'true' : !!value;
  return getFormattedSensorValueByName(exposeName, boolVal);
}

export function renderRangeTooltip(name: string, label: string, start: number, end: number): string {
  // HTML tooltip for binary range bars with readable time/duration.
  const durationMs = Math.max(0, end - start);
  const fmtOpts: Intl.DateTimeFormatOptions = {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  };
  const startStr = new Date(start).toLocaleString(undefined, fmtOpts);
  const endStr = new Date(end).toLocaleString(undefined, fmtOpts);
  const durStr = formatDuration(durationMs);
  return `<div style='background:#1f2937;color:#f8fafc;padding:8px 10px;border-radius:6px;font-size:12px;min-width:180px;'>
      <div style='font-weight:600;margin-bottom:4px;'>${name}: ${label}</div>
      <div><span style='color:#94a3b8;'>From:</span> ${startStr}</div>
      <div><span style='color:#94a3b8;'>To:</span> ${endStr}</div>
      <div><span style='color:#94a3b8;'>Duration:</span> ${durStr}</div>
    </div>`;
}

export function normalizeNumericData(data: NumericDataPoint[]): NumericDataPoint[] {
  // Clean and sort numeric points by time.
  if (!data?.length) return [];
  return data
    .filter((point) => Number.isFinite(point.x) && Number.isFinite(point.y))
    .slice()
    .sort((a, b) => a.x - b.x);
}

export function downsampleNumericData(data: NumericDataPoint[], maxPoints = 400): NumericDataPoint[] {
  // Reduce point count while preserving the last value.
  if (!data?.length || data.length <= maxPoints) return data;

  const step = Math.ceil(data.length / maxPoints);
  const sampled: NumericDataPoint[] = [];

  for (let i = 0; i < data.length; i += step) {
    sampled.push(data[i]);
  }

  const last = data[data.length - 1];
  if (sampled[sampled.length - 1] !== last) {
    sampled.push(last);
  }

  return sampled;
}

export function getNumericStats(data: NumericDataPoint[]): NumericStats | null {
  // Summary stats for quick KPI display.
  if (!data?.length) return null;

  const values = data.map((point) => point.y).filter((value) => Number.isFinite(value));
  if (values.length === 0) return null;

  const first = values[0];
  const last = values[values.length - 1];
  const min = Math.min(...values);
  const max = Math.max(...values);
  const avg = values.reduce((sum, value) => sum + value, 0) / values.length;
  const delta = last - first;
  const deltaPct = first === 0 ? 0 : (delta / Math.abs(first)) * 100;

  return {
    min,
    max,
    avg,
    first,
    last,
    delta,
    deltaPct,
  };
}

export function getNumericDomain(data: NumericDataPoint[], paddingRatio = 0.08): { min: number; max: number } {
  // Calculate y-axis bounds with a small padding.
  if (!data?.length) {
    return { min: 0, max: 1 };
  }

  const values = data.map((point) => point.y).filter((value) => Number.isFinite(value));
  if (values.length === 0) {
    return { min: 0, max: 1 };
  }

  const min = Math.min(...values);
  const max = Math.max(...values);
  if (min === max) {
    const pad = Math.abs(min) * paddingRatio || 1;
    return { min: min - pad, max: max + pad };
  }

  const pad = (max - min) * paddingRatio;
  return { min: min - pad, max: max + pad };
}

export function formatNumericXAxisLabel(
  value: string | number,
  showDayLabels: boolean,
  opts: { date?: number }
): string {
  // Keep labels readable for short vs multi-day ranges.
  const ts = typeof value === 'number' ? value : Number.isFinite(opts?.date) ? (opts?.date as number) : Number(value);
  if (!Number.isFinite(ts)) return String(value);
  const date = new Date(ts);
  if (showDayLabels) {
    return date.toLocaleDateString('en-GB', { day: '2-digit', month: 'short' });
  }
  return date.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });
}
