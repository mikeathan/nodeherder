import { Module } from "vuex";
import { RootState } from "../../state";
import { AppConfigModuleState } from "./state";

import { AppConfig, DeviceSettings } from "@/types/settings";

export const AppConfigModule: Module<AppConfigModuleState, RootState> = {
  namespaced: true,

  state: () => ({ appConfig: {} as AppConfig, initialized: false }),

  getters: {
    initialized: (state: AppConfigModuleState) => (): boolean =>
      state.initialized,

    findDeviceSetting:
      (state: AppConfigModuleState) =>
      (id: string): DeviceSettings => {
        return state.appConfig.devices[id];
      },
  },

  mutations: {
    setDeviceSetting(
      state: AppConfigModuleState,
      deviceSetting: DeviceSettings
    ) {
      state.appConfig.devices[deviceSetting.id] = deviceSetting;
    },

    clear(state: AppConfigModuleState) {
      Object.entries(state.appConfig.devices).forEach(([key, value]) => {
        delete state.appConfig.devices[key];
      });
      state.initialized = false;
    },
  },

  //LoadAppconfig
  actions: {
    init({ state, commit }, appConfig: AppConfig) {
      commit("clear", state);

      state.appConfig = appConfig;
      state.initialized = true;
    },

    saveDeviceSettings(
      { commit, dispatch, rootState },
      deviceSetting: DeviceSettings
    ) {
      commit("setDeviceSetting", deviceSetting);
      dispatch(
        "ws/emit",
        { event: "SaveDeviceConfig", message: deviceSetting },
        { root: true }
      );
    },
  },
};
