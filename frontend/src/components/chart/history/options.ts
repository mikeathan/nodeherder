import type { ChartOptions, ScriptableContext } from 'chart.js';
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
};
export type HistoryChartOptions = ChartOptions<'line' | 'bar'>;
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
function lineDataset(style: ChartAppearance, step: boolean, linear: boolean): NonNullable<Datasets['line']> {
  return {
    borderColor: style.color,
    borderWidth: 3,
    borderCapStyle: 'round',
    pointRadius: 0,
    pointHoverRadius: 4,
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
        items.length && items[0].parsed.x !== null ? new Date(items[0].parsed.x).toLocaleString() : '',
      label: (item) => (item.parsed.y === null ? 'Unknown' : `${Number(item.parsed.y.toFixed(2))} ${unit}`),
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
    ticks: {
      color: style.muted,
      maxTicksLimit: style.compact ? 3 : 4,
      maxRotation: 0,
      callback: (value) => historyTime(Number(value), to - from),
    },
  };
}
function yScale(style: ChartAppearance, unit: string, bars: boolean): NonNullable<Scales['y']> {
  return {
    display: !style.compact,
    beginAtZero: bars,
    grace: '10%',
    border: { display: false, dash: [4, 4] },
    grid: { color: style.grid, drawTicks: false },
    ticks: {
      color: style.muted,
      maxTicksLimit: 3,
      padding: 8,
      callback: (value) => `${Number(Number(value).toFixed(1))} ${unit}`,
    },
  };
}
function historyTime(value: number, duration: number): string {
  const date = new Date(value);
  return duration > 2 * 86400000
    ? date.toLocaleDateString([], { month: 'short', day: 'numeric' })
    : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
function historyGradient({ chart }: ScriptableContext<'line'>, color: string) {
  const area = chart.chartArea;
  if (!area) return 'transparent';
  const gradient = chart.ctx.createLinearGradient(0, area.top, 0, area.bottom);
  gradient.addColorStop(0, `${color}40`);
  gradient.addColorStop(1, `${color}00`);
  return gradient;
}
