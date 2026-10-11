/*
 * Read access to hub state for screens (adapter over the Vuex `hub` and `ws` modules).
 * Components use this instead of calling store getters directly, so the store API is
 * referenced in one place.
 */
import { computed } from 'vue';
import { store } from '@/store';
import { Device } from '@/types/device';
import { DeviceConfig } from '@/types/settings.type';
import { ConnectionStatusType } from '@/types/connection.type';
import { compareByName, deviceStatus, DeviceStatus, isOnline } from '@/domain/devices';

export function findDevice(id: string): Device | undefined {
  return store.getters['hub/findDevice'](id) as Device | undefined;
}

export function isDeviceDisabled(id: string): boolean {
  return (store.getters['hub/findDeviceSetting'](id) as DeviceConfig | undefined)?.disabled === true;
}

export function statusOf(device: Device): DeviceStatus {
  return deviceStatus(device, isDeviceDisabled(device.id));
}

export function useHub() {
  const devices = computed(() => [...(store.getters['hub/listAllDevices']() as Device[])].sort(compareByName));
  const connection = computed(() => store.getters['ws/getConnectionStatus'] as ConnectionStatusType);
  const isConnected = computed(() => connection.value === 'connected');
  const initialized = computed(() => store.getters['hub/isInitialized']() as boolean);

  /** A device accepts commands only when the hub link is up and it is online and enabled (FE-03). */
  const canControl = (device: Device | undefined): boolean =>
    !!device && isConnected.value && isOnline(device) && !isDeviceDisabled(device.id);

  return { devices, connection, isConnected, initialized, findDevice, isDeviceDisabled, statusOf, canControl };
}
