import type { DeviceConfig } from '@/types/settings.type';
// Generated from the backend canonical JSON; CI rejects drift. See scripts/sync-reporting-defaults.mjs.
import reportingDefaults from './reporting-defaults.json';

// Framework-free validation shared by settings forms and pending command handling.
// Raw values stay in their reported units; there are no sensor-name presets here.

/** Optional observation deltas in the expose's native unit; blank restores inheritance. */
export function updateReportingThreshold(current: Readonly<Record<string, number>>, property: string, input: string): Record<string, number> | null {
  const name = property.trim();
  if (!name) return null;
  const next = { ...current };
  if (!input.trim()) {
    delete next[name];
    return next;
  }
  const value = Number(input);
  if (!Number.isFinite(value) || value < 0) return null;
  return { ...current, [name]: value };
}

/** A browser timer must be positive and fit its signed 32-bit millisecond range. */
export function validConfirmationTimeout(value: unknown): value is number {
  return typeof value === 'number' && Number.isInteger(value) && value > 0 && value <= 2 ** 31 - 1;
}

/** Resolve the same device → hub → compatibility hierarchy used by the backend.
 * Invalid configured values block a command instead of silently choosing a new timeout. */
export function confirmationTimeoutMs(device: DeviceConfig | undefined, defaults: DeviceConfig | undefined): number | null {
  const value = device?.commandConfirmationTimeoutMs ?? defaults?.commandConfirmationTimeoutMs ?? reportingDefaults.commandConfirmationTimeoutMs;
  return validConfirmationTimeout(value) ? value : null;
}

import type { DeviceDebounce } from '@/types/settings.type';
import type { TimeInterval } from '@/types/types.type';

const unitSeconds: Record<string, number> = { milliseconds: .001, seconds: 1, minutes: 60, hours: 3600, days: 86400 };
export function intervalSeconds(interval: TimeInterval | undefined): number {
  return interval ? interval.value * (unitSeconds[interval.unit] ?? 0) : 0;
}

/** Empty input removes an override; zero explicitly requests immediate output.
 * Store integral milliseconds so editing seconds also supports subsecond intervals. */
export function updateReportingInterval(current: Readonly<DeviceDebounce>, name: string, input: string): DeviceDebounce | null {
  if (!name.trim()) return null;
  const next = { ...current };
  if (!input.trim()) { delete next[name]; return next; }
  const seconds = Number(input);
  const milliseconds = Math.round(seconds * 1000);
  if (!Number.isFinite(seconds) || seconds < 0 || seconds > 365 * 86400 || (seconds > 0 && milliseconds < 1)) return null;
  return { ...current, [name]: { value: milliseconds, unit: 'milliseconds' } };
}

export type EffectiveSetting<T> = { value: T; source: string };
/** Presence, not truthiness, controls inheritance: zero and false are real choices. */
export function resolveFieldSetting<T>(device: Record<string, T> | undefined, global: Record<string, T> | undefined, name: string, fallback: T, fallbackSource: string): EffectiveSetting<T> {
  if (device && Object.prototype.hasOwnProperty.call(device, name)) return { value: device[name], source: 'This device' };
  if (global && Object.prototype.hasOwnProperty.call(global, name)) return { value: global[name], source: 'All devices · field rule' };
  return { value: fallback, source: fallbackSource };
}

/** Category scope is resolved separately so summaries and field fallbacks agree.
 * Specific sensor rules remain more precise than a device-wide category choice. */
export function effectiveCategoryInterval(device: DeviceConfig | undefined, defaults: DeviceConfig | undefined, category: string, history = false): EffectiveSetting<TimeInterval> {
  const key = history ? 'defaultDebounceByCategory' : 'liveIntervalByCategory';
  const categories = defaults?.[key] ?? (history ? {} : reportingDefaults.liveIntervalByCategory as DeviceDebounce);
  if (device?.[key]?.[category] !== undefined) return { value: device[key][category], source: 'This device · category' };
  return { value: categories[category] ?? { value: 0, unit: 'seconds' }, source: `Shared ${category || 'other'} default` };
}

export function effectiveFieldInterval(device: DeviceConfig | undefined, defaults: DeviceConfig | undefined, name: string, category: string, history = false): EffectiveSetting<TimeInterval> {
  const key = history ? 'debounceOverrides' : 'liveIntervalOverrides';
  const fallback = effectiveCategoryInterval(device, defaults, category, history);
  return resolveFieldSetting(device?.[key], defaults?.[key], name, fallback.value, fallback.source);
}

export function effectiveLiveInterval(device: DeviceConfig | undefined, defaults: DeviceConfig | undefined, name: string, category: string): EffectiveSetting<number> {
  const result = effectiveFieldInterval(device, defaults, name, category);
  return { value: intervalSeconds(result.value), source: result.source };
}

export function formatLiveInterval(seconds: number): string {
  return seconds === 0 ? 'Every change' : `At most every ${seconds} seconds`;
}
