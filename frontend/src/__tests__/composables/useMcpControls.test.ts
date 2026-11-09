import { effectScope, nextTick } from 'vue';
import { store } from '@/store';
import { createAppconfig } from '@/contracts/settings';
import { MCP_ACTIONS, MCP_STATUS_TIMEOUT_MS, useMcpControls } from '@/composables/useMcpControls';
import { captureWsFrames } from '../helpers/ws-capture';

let scope: ReturnType<typeof effectScope>;
let capture: ReturnType<typeof captureWsFrames>;

function mountControls() {
  scope = effectScope();
  return scope.run(useMcpControls)!;
}

async function settle() {
  for (let i = 0; i < 8; i++) await Promise.resolve();
  await nextTick();
}

beforeEach(() => {
  jest.useFakeTimers();
  const config = createAppconfig();
  config.hub.mcp.enabled = true;
  store.commit('hub/setAppConfig', config);
  store.commit('hub/setMCPStatus', { running: false, enabled: true, name: 'MCP', version: '1', connectedClients: 0 });
  capture = captureWsFrames();
});

afterEach(() => {
  scope?.stop();
  capture.restore();
  jest.useRealTimers();
});

test('start dispatches, asks for status and reports the frame', async () => {
  const controls = mountControls();
  await controls.run(MCP_ACTIONS.start);
  expect(capture.frames.map((frame) => frame.type)).toEqual(['startMCP', 'loadMCPStatus']);
  expect(controls.pending.value).toBe(MCP_ACTIONS.start);
  expect(jest.getTimerCount()).toBe(1);
});

test('a new status frame clears the pending state', async () => {
  const controls = mountControls();
  await controls.run(MCP_ACTIONS.start);
  store.commit('hub/setMCPStatus', { running: true, enabled: true, name: 'MCP', version: '1', connectedClients: 1 });
  await settle();
  expect(controls.pending.value).toBeNull();
  expect(controls.error.value).toBe('');
  expect(jest.getTimerCount()).toBe(0);
});

test('an unanswered action times out with an explanatory error', async () => {
  const controls = mountControls();
  await controls.run(MCP_ACTIONS.restart);
  jest.advanceTimersByTime(MCP_STATUS_TIMEOUT_MS);
  await settle();
  expect(controls.pending.value).toBeNull();
  expect(controls.error.value).toContain('did not report a new MCP status');
});

test('a late status frame clears the timeout message as well as the pending action', async () => {
  const controls = mountControls();
  await controls.run(MCP_ACTIONS.restart);
  jest.advanceTimersByTime(MCP_STATUS_TIMEOUT_MS);
  await settle();
  expect(controls.error.value).not.toBe('');
  store.commit('hub/setMCPStatus', { running: true, enabled: true, name: 'MCP', version: '1', connectedClients: 1 });
  await settle();
  expect(controls.error.value).toBe('');
});

test('a second action is ignored while one is pending', async () => {
  const controls = mountControls();
  await controls.run(MCP_ACTIONS.start);
  const sent = capture.frames.length;
  await controls.run(MCP_ACTIONS.stop);
  expect(capture.frames).toHaveLength(sent);
  expect(controls.pending.value).toBe(MCP_ACTIONS.start);
});

test('disposing the scope releases the timer and the status watcher', async () => {
  const controls = mountControls();
  await controls.run(MCP_ACTIONS.stop);
  expect(jest.getTimerCount()).toBe(1);
  scope.stop();
  expect(jest.getTimerCount()).toBe(0);
  // A late frame after disposal must not touch released state.
  store.commit('hub/setMCPStatus', { running: true, enabled: true, name: 'MCP', version: '1', connectedClients: 2 });
  await settle();
  expect(controls.error.value).toBe('');
});

test('refresh and enabled reflect the hub configuration', () => {
  const controls = mountControls();
  controls.refresh();
  expect(capture.frames.map((frame) => frame.type)).toEqual(['loadMCPStatus']);
  expect(controls.enabled.value).toBe(true);
  store.commit('hub/setMCPSettings', false);
  expect(controls.enabled.value).toBe(false);
});
