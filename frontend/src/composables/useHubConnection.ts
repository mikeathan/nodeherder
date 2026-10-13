/*
 * Starts the hub link for a signed-in user: WebSocket connection, initial hub state
 * (GET hub state) and session validation, including foreground/network recovery.
 */
import { store } from '@/store';
import { fetchHubState } from '@/services/hubstate.service';
import { onMounted, onBeforeUnmount } from 'vue';

function ensureSocket() {
  const socket = store.getters['ws/getSocket'] as WebSocket | null;
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return;
  const status = store.getters['ws/getConnectionStatus'];
  if (status !== 'connected' && status !== 'connecting') store.dispatch('ws/connect');
}

async function loadHubState() {
  if (!store.getters['hub/isInitialized']()) {
    try {
      store.dispatch('hub/init', await fetchHubState());
    } catch (err) {
      console.error('Failed to load hub state:', err);
      store.commit('ws/setConnectionStatus', 'disconnected');
      return;
    }
  }
  store.dispatch('hub/loadMCPStatus');
}

let pendingConnection: Promise<void> | null = null;

export function startHubConnection(): Promise<void> {
  if (pendingConnection) return pendingConnection;
  pendingConnection = validateAndConnect().finally(() => { pendingConnection = null; });
  return pendingConnection;
}

async function validateAndConnect() {
  await store.dispatch('auth/restoreSession');
  if (store.getters['auth/isAuthenticated']() && store.getters['auth/validation'] === 'authenticated') {
    ensureSocket();
    // A slow hub-state request must not block later session validation/recovery.
    void loadHubState();
  } else if (!store.getters['auth/isAuthenticated']()) {
    store.commit('ws/setConnectionStatus', 'disconnected');
  }
}

/** Recheck after a mobile browser returns to the foreground or its network recovers. */
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
  });
}
