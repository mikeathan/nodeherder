import type { ChartOptions, ScriptableContext } from 'chart.js';
import { formatHistoryAxisLabel, formatHistoryMoment } from '@/domain/history/format';
import { formatWithUnit } from '@/utils/units';
export type ChartAppearance = {
  color: string;
  muted: string;
  grid: string;
  /** Theme surfaces; optional for callers with only the legacy appearance. */
  surface?: string;
  text?: string;
  border?: string;
  dark: boolean;
  compact: boolean;
  reducedMotion: boolean;
  fontFamily?: string;
  /**
   * Distinct series colours for charts that plot several measurements at once (the statistics
   * view). Optional so legacy appearance literals remain valid; `color` is the single-series
   * fallback.
   */
  series?: { power: string; temp: string; contact: string };
};
export type HistoryChartOptions = ChartOptions<'line' | 'bar'>;
export type HistoryChartPoint = { x: number; y: number | [number, number] | null };
export type HistoryChartSeries = {
  name: string;
  data: HistoryChartPoint[];
  type?: 'line' | 'bar';
  color?: string;
  outline?: string;
  pointStyle?: 'circle' | 'rectRot';
};
type Datasets = NonNullable<HistoryChartOptions['datasets']>;
type Scales = NonNullable<HistoryChartOptions['scales']>;
type Tooltip = NonNullable<NonNullable<HistoryChartOptions['plugins']>['tooltip']>;

/** One visual policy for compact/full history; transformations remain in the domain. */
export function historyOptions(
  style: ChartAppearance,
  from: number,
  to: number,
  unit: string,
  step: boolean,
  bars: boolean,
  linear = false
): HistoryChartOptions {
  return {
    responsive: true,
    maintainAspectRatio: false,
    parsing: false,
    animation: chartAnimation(style),
    color: style.muted,
    font: { family: style.fontFamily ?? 'Figtree, system-ui, sans-serif', size: 11 },
    interaction: { mode: 'nearest', axis: 'x', intersect: false },
    layout: { padding: { top: 8, right: 8 } },
    datasets: { line: lineDataset(style, step, linear), bar: barDataset(style) },
    plugins: { legend: { display: false }, tooltip: tooltip(style, unit) },
    scales: { x: xScale(style, from, to), y: yScale(style, unit, bars) },
  };
}
function chartAnimation(style: ChartAppearance): HistoryChartOptions['animation'] {
  return style.reducedMotion || style.compact ? false : { duration: 180 };
}

/** Floating ranges encode spread; mean/last markers use the same measured-value axis. */
export function statisticsHistoryOptions(style: ChartAppearance, from: number, to: number, unit: string): HistoryChartOptions {
  const options = historyOptions(style, from, to, unit, false, true);
  options.parsing = { xAxisKey: 'x', yAxisKey: 'y' };
  if (options.scales) options.scales.y = yScale(style, unit, false);
  if (options.scales?.x?.ticks) {
    options.scales.x.ticks.callback = (value) => new Date(Number(value)).toLocaleDateString([], { month: 'short', day: 'numeric' });
  }
  if (options.plugins?.tooltip?.callbacks) {
    // ONE formatter for every tooltip in the family; an inline copy here is what made the same
    // sample print differently between the two charts.
    options.plugins.tooltip.callbacks.label = (item) =>
      `${item.dataset.label ?? ''}: ${tooltipValue((item.raw as HistoryChartPoint).y, unit)}`;
  }
  return options;
}

/**
 * One tooltip value, formatted by the shared units formatter.
 *
 * The chart previously re-implemented rounding, spacing and the missing-value sentinel in three
 * places, so the same sample read "Unknown" here and "—" elsewhere (and a unitless series left a
 * trailing space). A spread range carries its unit once, at the end, rather than repeating it on
 * both bounds.
 */
function tooltipValue(value: HistoryChartPoint['y'], unit: string): string {
  if (Array.isArray(value)) {
    const [low, high] = value;
    return `${formatWithUnit(low, '', { digits: 2 })} – ${formatWithUnit(high, unit, { digits: 2 })}`;
  }
  return formatWithUnit(value, unit, { digits: 2 });
}
function lineDataset(style: ChartAppearance, step: boolean, linear: boolean): NonNullable<Datasets['line']> {
  return {
    borderColor: style.color,
    borderWidth: 3,
    borderCapStyle: 'round',
    pointRadius: 0,
    pointHoverRadius: 4,
    pointBackgroundColor: style.color,
    pointBorderColor: style.text || style.muted,
    pointBorderWidth: 2,
    // A sample on the range boundary still needs its complete marker visible.
    clip: 8,
    stepped: step ? 'before' : false,
    cubicInterpolationMode: linear ? 'default' : 'monotone',
    tension: linear ? 0 : 0.3,
    spanGaps: false,
    fill: true,
    backgroundColor: (context: ScriptableContext<'line'>) => historyGradient(context, style.color),
  };
}
function barDataset(style: ChartAppearance): NonNullable<Datasets['bar']> {
  return { backgroundColor: style.color, borderRadius: 4, borderSkipped: false, barPercentage: 0.65 };
}
function tooltip(style: ChartAppearance, unit: string): Tooltip {
  return {
    ...tooltipColors(style),
    displayColors: false,
    callbacks: {
      title: (items) =>
        items.length && items[0].parsed.x !== null ? formatHistoryMoment(items[0].parsed.x) : '',
      label: (item) => tooltipValue(item.parsed.y, unit),
    },
  };
}
function tooltipColors(style: ChartAppearance) {
  const fallback = style.dark
    ? { background: '#242529', text: '#f5f5f5' }
    : { background: '#ffffff', text: '#1a1a1a' };
  return {
    backgroundColor: style.surface || fallback.background,
    bodyColor: style.text || fallback.text,
    borderColor: style.border || 'transparent',
    borderWidth: 1,
  };
}
function xScale(style: ChartAppearance, from: number, to: number): NonNullable<Scales['x']> {
  return {
    type: 'linear',
    min: from,
    max: to,
    border: { display: false },
    grid: { display: false },
    // Numeric millisecond ticks otherwise produce irregular intervals and often
    // only two labels. These instants match the requested history range exactly.
    afterBuildTicks: (axis) => {
      axis.ticks = [{ value: from }, { value: (from + to) / 2 }, { value: to }];
    },
    ticks: {
      color: style.muted,
      maxTicksLimit: 3,
      autoSkip: false,
      maxRotation: 0,
      callback: (value) => formatHistoryAxisLabel(Number(value), to - from, false),
    },
  };
}
function yScale(style: ChartAppearance, unit: string, bars: boolean): NonNullable<Scales['y']> {
  return {
    type: 'linear',
    display: !style.compact,
    beginAtZero: bars,
    grace: '10%',
    border: { display: false, dash: [4, 4] },
    grid: { color: style.grid, drawTicks: false },
    ticks: {
      color: style.muted,
      maxTicksLimit: 3,
      padding: 8,
      // Same formatter as the tooltip: axis and tooltip must space the unit identically.
      callback: (value) => formatWithUnit(Number(value), unit, { digits: 1 }),
    },
  };
}
function historyGradient({ chart }: ScriptableContext<'line'>, color: string) {
  const area = chart.chartArea;
  if (!area) return 'transparent';
  const gradient = chart.ctx.createLinearGradient(0, area.top, 0, area.bottom);
  gradient.addColorStop(0, `${color}40`);
  gradient.addColorStop(1, `${color}00`);
  return gradient;
}
