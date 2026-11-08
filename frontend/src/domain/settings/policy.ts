/**
 * Hub settings policy: labels, bounds and payload builders shared by every settings
 * section. Components render these results; they never reshape settings themselves.
 *
 * Builders always return a new object. The Vuex getters return live state, so mutating a
 * getter result would write store state outside a mutation.
 */
import type { AssistantSettingsType, HistorySettingsType, LoggerSettingsType } from '@/types/settings.type';
import { TimeUnits, type TimeInterval, type TimeUnit } from '@/types/types.type';
import { intervalSeconds } from '@/domain/reporting-settings';
import { isValidHttpUrl } from '@/utils/http.utils';

export type RetentionFieldId = keyof HistorySettingsType;

/** One editable retention value: stable id, human label and explanation. */
export interface RetentionFieldSpec {
  id: RetentionFieldId;
  label: string;
  help: string;
}

/** Cleanup cadence and chart retention are the two values the hub stores for history. */
export const RETENTION_FIELDS: readonly RetentionFieldSpec[] = [
  {
    id: 'sleepTimeout',
    label: 'Cleanup interval',
    help: 'How often the hub checks for readings older than the retention period.',
  },
  {
    id: 'expireAt',
    label: 'Keep history for',
    help: 'Older readings are deleted. Charts and exports cannot reach beyond this window.',
  },
];

/** Units offered by the retention controls; seconds are too fine for a storage policy. */
export const RETENTION_UNITS: readonly { value: TimeUnit; label: string }[] = [
  { value: TimeUnits.UnitMinutes, label: 'minutes' },
  { value: TimeUnits.UnitHours, label: 'hours' },
  { value: TimeUnits.UnitDays, label: 'days' },
];

/** Floor keeps the hub's cleanup goroutine from being asked to run continuously. */
export const MIN_RETENTION_SECONDS = 60;
/** One year is the upper bound accepted by the backend interval maths. */
export const MAX_RETENTION_SECONDS = 365 * 86_400;

/** Documented hub defaults, used when the loaded configuration omits a value. */
export const RETENTION_DEFAULT: HistorySettingsType = {
  sleepTimeout: { value: 12, unit: TimeUnits.UnitHours },
  expireAt: { value: 10, unit: TimeUnits.UnitDays },
};

/** Logger levels accepted by the backend log-level filter. */
export type LoggerLevel = 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal';

export const LOGGER_LEVELS: readonly LoggerLevel[] = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'];

export const LOGGER_DEFAULT: LoggerSettingsType = { enableRemoteLogger: false, level: 'info' };

/** Assistant endpoints must be absolute HTTP(S) URLs; blank clears the setting. */
export const ASSISTANT_DEFAULT: AssistantSettingsType = { url: '' };

export const ASSISTANT_URL_ERROR = 'Enter an http:// or https:// address, or leave it empty.';

const UNIT_SECONDS: Readonly<Record<TimeUnit, number>> = {
  [TimeUnits.UnitMilliseconds]: 0.001,
  [TimeUnits.UnitSeconds]: 1,
  [TimeUnits.UnitMinutes]: 60,
  [TimeUnits.UnitHours]: 3_600,
  [TimeUnits.UnitDays]: 86_400,
};

/**
 * Convert an interval into a unit the retention control offers.
 *
 * Values already stored in minutes, hours or days are returned unchanged so the card never
 * rewrites a saved number. Sub-minute values are not selectable, so they round up to whole
 * minutes — rounding towards a longer retention never shortens what the hub keeps.
 */
function normaliseInterval(interval: TimeInterval): TimeInterval {
  if (RETENTION_UNITS.some((unit) => unit.value === interval.unit)) return interval;
  const minutes = Math.max(1, Math.ceil(intervalSeconds(interval) / UNIT_SECONDS[TimeUnits.UnitMinutes]));
  return { value: minutes, unit: TimeUnits.UnitMinutes };
}

/** Interval to render for a stored value; sub-minute precision is not offered. */
function displayRetentionInterval(interval: TimeInterval | undefined, fallback: TimeInterval): TimeInterval {
  // Only an omitted value falls back. An invalid stored value is shown as it is so the
  // section can flag it instead of claiming a different retention policy than the hub has.
  return normaliseInterval(interval ?? fallback);
}

