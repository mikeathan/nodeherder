import { Module } from 'vuex';
import { RootState } from '../../state';
import { HubStateModuleState } from './state';
import { Device, Devices, DeviceMap, DeviceUpdate } from '../../../types/device';
import {
  AppConfig,
  BridgeSettingsType,
  DeviceSettings,
  HistorySettingsType,
  LoggerSettingsType,
} from '../../../types/settings';
import { KeyValuePair } from '@/types/types';

export const HubStateModule: Module<HubStateModuleState, RootState> = {
  namespaced: true,

  state: () => ({
    deviceMap: {} as DeviceMap,
    appConfig: {} as AppConfig,
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
    history: (state) => (): HistorySettingsType => state.appConfig.history,
    logger: (state) => (): LoggerSettingsType => state.appConfig.logger,
    bridge: (state) => (): BridgeSettingsType => state.appConfig.bridge,
    findDeviceSetting:
      (state) =>
      (id: string): DeviceSettings | undefined => {
        return state.appConfig?.devices[id];
      },
  },

  mutations: {
    // Device mutations
    addDevice(state, device: Device) {
      state.deviceMap[device.id] = device;
    },
    updateDevices(state, devices: Devices) {
      devices.forEach((device: Device) => {
        if (device.id in state.deviceMap) state.deviceMap[device.id] = device;
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
      for (var key in deviceUpdate.properties) {
        if (key in device.properties) {
          device.properties[key] = deviceUpdate.properties[key];
        }
      }
      device.properties.last_seen = deviceUpdate.last_seen;
    },

    // AppConfig mutations
    setAppConfig(state, config: AppConfig) {
      state.appConfig = config;
      state.initialized = true;
    },
    setDeviceSetting(state, setting: DeviceSettings) {
      if (state.appConfig) {
        state.appConfig.devices[setting.id] = setting;
      }
    },
    setHistorySettings(state, historySetting: HistorySettingsType) {
      state.appConfig.history = historySetting;
    },
    setLoggerSettings(state, loggerSettings: LoggerSettingsType) {
      state.appConfig.logger = loggerSettings;
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
  },

  actions: {
    init({ commit }, payload: { devices: Devices; config: AppConfig }) {
      commit('clear');
      payload.devices.forEach((device) => commit('addDevice', device));
      commit('setAppConfig', payload.config);
    },

    // AppConfig actions
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
