import { InjectionKey } from 'vue';
import { createStore, useStore as baseUseStore, Store } from 'vuex';
import { RootState } from './state';
import { HubStateModule as hub } from './modules/hub-state/index';
import { DeviceModule as devices } from './modules/device/index';
import { AutomationModule as automations } from './modules/automation/index';
import { WSClientModule as ws } from './modules/ws/index';
import { AppConfigModule as appconfig } from './modules/appconfig/index';
import { MetricsModule as metrics } from './modules/metrics/index';
import { ConsoleModule as console } from './modules/console/index';
import { AlertsModule as alerts } from './modules/alerts/index';

import createPersistedState from 'vuex-persistedstate';

// https://blog.openreplay.com/integrate-vuex-and-typescript/

export const key: InjectionKey<Store<RootState>> = Symbol();

export type StoreType = Store<RootState>;

export const store = createStore<RootState>({
  plugins: [], // createPersistedState() - needs fixing . stopped workig with vuetify ?
  state: {},
  actions: {
    cleanup({ commit }) {
      commit('hub/clear');
      commit('automations/clear');
      commit('metrics/clear');
      commit('alerts/clear');
    },
  },
  modules: {
    devices,
    automations,
    appconfig,
    metrics,
    console,
    alerts,
    hub,
    ws,
  },
});

export function useStore_temp() {
  return baseUseStore(key);
}
