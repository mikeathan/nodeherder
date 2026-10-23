import { historyOptions, statisticsHistoryOptions } from '@/components/chart/history/options';
import type { TooltipItem, TooltipModel } from 'chart.js';
const style = { color: '#ffffff', muted: '#aaaaaa', grid: '#333333', dark: true, compact: false, reducedMotion: false };
test('shared visual policy preserves explicit bounds, monotone curves, steps, gaps and bars', () => {
  const line = historyOptions(style, 0, 100, '°C', false, false);
  expect(line.datasets?.line).toMatchObject({ cubicInterpolationMode: 'monotone', borderWidth: 3, spanGaps: false });
  expect(line.scales?.x).toMatchObject({ min: 0, max: 100 });
  const compact = historyOptions(
    { ...style, compact: true, dark: false, reducedMotion: true },
    0,
    100,
    '%',
    true,
    true
  );
  expect(compact.animation).toBe(false);
  expect(compact.datasets?.line?.stepped).toBe('before');
  expect(compact.datasets?.bar?.borderRadius).toBe(4);
  expect(compact.plugins?.tooltip?.backgroundColor).toBe('#ffffff');
});
test('instantaneous power has straight segments and consumption axes begin at zero', () => {
  expect(historyOptions(style, 0, 100, 'W', false, false, true).datasets?.line?.tension).toBe(0);
  expect(historyOptions(style, 0, 100, 'kWh', false, true).scales?.y).toMatchObject({ beginAtZero: true });
});
test('theme surface/text tokens drive tooltip colours when supplied', () => {
  const themed = historyOptions(
    { ...style, dark: false, surface: 'rgb(20, 20, 20)', text: 'rgb(240, 240, 240)', border: 'rgb(80, 80, 80)' },
    0,
    100,
    '%',
    false,
    false
  );
  expect(themed.plugins?.tooltip).toMatchObject({
    backgroundColor: 'rgb(20, 20, 20)',
    bodyColor: 'rgb(240, 240, 240)',
    borderColor: 'rgb(80, 80, 80)',
  });
});


test('final-sample markers retain contrast and chart guides stay subtle and dashed', () => {
  const chart = historyOptions({ ...style, color: '#26a69a', text: '#fafafa' }, 0, 100, 'ppm', false, false);
  expect(chart.datasets?.line).toMatchObject({ pointBackgroundColor: '#26a69a', pointBorderColor: '#fafafa', pointBorderWidth: 2, clip: 8 });
  expect(chart.scales?.y?.border).toMatchObject({ display: false, dash: [4, 4] });
  const axis = { ticks: [] as { value: number }[] };
  (chart.scales?.x?.afterBuildTicks as (scale: typeof axis) => void)(axis);
  expect(axis.ticks.map((tick) => tick.value)).toEqual([0, 50, 100]);
});
test('statistics parse floating ranges, fit the measured axis and name range/marker tooltips', () => {
  const options = statisticsHistoryOptions(style, 0, 86400000, '°C');
  expect(options.parsing).toEqual({ xAxisKey: 'x', yAxisKey: 'y' });
  expect(options.scales?.y).toMatchObject({ type: 'linear', beginAtZero: false });
  const label = options.plugins!.tooltip!.callbacks!.label!;
  const tooltip = (name: string, y: number | [number, number] | null) => label.call(
    {} as TooltipModel<'line' | 'bar'>,
    { dataset: { label: name }, raw: { x: 0, y } } as TooltipItem<'line' | 'bar'>
  );
  expect(tooltip('Minimum–Maximum', [18, 24])).toBe('Minimum–Maximum: 18 – 24 °C');
  expect(tooltip('Average', 21.25)).toBe('Average: 21.25 °C');
  expect(tooltip('Last', null)).toBe('Last: Unknown');
});
