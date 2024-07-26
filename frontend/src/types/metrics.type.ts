import { KeyValuePair } from './types';

export const MetricsTypes = {
  Integer: 'integer',
  Float: 'float32',
  Enums: 'enum',
};

export type DeviceMetrics = {
  deviceId: string;
  expose: DeviceExposeMetrics[];
};

export type DeviceExposeMetrics = {
  name: string;
  type: string;
  from?: string; // not implemented yet
  to?: string; // not implemented yet
  timestamp: string[];
  values: any[];
};

export type DeviceMetricsRequest = {
  id: string;
  expose?: string;
  from: number;
  to: number;
};

export type DeviceMetricsQuery = {
  request: DeviceMetricsRequest;
  results: DeviceMetrics;
};
export type DeviceMetricsMap = KeyValuePair<DeviceMetrics>;
