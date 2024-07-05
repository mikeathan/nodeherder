import { KeyyValuePair } from './types';

export type DeviceMetrics = {
  deviceId: string;
  expose: DeviceExposeMetrics[];
};

export type DeviceExposeMetrics = {
  name: string;
  type: string;
  timestamp: string[];
  values: any[];
};

export type DeviceMetricsRequest = {
  id: string;
  expose?: string;
  from: number;
  to: number;
};

export type DeviceMetricsMap = KeyyValuePair<DeviceMetrics>;
