import type { BinaryDataPoint } from '@/types/metrics.type';

export interface BinaryRange {
  value: string;
  start: number;
  end: number;
}


export interface ApexRangeBarDataPoint {
  x: string;
  y: [number, number];
  fillColor: string;
}

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


export function toBinaryRangeBarData(
  ranges: BinaryRange[],
  colorOn: string,
  colorOff: string
): ApexRangeBarDataPoint[] {
  return ranges.map((range) => ({
    x: range.value === 'true' ? 'On' : 'Off',
    y: [range.start, range.end] as [number, number],
    fillColor: range.value === 'true' ? colorOn : colorOff,
  }));
}

