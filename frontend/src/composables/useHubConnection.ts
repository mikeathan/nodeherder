/* Session recovery and initial data loading are independent of socket truth. */
import { store } from '@/store';
import { fetchHubState } from '@/services/hubstate.service';
import { onMounted, onBeforeUnmount } from 'vue';
import { createHubSnapshotReplay } from './hubSnapshotReplay';
import type { HubStateSnapshot } from '@/types/hub.type';

let generation = 0;
let pendingConnection: Promise<void> | null = null;
let pendingSnapshot: Promise<void> | null = null;
let snapshotController: AbortController | null = null;
function markSnapshotFailed(owner: number) {
  if (owner === generation) store.commit('hub/setLoadStatus', 'failed');
}

function createSnapshotAttempt(controller: AbortController, owner: number) {
  let failed = false;
  function fail() {
    failed = true;
    controller.abort();
    markSnapshotFailed(owner);
  }
  const replay = createHubSnapshotReplay(fail);
  const timer = setTimeout(fail, 8000);
  function dispose() {
    clearTimeout(timer);
    replay.stop();
    controller.signal.removeEventListener('abort', dispose);
  }
  controller.signal.addEventListener('abort', dispose, { once: true });
  return { replay, dispose, hasFailed: () => failed };
}

function publishSnapshot(snapshot: HubStateSnapshot) {
  const action = store.getters['hub/isInitialized']() ? 'hub/refresh' : 'hub/init';
  store.dispatch(action, snapshot);
  store.commit('hub/setLoadStatus', 'ready');
}

async function requestSnapshot(controller: AbortController, owner: number) {
  const attempt = createSnapshotAttempt(controller, owner);
  try {
    const snapshot = await fetchHubState(controller.signal);
    if (controller.signal.aborted || owner !== generation) return;
    const reconciled = attempt.replay.reconcile(snapshot);
    // Publishing the reconciled snapshot must not be captured as another live edit.
    attempt.replay.stop();
    publishSnapshot(reconciled);
  } catch {
    if (!controller.signal.aborted || attempt.hasFailed()) markSnapshotFailed(owner);
  } finally {
    attempt.dispose();
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

/** Coalesces session validation and starts the shared socket/snapshot work when authenticated. */
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

/** Owns shell recovery listeners and disposes shared connection work on unmount. */
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
