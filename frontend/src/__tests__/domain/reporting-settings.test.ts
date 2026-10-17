import { updateReportingThreshold } from '@/domain/reporting-settings';

test('edits an immutable override map, preserving native units and explicit zero', () => {
  const original = { target_distance: .1, illuminance: 5 };
  expect(updateReportingThreshold(original, 'target_distance', '0')).toEqual({ target_distance: 0, illuminance: 5 });
  expect(updateReportingThreshold(original, 'target_distance', '')).toEqual({ illuminance: 5 });
  expect(updateReportingThreshold(original, 'target_distance', '.2')).toEqual({ target_distance: .2, illuminance: 5 });
  expect(original.target_distance).toBe(.1);
});

test.each(['-1', 'NaN', 'Infinity', 'abc'])('rejects invalid delta %s', (value) => {
  expect(updateReportingThreshold({}, 'distance', value)).toBeNull();
});

test('requires a property and normalizes whitespace', () => {
  expect(updateReportingThreshold({}, '', '1')).toBeNull();
  expect(updateReportingThreshold({}, ' distance ', ' 1 ')).toEqual({ distance: 1 });
});

import { confirmationTimeoutMs, validConfirmationTimeout } from '@/domain/reporting-settings';
import { createDeviceConfig } from '@/contracts/settings';

test('confirmation timeout uses the device override, global value, then compatibility data', () => {
  const defaults = { ...createDeviceConfig(), commandConfirmationTimeoutMs: 12000 };
  const device = { ...createDeviceConfig(), commandConfirmationTimeoutMs: 2500 };
  expect(confirmationTimeoutMs(device, defaults)).toBe(2500);
  expect(confirmationTimeoutMs(undefined, defaults)).toBe(12000);
  expect(confirmationTimeoutMs(undefined, undefined)).toBeGreaterThan(0);
});

test.each([0, -1, 1.2, NaN, Infinity, 2 ** 31, '5000'])('rejects invalid confirmation timeout %p', (value) => {
  expect(validConfirmationTimeout(value)).toBe(false);
});

test('new property names cannot change an override map prototype', () => {
  const result = updateReportingThreshold({}, '__proto__', '1');
  expect(Object.getPrototypeOf(result)).toBe(Object.prototype);
  expect(Object.prototype.hasOwnProperty.call(result, '__proto__')).toBe(true);
});

import { effectiveLiveInterval, intervalSeconds, updateReportingInterval, formatLiveInterval } from '@/domain/reporting-settings';

test('category defaults apply without per-device setup; field rules and zero take precedence', () => {
  const defaults = { ...createDeviceConfig(), liveIntervalByCategory: { diagnostic: { value: 30, unit: 'seconds' as const } } };
  const device = createDeviceConfig();
  expect(effectiveLiveInterval(device, defaults, 'new_diagnostic', 'diagnostic')).toEqual({ value: 30, source: 'Shared diagnostic default' });
  expect(effectiveLiveInterval(device, defaults, 'distance', 'measurement').value).toBe(0);
  defaults.liveIntervalOverrides = { new_diagnostic: { value: 10, unit: 'seconds' } };
  expect(effectiveLiveInterval(device, defaults, 'new_diagnostic', 'diagnostic').value).toBe(10);
  device.liveIntervalOverrides = { new_diagnostic: { value: 0, unit: 'seconds' } };
  expect(effectiveLiveInterval(device, defaults, 'new_diagnostic', 'diagnostic')).toEqual({ value: 0, source: 'This device' });
  delete device.liveIntervalOverrides.new_diagnostic;
  expect(effectiveLiveInterval(device, defaults, 'new_diagnostic', 'diagnostic').source).toBe('All devices · field rule');
});

test('interval editing is immutable and blank restores inheritance while zero disables sampling', () => {
  const original = { distance: { value: 10, unit: 'seconds' as const } };
  expect(updateReportingInterval(original, 'distance', '0')).toEqual({ distance: { value: 0, unit: 'milliseconds' } });
  expect(updateReportingInterval(original, 'distance', '')).toEqual({});
  expect(intervalSeconds(updateReportingInterval(original, 'distance', '0.25')?.distance)).toBe(.25);
  expect(original.distance.value).toBe(10);
  for (const invalid of ['-1', 'NaN', 'Infinity', '0.00001', '31536001']) expect(updateReportingInterval({}, 'distance', invalid)).toBeNull();
  expect(updateReportingInterval({}, '', '1')).toBeNull();
  expect(formatLiveInterval(0)).toBe('Every change');
});


test('device categories preserve scope, zero, field priority and reset inheritance', () => {
  const defaults = createDeviceConfig();
  defaults.liveIntervalByCategory = { diagnostic: { value: 30, unit: 'seconds' } };
  const device = createDeviceConfig();
  device.liveIntervalByCategory = { diagnostic: { value: 10, unit: 'seconds' }, measurement: { value: 5, unit: 'seconds' } };
  expect(effectiveLiveInterval(device, defaults, 'signal', 'diagnostic').value).toBe(10);
  expect(effectiveLiveInterval(undefined, defaults, 'signal', 'diagnostic').value).toBe(30);
  expect(effectiveLiveInterval(device, defaults, 'distance', 'measurement').value).toBe(5);
  expect(effectiveLiveInterval(undefined, defaults, 'distance', 'measurement').value).toBe(0);
  device.liveIntervalByCategory.diagnostic.value = 0;
  expect(effectiveLiveInterval(device, defaults, 'signal', 'diagnostic').value).toBe(0);
  defaults.liveIntervalOverrides = { signal: { value: 7, unit: 'seconds' } };
  expect(effectiveLiveInterval(device, defaults, 'signal', 'diagnostic').value).toBe(7);
  device.liveIntervalOverrides = { signal: { value: 2, unit: 'seconds' } };
  expect(effectiveLiveInterval(device, defaults, 'signal', 'diagnostic').value).toBe(2);
  delete device.liveIntervalOverrides.signal;
  delete defaults.liveIntervalOverrides.signal;
  delete device.liveIntervalByCategory.diagnostic;
  defaults.liveIntervalByCategory.diagnostic.value = 12;
  expect(effectiveLiveInterval(device, defaults, 'signal', 'diagnostic').value).toBe(12);
});


import { effectiveFieldInterval } from '@/domain/reporting-settings';

test('history category inheritance is independent from live timing and field rules stay precise', () => {
  const defaults = createDeviceConfig();
  defaults.defaultDebounceByCategory = { diagnostic: { value: 300, unit: 'seconds' } };
  const device = createDeviceConfig();
  device.defaultDebounceByCategory = { diagnostic: { value: 20, unit: 'seconds' } };
  expect(intervalSeconds(effectiveFieldInterval(device, defaults, 'signal', 'diagnostic', true).value)).toBe(20);
  expect(effectiveLiveInterval(device, defaults, 'signal', 'diagnostic').value).toBe(30);
  defaults.debounceOverrides = { signal: { value: 7, unit: 'seconds' } };
  expect(intervalSeconds(effectiveFieldInterval(device, defaults, 'signal', 'diagnostic', true).value)).toBe(7);
  device.debounceOverrides = { signal: { value: 0, unit: 'seconds' } };
  expect(intervalSeconds(effectiveFieldInterval(device, defaults, 'signal', 'diagnostic', true).value)).toBe(0);
});
