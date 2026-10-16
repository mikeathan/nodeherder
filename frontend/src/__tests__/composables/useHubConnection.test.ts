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
afterEach(() => { stopHubConnection(); Object.assign(globalThis, { WebSocket: previous }); jest.restoreAllMocks(); });

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
