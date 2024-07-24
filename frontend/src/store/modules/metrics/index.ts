import { Module } from 'vuex';
import { MetricsModuleState } from './state';
import { RootState } from '@/store/state';
import {
  DeviceMetrics,
  DeviceMetricsMap,
  DeviceMetricsRequest,
} from '@/types/metrics.type';

export const MetricsModule: Module<
  MetricsModuleState,
  RootState
> = {
  namespaced: true,

  state: () => ({
    deviceMetricsQueryMap: {} as DeviceMetricsMap,
  }),

  getters: {
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
    set(state: MetricsModuleState, metrics: DeviceMetrics) {
      if (!metrics) {
        console.error('metrics is null');
        return;
      }

      console.log('metrics', metrics);
      state.deviceMetricsQueryMap[metrics.deviceId] =
        metrics;
    },
    delete(state: MetricsModuleState, id: string) {
      delete state.deviceMetricsQueryMap[id];
    },
    clear(state: MetricsModuleState) {
      Object.entries(state.deviceMetricsQueryMap).forEach(
        ([key, value]) => {
          delete state.deviceMetricsQueryMap[key];
        },
      );
    },
  },
  actions: {
    query(
      { commit, dispatch, rootState },
      request: DeviceMetricsRequest,
    ) {
      commit('delete', request.id); // remove any existing results
      dispatch(
        'ws/emit',
        { event: 'loadMetrics', message: request },
        { root: true },
      );
    },
    store(
      { commit, dispatch, rootState },
      metrics: DeviceMetrics,
    ) {
      commit('set', metrics);
    },
  },
};
