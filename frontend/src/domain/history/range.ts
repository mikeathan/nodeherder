export type HistoryRange = { from: number; to: number };

/** Calendar Today follows the browser's local zone; rolling hours are elapsed time. */
export function historyRange(period: string, now: number, customFrom = '', customTo = ''): HistoryRange {
  if (period === 'custom') return { from: Date.parse(customFrom), to: Date.parse(customTo) };
  if (period === 'today') {
    const start = new Date(now);
    start.setHours(0, 0, 0, 0);
    return { from: start.getTime(), to: now };
  }
  return { from: now - Number(period) * 3600000, to: now };
}
export function validHistoryRange(range: HistoryRange): boolean {
  if (!Number.isFinite(range.from) || !Number.isFinite(range.to)) return false;
  const duration = range.to - range.from;
  return duration > 0 && duration <= 31 * 86400000;
}
