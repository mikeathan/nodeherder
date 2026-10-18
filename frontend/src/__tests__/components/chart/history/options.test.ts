import { historyOptions } from '@/components/chart/history/options';
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
