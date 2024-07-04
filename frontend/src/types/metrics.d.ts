import { KeyyValuePair } from './types';

export type DeviceMetrics = {
  id: string;
  expose?: string;
  from: number;
  to: number;
};

export type DeviceMetricsMap = KeyyValuePair<DeviceMetrics>;
