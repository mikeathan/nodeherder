import { ValueOf } from './types';

export type ChartColor = {
  backgroundColor: string;
  borderColor: string;
};

export type AreaDataPoint = {
  x: number;
  y: number;
};

export type AreaChartEntry = {
  name: string;
  data: AreaDataPoint[];
};

export type TimelineDataPoint = {
  x: string;
  y: number[];
};

export type TimelineChartEntry = {
  name: string;
  data: TimelineDataPoint[];
};

export type ChartType = keyof typeof ChartTypes;
export const ChartTypes = {
  AreaChart: 'AreaChart',
  TimelineChart: 'TimelineChart',
  TimeRangeChart: 'TimeRangeChart',
  NumericChart: 'NumericChart',
} as const;

export type PeriodType = ValueOf<typeof PeriodTypes>;
export const PeriodTypes = {
  Today: 'Today',
  OneDay: '1 Day',
  ThreeDays: '3 Days',
  ThisWeek: 'This week',
  LastWeek: 'Last week',
} as const;

export const PeriodOptions = Object.values(PeriodTypes);
