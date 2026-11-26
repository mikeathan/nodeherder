import type { RouteLocationRaw } from 'vue-router';

/** Severity used for the inbox visual language; mirrors toast tones. */
export type NotificationSeverity = 'info' | 'success' | 'warn' | 'error';

/** Coarse source used for grouping, filtering and dedupe. */
export type NotificationCategory = 'connection' | 'device' | 'automation' | 'integration' | 'storage' | 'system';

/** One durable, browser-local inbox entry. */
export type AppNotification = {
  id: string;
  /** Epoch milliseconds when the event happened. */
  at: number;
  read: boolean;
  severity: NotificationSeverity;
  category: NotificationCategory;
  title: string;
  body?: string;
  /** Stable key for “this condition is still the same event”. */
  dedupeKey?: string;
  action?: { label: string; to: RouteLocationRaw };
};
