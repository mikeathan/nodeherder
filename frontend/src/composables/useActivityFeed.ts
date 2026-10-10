/*
 * Recent activity (spec 007 FR-05, FR-12 / AC-27). Observes device updates in the store and
 * keeps a bounded in-memory list (never persisted, FE-05). The on/off preference is
 * persisted; while off nothing is recorded and the list is cleared.
 */
import { readonly, ref, shallowRef } from 'vue';
import { store } from '@/store';
import { Device, DeviceUpdate } from '@/types/device';
import { ActivityEntry, recordUpdate, snapshotDevices, ValueSnapshot } from '@/domain/activity';
import { safeStorage } from '@/utils/storage';

export const ACTIVITY_STORAGE_KEY = 'nodeherder_activity';

const enabled = ref(safeStorage.getItem(ACTIVITY_STORAGE_KEY) !== 'off');
const entries = shallowRef<ActivityEntry[]>([]);
let snapshot: ValueSnapshot = new Map();
let unsubscribe: (() => void) | null = null;

const allDevices = () => store.getters['hub/listAllDevices']() as Device[];

function start() {
  if (unsubscribe) return;
  snapshot = snapshotDevices(allDevices());
  unsubscribe = store.subscribe((mutation) => {
    switch (mutation.type) {
      case 'hub/updateDevice':
        entries.value = recordUpdate(entries.value, snapshot, mutation.payload as DeviceUpdate, Date.now());
        break;
      case 'hub/setDevices':
      case 'hub/addDevice':
      case 'hub/clear':
        snapshot = snapshotDevices(allDevices());
        break;
    }
  });
}

function stop() {
  unsubscribe?.();
  unsubscribe = null;
  snapshot = new Map();
  entries.value = [];
}

export function useActivityFeed() {
  if (enabled.value) start();
  return {
    enabled: readonly(enabled),
    entries: readonly(entries),
    setEnabled(on: boolean) {
      enabled.value = on;
      safeStorage.setItem(ACTIVITY_STORAGE_KEY, on ? 'on' : 'off');
      if (on) start();
      else stop();
    },
    clear() {
      entries.value = [];
    },
  };
}
