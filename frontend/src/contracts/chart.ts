import { ChartTypes, PeriodType, PeriodTypes } from '@/types/chart.type';
import { alllowedExposeList, ExposeBinaryColor } from '@/types/device.type';
import {
  getDateRange,
  getLastWeekStartEndDate,
  getWeekStartEndDate,
  getTodayRange,
  getYesterdayRange,
} from '@/utils/date.utils';
import { KeyValuePair } from '@/types/types.type';
import { ColorTypes, ColorValue } from '@/types/color.type';

const HOURS = 24;

export const getPeriodOffset = (period: PeriodType): { from: Date; to: Date } => {
  const now = new Date();
  switch (period) {
    case PeriodTypes.OneHour:
      return getDateRange(-1);
    case PeriodTypes.SixHours:
      return getDateRange(-6);
    case PeriodTypes.TwelveHours:
      return getDateRange(-12);
    case PeriodTypes.Today:
      return getTodayRange();
    case PeriodTypes.Yesterday:
      return getYesterdayRange();
    case PeriodTypes.OneDay:
      return getDateRange(-HOURS);
    case PeriodTypes.ThreeDays:
      return getDateRange(-(HOURS * 3));
    case PeriodTypes.SevenDays:
      return getDateRange(-(HOURS * 7));
    case PeriodTypes.ThirtyDays:
      return getDateRange(-(HOURS * 30));
    case PeriodTypes.ThisWeek:
      return getWeekStartEndDate();
    case PeriodTypes.LastWeek:
      return getLastWeekStartEndDate();
    default:
      return { from: now, to: now };
  }
};

const buildExposeColors = (): KeyValuePair<string> => {
  const colors = Object.values(ColorTypes);
  const exposeColors: KeyValuePair<string> = {};
  for (let i = 0; i < alllowedExposeList.length; i++) {
    exposeColors[alllowedExposeList[i]] = colors[i % colors.length];
  }
  return exposeColors;
};

const exposeColors: KeyValuePair<string> = buildExposeColors();

export const getExposeColor = (exposeName: string): ColorValue => {
  return exposeColors[exposeName] ?? Object.values(ColorTypes)[0];
};

export const ExposeBinaryColours: KeyValuePair<ExposeBinaryColor> = {
  presence: {
    on: ColorTypes.Green500,
    off: ColorTypes.Slate700,
  },
  contact: {
    on: ColorTypes.Green500,
    off: ColorTypes.Slate700,
  },
  state: {
    on: ColorTypes.Green500,
    off: ColorTypes.Slate700,
  },
  tamper: {
    on: ColorTypes.Red500,
    off: ColorTypes.Slate700,
  },
};

export const getExposeBinaryColour = (exposeName: string): ExposeBinaryColor => {
  return (
    ExposeBinaryColours[exposeName] ?? {
      on: ColorTypes.Green500,
      off: ColorTypes.Slate700,
    }
  );
};

export function resolveChartOptions(chartType: string, extra?: Record<string, any>) {
  switch (chartType) {
    case ChartTypes.AreaChart:
    case ChartTypes.NumericChart:
      return {
        chart: {
          type: 'area',
          background: 'transparent',
          foreColor: '#ccc',
          toolbar: { show: false },
          zoom: { enabled: true, type: 'x', autoScaleYaxis: true },
          ...(extra?.chart ?? {}),
        },
        stroke: { curve: 'smooth', width: 2, ...(extra?.stroke ?? {}) },
        fill: { type: 'gradient', ...(extra?.fill ?? {}) },
        dataLabels: { enabled: false, ...(extra?.dataLabels ?? {}) },
        legend: { show: false, ...(extra?.legend ?? {}) },
        grid: {
          borderColor: 'rgba(255,255,255,0.12)',
          padding: { left: 0, right: 0, top: 0, bottom: 0 },
          ...(extra?.grid ?? {}),
        },
        markers: { ...(extra?.markers ?? {}) },
        xaxis: {
          type: 'datetime',
          labels: { style: { colors: '#b8c1cc' }, ...(extra?.xaxis?.labels ?? {}) },
          tooltip: { enabled: false },
          ...(extra?.xaxis ?? {}),
        },
        yaxis: {
          labels: { style: { colors: '#b8c1cc' }, ...(extra?.yaxis?.labels ?? {}) },
          ...(extra?.yaxis ?? {}),
        },
        tooltip: { theme: 'dark', ...(extra?.tooltip ?? {}) },
        colors: extra?.colors ?? ['#4FC3F7'],
      };
    case ChartTypes.BinaryChart:
      return {
        chart: {
          type: 'rangeBar',
          background: 'transparent',
          foreColor: '#ccc',
          toolbar: { show: false },
          zoom: { enabled: false },
          ...(extra?.chart ?? {}),
        },
        plotOptions: {
          bar: {
            horizontal: true,
            barHeight: '70%',
            borderRadius: 6,
            rangeBarGroupRows: true,
            distributed: false,
            ...(extra?.plotOptions?.bar ?? {}),
          },
        },
        xaxis: {
          type: 'datetime',
          labels: {
            datetimeUTC: false,
            style: { colors: '#ccc', fontSize: '11px' },
            ...(extra?.xaxis?.labels ?? {}),
          },
          ...(extra?.xaxis ?? {}),
        },
        yaxis: {
          labels: {
            style: { colors: '#ccc', fontSize: '12px' },
            ...(extra?.yaxis?.labels ?? {}),
          },
          ...(extra?.yaxis ?? {}),
        },
        tooltip: extra?.tooltip ?? { theme: 'dark' },
        grid: {
          borderColor: 'rgba(255,255,255,0.12)',
          padding: { left: 0, right: 0, top: 0, bottom: 0 },
          ...(extra?.grid ?? {}),
        },
        legend: { show: false },
      };

    default:
      throw new Error(`Unknown chart type: ${chartType}`);
  }
}
