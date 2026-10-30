import { formatDuration } from '@/utils/date.utils';
import type { ConnectionSchedule } from '@/types/integrations.type';

/** Units offered by the automatic-refresh editor; the API always stores seconds. */
export type ScheduleUnit = 'seconds' | 'minutes' | 'hours' | 'days';

/** One selectable schedule unit and its duration in seconds. */
export interface ScheduleUnitOption {
  value: ScheduleUnit;
  seconds: number;
  label: string;
}

/** Supported editor units, ordered from shortest to longest. */
export const SCHEDULE_UNITS: readonly ScheduleUnitOption[] = [
  { value: 'seconds', seconds: 1, label: 'seconds' },
  { value: 'minutes', seconds: 60, label: 'minutes' },
  { value: 'hours', seconds: 3_600, label: 'hours' },
  { value: 'days', seconds: 86_400, label: 'days' },
];

/** A duration split into an editor value and its unit. */
export interface IntervalValue {
  value: number;
  unit: ScheduleUnit;
}

/** Editor state derived from a confirmed schedule; a null schedule is manual with an hourly default. */
export interface ScheduleEditor extends IntervalValue {
  enabled: boolean;
}

/** Derive editor controls from the last confirmed schedule without mutating it. */
export function scheduleEditor(schedule: ConnectionSchedule | null): ScheduleEditor {
  const interval = splitInterval(schedule?.interval_seconds ?? 3600);
  return { enabled: schedule?.enabled ?? false, value: interval.value, unit: interval.unit };
}

/**
 * True when the editor still represents the last confirmed schedule. A fresh read may replace the
 * editor only in this state; otherwise the household has unsaved edits that must be preserved.
 */
export function matchesScheduleEditor(
  schedule: ConnectionSchedule | null,
  enabled: boolean,
  value: number,
  unit: ScheduleUnit
): boolean {
  const editor = scheduleEditor(schedule);
  return editor.enabled === enabled && editor.value === value && editor.unit === unit;
}

/** True when two confirmed schedules have the same user-facing cadence. */
export function sameScheduleConfig(a: ConnectionSchedule | null, b: ConnectionSchedule | null): boolean {
  return (
    (a?.enabled ?? false) === (b?.enabled ?? false) &&
    (a?.interval_seconds ?? 0) === (b?.interval_seconds ?? 0)
  );
}

/** Convert an editor value/unit pair to the integer seconds required by the API. */
export function toIntervalSeconds(value: number, unit: ScheduleUnit): number {
  const option = SCHEDULE_UNITS.find((entry) => entry.value === unit);
  if (!option || !Number.isFinite(value) || value <= 0) return 0;
  return Math.round(value * option.seconds);
}

/**
 * Split stored seconds into the largest unit that divides exactly, so a 5-day interval stays
 * "5 days" and a 61-second interval stays "61 seconds". Only fractional input is rounded.
 */
export function splitInterval(seconds: number): IntervalValue {
  if (!Number.isFinite(seconds) || seconds <= 0) return { value: 1, unit: 'hours' };
  for (const option of [...SCHEDULE_UNITS].reverse()) {
    if (seconds % option.seconds === 0) return { value: seconds / option.seconds, unit: option.value };
  }
  return { value: Math.max(1, Math.round(seconds)), unit: 'seconds' };
}

/** Format stored seconds as a human interval, for example "every 5 days". */
export function formatScheduleInterval(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) return 'manually';
  for (const option of [...SCHEDULE_UNITS].reverse()) {
    if (seconds >= option.seconds && seconds % option.seconds === 0) {
      const value = seconds / option.seconds;
      if (value === 1) return `every ${option.label.slice(0, -1)}`;
      return `every ${value} ${option.label}`;
    }
  }
  return `every ${formatDuration(seconds * 1_000)}`;
}
