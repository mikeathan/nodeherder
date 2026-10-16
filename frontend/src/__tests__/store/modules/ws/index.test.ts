import { describe, expect, test, beforeEach, afterEach, jest } from '@jest/globals';
import { store } from '@/store/index';
import { default as hubState } from '../../../../../../docs/hub_state.json';

// Characterization of inbound WebSocket routing (backend -> store). Spec 007 FR-03.
type Handler = ((ev: { data?: string }) => void) | null;
class FakeSocket {
  static OPEN = 1;
  static CONNECTING = 0;
  static CLOSED = 3;
  static CLOSING = 2;
  static instances: FakeSocket[] = [];
  static last: FakeSocket | null = null;
  readyState = 0;
  sent: string[] = [];
  onopen: Handler = null;
  onmessage: Handler = null;
  onclose: Handler = null;
  onerror: Handler = null;
  constructor(public url: string) {
    FakeSocket.last = this;
    FakeSocket.instances.push(this);
  }
  close() { this.readyState = 3; this.onclose?.({}); }
  open() { this.readyState = 1; this.onopen?.({}); }
  send(raw: string) {
    this.sent.push(raw);
  }
  receive(type: string, payload: unknown) {
    this.onmessage?.({ data: JSON.stringify({ type, payload }) });
  }
}

const payload = (hubState as { payload: { devices: unknown[]; config: unknown } }).payload;

describe('ws inbound routing', () => {
  const g = globalThis as unknown as { WebSocket: unknown };
  let previous: unknown;

  beforeEach(() => {
    jest.useFakeTimers();
    store.dispatch('ws/disconnect');
    FakeSocket.instances = [];
    previous = g.WebSocket;
    g.WebSocket = FakeSocket;
    store.dispatch('hub/init', JSON.parse(JSON.stringify(payload)));
    store.commit('automations/clear');
    store.dispatch('ws/connect');
    jest.spyOn(console, 'log').mockImplementation(() => {});
    jest.spyOn(console, 'error').mockImplementation(() => {});
  });
  afterEach(() => {
    store.dispatch('ws/disconnect');
    jest.clearAllTimers();
    jest.useRealTimers();
    g.WebSocket = previous;
    jest.restoreAllMocks();
  });

  test('open marks connected and requests MCP status', () => {
    FakeSocket.last!.open();
    expect(store.getters['ws/getConnectionStatus']).toBe('connected');
    expect(FakeSocket.last!.sent.map((s) => JSON.parse(s).type)).toContain('loadMCPStatus');
  });

  test('deviceUpdated merges known exposes, last_seen and availability', () => {
    const id = '0xa4c1381b6fd53fc4';
    FakeSocket.last!.receive('deviceUpdated', { id, last_seen: '2026-10-10T10:00:00Z', availability: 'offline', data: { power: 77, unknown_key: 1 } });
    const d = store.getters['hub/findDevice'](id);
    expect(d.exposes.power.data).toBe(77);
    expect(d.exposes.unknown_key).toBeUndefined();
    expect(d.last_seen).toBe('2026-10-10T10:00:00Z');
    expect(d.availability).toBe('offline');
  });

  test('automations, automationUpdated, dashboardGroups, bridgeConfig and mcpStatus update state', () => {
    const s = FakeSocket.last!;
    s.receive('automations', [{ id: 'a1', triggers: [] }]);
    expect(store.getters['automations/initialized']()).toBe(true);
    s.receive('automationUpdated', { id: 'a1', triggers: [], description: 'changed' });
    expect(store.getters['automations/find']('a1').description).toBe('changed');
    s.receive('dashboardGroups', { X: { name: 'X', deviceGroup: {} } });
    expect(Object.keys(store.getters['hub/dashboardGroups']())).toEqual(['X']);
    s.receive('bridgeConfig', { permitJoin: true, maxTimeAllowed: { value: 120, unit: 'seconds' } });
    expect(store.getters['hub/bridge']().permitJoin).toBe(true);
    s.receive('mcpStatus', { running: true, enabled: true, name: 'm', version: '1', connectedClients: 2 });
    expect(store.getters['hub/mcpStatus']().connectedClients).toBe(2);
  });

  test('operationSuccess and operationFailed raise alerts', () => {
    const before = store.getters['alerts/messages']().length;
    FakeSocket.last!.receive('operationFailed', 'boom');
    FakeSocket.last!.receive('operationSuccess', null);
    const msgs = store.getters['alerts/messages']();
    expect(msgs.length).toBe(before + 2);
  });

  test('invalid JSON and envelopes cannot mutate hub state, unknown events are observable', () => {
    const socket = FakeSocket.last!;
    const before = store.getters['hub/listAllDevices']().length;
    for (const data of ['not-json', 'null', '42', '{"type":4}']) socket.onmessage?.({ data });
    socket.receive('unknown-type', {});
    expect(store.getters['hub/listAllDevices']()).toHaveLength(before);
    expect(console.error).toHaveBeenCalledWith('Invalid hub message');
    expect(console.error).toHaveBeenCalledWith('Unhandled hub message type:', 'unknown-type');
  });
  test('sending while the socket is not open refuses the message', async () => {
    await store.dispatch('ws/emit', { event: 'setValue', message: {} });
    expect(FakeSocket.last!.sent).toEqual([]);
    expect(console.error).toHaveBeenCalledWith('Cannot send message: WebSocket is not open.');
  });

  test('close moves to connecting and schedules a reconnect', () => {
    jest.useFakeTimers();
    FakeSocket.last!.onclose?.({});
    expect(store.getters['ws/getConnectionStatus']).toBe('reconnecting');
  });
});

