import { onMounted, onBeforeUnmount } from 'vue';
import { store } from '@/store';
import { startHubConnection, stopHubConnection, useHubConnection } from '@/composables/useHubConnection';
import { fetchHubState } from '@/services/hubstate.service';
import { restoreSession } from '@/services/auth.service';
import hubState from '../../../../docs/hub_state.json';

jest.mock('vue', () => ({ ...jest.requireActual('vue'), onMounted: jest.fn(), onBeforeUnmount: jest.fn() }));
jest.mock('@/services/hubstate.service', () => ({ fetchHubState: jest.fn() }));
jest.mock('@/services/auth.service', () => ({ restoreSession: jest.fn() }));
const session = { user: { id: 'test', username: 'Test' }, isAuthenticated: true };
const payload = hubState.payload as unknown as Awaited<ReturnType<typeof fetchHubState>>;

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: Error) => void;
  const promise = new Promise<T>((done, fail) => { resolve = done; reject = fail; });
  return { promise, resolve, reject };
}
async function settle() { for (let i = 0; i < 12; i++) await Promise.resolve(); }
class Socket {
  static OPEN = 1;
  static CLOSING = 2;
  static latest: Socket;
  readyState = 0;
  onopen: (() => void) | null = null;
  onclose: (() => void) | null = null;
  constructor() { Socket.latest = this; }
  close() { this.readyState = 3; this.onclose?.(); }
  send() {}
  open() { this.readyState = 1; this.onopen?.(); }
}
let previous: unknown;
let events: EventTarget;
beforeEach(() => {
  jest.clearAllMocks();
  previous = globalThis.WebSocket;
  Object.assign(globalThis, { WebSocket: Socket });
  stopHubConnection();
  store.commit('hub/clear');
  store.commit('auth/logout');
  jest.mocked(restoreSession).mockResolvedValue(session);
  jest.mocked(fetchHubState).mockResolvedValue(payload);
  jest.spyOn(console, 'error').mockImplementation(() => {});
  events = new EventTarget();
  Object.assign(window, { addEventListener: jest.fn(events.addEventListener.bind(events)), removeEventListener: jest.fn(events.removeEventListener.bind(events)) });
  Object.assign(document, { visibilityState: 'visible', addEventListener: jest.fn(events.addEventListener.bind(events)), removeEventListener: jest.fn(events.removeEventListener.bind(events)) });
});
afterEach(() => {
  for (const [unmount] of jest.mocked(onBeforeUnmount).mock.calls) unmount();
  stopHubConnection(); Object.assign(globalThis, { WebSocket: previous }); jest.restoreAllMocks();
});

test('coalesces validation and waits before starting the first socket', async () => {
  const check = deferred<typeof session>();
  jest.mocked(restoreSession).mockReturnValue(check.promise);
  const first = startHubConnection();
  expect(startHubConnection()).toBe(first);
  expect(store.getters['ws/getConnectionStatus']).toBe('idle');
  check.resolve(session);
  await first; await settle();
  expect(restoreSession).toHaveBeenCalledTimes(1);
  expect(store.getters['ws/getConnectionStatus']).toBe('connecting');
  expect(store.getters['hub/loadStatus']).toBe('ready');
});

test('snapshot failure preserves open socket status and retry clears only the data error', async () => {
  const data = deferred<typeof payload>();
  jest.mocked(fetchHubState).mockReturnValueOnce(data.promise);
  await startHubConnection(); Socket.latest.open();
  data.resolve(payload); await settle();
  expect(store.getters['ws/getConnectionStatus']).toBe('connected');
  store.commit('hub/clear');
  jest.mocked(fetchHubState).mockRejectedValueOnce(new Error('503'));
  await startHubConnection(); await settle();
  expect(store.getters['hub/loadStatus']).toBe('failed');
  expect(store.getters['ws/getConnectionStatus']).toBe('connected');
  const socket = store.getters['ws/getSocket'];
  await startHubConnection(); await settle();
  expect(store.getters['hub/loadStatus']).toBe('ready');
  expect(store.getters['ws/getSocket']).toBe(socket);
});

