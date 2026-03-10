import { Module } from 'vuex';
import { RootState } from '../../state';
import { HubStateModuleState } from './state';
import { Device, Devices, DeviceMap, DeviceUpdate, Expose } from '../../../types/device';
import {
  AppConfig,
  BridgeSettingsType,
  DashboardGroup,
  DashboardGroups,
  DeviceConfig,
  HistorySettingsType,
  AssistantSettingsType,
  LoggerSettingsType,
  MCPStatusType,
} from '../../../types/settings.type';
import { KeyValuePair } from '@/types/types.type';
import { createAppconfig } from '@/contracts/settings';

export const HubStateModule: Module<HubStateModuleState, RootState> = {
  namespaced: true,

  state: () => ({
    deviceMap: {} as DeviceMap,
    appConfig: createAppconfig(),
    initialized: false,
    mcpStatus: null as MCPStatusType | null,
  }),

  getters: {
    isInitialized: (state) => (): boolean => {
      return state.initialized;
    },

    // Device getters
    listAllDevices: (state) => (): Devices => {
      return Object.values(state.deviceMap) as Devices;
    },
    findDevice:
      (state) =>
      (id: string): Device => {
        return state.deviceMap[id];
      },
    deviceExists:
      (state) =>
      (id: string): boolean => {
        return state.deviceMap[id] != null;
      },

    deviceDefaults: (state) => (): DeviceConfig => {
      return state.appConfig?.hub.devices?.defaults;
    },

    mcpConfig: (state) => (): boolean => {
      return state.appConfig?.hub.mcp?.enabled;
    },

    // AppConfig getters
    history: (state) => (): HistorySettingsType => {
      return state.appConfig.hub.history;
    },
    assistant: (state) => (): AssistantSettingsType => state.appConfig.hub.assistant,
    logger: (state) => (): LoggerSettingsType => state.appConfig.hub.logger,
    bridge: (state) => (): BridgeSettingsType => state.appConfig.bridge,
    dashboardGroups: (state) => (): DashboardGroups => state.appConfig.hub.dashboardGroups,
    mcpStatus: (state) => (): MCPStatusType | null => state.mcpStatus,
    findDeviceSetting:
      (state) =>
      (id: string): DeviceConfig | undefined => {
        var override = state.appConfig?.hub.devices?.overrides?.[id];
        var defaults = state.appConfig?.hub.devices?.defaults;
        return override ?? defaults;
      },
    hasDeviceConfigOverride:
      (state) =>
      (id: string): boolean => {
        return state.appConfig?.hub.devices?.overrides?.[id] != null;
      },
  },

  mutations: {
    // Device mutations
    addDevice(state, device: Device) {
      state.deviceMap[device.id] = device;
    },
    setDevices(state, devices: Devices) {
      // clear the device map
      Object.entries(state.deviceMap).forEach(([key, value]) => {
        delete state.deviceMap[key];
      });

      devices.forEach((device: Device) => {
        state.deviceMap[device.id] = device;
      });
    },

    updateDevice(state, deviceUpdate: DeviceUpdate) {
      const current = state.deviceMap[deviceUpdate.id];
      if (!current) {
        console.error('device ', deviceUpdate.id, ' not found');
        return;
      }
      // Create a shallow cloned device object to ensure Vue reactivity triggers
      let changed = false;
      const updated = { ...current };

      // Clone exposes map only if needed
      const newExposes = { ...current.exposes };

      for (const key in deviceUpdate.data) {
        const incoming = deviceUpdate.data[key];
        const expose = newExposes[key];
        if (!expose) continue;

        if (expose.data !== incoming) {
          newExposes[key] = { ...expose, data: incoming };
          changed = true;
        }
      }

      if (changed) {
        updated.exposes = newExposes;
      }

      if (deviceUpdate.last_seen && deviceUpdate.last_seen !== current.last_seen) {
        updated.last_seen = deviceUpdate.last_seen;
        changed = true;
      }

      if (deviceUpdate.availability && deviceUpdate.availability !== current.availability) {
        updated.availability = deviceUpdate.availability;
        changed = true;
      }

      if (changed) {
        state.deviceMap[deviceUpdate.id] = updated;
      }
    },
    // AppConfig mutations
    setMCPSettings(state, enabled: boolean) {
      if (state.appConfig.hub.mcp) {
        state.appConfig.hub.mcp.enabled = enabled;
      }
    },
    setAppConfig(state, config: AppConfig) {
      state.appConfig = config;
    },
    setDeviceDeConfigfaults(state, defaults: DeviceConfig) {
      state.appConfig.hub.devices.defaults = defaults;
    },
    removeDeviceConfigOverride(state, id: string) {
      delete state.appConfig.hub.devices.overrides[id];
    },
    setDeviceConfigOverride(state, setting: DeviceConfig) {
      state.appConfig.hub.devices.overrides[setting.id] = setting;
    },
    setDashboardGroups(state, dashboardGroups: DashboardGroups) {
      state.appConfig.hub.dashboardGroups = dashboardGroups;
    },
    setDashboardGroup(state, dashboardGroup: DashboardGroup) {
      state.appConfig.hub.dashboardGroups[dashboardGroup.name] = dashboardGroup;
    },
    removeDashboardGroup(state, name: string) {
      delete state.appConfig.hub.dashboardGroups[name];
    },
    setHistorySettings(state, historySetting: HistorySettingsType) {
      state.appConfig.hub.history = historySetting;
    },
    setAssistantSettings(state, assistantSettings: AssistantSettingsType) {
      state.appConfig.hub.assistant.url = assistantSettings.url;
    },
    setLoggerSettings(state, loggerSettings: LoggerSettingsType) {
      state.appConfig.hub.logger = loggerSettings;
    },
    setBridgeSettings(state, bridgeSettings: BridgeSettingsType) {
      state.appConfig.bridge = bridgeSettings;
    },
    setMCPStatus(state, mcpStatus: MCPStatusType) {
      state.mcpStatus = mcpStatus;
    },
    clear(state) {
      Object.entries(state.deviceMap).forEach(([key, value]) => {
        delete state.deviceMap[key];
      });

      state.appConfig = {} as AppConfig;
      state.initialized = false;
      state.mcpStatus = null;
    },

    setInitialized(state, initialized: boolean) {
      state.initialized = initialized;
    },
  },

  actions: {
    init({ commit }, payload: { devices: Devices; config: AppConfig }) {
      commit('clear');
      commit('setAppConfig', payload.config);
      commit('setDevices', payload.devices);
      commit('setInitialized', true);
    },

    // AppConfig actions
    setAppConfig({ state, commit }, appConfig: AppConfig) {
      commit('setAppConfig', appConfig);
    },

    saveDeviceConfigOverride({ commit, dispatch }, deviceSetting: DeviceConfig) {
      commit('setDeviceConfigOverride', deviceSetting);
      dispatch(
        'ws/emit',
        {
          event: 'saveDeviceConfigOverride',
          message: deviceSetting,
        },
        { root: true }
      );
    },
    deleteDeviceConfigOverride({ commit, dispatch }, id: string) {
      commit('removeDeviceConfigOverride', id);

      var payload = {
        id: id,
      };
      dispatch(
        'ws/emit',
        {
          event: 'deleteDeviceConfigOverride',
          message: payload,
        },
        { root: true }
      );
    },

    saveDeviceConfigDefaults({ commit, dispatch }, config: DeviceConfig) {
      commit('setDeviceDeConfigfaults', config);
      dispatch(
        'ws/emit',
        {
          event: 'saveDeviceConfigDefaults',
          message: config,
        },
        { root: true }
      );
    },

    renameDashboardGroup({ commit, dispatch }, { oldName, newName }: { oldName: string; newName: string }) {
      commit('renameDashboardGroup', { oldName, newName });
      dispatch(
        'ws/emit',
        {
          event: 'renameDashboardGroup',
          message: { oldName, newName },
        },
        { root: true }
      );
    },

    saveDashboardGroup({ commit, dispatch }, dashboardGroup: DashboardGroup) {
      commit('setDashboardGroup', dashboardGroup);
      dispatch(
        'ws/emit',
        {
          event: 'saveDashboardGroup',
          message: dashboardGroup,
        },
        { root: true }
      );
    },

    importDashboardGroups({ commit, dispatch }, dashboardGroups: DashboardGroups) {
      dispatch(
        'ws/emit',
        {
          event: 'importDashboardGroups',
          message: dashboardGroups,
        },
        { root: true }
      );
    },

    deleteDashboardGroup({ commit, dispatch }, name: string) {
      commit('removeDashboardGroup', name);

      var payload = {
        groupName: name,
      };
      dispatch(
        'ws/emit',
        {
          event: 'deleteDashboardGroup',
          message: payload,
        },
        { root: true }
      );
    },

    saveHistorySettings({ commit, dispatch }, historySettings: HistorySettingsType) {
      commit('setHistorySettings', historySettings);
      dispatch(
        'ws/emit',
        {
          event: 'saveHistoryConfig',
          message: historySettings,
        },
        { root: true }
      );
    },
    saveAssistantSettings({ commit, dispatch }, assistantSettings: AssistantSettingsType) {
      commit('setAssistantSettings', assistantSettings);
      dispatch(
        'ws/emit',
        {
          event: 'saveAssistantConfig',
          message: assistantSettings,
        },
        { root: true }
      );
    },
    saveLoggerSettings({ commit, dispatch }, loggerSetings: LoggerSettingsType) {
      commit('setLoggerSettings', loggerSetings);
      dispatch(
        'ws/emit',
        {
          event: 'saveLoggerConfig',
          message: loggerSetings,
        },
        { root: true }
      );
    },
    loadMCPStatus({ dispatch }) {
      dispatch('ws/emit', { event: 'loadMCPStatus', message: {} }, { root: true });
    },
    restartMCP({ dispatch }) {
      dispatch('ws/emit', { event: 'restartMCP', message: {} }, { root: true });
    },
    stopMCP({ commit, dispatch }) {
      commit('setMCPSettings', false);
      dispatch('ws/emit', { event: 'stopMCP', message: {} }, { root: true });
    },
    startMCP({ commit, dispatch }) {
      commit('setMCPSettings', true);
      dispatch('ws/emit', { event: 'startMCP', message: {} }, { root: true });
    },
    enablePermitJoin({ dispatch }, timeout: number) {
      const cfg: BridgeSettingsType = {
        permitJoin: true,
        maxTimeAllowed: { value: timeout, unit: 'seconds' },
      };
      dispatch(
        'ws/emit',
        {
          event: 'bridgePermitJoin',
          message: cfg,
        },
        { root: true }
      );
    },

    disablePermitJoin({ dispatch }) {
      const cfg: BridgeSettingsType = {
        permitJoin: false,
        maxTimeAllowed: { value: 0, unit: 'seconds' },
      };

      dispatch(
        'ws/emit',
        {
          event: 'bridgePermitJoin',
          message: cfg,
        },
        { root: true }
      );
    },

    // Device actions
    setDevices({ commit, dispatch }, devices: Devices) {
      commit('setDevices', devices);
    },

    setDeviceValue({ dispatch }, payload: KeyValuePair<any>) {
      dispatch('ws/emit', { event: 'deviceSetValue', message: payload }, { root: true });
    },

    renameDevice({ dispatch }, { name, newName }) {
      var payload = {
        from: name,
        to: newName,
      };

      dispatch('ws/emit', { event: 'deviceRename', message: payload }, { root: true });
    },

    interviewDevice({ dispatch }, { id }) {
      dispatch('ws/emit', { event: 'deviceInterview', message: { id } }, { root: true });
    },

    removeDevice({ dispatch }, { id, force, block }) {
      var payload = {
        id: id,
        force: force,
        block: block,
      };
      dispatch('ws/emit', { event: 'deviceRemove', message: payload }, { root: true });
    },
  },
};
