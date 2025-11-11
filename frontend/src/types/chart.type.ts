import { ColorValue } from './color.type';
import { KeyValuePair, ValueOf } from './types.type';

export type AreaDataPoint = {
  x: number;
  y: number;
};

export type AreaChartEntry = {
  name: string;
  color: ColorValue;
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
  OneHour: '1 Hour',
  SixHours: '6 Hours',
  TwelveHours: '12 Hours',
  Today: 'Today',
  OneDay: '1 Day',
  ThreeDays: '3 Days',
  SevenDays: '7 Days',
  ThisWeek: 'This Week',
  ThirtyDays: '30 Days',
  LastWeek: 'Last Week',
} as const;

export const PeriodOptions = Object.values(PeriodTypes);
