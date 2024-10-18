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
import { ConsoleModule as console } from './modules/console/index';
import createPersistedState from 'vuex-persistedstate';

// https://blog.openreplay.com/integrate-vuex-and-typescript/

export const key: InjectionKey<Store<RootState>> = Symbol();

export type StoreType = Store<RootState>;

export const store = createStore<RootState>({
  plugins: [createPersistedState()],
  state: {},
  actions: {
    cleanup({ commit }) {
      commit('devices/clear');
      commit('automations/clear');
      commit('appconfig/clear');
      commit('metrics/clear');
      // commit('console/clear');
    },
  },
  modules: {
    devices,
    automations,
    appconfig,
    metrics,
    console,
    ws,
  },
});

export function useStore_temp() {
  return baseUseStore(key);
}
