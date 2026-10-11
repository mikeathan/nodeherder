/*
 * Starts the hub link for a signed-in user: WebSocket connection, initial hub state
 * (GET hub state) and session validation. Moved unchanged from the previous MainLayout.
 */
import { store } from '@/store';
import { fetchHubState } from '@/services/hubstate.service';

function ensureSocket() {
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

export async function startHubConnection() {
  if (store.getters['auth/isAuthenticated']()) {
    // Persisted session: show the UI at once and validate in the background. If the session
    // is no longer valid, auth/restoreSession logs out and the sign-in page is shown.
    ensureSocket();
    loadHubState();
    store.dispatch('auth/restoreSession');
    return;
  }
  await store.dispatch('auth/restoreSession');
  if (store.getters['auth/isAuthenticated']()) {
    ensureSocket();
    await loadHubState();
  } else {
    store.commit('ws/setConnectionStatus', 'disconnected');
  }
}
