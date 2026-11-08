import {
  LOGGER_DEFAULT,
  RETENTION_DEFAULT,
  RETENTION_FIELDS,
  canPersistRetention,
  isLoggerLevel,
  isSameRetention,
  isValidRetentionInterval,
  parseRetentionInterval,
  retentionValues,
  withAssistantUrl,
  withLoggerSetting,
  withRetentionField,
} from '@/domain/settings/policy';
import { TimeUnits } from '@/types/types.type';
import type { HistorySettingsType, LoggerSettingsType } from '@/types/settings.type';

const history: HistorySettingsType = {
  sleepTimeout: { value: 12, unit: TimeUnits.UnitHours },
  expireAt: { value: 10, unit: TimeUnits.UnitDays },
};

test('retention fields label both stored values', () => {
  expect(RETENTION_FIELDS.map((field) => field.id)).toEqual(['sleepTimeout', 'expireAt']);
  expect(RETENTION_FIELDS.every((field) => field.label.length > 0 && field.help.length > 0)).toBe(true);
});

test('retention values fall back to documented defaults per field', () => {
  expect(retentionValues(undefined)).toEqual(RETENTION_DEFAULT);
  expect(retentionValues({ sleepTimeout: { value: 90, unit: TimeUnits.UnitMinutes }, expireAt: { value: 0, unit: TimeUnits.UnitDays } }))
    .toEqual({ sleepTimeout: { value: 90, unit: TimeUnits.UnitMinutes }, expireAt: { value: 0, unit: TimeUnits.UnitDays } });
  expect(retentionValues({} as HistorySettingsType)).toEqual(RETENTION_DEFAULT);
});

test('retention display coarsens sub-minute values to whole minutes', () => {
  expect(retentionValues({ sleepTimeout: { value: 30, unit: TimeUnits.UnitSeconds }, expireAt: { value: 3, unit: TimeUnits.UnitDays } }).sleepTimeout)
    .toEqual({ value: 1, unit: TimeUnits.UnitMinutes });
});

test('retention validation rejects empty, fractional, negative and out-of-range values', () => {
  expect(parseRetentionInterval('', TimeUnits.UnitDays)).toBeNull();
  expect(parseRetentionInterval('0', TimeUnits.UnitDays)).toBeNull();
  expect(parseRetentionInterval('-3', TimeUnits.UnitDays)).toBeNull();
  expect(parseRetentionInterval('1.5', TimeUnits.UnitHours)).toBeNull();
  expect(parseRetentionInterval('abc', TimeUnits.UnitDays)).toBeNull();
  expect(parseRetentionInterval('1', TimeUnits.UnitSeconds)).toBeNull();
  expect(parseRetentionInterval('400', TimeUnits.UnitDays)).toBeNull();
  expect(parseRetentionInterval(' 5 ', TimeUnits.UnitDays)).toEqual({ value: 5, unit: TimeUnits.UnitDays });
  expect(isValidRetentionInterval({ value: 1, unit: TimeUnits.UnitMinutes })).toBe(true);
});

test('withRetentionField keeps the other field and never mutates its input', () => {
  const original = structuredClone(history);
  const next = withRetentionField(history, 'expireAt', { value: 30, unit: TimeUnits.UnitDays });
  expect(next).toEqual({ sleepTimeout: history.sleepTimeout, expireAt: { value: 30, unit: TimeUnits.UnitDays } });
  expect(history).toEqual(original);
  expect(next).not.toBe(history);
  expect(isSameRetention(history, next)).toBe(false);
  expect(isSameRetention(history, { ...history })).toBe(true);
});

test('a placeholder policy cannot be persisted, so an untouched zero never reaches the hub', () => {
  // `createAppconfig()` starts both intervals at 0 days before the snapshot loads.
  const placeholder: HistorySettingsType = {
    sleepTimeout: { value: 0, unit: TimeUnits.UnitDays },
    expireAt: { value: 0, unit: TimeUnits.UnitDays },
  };
  expect(canPersistRetention(placeholder)).toBe(false);
  expect(canPersistRetention({ ...placeholder, expireAt: { value: 30, unit: TimeUnits.UnitDays } })).toBe(false);
  expect(canPersistRetention(history)).toBe(true);
});

test('logger payload builders stay typed and complete', () => {
  const current: LoggerSettingsType = { enableRemoteLogger: true, level: 'debug' };
  expect(withLoggerSetting(current, 'enableRemoteLogger', false)).toEqual({ enableRemoteLogger: false, level: 'debug' });
  expect(withLoggerSetting(current, 'level', 'warn')).toEqual({ enableRemoteLogger: true, level: 'warn' });
  expect(withLoggerSetting(undefined, 'level', 'error')).toEqual({ ...LOGGER_DEFAULT, level: 'error' });
  expect(isLoggerLevel('fatal')).toBe(true);
  expect(isLoggerLevel('verbose')).toBe(false);
  expect(isLoggerLevel(3)).toBe(false);
});

test('assistant URL validation blocks bad addresses and accepts clearing', () => {
  expect(withAssistantUrl(undefined, 'ftp://nope')).toBeNull();
  expect(withAssistantUrl(undefined, 'not a url')).toBeNull();
  expect(withAssistantUrl(undefined, '  https://home.local/chat  ')).toEqual({ url: 'https://home.local/chat' });
  expect(withAssistantUrl({ url: 'https://a.local' }, '')).toEqual({ url: '' });
});
