import { Module } from 'vuex';
import { MetricsModuleState } from './state';
import { RootState } from '@/store/state';
import { resetHistoryRequests } from '@/services/history.service';
import type { HistoryEntry } from '@/domain/history/types';
import { DeviceMetrics, DeviceMetricsMap, DeviceMetricsRequest } from '@/types/metrics.type';

export const MetricsModule: Module<MetricsModuleState, RootState> = {
  namespaced: true,

  state: () => ({
    epoch: 0,
    history: {},
    deviceMetricsQueryMap: {} as DeviceMetricsMap,
  }),

  getters: {
    historyEpoch: (state: MetricsModuleState) => state.epoch,
    history:
      (state: MetricsModuleState) =>
      (key: string): HistoryEntry | undefined =>
        state.history[key],
    view:
      (state: MetricsModuleState) =>
      (id: string): DeviceMetrics | null => {
        if (!state.deviceMetricsQueryMap[id]) {
          return null;
        }
        return state.deviceMetricsQueryMap[id];
      },
  },

  mutations: {
    historySet(state, payload: { key: string; entry: HistoryEntry }) {
      delete state.history[payload.key];
      state.history[payload.key] = payload.entry;
      const keys = Object.keys(state.history);
      if (keys.length > 10) delete state.history[keys[0]];
    },
    set(state: MetricsModuleState, metrics: DeviceMetrics) {
      if (!metrics) {
        console.error('metrics is null');
        return;
      }

      state.deviceMetricsQueryMap[metrics.deviceId] = metrics;
    },
    delete(state: MetricsModuleState, id: string) {
      delete state.deviceMetricsQueryMap[id];
    },
    clear(state: MetricsModuleState) {
      state.epoch++;
      resetHistoryRequests();
      state.history = {};
      Object.entries(state.deviceMetricsQueryMap).forEach(([key, value]) => {
        delete state.deviceMetricsQueryMap[key];
      });
    },
  },
  actions: {
    query({ commit, dispatch, rootState }, request: DeviceMetricsRequest) {
      commit('delete', request.id); // remove any existing results
      dispatch('ws/emit', { event: 'loadMetrics', message: request }, { root: true });
    },
    store({ commit, dispatch, rootState }, metrics: DeviceMetrics) {
      commit('set', metrics);
    },
  },
};
