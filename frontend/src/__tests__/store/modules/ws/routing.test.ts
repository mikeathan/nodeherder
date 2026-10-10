import { describe, expect, test, beforeEach, afterEach, jest } from '@jest/globals';
import { store } from '@/store/index';
import { default as hubState } from '../../../../../../docs/hub_state.json';

// Characterization of inbound WebSocket routing (backend -> store). Spec 007 FR-03.
type Handler = ((ev: { data?: string }) => void) | null;
class FakeSocket {
  static OPEN = 1;
  static last: FakeSocket | null = null;
  readyState = 1;
  sent: string[] = [];
  onopen: Handler = null;
  onmessage: Handler = null;
  onclose: Handler = null;
  onerror: Handler = null;
  constructor(public url: string) {
    FakeSocket.last = this;
  }
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
    previous = g.WebSocket;
    g.WebSocket = FakeSocket;
    store.dispatch('hub/init', JSON.parse(JSON.stringify(payload)));
    store.commit('automations/clear');
    store.dispatch('ws/connect');
    jest.spyOn(console, 'log').mockImplementation(() => {});
    jest.spyOn(console, 'error').mockImplementation(() => {});
  });
  afterEach(() => {
    g.WebSocket = previous;
    jest.restoreAllMocks();
  });

  test('open marks connected and requests MCP status', () => {
    FakeSocket.last!.onopen?.({});
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

  test('close moves to connecting and schedules a reconnect', () => {
    jest.useFakeTimers();
    FakeSocket.last!.onclose?.({});
    expect(store.getters['ws/getConnectionStatus']).toBe('connecting');
    jest.useRealTimers();
  });
});
