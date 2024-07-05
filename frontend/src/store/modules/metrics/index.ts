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

  getters: {},

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
      dispatch(
        'ws/emit',
        { event: 'loadMetrics', message: request },
        { root: true },
      );
    },
  },
};
