import { Module } from 'vuex';
import { RootState } from '../../state';
import { HubStateModuleState } from './state';
import { Device, Devices, DeviceMap, DeviceUpdate } from '../../../types/device';
import {
  AppConfig,
  BridgeSettingsType,
  DashboardGroups,
  DeviceSettings,
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

    isInitialized: (state) => (): boolean => {
      return state.initialized;
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
      (id: string): DeviceSettings | undefined => {
        return state.appConfig?.hub.devices[id];
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
    setDeviceSetting(state, setting: DeviceSettings) {
      if (state.appConfig) {
        state.appConfig.hub.devices[setting.id] = setting;
      }
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

    saveDeviceSettings({ commit, dispatch }, deviceSetting: DeviceSettings) {
      commit('setDeviceSetting', deviceSetting);
      dispatch(
        'ws/emit',
        {
          event: 'saveDeviceConfig',
          message: deviceSetting,
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
