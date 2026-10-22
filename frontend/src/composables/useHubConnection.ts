/* Session recovery and initial data loading are independent of socket truth. */
import { store } from '@/store';
import { fetchHubState } from '@/services/hubstate.service';
import { onMounted, onBeforeUnmount } from 'vue';
import type { MutationPayload } from 'vuex';
import { appConfigMutations } from '@/store/modules/hub-state';
import type { Device, DeviceUpdate } from '@/types/device';
import type { HubStateSnapshot } from '@/types/hub.type';

let generation = 0;
let pendingConnection: Promise<void> | null = null;
let pendingSnapshot: Promise<void> | null = null;
let snapshotController: AbortController | null = null;
const maxSnapshotConfigMutations = 256;

function preserveLiveReports(data: HubStateSnapshot, updates: Map<string, DeviceUpdate>, added: Map<string, Device>): HubStateSnapshot {
  const devices = new Map(data.devices.map((device) => [device.id, device]));
  for (const [id, device] of added) devices.set(id, device);
  for (const [id, update] of updates) {
    const device = devices.get(id);
    if (!device) continue;
    const exposes = { ...device.exposes };
    for (const [name, value] of Object.entries(update.data)) {
      if (exposes[name]) exposes[name] = { ...exposes[name], data: value };
    }
    devices.set(id, { ...device, exposes, last_seen: update.last_seen || device.last_seen, availability: update.availability || device.availability });
  }
  return { ...data, devices: [...devices.values()] };
}

async function requestSnapshot(controller: AbortController, owner: number) {
  const { signal } = controller;
  // Coalesce by known device/field, bounded by the live registry rather than message count.
  const updates = new Map<string, DeviceUpdate>();
  const added = new Map<string, Device>();
  const configEdits: MutationPayload[] = [];
  let configOverflow = false;
  let listChanged = false;
  let timedOut = false;
  const timer = setTimeout(() => { timedOut = true; controller.abort(); }, 8000);
  const unsubscribe = store.subscribe((mutation) => {
    if (mutation.type === 'hub/updateDevice') {
      const update = mutation.payload as DeviceUpdate;
      const device = store.getters['hub/findDevice'](update.id) as Device | undefined;
      if (!device) return;
      const previous = updates.get(update.id);
      const data = { ...previous?.data };
      for (const [name, value] of Object.entries(update.data ?? {})) if (device.exposes[name]) data[name] = value;
      updates.set(update.id, { ...previous, ...update, data, last_seen: update.last_seen || previous?.last_seen || '', availability: update.availability || previous?.availability || null });
    } else if (mutation.type === 'hub/addDevice') {
      const device = mutation.payload as Device;
      added.set(device.id, device);
    } else if (mutation.type === 'hub/setDevices') {
      listChanged = true;
    } else if (mutation.type.startsWith('hub/') && Object.prototype.hasOwnProperty.call(appConfigMutations, mutation.type.slice(4))) {
      if (configEdits.length >= maxSnapshotConfigMutations) {
        configOverflow = true;
        controller.abort();
        if (owner === generation) store.commit('hub/setLoadStatus', 'failed');
        return;
      }
      configEdits.push({
        type: mutation.type.slice(4),
        payload: mutation.payload === undefined ? undefined : JSON.parse(JSON.stringify(mutation.payload)),
      });
    }
  });
  const dispose = () => {
    clearTimeout(timer);
    unsubscribe();
    signal.removeEventListener('abort', dispose);
  };
  signal.addEventListener('abort', dispose, { once: true });
  try {
    let data = await fetchHubState(signal);
    if (signal.aborted || owner !== generation) return;
    if (listChanged) data = { ...data, devices: store.getters['hub/listAllDevices']() as Device[] };
    else data = preserveLiveReports(data, updates, added);
    if (configEdits.length > 0) {
      const next = { appConfig: JSON.parse(JSON.stringify(data.config)) as HubStateSnapshot['config'] };
      for (const edit of configEdits) appConfigMutations[edit.type](next, edit.payload);
      data = { ...data, config: next.appConfig };
    }
    unsubscribe();
    store.dispatch(store.getters['hub/isInitialized']() ? 'hub/refresh' : 'hub/init', data);
    store.commit('hub/setLoadStatus', 'ready');
  } catch {
    if ((signal.aborted && !timedOut && !configOverflow) || owner !== generation) return;
    store.commit('hub/setLoadStatus', 'failed');
  } finally {
    dispose();
  }
}

function loadHubState(force = false) {
  if (!force && (pendingSnapshot || (store.getters['hub/isInitialized']() && store.getters['hub/loadStatus'] !== 'failed'))) return;
  if (force) snapshotController?.abort();
  snapshotController = new AbortController();
  store.commit('hub/setLoadStatus', 'loading');
  const request = requestSnapshot(snapshotController, generation).finally(() => {
    if (pendingSnapshot !== request) return;
    pendingSnapshot = null;
    snapshotController = null;
  });
  pendingSnapshot = request;
}

async function validateAndConnect(owner: number) {
  await store.dispatch('auth/restoreSession');
  if (owner !== generation) return;
  if (!store.getters['auth/isAuthenticated']()) {
    stopHubConnection();
    return;
  }
  if (store.getters['auth/validation'] !== 'authenticated') return;
  store.dispatch('ws/connect');
  // The socket action owns deduplication; slow HTTP does not block validation.
  loadHubState();
}

export function startHubConnection(): Promise<void> {
  if (pendingConnection) return pendingConnection;
  const check = validateAndConnect(generation).finally(() => {
    if (pendingConnection === check) pendingConnection = null;
  });
  pendingConnection = check;
  return check;
}

/** Dispose the shell's network work before another session/link can take ownership. */
export function stopHubConnection() {
  generation += 1;
  pendingConnection = null;
  snapshotController?.abort();
  snapshotController = null;
  pendingSnapshot = null;
  store.dispatch('ws/disconnect');
  if (store.getters['hub/loadStatus'] === 'loading') store.commit('hub/setLoadStatus', 'idle');
}

export function useHubConnection() {
  let connected = false;
  const unsubscribe = store.subscribe((mutation) => {
    if (mutation.type === 'auth/logout') {
      connected = false;
      stopHubConnection();
      return;
    }
    if (mutation.type !== 'ws/setConnectionStatus' || mutation.payload !== 'connected') return;
    if (connected) loadHubState(true);
    connected = true;
  });
  const retry = () => { void startHubConnection(); };
  const visible = () => { if (document.visibilityState === 'visible') retry(); };
  onMounted(() => {
    retry();
    window.addEventListener('online', retry);
    window.addEventListener('focus', retry);
    document.addEventListener('visibilitychange', visible);
  });
  onBeforeUnmount(() => {
    unsubscribe();
    window.removeEventListener('online', retry);
    window.removeEventListener('focus', retry);
    document.removeEventListener('visibilitychange', visible);
    stopHubConnection();
  });
}
