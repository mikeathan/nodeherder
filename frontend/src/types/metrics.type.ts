import { KeyValuePair } from './types.type';

export const MetricsTypes = {
  Binary: 'binary',
  Numeric: 'numeric',
  Enum: 'enum',
};

export type DeviceMetrics = {
  deviceId: string;
  exposes: DeviceExposeMetrics[];
};

export type DeviceExposeMetrics = {
  name: string;
  type: string;
  from?: string;
  to?: string;
};

export type NumericDataPoint = {
  x: number;
  y: number;
};

export type BinaryDataPoint = {
  timestamp: number;
  value: string;
};

export type DeviceExposeNumericMetrics = {
  name: string;
  type: string;
  from?: string;
  to?: string;
  data: NumericDataPoint[];
};

export type DeviceExposeBinaryMetrics = {
  name: string;
  type: string;
  from?: string;
  to?: string;
  data: BinaryDataPoint[];
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