test('slow snapshot is shared and never blocks subsequent session recovery', async () => {
  const data = deferred<typeof payload>();
  jest.mocked(fetchHubState).mockReturnValue(data.promise);
  await startHubConnection();
  await startHubConnection();
  expect(restoreSession).toHaveBeenCalledTimes(2);
  expect(fetchHubState).toHaveBeenCalledTimes(1);
  expect(store.getters['hub/loadStatus']).toBe('loading');
  data.resolve(payload); await settle();
  expect(store.getters['hub/loadStatus']).toBe('ready');
});

test('unmount aborts a pending snapshot and ignores late success', async () => {
  const data = deferred<typeof payload>();
  jest.mocked(fetchHubState).mockReturnValue(data.promise);
  useHubConnection();
  jest.mocked(onMounted).mock.calls[0][0](); await settle();
  const signal = jest.mocked(fetchHubState).mock.calls[0][0];
  jest.mocked(onBeforeUnmount).mock.calls[0][0]();
  expect(signal?.aborted).toBe(true);
  expect(store.getters['ws/getConnectionStatus']).toBe('idle');
  data.resolve(payload); await settle();
  expect(store.getters['hub/isInitialized']()).toBe(false);
  expect(store.getters['hub/loadStatus']).toBe('idle');
  expect(window.removeEventListener).toHaveBeenCalledTimes(2);
  expect(document.removeEventListener).toHaveBeenCalledTimes(1);
});

test('late validation after disposal cannot open a socket', async () => {
  const check = deferred<typeof session>();
  jest.mocked(restoreSession).mockReturnValue(check.promise);
  const first = startHubConnection();
  stopHubConnection();
  check.resolve(session); await first;
  expect(store.getters['ws/getConnectionStatus']).toBe('idle');
  expect(fetchHubState).not.toHaveBeenCalled();
});

test('signed-out and unavailable validation do not start a connection', async () => {
  jest.mocked(restoreSession).mockResolvedValueOnce(null);
  await startHubConnection();
  expect(store.getters['ws/getConnectionStatus']).toBe('idle');
  await store.dispatch('auth/loginUser', session);
  jest.mocked(restoreSession).mockRejectedValueOnce(new Error('offline'));
  await startHubConnection();
  expect(store.getters['auth/isAuthenticated']()).toBe(true);
  expect(store.getters['auth/validation']).toBe('unavailable');
  expect(fetchHubState).not.toHaveBeenCalled();
});

test('foreground/online recovery runs until disposal, hidden documents do not', async () => {
  useHubConnection();
  jest.mocked(onMounted).mock.calls[0][0](); await settle();
  Object.assign(document, { visibilityState: 'hidden' });
  events.dispatchEvent(new Event('visibilitychange')); await settle();
  expect(restoreSession).toHaveBeenCalledTimes(1);
  Object.assign(document, { visibilityState: 'visible' });
  events.dispatchEvent(new Event('visibilitychange')); await settle();
  events.dispatchEvent(new Event('focus')); await settle();
  events.dispatchEvent(new Event('online')); await settle();
  expect(restoreSession).toHaveBeenCalledTimes(4);
  jest.mocked(onBeforeUnmount).mock.calls[0][0]();
  events.dispatchEvent(new Event('online')); await settle();
  expect(restoreSession).toHaveBeenCalledTimes(4);
});


test('a disposed failed request cannot overwrite the next successful load', async () => {
 const old = deferred<typeof payload>();
 jest.mocked(fetchHubState).mockReturnValueOnce(old.promise);
 await startHubConnection();
 stopHubConnection();
 await startHubConnection(); await settle();
 expect(store.getters['hub/loadStatus']).toBe('ready');
 old.reject(new Error('late failure')); await settle();
 expect(store.getters['hub/loadStatus']).toBe('ready');
 expect(store.getters['hub/isInitialized']()).toBe(true);
});

