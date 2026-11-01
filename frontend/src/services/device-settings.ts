import { store } from '@/store';
import type { DeviceConfigSaveResult } from '@/contracts/device-settings';
import type { DeviceConfig } from '@/types/settings.type';
import { ConnectionStatus } from '@/types/connection.type';
import { createUUID } from '@/utils/unique';

/** Maximum time to wait for a correlated persistence acknowledgement. */
export const DEVICE_SETTINGS_SAVE_TIMEOUT_MS = 10_000;
/** Recovery copy for every incomplete save path; drafts and Undo are never discarded. */
export const DEVICE_SETTINGS_MESSAGES = {
  canceled: 'Settings save canceled.',
  reconnect: 'Reconnect to the hub before saving. Your draft and Undo are unchanged.',
  submitFailed: 'Unable to submit settings. Your draft and Undo are unchanged.',
  timeout:
    'Save was not confirmed. It may have reached the hub. Your draft and Undo are unchanged; check the connection before retrying.',
  disconnected: 'Connection lost before save confirmation. Your draft and Undo are unchanged.',
  failed: 'Settings could not be saved. Your draft and Undo are unchanged.',
} as const;

/**
 * Persists one device override over the existing socket and publishes it only after a matching acknowledgement.
 * Owns a ten-second timer/subscription; abort, disconnect, hub reset and completion release both.
 * A timeout has an unknown persistence outcome: no automatic retry or optimistic Vuex update occurs.
 */
export function saveConfirmedDeviceSettings(config: DeviceConfig, signal: AbortSignal): Promise<void> {
  if (signal.aborted) return Promise.reject(new Error(DEVICE_SETTINGS_MESSAGES.canceled));
  if (
    store.getters['ws/getConnectionStatus'] !== ConnectionStatus.connected ||
    (store.getters['ws/getSocket'] as WebSocket | null)?.readyState !== WebSocket.OPEN
  ) {
    return Promise.reject(new Error(DEVICE_SETTINGS_MESSAGES.reconnect));
  }
  // Vuex returns reactive proxies; JSON copying snapshots the plain persisted shape.
  const submitted = JSON.parse(JSON.stringify(config)) as DeviceConfig;
  const requestId = createUUID();
  return new Promise((resolve, reject) => {
    let settled = false;
    const timer = setTimeout(
      () => finish(new Error(DEVICE_SETTINGS_MESSAGES.timeout)),
      DEVICE_SETTINGS_SAVE_TIMEOUT_MS
    );
    const unsubscribe = store.subscribe((mutation) => {
      if (
        mutation.type === 'auth/logout' ||
        mutation.type === 'hub/clear' ||
        (mutation.type === 'ws/setConnectionStatus' && mutation.payload !== ConnectionStatus.connected)
      ) {
        finish(new Error(DEVICE_SETTINGS_MESSAGES.disconnected));
        return;
      }
      if (mutation.type !== 'ws/setDeviceConfigSaveResult') return;
      const result = store.getters['ws/getDeviceConfigSaveResult'] as DeviceConfigSaveResult | null;
      if (!result || result.requestId !== requestId || result.deviceId !== submitted.id) return;
      if (!result.success) {
        finish(new Error(result.error || DEVICE_SETTINGS_MESSAGES.failed));
        return;
      }
      finish();
    });
    function abort() {
      finish(new Error(DEVICE_SETTINGS_MESSAGES.canceled));
    }
    /** Releases the timer, subscription and abort listener exactly once. */
    function finish(error?: Error) {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      unsubscribe();
      signal.removeEventListener('abort', abort);
      if (error) {
        reject(error);
        return;
      }
      store.commit('hub/setDeviceConfigOverride', submitted);
      resolve();
    }
    signal.addEventListener('abort', abort, { once: true });
    void store
      .dispatch('ws/emit', { event: 'saveDeviceConfigOverride', message: { ...submitted, requestId } })
      .catch(() => finish(new Error(DEVICE_SETTINGS_MESSAGES.submitFailed)));
  });
}
