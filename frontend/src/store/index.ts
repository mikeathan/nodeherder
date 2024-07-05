import { InjectionKey } from 'vue';
import {
  createStore,
  useStore as baseUseStore,
  Store,
} from 'vuex';
import { RootState } from './state';
import { DeviceModule as devices } from './modules/device/index';
import { AutomationModule as automations } from './modules/automation/index';
import { WSClientModule as ws } from './modules/ws/index';
import { AppConfigModule as appconfig } from './modules/appconfig/index';
import { MetricsModule as metrics } from './modules/metrics/index';

// https://blog.openreplay.com/integrate-vuex-and-typescript/

export const key: InjectionKey<Store<RootState>> = Symbol();

export const store = createStore<RootState>({
  state: {},
  actions: {
    cleanup({ commit }) {
      commit('devices/clear');
      commit('automations/clear');
      commit('appconfig/clear');
      commit('metrics/clear');
    },
  },
  modules: {
    devices,
    automations,
    appconfig,
    metrics,
    ws,
  },
});

export function useStore_temp() {
  return baseUseStore(key);
}
