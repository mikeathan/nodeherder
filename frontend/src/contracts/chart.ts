import {
  ChartColor,
  PeriodType,
  PeriodTypes,
} from '@/types/chart.type';
import { AlllowedExposeList } from '@/types/device.type';
import {
  getDateRange,
  getLastWeekStartEndDate,
  getWeekStartEndDate,
} from '@/utils/date.utils';
import { KeyValuePair } from '@/types/types';

const HOURS = 24;

export const chartColors: ChartColor[] = buildColors(20);

export const getPeriodOffset = (
  period: PeriodType,
): { from: Date; to: Date } => {
  const now = new Date();
  switch (period) {
    case PeriodTypes.Today:
      return getDateRange(-(HOURS - now.getHours()));
    case PeriodTypes.OneDay:
      return getDateRange(-HOURS);
    case PeriodTypes.ThreeDays:
      return getDateRange(-(HOURS * 3));
    case PeriodTypes.ThisWeek:
      return getWeekStartEndDate();
    case PeriodTypes.LastWeek:
      return getLastWeekStartEndDate();
    default:
      return { from: now, to: now };
  }
};

export function dynamicColors() {
  var r = Math.floor(Math.random() * 255);
  var g = Math.floor(Math.random() * 255);
  var b = Math.floor(Math.random() * 255);

  return {
    backgroundColor: 'rgb(' + r + ',' + g + ',' + b + ')',
    borderColor:
      'rgba(' + r + ',' + g + ',' + b + ',' + 0.5 + ')',
  };
}

function buildColors(n: number): ChartColor[] {
  let chartColors: ChartColor[] = [];
  for (let i = 0; i < n; i++) {
    chartColors[i] = dynamicColors();
  }
  return chartColors;
}

const colors: string[] = [
  'red',
  'green',
  'blue',
  'orange',
  'purple',
  'yellow',
  'pink',
  'cyan',
  'lightblue',
  'lime',
  'darkgreen',
  'darkblue',
  'darkorange',
  'darkred',
  'darkgreen',
  'darkblue',
  'darkorange',
  'darkred',
  'darkgreen',
  'darkblue',
  'darkorange',
  'darkred',
  'darkgreen',
  'darkblue',
  'darkorange',
];

const buildExposeColors = (): KeyValuePair<string> => {
  const exposeColors: KeyValuePair<string> = {};
  for (let i = 0; i < AlllowedExposeList.length; i++) {
    exposeColors[AlllowedExposeList[i]] =
      colors[i % colors.length];
  }
  return exposeColors;
};

const exposeColors: KeyValuePair<string> =
  buildExposeColors();

export const getExposeColor = (
  exposeName: string,
): string => {
  return exposeColors[exposeName] ?? 'blue';
};
