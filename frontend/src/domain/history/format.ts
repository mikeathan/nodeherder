/*
 * User-facing history formatting. Durations, percentages and moments appear in several
 * history surfaces (chart tooltips, compact strip, summary, coverage note), so they live
 * here as one tested source of truth instead of per-component copies.
 */

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;

/** "4 min", "1 h 5 min", "2 h" — rounded to whole minutes and never "0 min". */
export function formatHistoryMinutes(ms: number): string {
  const total = Math.max(1, Math.round(ms / MINUTE));
  if (total < 60) return `${total} min`;
  const hours = Math.floor(total / 60);
  const minutes = total % 60;
  return minutes ? `${hours} h ${minutes} min` : `${hours} h`;
}

/** "under 1 s", "35 s", "4 min" — for inspecting an individual interval. */
export function formatHistoryPreciseDuration(ms: number): string {
  if (ms < 1000) return 'under 1 s';
  if (ms < MINUTE) return `${Math.round(ms / 1000)} s`;
  return formatHistoryMinutes(ms);
}

/** "under 1 min", "4 min", "23 h 55 min" — for covered/no-data spans. */
export function formatHistorySpan(ms: number): string {
  return ms < MINUTE ? 'under 1 min' : formatHistoryMinutes(ms);
}

/** "Oct 24, 18:20" in the viewer's locale. */
export function formatHistoryMoment(timestamp: number): string {
  return new Date(timestamp).toLocaleString([], {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/** "18:20" in the viewer's locale. */
export function formatHistoryClock(timestamp: number): string {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

/**
 * Known-coverage percentage. Sub-1% coverage stays distinguishable from zero, and a
 * range with no duration at all is "Unknown" rather than a fake 0%.
 */
export function formatHistoryCoverage(known: number, total: number): string {
  if (total <= 0) return 'Unknown';
  if (known <= 0) return '0%';
  const percent = (known / total) * 100;
  return percent < 1 ? '<1%' : `${Math.round(percent)}%`;
}

/** Rounded share of a bin ("42%"), or null when the bin has no known time. */
export function formatHistorySharePercent(percent: number | null): string | null {
  return percent === null ? null : `${Math.round(percent)}%`;
}

/** Axis/moment label: dates for multi-day ranges, clock labels for short ones. */
export function formatHistoryAxisLabel(timestamp: number, range: number, endpoint: boolean): string {
  const date = new Date(timestamp);
  return range > 2 * 24 * HOUR || (endpoint && range >= 24 * HOUR)
    ? date.toLocaleDateString([], { month: 'short', day: 'numeric' })
    : formatHistoryClock(timestamp);
}
