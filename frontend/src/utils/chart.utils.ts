import type {
  RangeBarDataPoint,
  BinaryDataPoint,
  BinaryBucket,
  BinaryRange,
  NumericDataPoint,
  NumericStats,
} from '@/types/metrics.type';
import { formatDuration, formatTime } from './date.utils';
import { escapeHTML } from './html';
import { getFormattedSensorValueByName } from '@/modules/formatters/sensor-formatter';
import { MetricsTypes, type MetricsType, type MiniChartComponentKey } from '@/types/metrics.type';
import { ColorTypes } from '@/types/color.type';

export const DENSITY_COLORS = {
  idle: ColorTypes.Slate700,
  low: ColorTypes.Green800,
  medium: ColorTypes.Green500,
  high: ColorTypes.Green400,
} as const;

export function normalizeBinaryEvents(events: BinaryDataPoint[], from: number, to: number): BinaryRange[] {
  if (!events?.length) {
    return [{ value: 'unknown', start: from, end: to }];
  }

  const ranges: BinaryRange[] = [];

  // First segment: chart start → first event
  // We use 'unknown' instead of assuming the inverse, so we don't incorrectly
  // paint the entire leading portion of a chart if it just started tracking.
  if (events[0].timestamp > from) {
    ranges.push({
      value: 'unknown',
      start: from,
      end: events[0].timestamp,
    });
  }

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
  if (lastEvent.timestamp < to) {
    ranges.push({
      value: lastEvent.value,
      start: lastEvent.timestamp,
      end: to,
    });
  }

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

export function getBinaryRanges(data: BinaryDataPoint[], from: number, to: number, minDurationMs = 0): BinaryRange[] {
  if (!data?.length) return [];
  const sorted = data
    .slice()
    .filter((point) => Number.isFinite(point.timestamp))
    .sort((a, b) => a.timestamp - b.timestamp);
  if (sorted.length === 0) return [];

  let normalized = normalizeBinaryEvents(sorted, from, to);
  normalized = mergeContiguousStates(normalized);

  return minDurationMs > 0 ? mergeBinaryFlickers(normalized, minDurationMs) : normalized;
}

function mergeContiguousStates(ranges: BinaryRange[]): BinaryRange[] {
  if (!ranges.length) return [];
  const merged: BinaryRange[] = [{ ...ranges[0] }];
  for (let i = 1; i < ranges.length; i++) {
    const current = ranges[i];
    const last = merged[merged.length - 1];

    if (current.value === 'unknown' || last.value === 'unknown') {
      if (current.value === last.value) {
        last.end = current.end;
      } else {
        merged.push({ ...current });
      }
    } else if (isBinaryOn(current.value) === isBinaryOn(last.value)) {
      last.end = current.end;
    } else {
      merged.push({ ...current });
    }
  }
  return merged;
}

export interface TimelineSegment {
  width: string;
  color: string;
  start: number;
  end: number;
  stateLabel: string;
  isOn: boolean;
  isUnknown: boolean;
}

export function computeTimelineSegments(
  ranges: BinaryRange[],
  totalMs: number,
  baseMinWidthPct: number,
  options: {
    colorOn: string;
    colorOff: string;
    exposeName: string;
  }
): TimelineSegment[] {
  if (ranges.length === 0 || totalMs <= 0) return [];

  // 1. Calculate dynamic minimum width
  // To avoid a "barcode" effect where all segments look identical due to flexbox over-shrinking,
  // we cap the minimum width so that all ticks combined consume at most 50% of the timeline.
  // The remaining 50% guarantees proportional differences between long and short spaces.
  // Hard floor of 0.5% ensures thin ticks remain visible even in highly dense charts.
  const dynamicMinWidth = Math.max(0.5, Math.min(baseMinWidthPct, 50 / ranges.length));

  let totalExtra = 0;

  // 2. First pass: compute real % widths and bump small elements to the dynamic minimum
  const segments = ranges.map((r) => {
    const rawPct = ((r.end - r.start) / totalMs) * 100;
    const widthPct = Math.max(rawPct, dynamicMinWidth);

    totalExtra += widthPct - rawPct;

    return {
      widthPct,
      isOn: r.value !== 'unknown' && isBinaryOn(r.value),
      isUnknown: r.value === 'unknown',
      start: r.start,
      end: r.end,
    };
  });

  // 3. Second pass: to keep total at 100%, steal the extra added percentage from genuinely large segments
  if (totalExtra > 0) {
    const largeSegments = segments.filter((s) => s.widthPct > dynamicMinWidth * 2);
    const totalLargeWidth = largeSegments.reduce((sum, s) => sum + s.widthPct, 0);

    if (totalLargeWidth > 0) {
      for (const seg of largeSegments) {
        const stolen = (seg.widthPct / totalLargeWidth) * totalExtra;
        seg.widthPct = Math.max(dynamicMinWidth, seg.widthPct - stolen);
      }
    }
  }

  const UNKNOWN_COLOR = 'rgba(148, 163, 184, 0.15)'; // Slate-400 with opacity 0.15 roughly matching background panel gap.

  // 4. Map to final display objects
  return segments.map((seg) => {
    let color = seg.isOn ? options.colorOn : options.colorOff;
    if (seg.isUnknown) {
      color = UNKNOWN_COLOR;
    }
    return {
      width: `${seg.widthPct}%`,
      color,
      start: seg.start,
      end: seg.end,
      stateLabel: seg.isUnknown ? 'Unknown' : resolveBinaryLabel(options.exposeName, seg.isOn ? 'true' : 'false'),
      isOn: seg.isOn,
      isUnknown: seg.isUnknown,
    };
  });
}

export function toBinaryRangeBarData(ranges: BinaryRange[], colorOn: string, colorOff: string): RangeBarDataPoint[] {
  // Convert normalized ranges into Apex-compatible range-bar points.
  const fadedOff = withAlpha(colorOff, 0.8);
  return ranges.map((range) => ({
    x: isBinaryOn(range.value) ? 'On' : 'Off',
    y: [range.start, range.end] as [number, number],
    fillColor: isBinaryOn(range.value) ? colorOn : fadedOff,
  }));
}

const UNIT_SETS = {
  energy: new Set(['wh', 'kwh']),
  power: new Set(['w', 'kw']),
  realtime: new Set(['a', 'ma', 'v', 'kv', 'mv', 'lux']),
  percent: new Set(['%']),
};

const EXPOSE_KEYWORDS = {
  energy: ['energy', 'consumption', 'energy_total', 'total_energy', 'daily_energy', 'monthly_energy'],
  realtime: ['current', 'voltage', 'illuminance'],
  percent: ['battery', 'battery_percentage', 'battery_percent', 'battery_level', 'battpercentage'],
};

const DEFAULT_GAUGE_MAX_BY_UNIT: Record<string, number> = {
  w: 1000,
  kw: 5,
  wh: 1000,
  kwh: 10,
};

function normalizeUnit(unit?: string): string {
  return (unit ?? '').toLowerCase().trim();
}

function normalizeExposeName(exposeName?: string): string {
  return (exposeName ?? '').toLowerCase();
}

export function isEnergyUnit(unit?: string): boolean {
  return UNIT_SETS.energy.has(normalizeUnit(unit));
}

export function isPowerUnit(unit?: string): boolean {
  return UNIT_SETS.power.has(normalizeUnit(unit));
}

export function isRealtimeUnit(unit?: string): boolean {
  return UNIT_SETS.realtime.has(normalizeUnit(unit));
}

export function isPercentUnit(unit?: string): boolean {
  return UNIT_SETS.percent.has(normalizeUnit(unit));
}

export function isEnergyExpose(exposeName?: string, unit?: string): boolean {
  if (isEnergyUnit(unit)) return true;
  const name = normalizeExposeName(exposeName);
  return EXPOSE_KEYWORDS.energy.some((keyword) => name.includes(keyword));
}

export function isRealtimeExpose(exposeName?: string, unit?: string): boolean {
  if (isRealtimeUnit(unit)) return true;
  const name = normalizeExposeName(exposeName);
  return EXPOSE_KEYWORDS.realtime.some((keyword) => name.includes(keyword));
}

export function isPercentExpose(exposeName?: string, unit?: string): boolean {
  const name = normalizeExposeName(exposeName);
  const matches = EXPOSE_KEYWORDS.percent.some((keyword) => name === keyword);
  if (!matches) return false;
  if (!unit) return true;
  return isPercentUnit(unit);
}

export function getDefaultGaugeMax(unit?: string): number {
  const normalized = normalizeUnit(unit);
  return DEFAULT_GAUGE_MAX_BY_UNIT[normalized] ?? 100;
}

export function isBinaryChartableExpose(type?: MetricsType | string, exposeName?: string): boolean {
  return type === MetricsTypes.Binary || (type === MetricsTypes.Enum && exposeName === 'state');
}

export function resolveMiniChartComponentKey(
  type?: MetricsType | string,
  exposeName?: string,
  unit?: string
): MiniChartComponentKey | null {
  if (isBinaryChartableExpose(type, exposeName)) {
    return 'MiniDynamicBinaryChart';
  }
  if (type !== MetricsTypes.Numeric) return null;
  if (isEnergyExpose(exposeName, unit)) return 'MiniEnergyChart';
  if (isPercentExpose(exposeName, unit)) return 'MiniPercentChart';
  if (isRealtimeExpose(exposeName, unit)) return 'MiniRealtimeChart';
  return 'MiniNumericChart';
}

// Resolve human-friendly label for a binary expose based on its name and value
export function resolveBinaryLabel(exposeName: string, value: string | boolean): string {
  const boolVal = isBinaryOn(value);
  return getFormattedSensorValueByName(exposeName, boolVal);
}

export function isBinaryOn(value: string | boolean): boolean {
  // Normalize common true/false string values.
  if (typeof value === 'boolean') return value;
  const normalized = value.toLowerCase();
  return normalized === 'true' || normalized === 'on' || normalized === '1';
}

export function withAlpha(color: string, alpha: number): string {
  // Convert #RRGGBB to rgba while leaving other formats untouched.
  if (!color?.startsWith('#') || color.length !== 7) return color;
  const r = parseInt(color.slice(1, 3), 16);
  const g = parseInt(color.slice(3, 5), 16);
  const b = parseInt(color.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

export function getBinaryStats(
  data: BinaryDataPoint[],
  endTimestamp = Date.now()
): {
  onCount: number;
  offCount: number;
  onPercentage: number;
  onDurationMs: number;
  offDurationMs: number;
} {
  // Aggregate counts and time-on vs time-off over the provided range.
  if (!data?.length) {
    return { onCount: 0, offCount: 0, onPercentage: 0, onDurationMs: 0, offDurationMs: 0 };
  }

  const points = data
    .slice()
    .sort((a, b) => a.timestamp - b.timestamp)
    .filter((point) => Number.isFinite(point.timestamp));

  let onDuration = 0;
  let offDuration = 0;

  for (let i = 0; i < points.length - 1; i++) {
    const current = points[i];
    const next = points[i + 1];
    const duration = Math.max(0, next.timestamp - current.timestamp);

    if (current.value !== 'unknown') {
      if (isBinaryOn(current.value)) {
        onDuration += duration;
      } else {
        offDuration += duration;
      }
    }
  }

  const last = points[points.length - 1];
  const tailDuration = Math.max(0, endTimestamp - last.timestamp);
  if (last.value !== 'unknown') {
    if (isBinaryOn(last.value)) {
      onDuration += tailDuration;
    } else {
      offDuration += tailDuration;
    }
  }

  const total = onDuration + offDuration;
  const onPercentage = total > 0 ? (onDuration / total) * 100 : 0;

  return {
    onCount: points.filter((d) => d.value !== 'unknown' && isBinaryOn(d.value)).length,
    offCount: points.filter((d) => d.value !== 'unknown' && !isBinaryOn(d.value)).length,
    onPercentage,
    onDurationMs: onDuration,
    offDurationMs: offDuration,
  };
}

export function renderRangeTooltip(name: string, label: string, start: number, end: number, color?: string): string {
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
  const swatch = color
    ? `<span style='display:inline-block;width:8px;height:8px;border-radius:999px;background:${color};margin-right:6px;'></span>`
    : '';
  return `<div style='background:#1f2937;color:#f8fafc;padding:6px 8px;border-radius:6px;font-size:11px;min-width:140px;'>
      <div style='font-weight:600;margin-bottom:4px;'>${swatch}${escapeHTML(name)}: ${escapeHTML(label)}</div>
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

// ── Binary heatmap utilities ──

/** Event count threshold above which heatmap mode is preferred over timeline. */
export const BINARY_NOISE_THRESHOLD = 50;

/**
 * Group binary events into fixed-width time buckets.
 * Each bucket records the number of state transitions and total active-state
 * duration, then normalizes intensity to 0..1 across all buckets.
 */
export function bucketBinaryEvents(
  data: BinaryDataPoint[],
  from: number,
  to: number,
  bucketMinutes = 15
): BinaryBucket[] {
  if (!data?.length || from >= to) return [];

  const bucketMs = bucketMinutes * 60_000;
  const sorted = data
    .slice()
    .filter((p) => Number.isFinite(p.timestamp))
    .sort((a, b) => a.timestamp - b.timestamp);

  if (sorted.length === 0) return [];

  const ranges = normalizeBinaryEvents(sorted, from, to);
  const bucketCount = Math.ceil((to - from) / bucketMs);
  const buckets: BinaryBucket[] = [];

  for (let i = 0; i < bucketCount; i++) {
    const bStart = from + i * bucketMs;
    const bEnd = Math.min(bStart + bucketMs, to);
    let count = 0;
    let activeMs = 0;

    for (const range of ranges) {
      // Skip ranges that don't overlap this bucket
      if (range.end <= bStart || range.start >= bEnd) continue;

      const overlapStart = Math.max(range.start, bStart);
      const overlapEnd = Math.min(range.end, bEnd);

      if (isBinaryOn(range.value)) {
        activeMs += overlapEnd - overlapStart;
      }

      // Count transitions that START within this bucket
      if (range.start >= bStart && range.start < bEnd && range.start > from) {
        count++;
      }
    }

    buckets.push({ start: bStart, end: bEnd, count, activeMs, intensity: 0 });
  }

  // Normalize intensity across all buckets
  const maxCount = Math.max(...buckets.map((b) => b.count), 1);
  for (const bucket of buckets) {
    bucket.intensity = bucket.count / maxCount;
  }

  return buckets;
}

/**
 * Generates percentage-based offsets and formatted time strings for density strip X-axis labels.
 */
export function getDensityTimeLabels(buckets: BinaryBucket[]): Array<{ offset: number; text: string }> {
  if (!buckets?.length) return [];
  const labels: Array<{ offset: number; text: string }> = [];

  const startTime = buckets[0].start;
  const endTime = buckets[buckets.length - 1].end;
  const totalMs = endTime - startTime;
  if (totalMs <= 0) return [];

  const showDayLabels = totalMs > 86400000; // > 24 hours

  // Decide how many labels to generate based on scale
  let numLabels = 5;
  if (showDayLabels) {
    const totalDays = totalMs / 86400000;
    if (totalDays <= 6) {
      // 1-6 days: try to align exactly 1 interval per day
      numLabels = Math.max(3, Math.round(totalDays) + 1);
    }
  }

  let lastText = '';
  for (let i = 0; i < numLabels; i++) {
    const fraction = i / (numLabels - 1); // 0.0 to 1.0
    const timeMs = startTime + fraction * totalMs;
    const t = new Date(timeMs);

    let text = '';
    if (showDayLabels) {
      text = t.toLocaleDateString('en-GB', { day: '2-digit', month: 'short' });
    } else {
      text = t.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });
    }

    if (text !== lastText) {
      labels.push({ offset: fraction * 100, text });
      lastText = text;
    }
  }

  return labels;
}

/**
 * HTML tooltip for a heatmap bucket cell.
 */
export function renderHeatmapTooltip(bucket: BinaryBucket, exposeName: string): string {
  const startStr = formatTime(bucket.start);
  const endStr = formatTime(bucket.end);
  const activeStr = formatDuration(bucket.activeMs);
  const label = resolveBinaryLabel(exposeName, true);
  return `<div style='background:#1f2937;color:#f8fafc;padding:6px 8px;border-radius:6px;font-size:11px;min-width:140px;'>
      <div style='font-weight:600;margin-bottom:4px;'>${escapeHTML(exposeName)}</div>
      <div><span style='color:#94a3b8;'>Time:</span> ${startStr} – ${endStr}</div>
      <div><span style='color:#94a3b8;'>Triggers:</span> ${bucket.count}</div>
      <div><span style='color:#94a3b8;'>${escapeHTML(label)}:</span> ${activeStr}</div>
    </div>`;
}