// Lifecycle assertions share the same production owner as inbound routing.
describe('connection lifecycle', () => {
 let previous: unknown;
 beforeEach(() => {
  jest.useFakeTimers();
  previous = globalThis.WebSocket;
  Object.assign(globalThis, { WebSocket: FakeSocket });
  store.dispatch('ws/disconnect');
  FakeSocket.instances = [];
  jest.spyOn(console, 'log').mockImplementation(() => {});
  jest.spyOn(console, 'error').mockImplementation(() => {});
 });
 afterEach(() => {
  store.dispatch('ws/disconnect');
  jest.clearAllTimers(); jest.useRealTimers();
  Object.assign(globalThis, { WebSocket: previous });
  jest.restoreAllMocks();
 });
 test('startup is idle then connecting, with one socket for repeated clicks', async () => {
  expect(store.getters['ws/getConnectionStatus']).toBe('idle');
  await store.dispatch('ws/connect');
  expect(store.getters['ws/getConnectionStatus']).toBe('connecting');
  await store.dispatch('ws/connect');
  expect(FakeSocket.instances).toHaveLength(1);
  FakeSocket.last!.open();
  await store.dispatch('ws/connect');
  expect(FakeSocket.instances).toHaveLength(1);
 });
 test('hung first handshake times out at ten seconds and recovers', async () => {
  await store.dispatch('ws/connect');
  jest.advanceTimersByTime(9999);
  expect(store.getters['ws/getConnectionStatus']).toBe('connecting');
  jest.advanceTimersByTime(1);
  expect(store.getters['ws/getConnectionStatus']).toBe('reconnecting');
  jest.advanceTimersByTime(1000);
  expect(FakeSocket.instances).toHaveLength(2);
  FakeSocket.last!.open();
  expect(store.getters['ws/getConnectionStatus']).toBe('connected');
  expect(jest.getTimerCount()).toBe(0);
 });
 test('cleanup ignores late callbacks and cancels pending reconnect', async () => {
  await store.dispatch('ws/connect');
  const old = FakeSocket.last!;
  old.open(); old.close();
  await store.dispatch('ws/disconnect');
  old.onopen?.({}); old.onclose?.({});
  old.receive('deviceList', []);
  jest.advanceTimersByTime(120000);
  expect(store.getters['ws/getConnectionStatus']).toBe('idle');
  expect(store.getters['ws/getSocket']).toBeNull();
  expect(FakeSocket.instances).toHaveLength(1);
  expect(jest.getTimerCount()).toBe(0);
 });
 test('replaced sockets cannot change the current connection', async () => {
  await store.dispatch('ws/connect'); const old = FakeSocket.last!;
  old.close();
  await store.dispatch('ws/connect'); const current = FakeSocket.last!;
  current.open(); old.onopen?.({}); old.onclose?.({});
  expect(store.getters['ws/getSocket']).toBe(current);
  expect(store.getters['ws/getConnectionStatus']).toBe('connected');
  expect(jest.getTimerCount()).toBe(0);
 });
 test('constructor failures use bounded retry and disconnect cancels it', async () => {
  Object.assign(globalThis, { WebSocket: class { constructor() { throw new Error('invalid endpoint'); } } });
  await store.dispatch('ws/connect');
  expect(store.getters['ws/getConnectionStatus']).toBe('reconnecting');
  expect(jest.getTimerCount()).toBe(1);
  await store.dispatch('ws/disconnect');
  expect(jest.getTimerCount()).toBe(0);
 });
 test('ten failed retries stop until explicit retry, which resets the budget', async () => {
  await store.dispatch('ws/connect');
  for (let i=0; i<10; i++) {
   FakeSocket.last!.close();
   const delay = Math.min(1000 * 2 ** i, 30000);
   jest.advanceTimersByTime(delay - 1);
   expect(FakeSocket.instances).toHaveLength(i + 1);
   jest.advanceTimersByTime(1);
   expect(FakeSocket.instances).toHaveLength(i + 2);
  }
  FakeSocket.last!.close();
  expect(store.getters['ws/getConnectionStatus']).toBe('disconnected');
  expect(jest.getTimerCount()).toBe(0);
  await store.dispatch('ws/connect');
  FakeSocket.last!.open();
  expect(store.getters['ws/getConnectionStatus']).toBe('connected');
 });
});
