import { formatScheduleInterval, matchesScheduleEditor, scheduleEditor, splitInterval, toIntervalSeconds } from '@/domain/schedule';
import type { ConnectionSchedule } from '@/types/integrations.type';

function schedule(overrides: Partial<ConnectionSchedule> = {}): ConnectionSchedule {
  return {
    enabled: true,
    interval_seconds: 432000,
    next_due_at: null,
    last_attempt_at: null,
    last_success_at: null,
    last_failure_at: null,
    consecutive_failures: 0,
    ...overrides,
  };
}

test('stored seconds split into the largest exact unit', () => {
  expect(splitInterval(432_000)).toEqual({ value: 5, unit: 'days' });
  expect(splitInterval(3_600)).toEqual({ value: 1, unit: 'hours' });
  expect(splitInterval(1_800)).toEqual({ value: 30, unit: 'minutes' });
  expect(splitInterval(5_400)).toEqual({ value: 90, unit: 'minutes' });
  expect(splitInterval(61)).toEqual({ value: 61, unit: 'seconds' });
  expect(splitInterval(90)).toEqual({ value: 90, unit: 'seconds' });
});

test('non-positive values stay usable without inventing zero', () => {
  expect(splitInterval(0)).toEqual({ value: 1, unit: 'hours' });
  expect(splitInterval(Number.NaN)).toEqual({ value: 1, unit: 'hours' });
  expect(splitInterval(3_601)).toEqual({ value: 3_601, unit: 'seconds' });
});

test('editor values convert back to API seconds and format for humans', () => {
  expect(toIntervalSeconds(5, 'days')).toBe(432_000);
  expect(toIntervalSeconds(1.5, 'hours')).toBe(5_400);
  expect(toIntervalSeconds(61, 'seconds')).toBe(61);
  expect(toIntervalSeconds(0, 'hours')).toBe(0);
  expect(formatScheduleInterval(432_000)).toBe('every 5 days');
  expect(formatScheduleInterval(3_600)).toBe('every hour');
  expect(formatScheduleInterval(1_800)).toBe('every 30 minutes');
  expect(formatScheduleInterval(61)).toBe('every 61 seconds');
  expect(formatScheduleInterval(0)).toBe('manually');
});

test('confirmed schedule derives editor state and detects a pristine editor', () => {
  expect(scheduleEditor(null)).toEqual({ enabled: false, value: 1, unit: 'hours' });
  expect(scheduleEditor(schedule())).toEqual({ enabled: true, value: 5, unit: 'days' });
  expect(scheduleEditor(schedule({ interval_seconds: 61 }))).toEqual({
    enabled: true,
    value: 61,
    unit: 'seconds',
  });
  expect(matchesScheduleEditor(schedule(), true, 5, 'days')).toBe(true);
  expect(matchesScheduleEditor(schedule(), false, 5, 'days')).toBe(false);
  expect(matchesScheduleEditor(schedule(), true, 5, 'hours')).toBe(false);
  expect(matchesScheduleEditor(schedule({ interval_seconds: 61 }), true, 61, 'seconds')).toBe(true);
  expect(matchesScheduleEditor(null, false, 1, 'hours')).toBe(true);
});
