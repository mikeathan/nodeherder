import { describe, expect, test, beforeEach, afterEach, jest } from '@jest/globals';
import { store } from '@/store/index';
import { default as hubState } from '../../../../../../docs/hub_state.json';
import { captureWsFrames, SentFrame } from '../../../helpers/ws-capture';
import { DashboardGroup, DeviceConfig } from '@/types/settings.type';
import { createAppconfig } from '@/contracts/settings';

// Characterization of the WebSocket commands the hub store sends. The redesigned UI must
// keep producing exactly these payloads (spec 007, FR-03).
const payload = (hubState as { payload: { devices: unknown[]; config: unknown } }).payload;

describe('hub store commands (wire contract)', () => {
  let capture: { frames: SentFrame[]; restore: () => void };

  beforeEach(() => {
    store.dispatch('hub/init', JSON.parse(JSON.stringify({ devices: payload.devices, config: payload.config })));
    capture = captureWsFrames();
    jest.spyOn(console, 'error').mockImplementation(() => {});
  });
  afterEach(() => {
    capture.restore();
    jest.restoreAllMocks();
  });

  test('setDeviceValue sends deviceSetValue with id, name and value', () => {
    store.dispatch('hub/setDeviceValue', { id: '0x70ac08fffefafeca', name: 'state', value: 'ON' });
    expect(capture.frames).toEqual([{ type: 'deviceSetValue', payload: { id: '0x70ac08fffefafeca', name: 'state', value: 'ON' } }]);
  });

  test('renameDevice sends deviceRename from/to friendly names', () => {
    store.dispatch('hub/renameDevice', { name: 'Attic Light', newName: 'Attic lamp' });
    expect(capture.frames).toEqual([{ type: 'deviceRename', payload: { from: 'Attic Light', to: 'Attic lamp' } }]);
  });

  test('interviewDevice and removeDevice send their requests', () => {
    store.dispatch('hub/interviewDevice', { id: 'x1' });
    store.dispatch('hub/removeDevice', { id: 'x1', force: true, block: false });
    expect(capture.frames).toEqual([
      { type: 'deviceInterview', payload: { id: 'x1' } },
      { type: 'deviceRemove', payload: { id: 'x1', force: true, block: false } },
    ]);
  });

  test('permit join enable/disable sends bridgePermitJoin', () => {
    store.dispatch('hub/enablePermitJoin', 120);
    store.dispatch('hub/disablePermitJoin');
    expect(capture.frames).toEqual([
      { type: 'bridgePermitJoin', payload: { permitJoin: true, maxTimeAllowed: { value: 120, unit: 'seconds' } } },
      { type: 'bridgePermitJoin', payload: { permitJoin: false, maxTimeAllowed: { value: 0, unit: 'seconds' } } },
    ]);
  });

  test('saveDashboardGroup stores the group locally and sends it', () => {
    const group: DashboardGroup = { name: 'Garage', deviceGroup: { d1: { deviceId: 'd1', exposes: ['temperature'] } } };
    store.dispatch('hub/saveDashboardGroup', group);
    expect(store.getters['hub/dashboardGroups']().Garage).toEqual(group);
    expect(capture.frames).toEqual([{ type: 'saveDashboardGroup', payload: group }]);
  });

  test('deleteDashboardGroup removes locally and sends groupName', () => {
    store.dispatch('hub/deleteDashboardGroup', 'Kitchen');
    expect(store.getters['hub/dashboardGroups']().Kitchen).toBeUndefined();
    expect(capture.frames).toEqual([{ type: 'deleteDashboardGroup', payload: { groupName: 'Kitchen' } }]);
  });

  test('renameDashboardGroup renames locally (keeping content) and sends old/new names', () => {
    const before = JSON.parse(JSON.stringify(store.getters['hub/dashboardGroups']().Kitchen));
    store.dispatch('hub/renameDashboardGroup', { oldName: 'Kitchen', newName: 'Galley' });
    const groups = store.getters['hub/dashboardGroups']();
    expect(groups.Kitchen).toBeUndefined();
    expect(groups.Galley).toEqual({ ...before, name: 'Galley' });
    expect(capture.frames).toEqual([{ type: 'renameDashboardGroup', payload: { oldName: 'Kitchen', newName: 'Galley' } }]);
  });

  test('importDashboardGroups sends the map unchanged', () => {
    const groups = { A: { name: 'A', deviceGroup: {} } };
    store.dispatch('hub/importDashboardGroups', groups);
    expect(capture.frames).toEqual([{ type: 'importDashboardGroups', payload: groups }]);
  });

  test('saveDeviceConfigDefaults updates local defaults and sends them', () => {
    const defaults: DeviceConfig = { ...createAppconfig().hub.devices.defaults, id: 'defaults', metricsEnabled: false, disabled: false };
    store.dispatch('hub/saveDeviceConfigDefaults', defaults);
    expect(store.getters['hub/deviceDefaults']()).toEqual(defaults);
    expect(capture.frames).toEqual([{ type: 'saveDeviceConfigDefaults', payload: defaults }]);
  });

  test('device config override save/delete', () => {
    const cfg: DeviceConfig = { id: 'd9', disabled: true, metricsEnabled: true, defaultDebounceByCategory: {}, debounceOverrides: {} };
    store.dispatch('hub/saveDeviceConfigOverride', cfg);
    expect(store.getters['hub/hasDeviceConfigOverride']('d9')).toBe(true);
    store.dispatch('hub/deleteDeviceConfigOverride', 'd9');
    expect(store.getters['hub/hasDeviceConfigOverride']('d9')).toBe(false);
    expect(capture.frames.map((f) => f.type)).toEqual(['saveDeviceConfigOverride', 'deleteDeviceConfigOverride']);
    expect(capture.frames[1].payload).toEqual({ id: 'd9' });
  });

  test('settings and MCP commands', () => {
    store.dispatch('hub/saveLoggerSettings', { enableRemoteLogger: true, level: 'debug' });
    store.dispatch('hub/saveHistorySettings', { sleepTimeout: { value: 1, unit: 'minutes' }, expireAt: { value: 2, unit: 'days' } });
    store.dispatch('hub/saveAssistantSettings', { url: 'http://a' });
    store.dispatch('hub/startMCP');
    store.dispatch('hub/stopMCP');
    store.dispatch('hub/restartMCP');
    store.dispatch('hub/loadMCPStatus');
    expect(capture.frames.map((f) => f.type)).toEqual(['saveLoggerConfig', 'saveHistoryConfig', 'saveAssistantConfig', 'startMCP', 'stopMCP', 'restartMCP', 'loadMCPStatus']);
  });
});

describe('automation store commands (wire contract)', () => {
  let capture: { frames: SentFrame[]; restore: () => void };
  beforeEach(() => {
    store.commit('automations/clear');
    capture = captureWsFrames();
  });
  afterEach(() => capture.restore());

  test('save commits and sends saveAutomation with the automation unchanged', () => {
    const a = { id: 'a1', friendlyname: 'A', type: 'device', description: '', enabled: true, schedules: [], triggers: [] };
    store.dispatch('automations/save', a);
    expect(store.getters['automations/find']('a1')).toEqual(a);
    expect(capture.frames).toEqual([{ type: 'saveAutomation', payload: a }]);
  });

  test('delete removes locally and sends deleteAutomation id', () => {
    store.commit('automations/add', { id: 'a2', triggers: [] });
    store.dispatch('automations/delete', 'a2');
    expect(store.getters['automations/find']('a2')).toBeUndefined();
    expect(capture.frames).toEqual([{ type: 'deleteAutomation', payload: { id: 'a2' } }]);
  });
});
