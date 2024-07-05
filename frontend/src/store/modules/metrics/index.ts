import { Module } from 'vuex';
import { MetricsModuleState } from './state';
import { RootState } from '@/store/state';
import {
  DeviceMetrics,
  DeviceMetricsQueryMap,
  DeviceMetricsRequest,
} from '@/types/metrics';

export const MetricsModule: Module<
  MetricsModuleState,
  RootState
> = {
  namespaced: true,

  state: () => ({
    deviceMetricsQueryMap: {} as DeviceMetricsQueryMap,
  }),

  getters: {
    view(
      state: MetricsModuleState,
      id: string,
    ): DeviceMetrics | null {
      if (!state.deviceMetricsQueryMap[id]) {
        return null;
      }
      return state.deviceMetricsQueryMap[id].results;
    },
  },

  mutations: {
    // addRequest(
    //   state: MetricsModuleState,
    //   request: DeviceMetricsRequest,
    // ) {
    //   state.deviceMetricsQueryMap[request.id].request =
    //     request;
    // },
    store(
      state: MetricsModuleState,
      metrics: DeviceMetrics,
    ) {
      state.deviceMetricsQueryMap[
        metrics.deviceId
      ].results = metrics;
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
  },
};
