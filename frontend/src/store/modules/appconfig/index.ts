import { Logger, Module } from 'vuex';
import { RootState } from '../../state';
import { AppConfigModuleState } from './state';
import {
  AppConfig,
  BridgeSettingsType,
  DeviceSettings,
  DeviceSettingsMap,
  HistorySettingsType,
  LoggerSettingsType,
} from '@/types/settings.type';
import { key } from '@/store';

export const AppConfigModule: Module<AppConfigModuleState, RootState> = {
  namespaced: true,

  state: () => ({
    deviceSettingsMap: {} as DeviceSettingsMap,
    appConfig: {} as AppConfig,
    initialized: false,
  }),

  getters: {
    initialized: (state: AppConfigModuleState) => (): boolean => state.initialized,

    history: (state: AppConfigModuleState) => (): HistorySettingsType => state.appConfig.history,
    logger: (state: AppConfigModuleState) => (): LoggerSettingsType => state.appConfig.logger,
    bridge: (state: AppConfigModuleState) => (): BridgeSettingsType => state.appConfig.bridge,
    findDeviceSetting:
      (state: AppConfigModuleState) =>
      (id: string): DeviceSettings => {
        return state.deviceSettingsMap[id];
      },
  },

  mutations: {
    setDeviceSetting(state: AppConfigModuleState, deviceSetting: DeviceSettings) {
      state.deviceSettingsMap[deviceSetting.id] = deviceSetting;
    },

    setHistorySettings(state: AppConfigModuleState, historySetting: HistorySettingsType) {
      state.appConfig.history = historySetting;
    },
    setLoggerSettings(state: AppConfigModuleState, loggerSettings: LoggerSettingsType) {
      state.appConfig.logger = loggerSettings;
    },
    setBridgeSettings(state: AppConfigModuleState, bridgeSettings: BridgeSettingsType) {
      state.appConfig.bridge = bridgeSettings;
    },
    clear(state: AppConfigModuleState) {
      Object.entries(state.deviceSettingsMap).forEach(([key, value]) => {
        delete state.deviceSettingsMap[key];
      });

      state.appConfig = {} as AppConfig;
      state.initialized = false;
    },
  },

  actions: {
    init({ state, commit }, appConfig: AppConfig) {
      commit('clear', state);

      state.appConfig = appConfig;
      console.log(state.appConfig);
      Object.values(appConfig.devices).forEach((value) => {
        commit('setDeviceSetting', value);
      });

      state.initialized = true;
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

    // TO MOVE - ws !!!!!!!!!!!!!!!!!!!
    enablePermitJoin({ commit, dispatch }, timeout: number) {
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
    // TO MOVE - ws !!!!!!!!!!!!!!!!!!!
    disablePermitJoin({ commit, dispatch }) {
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
  },
};