test('reconnect refreshes missed state and preserves reports arriving during the snapshot', async () => {
  useHubConnection();
  jest.mocked(onMounted).mock.calls[0][0](); await settle();
  Socket.latest.open(); await settle();
  expect(fetchHubState).toHaveBeenCalledTimes(1);
  const staleSnapshot = JSON.parse(JSON.stringify(payload)) as typeof payload;
  const snapshot = deferred<typeof payload>();
  jest.mocked(fetchHubState).mockReturnValueOnce(snapshot.promise);
  store.commit('ws/setConnectionStatus', 'reconnecting');
  store.commit('ws/setConnectionStatus', 'connected');
  expect(fetchHubState).toHaveBeenCalledTimes(2);
  const device = payload.devices.find((d) => Object.keys(d.exposes).length > 0)!;
  const field = Object.keys(device.exposes)[0];
  store.commit('hub/updateDevice', { id: device.id, data: { [field]: 123 }, last_seen: 'new', availability: 'online' });
  const bridge = { permitJoin: true, maxTimeAllowed: { value: 120, unit: 'seconds' } };
  store.commit('hub/setBridgeSettings', bridge);
  store.commit('hub/setMCPStatus', { running: true, connectedClients: 1 });
  snapshot.resolve(staleSnapshot); await settle();
  expect(store.getters['hub/findDevice'](device.id).exposes[field].data).toBe(123);
  expect(store.getters['hub/findDevice'](device.id).last_seen).toBe('new');
  expect(store.getters['hub/bridge']()).toEqual(bridge);
  expect(store.getters['hub/mcpStatus']()).toEqual({ running: true, connectedClients: 1 });
  expect(store.getters['hub/loadStatus']).toBe('ready');
  jest.mocked(onBeforeUnmount).mock.calls[0][0]();
});

test('a later reconnect invalidates an older pending snapshot', async () => {
  useHubConnection();
  jest.mocked(onMounted).mock.calls[0][0](); await settle(); Socket.latest.open();
  const first = deferred<typeof payload>(), second = deferred<typeof payload>();
  jest.mocked(fetchHubState).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
  store.commit('ws/setConnectionStatus', 'connected');
  const oldSignal = jest.mocked(fetchHubState).mock.calls[1][0];
  store.commit('ws/setConnectionStatus', 'connected');
  expect(oldSignal?.aborted).toBe(true);
  second.resolve(payload); await settle();
  first.resolve({ ...payload, devices: [] }); await settle();
  expect(store.getters['hub/listAllDevices']()).toHaveLength(payload.devices.length);
  jest.mocked(onBeforeUnmount).mock.calls[0][0]();
});

test('snapshot timeout reports failure while preserving socket truth and supports retry', async () => {
  jest.useFakeTimers();
  jest.mocked(fetchHubState).mockImplementationOnce((signal) => new Promise((_resolve, reject) => {
    signal?.addEventListener('abort', () => reject(new Error('aborted')), { once: true });
  }));
  useHubConnection();
  jest.mocked(onMounted).mock.calls[0][0](); await settle(); Socket.latest.open();
  jest.advanceTimersByTime(8000); await settle();
  expect(store.getters['hub/loadStatus']).toBe('failed');
  expect(store.getters['ws/getConnectionStatus']).toBe('connected');
  jest.mocked(fetchHubState).mockResolvedValueOnce(payload);
  await startHubConnection(); await settle();
  expect(store.getters['hub/loadStatus']).toBe('ready');
  jest.mocked(onBeforeUnmount).mock.calls[0][0]();
  expect(jest.getTimerCount()).toBe(0);
  jest.useRealTimers();
});

