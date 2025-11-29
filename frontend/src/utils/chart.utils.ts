import type { RangeBarDataPoint, BinaryDataPoint, BinaryRange } from '@/types/metrics.type';
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
