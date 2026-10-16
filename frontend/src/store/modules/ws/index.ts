import { markRaw } from 'vue';
import type { ActionContext, Module } from 'vuex';
import type { RootState } from '../../state';
import type { WSClientState } from './state';
import type { ConnectionStatusType } from '@/types/connection.type';
import { getEnv } from '@/config/env';

type Context = ActionContext<WSClientState, RootState>;
type Timer = ReturnType<typeof setTimeout>;
type Resources = { retry: Timer | null; handshake: Timer | null; attempts: number };
// Transport handles are not reactive application data. Each module instance owns
// its resources, and disconnect disposes them before invalidating the socket.
const owned = new WeakMap<WSClientState, Resources>();
const maxRetries = 10;
const handshakeTimeout = 10_000;

function resources(state: WSClientState): Resources {
  let current = owned.get(state);
  if (!current) {
    current = { retry: null, handshake: null, attempts: 0 };
    owned.set(state, current);
  }
  return current;
}

function clearTimers(current: Resources) {
  if (current.retry !== null) clearTimeout(current.retry);
  if (current.handshake !== null) clearTimeout(current.handshake);
  current.retry = null;
  current.handshake = null;
}

function scheduleReconnect(context: Context) {
  const current = resources(context.state);
  clearTimers(current);
  if (current.attempts >= maxRetries) {
    context.commit('setConnectionStatus', 'disconnected');
    return;
  }
  context.commit('setConnectionStatus', 'reconnecting');
  const delay = Math.min(1000 * 2 ** current.attempts, 30_000);
  current.retry = setTimeout(() => {
    current.retry = null;
    current.attempts += 1;
    context.dispatch('connect', { automatic: true });
  }, delay);
}

const actionEvents = new Map([
  ['automations', 'automations/init'], ['deviceList', 'hub/setDevices'],
  ['appConfig', 'hub/setAppConfig'], ['metrics', 'metrics/store'],
  ['logger', 'console/addMessage'], ['operationFailed', 'alerts/showError'],
]);
const mutationEvents = new Map([
  ['automationUpdated', 'automations/update'], ['deviceUpdated', 'hub/updateDevice'],
  ['bridgeConfig', 'hub/setBridgeSettings'], ['deviceAdded', 'hub/addDevice'],
  ['dashboardGroups', 'hub/setDashboardGroups'], ['mcpStatus', 'hub/setMCPStatus'],
]);

type HubMessage = { type: string; payload?: unknown };
function decodeMessage(raw: string): HubMessage | null {
  try {
    const value: unknown = JSON.parse(raw);
    if (typeof value === 'object' && value !== null && 'type' in value && typeof value.type === 'string') return { type: value.type, payload: Reflect.get(value, 'payload') };
  } catch {
    console.error('Invalid hub message');
  }
  return null;
}

function routeMessage(context: Context, message: HubMessage) {
  const action = actionEvents.get(message.type);
  if (action) {
    context.dispatch(action, message.payload, { root: true });
    return;
  }
  const mutation = mutationEvents.get(message.type);
  if (mutation) {
    context.commit(mutation, message.payload, { root: true });
    return;
  }
  if (message.type === 'operationSuccess') {
    context.dispatch('alerts/showSuccess', message.payload || 'Operation was successful.', { root: true });
    return;
  }
  console.error('Unhandled hub message type:', message.type);
}

function bindSocket(context: Context, socket: WebSocket) {
  const isCurrent = () => context.state.socket === socket;
  socket.onopen = () => {
    if (!isCurrent()) return;
    const current = resources(context.state);
    clearTimers(current);
    current.attempts = 0;
    context.commit('setConnectionStatus', 'connected');
    context.dispatch('emit', { event: 'loadMCPStatus', message: {} });
  };
  socket.onmessage = (event) => {
    if (!isCurrent()) return;
    const message = decodeMessage(event.data);
    if (message) routeMessage(context, message);
  };
  socket.onclose = () => {
    if (isCurrent()) scheduleReconnect(context);
  };
  // Browsers follow a failed handshake with close; close owns retry scheduling.
  socket.onerror = () => {};
  resources(context.state).handshake = setTimeout(() => {
    if (isCurrent()) socket.close();
  }, handshakeTimeout);
}

function connect(context: Context, { automatic = false } = {}) {
  const socket = context.state.socket;
  if (socket && socket.readyState <= WebSocket.OPEN) return;
  const current = resources(context.state);
  clearTimers(current);
  if (!automatic) current.attempts = 0;
  const status = context.state.connectionStatus === 'idle' ? 'connecting' : 'reconnecting';
  context.commit('setConnectionStatus', status);
  try {
    const next = markRaw(new WebSocket(getEnv('VITE_WS_BASE_URL')));
    context.commit('setSocket', next);
    bindSocket(context, next);
  } catch {
    scheduleReconnect(context);
  }
}

function disconnect(context: Context) {
  const current = resources(context.state);
  clearTimers(current);
  current.attempts = 0;
  const socket = context.state.socket;
  context.commit('setSocket', null);
  context.commit('setConnectionStatus', 'idle');
  if (socket && socket.readyState < WebSocket.CLOSING) socket.close();
}

export const WSClientModule: Module<WSClientState, RootState> = {
  namespaced: true,
  state: () => ({ socket: null, connectionStatus: 'idle' }),
  getters: {
    getConnectionStatus: (state) => state.connectionStatus,
    getSocket: (state) => state.socket,
  },
  mutations: {
    setSocket(state, socket: WebSocket | null) { state.socket = socket; },
    setConnectionStatus(state, status: ConnectionStatusType) { state.connectionStatus = status; },
    sendMessage(state, { event, message }) {
      if (state.socket?.readyState === WebSocket.OPEN) {
        state.socket.send(JSON.stringify({ type: event, payload: message }));
      } else {
        console.error('Cannot send message: WebSocket is not open.');
      }
    },
  },
  actions: {
    connect,
    disconnect,
    emit({ commit }, { event, message }) { commit('sendMessage', { event, message }); },
  },
};
