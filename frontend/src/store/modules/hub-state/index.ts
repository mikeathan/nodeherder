import { Module } from 'vuex';
import { RootState } from '../../state';
import { HubStateModuleState } from './state';
import { Device, Devices, DeviceMap, DeviceUpdate } from '../../../types/device';
import {
  AppConfig,
  BridgeSettingsType,
  DashboardGroup,
  DashboardGroups,
  DeviceConfig,
  HistorySettingsType,
  LoggerSettingsType,
} from '../../../types/settings.type';
import { KeyValuePair } from '@/types/types.type';
import { createAppconfig } from '@/contracts/settings';

export const HubStateModule: Module<HubStateModuleState, RootState> = {
  namespaced: true,

  state: () => ({
    deviceMap: {} as DeviceMap,
    appConfig: createAppconfig(),
    initialized: false,
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

    // AppConfig getters
    history: (state) => (): HistorySettingsType => {
      return state.appConfig.hub.history;
    },
    logger: (state) => (): LoggerSettingsType => state.appConfig.hub.logger,
    bridge: (state) => (): BridgeSettingsType => state.appConfig.bridge,
    dashboardGroups: (state) => (): DashboardGroups => state.appConfig.hub.dashboardGroups,
    findDeviceSetting:
      (state) =>
      (id: string): DeviceConfig | undefined => {
        return state.appConfig?.hub.devices?.overrides[id];
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
        //if (device.id in state.deviceMap)
        state.deviceMap[device.id] = device;
      });
    },

    updateDevice(state, deviceUpdate: DeviceUpdate) {
      if (deviceUpdate.id in state.deviceMap == false) {
        console.error('device ', deviceUpdate.id, ' not found');
        return;
      }
      var device = state.deviceMap[deviceUpdate.id];
      for (var key in deviceUpdate.data) {
        if (key in device.exposes) {
          device.exposes[key].data = deviceUpdate.data[key];
        }
      }
      device.last_seen = deviceUpdate.last_seen;
      if (deviceUpdate.availability) {
        device.availability = deviceUpdate.availability;
      }
    },

    // AppConfig mutations
    setAppConfig(state, config: AppConfig) {
      state.appConfig = config;
    },
    setDeviceDeConfigfaults(state, defaults: DeviceConfig) {
      state.appConfig.hub.devices.defaults = defaults;
    },
    removeDeviceConfigOverrides(state, id: string) {
      if (state.appConfig) {
        delete state.appConfig.hub.devices.overrides[id];
      }
    },
    setDeviceConfigOverrides(state, setting: DeviceConfig) {
      if (state.appConfig) {
        state.appConfig.hub.devices.overrides[setting.id] = setting;
      }
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
    setLoggerSettings(state, loggerSettings: LoggerSettingsType) {
      state.appConfig.hub.logger = loggerSettings;
    },
    setBridgeSettings(state, bridgeSettings: BridgeSettingsType) {
      state.appConfig.bridge = bridgeSettings;
    },
    clear(state) {
      Object.entries(state.deviceMap).forEach(([key, value]) => {
        delete state.deviceMap[key];
      });

      state.appConfig = {} as AppConfig;
      state.initialized = false;
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

    saveDeviceConfigOverrides({ commit, dispatch }, deviceSetting: DeviceConfig) {
      commit('setDeviceConfigOverrides', deviceSetting);
      dispatch(
        'ws/emit',
        {
          event: 'saveDeviceConfigOverrides',
          message: deviceSetting,
        },
        { root: true }
      );
    },
    deleteDeviceConfigOverrides({ commit, dispatch }, id: string) {
      commit('removeDeviceConfigOverrides', id);

      var payload = {
        id: id,
      };
      dispatch(
        'ws/emit',
        {
          event: 'deleteDeviceConfigOverrides',
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
