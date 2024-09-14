import {
  PeriodType,
  PeriodTypes,
} from '@/types/chart.type';
import {
  alllowedExposeList,
  ExposeBinaryColor,
} from '@/types/device.type';
import {
  getDateRange,
  getLastWeekStartEndDate,
  getWeekStartEndDate,
} from '@/utils/date.utils';
import { KeyValuePair } from '@/types/types';
import { ColorTypes, ColorValue } from '@/types/color.type';

const HOURS = 24;

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

const buildExposeColors = (): KeyValuePair<string> => {
  const colors = Object.values(ColorTypes);
  const exposeColors: KeyValuePair<string> = {};
  for (let i = 0; i < alllowedExposeList.length; i++) {
    exposeColors[alllowedExposeList[i]] =
      colors[i % colors.length];
  }
  return exposeColors;
};

const exposeColors: KeyValuePair<string> =
  buildExposeColors();

export const getExposeColor = (
  exposeName: string,
): ColorValue => {
  return (
    exposeColors[exposeName] ?? Object.values(ColorTypes)[0]
  );
};

export const ExposeBinaryColours: KeyValuePair<ExposeBinaryColor> =
  {
    presence: {
      on: ColorTypes.SkyBlue,
      off: ColorTypes.Grey,
    },
    state: { on: ColorTypes.Yellow, off: ColorTypes.Grey },
    tamper: { on: ColorTypes.Red, off: ColorTypes.SkyBlue },
  };

export const getExposeBinaryColour = (
  exposeName: string,
): ExposeBinaryColor => {
  return (
    ExposeBinaryColours[exposeName] ?? {
      on: ColorTypes.Blue,
      off: ColorTypes.Grey,
    }
  );
};
