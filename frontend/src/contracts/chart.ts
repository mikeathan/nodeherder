import { ChartTypes, PeriodType, PeriodTypes } from '@/types/chart.type';
import { alllowedExposeList, ExposeBinaryColor } from '@/types/device.type';
import { getDateRange, getLastWeekStartEndDate, getWeekStartEndDate } from '@/utils/date.utils';
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
      return getDateRange(-(HOURS - now.getHours()));
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

export function dynamicColors() {
  var r = Math.floor(Math.random() * 255);
  var g = Math.floor(Math.random() * 255);
  var b = Math.floor(Math.random() * 255);

  return {
    backgroundColor: 'rgb(' + r + ',' + g + ',' + b + ')',
    borderColor: 'rgba(' + r + ',' + g + ',' + b + ',' + 0.5 + ')',
  };
}

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
    on: ColorTypes.EmeraldGreen,
    off: ColorTypes.SlateGrey,
  },
  state: {
    on: ColorTypes.EmeraldGreen,
    off: ColorTypes.SlateGrey,
  },
  tamper: {
    on: ColorTypes.Red,
    off: ColorTypes.SlateGrey,
  },
};

export const getExposeBinaryColour = (exposeName: string): ExposeBinaryColor => {
  return (
    ExposeBinaryColours[exposeName] ?? {
      on: ColorTypes.EmeraldGreen,
      off: ColorTypes.SlateGrey,
    }
  );
};

export function resolveChartOptions(chartType: string, extra?: Record<string, any>) {
  switch (chartType) {
    case ChartTypes.TimelineChart:
      return {
        chart: {
          type: 'rangeBar',
          background: 'transparent',
          foreColor: '#ccc',
          toolbar: {
            show: false,
          },
          zoom: { enabled: false, type: 'x' },
          width: '100%',
          height: '100%',
          animations: { enabled: false },
          parentHeightOffset: 0,
          offsetX: 0,
          ...(extra?.chart ?? {}),
        },
        grid: {
          borderColor: 'rgba(255,255,255,0.15)',
          padding: {
            left: 0,
            right: 0,
            top: 0,
            bottom: 0,
          },
        },
        plotOptions: {
          bar: {
            horizontal: true,
            barHeight: '60%',
            rangeBarGroupRows: true,
          },
        },
        stroke: { width: 0 },
        fill: { type: 'solid', opacity: 0.7 },
        legend: { show: false },
        xaxis: {
          type: 'datetime',
          labels: {
            style: { colors: '#ccc', fontSize: '11px' },
            datetimeFormatter: { day: 'dd MMM', hour: 'HH:mm', minute: 'HH:mm' },
          },
          ...(extra?.xaxis ?? {}),
        },
        yaxis: {
          title: { text: undefined },
          labels: { show: true },
          show: false,
        },
        tooltip: extra?.tooltip ?? {},
      };

    case ChartTypes.AreaChart:
    case ChartTypes.NumericChart:
      return {
        chart: {
          type: 'area',
          background: 'transparent',
          foreColor: '#ccc',
          toolbar: { show: false, autoselected: 'pan' },
          zoom: { enabled: true, type: 'x', autoScaleYaxis: true },
          ...(extra?.chart ?? {}),
        },
        stroke: { curve: 'smooth', width: 2, ...(extra?.stroke ?? {}) },
        fill: {
          type: 'gradient',
          gradient: {
            shadeIntensity: 1,
            inverseColors: false,
            opacityFrom: 0.6,
            opacityTo: 0,
            stops: [0, 100],
          },
        },
        dataLabels: { enabled: false },
        legend: { showForSingleSeries: true, position: 'top' },
        grid: { borderColor: 'rgba(255,255,255,0.15)' },
        xaxis: {
          type: 'datetime',
          labels: {
            datetimeUTC: false,
            style: { colors: '#ccc' },
            datetimeFormatter: {
              year: 'yyyy',
              month: "MMM 'yy",
              day: 'dd MMM',
              hour: 'HH:mm',
            },
          },
        },
        yaxis: {
          decimalsInFloat: 1,
          labels: {
            style: { colors: '#ccc' },
            formatter: (value: number) => {
              return value !== null ? value.toFixed(1) : '';
            },
          },
        },
        tooltip: {
          theme: 'dark',
          shared: false,
          followCursor: false,
          x: {
            format: 'dd MMM yyyy HH:mm:ss',
            formatter: function (value: number) {
              const date = new Date(value);
              const now = new Date();
              const diffMs = now.getTime() - date.getTime();
              const diffMins = Math.floor(diffMs / 60000);
              const diffHours = Math.floor(diffMs / 3600000);
              const diffDays = Math.floor(diffMs / 86400000);

              const timeStr = date.toLocaleTimeString('en-GB', {
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
              });

              if (diffMins < 60) {
                return `${diffMins} min ago (${timeStr})`;
              } else if (diffHours < 24) {
                return `${diffHours}h ago (${timeStr})`;
              } else if (diffDays === 1) {
                return `Yesterday ${timeStr}`;
              } else if (diffDays < 7) {
                return `${diffDays} days ago (${timeStr})`;
              }

              const dateStr = date.toLocaleDateString('en-GB', {
                day: '2-digit',
                month: 'short',
                year: 'numeric',
              });
              return `${dateStr} ${timeStr}`;
            },
          },
          y: {
            formatter: (value: number) => {
              return value !== null ? value.toFixed(2) : '';
            },
          },
          ...(extra?.tooltip ?? {}),
        },
        colors: extra?.colors ?? ['#4FC3F7'],
      };

    case ChartTypes.TimeRangeChart:
      return {
        chart: {
          type: 'line',
          background: 'transparent',
          foreColor: '#ccc',
          toolbar: { show: false },
          ...(extra?.chart ?? {}),
        },
        stroke: { width: 2 },
        markers: { size: 3 },
        grid: { borderColor: 'rgba(255,255,255,0.15)' },
        xaxis: { type: 'datetime', labels: { style: { colors: '#ccc' } } },
        colors: extra?.colors ?? ['#FFB300'],
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
          },
        },
        yaxis: {
          labels: {
            style: { colors: '#ccc', fontSize: '12px' },
          },
        },
        tooltip: extra?.tooltip ?? { theme: 'dark', x: { format: 'dd MMM HH:mm' } },
        grid: { borderColor: 'rgba(255,255,255,0.15)' },
        legend: { show: false },
      };

    default:
      throw new Error(`Unknown chart type: ${chartType}`);
  }
}