test('logout invalidates a pending snapshot before shell unmount', async () => {
  const snapshot = deferred<typeof payload>();
  jest.mocked(fetchHubState).mockReturnValueOnce(snapshot.promise);
  useHubConnection();
  jest.mocked(onMounted).mock.calls[0][0](); await settle();
  const signal = jest.mocked(fetchHubState).mock.calls[0][0];
  store.commit('auth/logout');
  expect(signal?.aborted).toBe(true);
  snapshot.resolve(payload); await settle();
  expect(store.getters['hub/isInitialized']()).toBe(false);
  expect(store.getters['hub/listAllDevices']()).toEqual([]);
  jest.mocked(onBeforeUnmount).mock.calls[0][0]();
});

test('failed initialized refresh can be retried without another reconnect', async () => {
  useHubConnection();
  jest.mocked(onMounted).mock.calls[0][0](); await settle(); Socket.latest.open();
  jest.mocked(fetchHubState).mockRejectedValueOnce(new Error('503'));
  store.commit('ws/setConnectionStatus', 'connected'); await settle();
  expect(store.getters['hub/isInitialized']()).toBe(true);
  expect(store.getters['hub/loadStatus']).toBe('failed');
  await startHubConnection(); await settle();
  expect(fetchHubState).toHaveBeenCalledTimes(3);
  expect(store.getters['hub/loadStatus']).toBe('ready');
  jest.mocked(onBeforeUnmount).mock.calls[0][0]();
});

test('live config mutations preserve unrelated server fields during recovery', async () => {
  useHubConnection();
  jest.mocked(onMounted).mock.calls[0][0](); await settle(); Socket.latest.open();
  const snapshot = deferred<typeof payload>();
  const recovered = JSON.parse(JSON.stringify(payload)) as typeof payload;
  recovered.config.hub.assistant.url = 'http://updated-assistant.invalid';
  recovered.config.hub.logger.level = 'error';
  recovered.config.hub.dashboardGroups = {
    recovered: { name: 'recovered', deviceGroup: {} },
  };
  jest.mocked(fetchHubState).mockReturnValueOnce(snapshot.promise);
  store.commit('ws/setConnectionStatus', 'connected');
  const bridge = { permitJoin: true, maxTimeAllowed: { value: 120, unit: 'seconds' } };
  store.commit('hub/setBridgeSettings', bridge);
  store.commit('hub/setDashboardGroup', { name: 'live', deviceGroup: {} });
  snapshot.resolve(recovered); await settle();
  expect(store.getters['hub/bridge']()).toEqual(bridge);
  expect(store.getters['hub/assistant']().url).toBe('http://updated-assistant.invalid');
  expect(store.getters['hub/logger']().level).toBe('error');
  expect(Object.keys(store.getters['hub/dashboardGroups']()).sort()).toEqual(['live', 'recovered']);
  jest.mocked(onBeforeUnmount).mock.calls[0][0]();
});

test('config replay overflow fails safely and can be retried', async () => {
  useHubConnection();
  jest.mocked(onMounted).mock.calls[0][0](); await settle(); Socket.latest.open();
  const snapshot = deferred<typeof payload>();
  jest.mocked(fetchHubState).mockReturnValueOnce(snapshot.promise);
  store.commit('ws/setConnectionStatus', 'connected');
  const signal = jest.mocked(fetchHubState).mock.calls[1][0];
  for (let i = 0; i <= 256; i++) store.commit('hub/setMCPSettings', true);
  expect(signal?.aborted).toBe(true);
  expect(store.getters['hub/loadStatus']).toBe('failed');
  expect(store.getters['hub/isInitialized']()).toBe(true);
  snapshot.resolve({ ...payload, devices: [] }); await settle();
  expect(store.getters['hub/listAllDevices']()).toHaveLength(payload.devices.length);
  await startHubConnection(); await settle();
  expect(store.getters['hub/loadStatus']).toBe('ready');
});
