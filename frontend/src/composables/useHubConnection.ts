/* Session recovery and initial data loading are independent of socket truth. */
import { store } from '@/store';
import { fetchHubState } from '@/services/hubstate.service';
import { onMounted, onBeforeUnmount } from 'vue';

let generation = 0;
let pendingConnection: Promise<void> | null = null;
let pendingSnapshot: Promise<void> | null = null;
let snapshotController: AbortController | null = null;

async function requestSnapshot(signal: AbortSignal, owner: number) {
  try {
    const data = await fetchHubState(signal);
    if (signal.aborted || owner !== generation) return;
    store.dispatch('hub/init', data);
    store.commit('hub/setLoadStatus', 'ready');
  } catch {
    if (signal.aborted || owner !== generation) return;
    store.commit('hub/setLoadStatus', 'failed');
  }
}

function loadHubState() {
  if (store.getters['hub/isInitialized']()) return;
  if (pendingSnapshot) return;
  snapshotController = new AbortController();
  store.commit('hub/setLoadStatus', 'loading');
  const request = requestSnapshot(snapshotController.signal, generation).finally(() => {
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
  const retry = () => { void startHubConnection(); };
  const visible = () => { if (document.visibilityState === 'visible') retry(); };
  onMounted(() => {
    retry();
    window.addEventListener('online', retry);
    window.addEventListener('focus', retry);
    document.addEventListener('visibilitychange', visible);
  });
  onBeforeUnmount(() => {
    window.removeEventListener('online', retry);
    window.removeEventListener('focus', retry);
    document.removeEventListener('visibilitychange', visible);
    stopHubConnection();
  });
}
