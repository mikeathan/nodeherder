import { computed } from 'vue';
import { store } from '@/store';
import type { DeviceConfig } from '@/types/settings.type';
import { createDeviceConfigOverride } from '@/contracts/settings';

/** Shared adapter for hub defaults and per-device overrides. Vuex is the only form
 * state: changing routes cannot retain another device's draft, and related policy
 * fields are merged into a single immutable configuration update. */
export function useDeviceSettings(deviceId: () => string = () => '') {
  const settings = computed<DeviceConfig>(() => (
    deviceId() ? store.getters['hub/findDeviceSetting'](deviceId()) : store.getters['hub/deviceDefaults']()
  ));
  const hasOverride = computed(() => !!deviceId() && store.getters['hub/hasDeviceConfigOverride'](deviceId()));

  function save(patch: Partial<DeviceConfig>) {
    const id = deviceId();
    const action = id ? 'hub/saveDeviceConfigOverride' : 'hub/saveDeviceConfigDefaults';
    const base = id && !hasOverride.value
      ? { ...createDeviceConfigOverride(id), disabled: settings.value.disabled, metricsEnabled: settings.value.metricsEnabled }
      : settings.value;
    return store.dispatch(action, { ...base, ...patch, id });
  }
  function updateField<K extends keyof DeviceConfig>(key: K, value: DeviceConfig[K]) {
    if (settings.value[key] !== value) save({ [key]: value });
  }
  function createOverride() {
    store.dispatch('hub/saveDeviceConfigOverride', createDeviceConfigOverride(deviceId()));
  }
  function deleteOverride() {
    store.dispatch('hub/deleteDeviceConfigOverride', deviceId());
  }
  return { settings, hasOverride, save, updateField, createOverride, deleteOverride };
}
