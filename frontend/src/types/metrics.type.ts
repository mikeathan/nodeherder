import { KeyValuePair } from './types.type';

export const MetricsTypes = {
  Binary: 'binary',
  Numeric: 'numeric',
  Enum: 'enum',
} as const;

export type MetricsType = (typeof MetricsTypes)[keyof typeof MetricsTypes];

export type MiniChartComponentKey =
  | 'MiniEnergyChart'
  | 'MiniPercentChart'
  | 'MiniRealtimeChart'
  | 'MiniNumericChart'
  | 'MiniBinaryChart';

export interface BinaryRange {
  value: string;
  start: number;
  end: number;
}
export type NumericStats = {
  min: number;
  max: number;
  avg: number;
  first: number;
  last: number;
  delta: number;
  deltaPct: number;
};

export type NumericChartHeaderData = DeviceExposeNumericMetrics & {
  label: string;
  unit: string | null;
  color: string;
  stats: NumericStats | null;
};
export interface RangeBarDataPoint {
  x: string;
  y: [number, number];
  fillColor: string;
}

export type DeviceMetrics = {
  deviceId: string;
  exposes: DeviceExposeMetrics[];
};

export type DeviceExposeMetrics = {
  name: string;
  type: MetricsType;
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
  type: MetricsType;
  from?: string;
  to?: string;
  data: NumericDataPoint[];
};

export type DeviceExposeBinaryMetrics = {
  name: string;
  type: MetricsType;
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