/** True when an interval is a whole number of the given unit inside the accepted range. */
export function isValidRetentionInterval(interval: TimeInterval): boolean {
  if (!Number.isFinite(interval.value) || !Number.isInteger(interval.value) || interval.value <= 0) return false;
  const seconds = interval.value * (UNIT_SECONDS[interval.unit] ?? 0);
  return seconds >= MIN_RETENTION_SECONDS && seconds <= MAX_RETENTION_SECONDS;
}

/**
 * Build a retention interval from raw control input.
 *
 * Returns `null` when the text is not a positive whole number in range so callers can
 * report an inline error without sending a frame.
 */
export function parseRetentionInterval(input: string, unit: TimeUnit): TimeInterval | null {
  const trimmed = input.trim();
  if (!trimmed) return null;
  const value = Number(trimmed);
  if (!Number.isFinite(value) || !Number.isInteger(value)) return null;
  const interval: TimeInterval = { value, unit };
  return isValidRetentionInterval(interval) ? interval : null;
}

/** Replace one retention field, keeping the other and all unrelated settings untouched. */
export function withRetentionField(
  current: HistorySettingsType,
  field: RetentionFieldId,
  interval: TimeInterval
): HistorySettingsType {
  return { ...current, [field]: interval };
}

/** True when two policies describe the same cleanup cadence and retention period. */
export function isSameRetention(a: HistorySettingsType, b: HistorySettingsType): boolean {
  return a.sleepTimeout?.value === b.sleepTimeout?.value
    && a.sleepTimeout?.unit === b.sleepTimeout?.unit
    && a.expireAt?.value === b.expireAt?.value
    && a.expireAt?.unit === b.expireAt?.unit;
}

/**
 * True when a policy may be sent to the hub, i.e. both intervals are valid.
 *
 * This refuses a partial save built from the placeholder configuration, where an untouched
 * field is `0 days`. The hub treats a zero retention as "keep nothing" and a zero cleanup
 * interval as "run continuously", so an invalid sibling must block the write rather than
 * ride along and be silently persisted.
 */
export function canPersistRetention(policy: HistorySettingsType): boolean {
  return isValidRetentionInterval(policy.sleepTimeout) && isValidRetentionInterval(policy.expireAt);
}

/** Values to show for the retention card, filling hub omissions with documented defaults. */
export function retentionValues(
  history: HistorySettingsType | undefined
): Record<RetentionFieldId, TimeInterval> {
  const source = history ?? RETENTION_DEFAULT;
  return {
    sleepTimeout: displayRetentionInterval(source.sleepTimeout, RETENTION_DEFAULT.sleepTimeout),
    expireAt: displayRetentionInterval(source.expireAt, RETENTION_DEFAULT.expireAt),
  };
}

/** True for a value the hub would accept as a log level filter. */
export function isLoggerLevel(value: unknown): value is LoggerLevel {
  return typeof value === 'string' && (LOGGER_LEVELS as readonly string[]).includes(value);
}

/** Complete logger settings from current values plus the edited field. */
export function withLoggerSetting<K extends keyof LoggerSettingsType>(
  current: LoggerSettingsType | undefined,
  key: K,
  value: LoggerSettingsType[K]
): LoggerSettingsType {
  const base = current ?? LOGGER_DEFAULT;
  const level = key === 'level' && isLoggerLevel(value) ? value : base.level;
  return {
    enableRemoteLogger: key === 'enableRemoteLogger' ? Boolean(value) : base.enableRemoteLogger,
    level,
  };
}

/**
 * Complete assistant settings from current values plus a new URL.
 *
 * Returns `null` for an invalid address so the section can show an inline error and
 * send nothing. A blank string is a valid way to clear the setting.
 */
export function withAssistantUrl(
  current: AssistantSettingsType | undefined,
  url: string
): AssistantSettingsType | null {
  const trimmed = url.trim();
  if (trimmed !== '' && !isValidHttpUrl(trimmed)) return null;
  return { ...(current ?? ASSISTANT_DEFAULT), url: trimmed };
}
