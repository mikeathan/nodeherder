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

// ── Numeric Semantic Chart Colors ──

const SEMANTIC_NUMERIC_COLORS: KeyValuePair<ColorValue> = {
  // Temperature (Warm)
  temperature: ColorTypes.Coral,
  device_temperature: ColorTypes.Red500,
  local_temperature: ColorTypes.Orange,

  // Humidity/Water (Cool/Blue)
  humidity: ColorTypes.SkyBlue,
  water_leak: ColorTypes.AzureBlue,
  soil_moisture: ColorTypes.Turquoise,

  // Light/Illuminance (Yellow/Amber)
  illuminance: ColorTypes.Amber,
  illuminance_lux: ColorTypes.Yellow,
  brightness: ColorTypes.Amber,
  color_temp: ColorTypes.Yellow,

  // Power/Energy (Purple/Neon)
  power: ColorTypes.ElectricViolet,
  energy: ColorTypes.Purple,
  voltage: ColorTypes.Pink,
  current: ColorTypes.SkyBlue,

  // Health/Battery (Green)
  battery: ColorTypes.Green500,
  battpercentage: ColorTypes.Green400,
  linkquality: ColorTypes.PineGreen,

  // Air Quality/Pressure
  pressure: ColorTypes.Turquoise,
  co2: ColorTypes.SlateGrey,
  voc: ColorTypes.LightBrown,

  // Danger/Critical
  smoke_concentration: ColorTypes.Red,
};

// Fallback palette: vibrant, highly visible colors only (no grays/whites/dull slates)
const VIBRANT_FALLBACK_COLORS = [
  ColorTypes.SkyBlue,
  ColorTypes.SpringGreen,
  ColorTypes.Coral,
  ColorTypes.Yellow,
  ColorTypes.ElectricViolet,
  ColorTypes.Turquoise,
  ColorTypes.Pink,
  ColorTypes.Amber,
  ColorTypes.BrightGreen,
  ColorTypes.Blue,
];

export const getExposeColor = (exposeName: string): ColorValue => {
  // 1. Check for a specific semantic color mapping
  const normalized = (exposeName || '').toLowerCase();

  for (const [key, color] of Object.entries(SEMANTIC_NUMERIC_COLORS)) {
    if (normalized.includes(key)) {
      return color;
    }
  }

  // 2. Fallback to a consistent vibrant color based on its position in the allowed list
  let index = alllowedExposeList.indexOf(exposeName);
  if (index === -1) {
    // String hash for entirely unknown custom exposes to ensure consistent coloring
    index = exposeName.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
  }

  return VIBRANT_FALLBACK_COLORS[index % VIBRANT_FALLBACK_COLORS.length];
};

export const ExposeBinaryColours: KeyValuePair<ExposeBinaryColor> = {
  presence: {
    on: ColorTypes.Green500,
    off: ColorTypes.Slate500,
  },
  contact: {
    on: ColorTypes.Green500,
    off: ColorTypes.Slate500,
  },
  state: {
    on: ColorTypes.Green500,
    off: ColorTypes.Slate500,
  },
  tamper: {
    on: ColorTypes.Red500,
    off: ColorTypes.Slate500,
  },
};

const DANGER_KEYWORDS = ['smoke', 'alarm', 'tamper', 'leak', 'gas', 'carbon_monoxide'];
const WARNING_KEYWORDS = ['silence', 'override', 'child_lock', 'bypass'];

export const getExposeBinaryColour = (exposeName: string): ExposeBinaryColor => {
  if (ExposeBinaryColours[exposeName]) {
    return ExposeBinaryColours[exposeName];
  }

  const normalizedName = (exposeName || '').toLowerCase();

  if (DANGER_KEYWORDS.some((kw) => normalizedName.includes(kw))) {
    return { on: ColorTypes.Red500, off: ColorTypes.Slate500 };
  }

  if (WARNING_KEYWORDS.some((kw) => normalizedName.includes(kw))) {
    return { on: ColorTypes.Amber, off: ColorTypes.Slate500 };
  }

  return {
    on: ColorTypes.Green500,
    off: ColorTypes.Slate500,
  };
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
