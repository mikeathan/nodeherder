/* Human-readable times for "last seen" and activity (no dependencies, locale-neutral English). */
const MIN = 60_000;
const HOUR = 60 * MIN;
const DAY = 24 * HOUR;

/** "just now", "5 min ago", "3 h ago", "2 days ago"; "—" when unknown. */
export function relativeTime(iso: string | number | null | undefined, now: number): string {
  const t = typeof iso === 'number' ? iso : iso ? Date.parse(iso) : NaN;
  if (!Number.isFinite(t)) return '—';
  const diff = Math.max(0, now - t);
  if (diff < 45_000) return 'just now';
  if (diff < HOUR) return `${Math.round(diff / MIN)} min ago`;
  if (diff < DAY) return `${Math.round(diff / HOUR)} h ago`;
  const days = Math.round(diff / DAY);
  return days === 1 ? 'yesterday' : `${days} days ago`;
}

/** 24-hour clock time, e.g. "07:05:09" (activity feed). */
export function clockTime(at: number): string {
  const d = new Date(at);
  return [d.getHours(), d.getMinutes(), d.getSeconds()].map((n) => String(n).padStart(2, '0')).join(':');
}
