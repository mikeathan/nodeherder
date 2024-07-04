import { Module } from 'vuex';
import { MetricsModuleState } from './state';
import { RootState } from '@/store/state';
import { DeviceMetricsMap } from '@/types/metrics';

export const DeviceModule: Module<
  MetricsModuleState,
  RootState
> = {
  namespaced: true,

  state: () => ({
    deviceMetricseMap: {} as DeviceMetricsMap,
  }),
};
