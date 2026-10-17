import { ref } from 'vue';
import { store } from '@/store';
import { useDeviceSettings } from '@/composables/useDeviceSettings';
import { hubConfig } from '../helpers/devices';
import { captureWsFrames } from '../helpers/ws-capture';

let capture: ReturnType<typeof captureWsFrames>;
beforeEach(() => {
  store.dispatch('hub/init', { devices: [], config: hubConfig() });
  capture = captureWsFrames();
});
afterEach(() => capture.restore());

test('one atomic policy edit produces one update and uses current Vuex state', () => {
  const settings = useDeviceSettings(() => 'sensor');
  settings.createOverride();
  expect(settings.hasOverride.value).toBe(true);
  const before = capture.frames.length;
  settings.save({ reportingThresholds: { distance: .2 }, metricOverrides: { distance: true }, commandConfirmationTimeoutMs: 12000 });
  expect(capture.frames).toHaveLength(before + 1);
  expect(settings.settings.value).toEqual(expect.objectContaining({ id: 'sensor', reportingThresholds: { distance: .2 }, metricOverrides: { distance: true }, commandConfirmationTimeoutMs: 12000 }));
  settings.updateField('metricsEnabled', true);
  expect(settings.settings.value.reportingThresholds).toEqual({ distance: .2 });
  settings.deleteOverride();
  expect(settings.hasOverride.value).toBe(false);
});

test('switching the device cannot reuse another device draft', () => {
  const id = ref('first');
  const settings = useDeviceSettings(() => id.value);
  settings.createOverride();
  settings.save({ reportingThresholds: { distance: .2 } });
  id.value = 'second';
  expect(settings.hasOverride.value).toBe(false);
  settings.createOverride();
  settings.save({ reportingThresholds: { temperature: .5 } });
  expect(store.getters['hub/findDeviceSetting']('first').reportingThresholds).toEqual({ distance: .2 });
  expect(settings.settings.value.reportingThresholds).toEqual({ temperature: .5 });
});

test('global and device forms share save behavior without redundant unchanged writes', () => {
  const settings = useDeviceSettings();
  const before = capture.frames.length;
  settings.updateField('metricsEnabled', settings.settings.value.metricsEnabled);
  expect(capture.frames).toHaveLength(before);
  settings.save({ commandConfirmationTimeoutMs: 2500 });
  expect(capture.frames.at(-1)?.type).toBe('saveDeviceConfigDefaults');
  expect(settings.settings.value.commandConfirmationTimeoutMs).toBe(2500);
});

test('first field edit preserves device flags and inherits reporting rules without copying them', () => {
  const defaults = useDeviceSettings();
  defaults.save({ metricsEnabled: true, liveIntervalOverrides: { linkquality: { value: 10, unit: 'seconds' } } });
  const device = useDeviceSettings(() => 'new-sensor');
  device.save({ reportingThresholds: { distance: .2 } });
  expect(device.hasOverride.value).toBe(true);
  expect(device.settings.value.metricsEnabled).toBe(true);
  expect(device.settings.value.liveIntervalOverrides).toBeUndefined();
  defaults.save({ liveIntervalOverrides: { linkquality: { value: 30, unit: 'seconds' } } });
  expect(device.settings.value.liveIntervalOverrides).toBeUndefined();
});
