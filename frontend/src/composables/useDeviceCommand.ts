/*
 * Device commands with pending state (spec 007 FR-04, AC-04; NH-02, FE-03).
 * A command stays "pending" until a deviceUpdated message reports the requested value or
 * COMMAND_TIMEOUT_MS passes; then the user is told the device did not respond. Commands are
 * never retried, and a second command for the same expose is refused while one is pending.
 */
import { reactive } from 'vue';
import { store } from '@/store';
import { Device, DeviceUpdate } from '@/types/device';
import { COMMAND_TIMEOUT_MS, commandKey, confirmedBy, PendingCommand } from '@/domain/commands';
import { useHub } from './useHub';

const pending = reactive(new Map<string, PendingCommand>());
const timers = new Map<string, ReturnType<typeof setTimeout>>();
let subscribed = false;

function settle(key: string) {
  clearTimeout(timers.get(key));
  timers.delete(key);
  pending.delete(key);
}

function subscribe() {
  if (subscribed) return;
  subscribed = true;
  store.subscribe((mutation) => {
    if (mutation.type !== 'hub/updateDevice' || pending.size === 0) return;
    confirmedBy(pending.values(), mutation.payload as DeviceUpdate).forEach(settle);
  });
}

export function useDeviceCommand() {
  subscribe();
  const { canControl } = useHub();

  const isPending = (deviceId: string, expose: string) => pending.has(commandKey(deviceId, expose));

  /** Sends `expose = value` to the device. Returns false when the command was refused. */
  function send(device: Device | undefined, expose: string, value: unknown): boolean {
    if (!device || !canControl(device)) return false;
    const key = commandKey(device.id, expose);
    if (pending.has(key)) return false;

    pending.set(key, { key, deviceId: device.id, expose, value, sentAt: Date.now() });
    timers.set(
      key,
      setTimeout(() => {
        if (!pending.has(key)) return;
        settle(key);
        store.dispatch('alerts/showError', `${device.friendly_name} did not confirm the change. Check that it is reachable and try again.`);
      }, COMMAND_TIMEOUT_MS)
    );
    store.dispatch('hub/setDeviceValue', { id: device.id, name: expose, value });
    return true;
  }

  return { send, isPending, canControl };
}
