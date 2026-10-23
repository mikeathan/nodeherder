export type HistoryRange = { from: number; to: number };
export const MAX_HISTORY_DURATION = 31 * 86400000;
export const MIN_HISTORY_DURATION = 5 * 60000;
export const HISTORY_PRESETS = [
  { value: '1', label: '1h' }, { value: '6', label: '6h' }, { value: '24', label: '24h' },
  { value: '72', label: '3d' }, { value: '168', label: '7d' }, { value: '720', label: '30d' },
  { value: 'today', label: 'Today' }, { value: 'yesterday', label: 'Yesterday' }, { value: 'custom', label: 'Custom' },
] as const;

/** Calendar Today follows the browser's local zone; rolling hours are elapsed time. */
export function historyRange(period: string, now: number, customFrom = '', customTo = ''): HistoryRange {
  if (period === 'custom') return { from: Date.parse(customFrom), to: Date.parse(customTo) };
  if (period === 'today' || period === 'yesterday') {
    const start = new Date(now);
    start.setHours(0, 0, 0, 0);
    if (period === 'yesterday') {
      const end = start.getTime();
      start.setDate(start.getDate() - 1);
      return { from: start.getTime(), to: end };
    }
    return { from: start.getTime(), to: now };
  }
  return { from: now - Number(period) * 3600000, to: now };
}
export function validHistoryRange(range: HistoryRange): boolean {
  if (!Number.isFinite(range.from) || !Number.isFinite(range.to)) return false;
  if (!Number.isFinite(new Date(range.from).getTime()) || !Number.isFinite(new Date(range.to).getTime())) return false;
  const duration = range.to - range.from;
  return duration > 0 && duration <= MAX_HISTORY_DURATION;
}

/** Never expand beyond the fetched window, including Today in its first five minutes. */
export function clampHistoryRange(range: HistoryRange, bounds: HistoryRange): HistoryRange {
  if (!validHistoryRange(bounds) || !Number.isFinite(range.from) || !Number.isFinite(range.to)) return { ...bounds };
  const available = bounds.to - bounds.from;
  const duration = Math.min(available, MAX_HISTORY_DURATION,
    Math.max(Math.min(MIN_HISTORY_DURATION, available), range.to - range.from));
  const center = (range.from + range.to) / 2;
  const from = Math.max(bounds.from, Math.min(bounds.to - duration, center - duration / 2));
  return { from, to: from + duration };
}

export function zoomHistoryRange(range: HistoryRange, bounds: HistoryRange, factor: number, anchor = 0.5): HistoryRange {
  const duration = range.to - range.from;
  const at = range.from + duration * anchor;
  return clampHistoryRange({ from: at - duration * factor * anchor, to: at + duration * factor * (1 - anchor) }, bounds);
}

export function panHistoryRange(range: HistoryRange, bounds: HistoryRange, direction: number): HistoryRange {
  const step = (range.to - range.from) * 0.2 * direction;
  return clampHistoryRange({ from: range.from + step, to: range.to + step }, bounds);
}
