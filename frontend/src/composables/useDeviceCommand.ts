/*
 * Device commands with pending state (spec 007 FR-04, AC-04; NH-02, FE-03).
 * A command stays "pending" until a deviceUpdated message reports the requested value or
 * the configured confirmation timeout passes; then the user is told the device did not respond. Commands are
 * never retried, and a second command for the same expose is refused while one is pending.
 */
import { reactive } from 'vue';
import { store } from '@/store';
import { Device, DeviceUpdate } from '@/types/device';
import { commandKey, confirmedBy, PendingCommand } from '@/domain/commands';
import { confirmationTimeoutMs } from '@/domain/reporting-settings';
import type { DeviceConfig } from '@/types/settings.type';
import { useHub } from './useHub';

const pending = reactive(new Map<string, PendingCommand>());
const timers = new Map<string, ReturnType<typeof setTimeout>>();
let subscribed = false;

function settle(key: string) {
  clearTimeout(timers.get(key));
  timers.delete(key);
  pending.delete(key);
}

function trackPending(device: Device, expose: string, value: unknown, timeout: number, onTimeout: () => void) {
  const key = commandKey(device.id, expose);
  pending.set(key, { key, deviceId: device.id, expose, value, sentAt: Date.now() });
  timers.set(key, setTimeout(() => {
    if (!pending.has(key)) return;
    settle(key);
    onTimeout();
  }, timeout));
}

function reportConfirmationTimeout(device: Device) {
  store.dispatch('alerts/showError', `${device.friendly_name} did not confirm the change. Check that it is reachable and try again.`);
}

function subscribe() {
  if (subscribed) return;
  subscribed = true;
  store.subscribe((mutation) => {
    if (['hub/clear', 'auth/logout', 'ws/setConnectionStatus'].includes(mutation.type)) {
      if (mutation.type === 'ws/setConnectionStatus' && mutation.payload === 'connected') return;
      for (const key of pending.keys()) settle(key);
      return;
    }
    if (mutation.type !== 'hub/updateDevice' || pending.size === 0) return;
    confirmedBy(pending.values(), mutation.payload as DeviceUpdate).forEach(settle);
  });
}

/** Shares device command submission and pending guards across views; owns bounded timers cleared on disconnect/reset. */
export function useDeviceCommand() {
  subscribe();
  const { canControl } = useHub();

  /** Reports an outstanding state confirmation or short stateless submission guard. */
  const isPending = (deviceId: string, expose: string) => pending.has(commandKey(deviceId, expose));

  /** Sends `expose = value` to the device. Returns false when the command was refused. */
  function send(device: Device | undefined, expose: string, value: unknown): boolean {
    if (!device || !canControl(device)) return false;
    const key = commandKey(device.id, expose);
    if (pending.has(key)) return false;

    const timeout = confirmationTimeoutMs(
      store.getters['hub/findDeviceSetting'](device.id) as DeviceConfig | undefined,
      store.getters['hub/deviceDefaults']() as DeviceConfig | undefined
    );
    if (timeout === null) {
      store.dispatch('alerts/showError', 'Configure a valid command confirmation timeout in hub settings.');
      return false;
    }
    trackPending(device, expose, value, timeout, () => reportConfirmationTimeout(device));
    store.dispatch('hub/setDeviceValue', { id: device.id, name: expose, value });
    return true;
  }

  /** Requests a stateless action with a one-second duplicate guard; no execution confirmation is inferred. */
  function sendAction(device: Device | undefined, expose: string): boolean {
    if (!device || !canControl(device)) return false;
    const key = commandKey(device.id, expose);
    if (pending.has(key)) return false;
    trackPending(device, expose, null, 1000, () => {});
    store.dispatch('hub/setDeviceValue', { id: device.id, name: expose, value: null });
    store.dispatch('alerts/showSuccess', 'Action requested. Device execution cannot be confirmed.');
    return true;
  }

  return { send, sendAction, isPending, canControl };
}
