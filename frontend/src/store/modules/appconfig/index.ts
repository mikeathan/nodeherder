import { Module } from 'vuex';
import { RootState } from '../../state';
import { AppConfigModuleState } from './state';
import {
  AppConfig,
  DeviceSettings,
  DeviceSettingsMap,
  HistorySettings,
} from '@/types/settings';
import { key } from '@/store';

export const AppConfigModule: Module<
  AppConfigModuleState,
  RootState
> = {
  namespaced: true,

  state: () => ({
    deviceSettingsMap: {} as DeviceSettingsMap,
    appConfig: {} as AppConfig,
    initialized: false,
  }),

  getters: {
    initialized:
      (state: AppConfigModuleState) => (): boolean =>
        state.initialized,

    history:
      (state: AppConfigModuleState) =>
      (): HistorySettings =>
        state.appConfig.history,

    findDeviceSetting:
      (state: AppConfigModuleState) =>
      (id: string): DeviceSettings => {
        return state.deviceSettingsMap[id];
      },
  },

  mutations: {
    setDeviceSetting(
      state: AppConfigModuleState,
      deviceSetting: DeviceSettings,
    ) {
      state.deviceSettingsMap[deviceSetting.id] =
        deviceSetting;
    },

    clear(state: AppConfigModuleState) {
      Object.entries(state.deviceSettingsMap).forEach(
        ([key, value]) => {
          delete state.deviceSettingsMap[key];
        },
      );

      state.appConfig = {} as AppConfig;
      state.initialized = false;
    },
  },

  actions: {
    init({ state, commit }, appConfig: AppConfig) {
      commit('clear', state);
      state.appConfig = appConfig;
      Object.values(appConfig.devices).forEach((value) => {
        commit('setDeviceSetting', value);
      });

      state.initialized = true;
    },

    saveDeviceSettings(
      { commit, dispatch, rootState },
      deviceSetting: DeviceSettings,
    ) {
      commit('setDeviceSetting', deviceSetting);
      dispatch(
        'ws/emit',
        {
          event: 'saveDeviceConfig',
          message: deviceSetting,
        },
        { root: true },
      );
    },
  },
};
