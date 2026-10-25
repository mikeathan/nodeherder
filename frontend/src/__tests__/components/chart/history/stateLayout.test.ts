import {
  FOCUS_CLUSTER_MAX_PIXELS,
  alignedAxisInstants,
  isDenseStateChart,
  stateChartLayout,
  stateChartTicks,
  stateChartX,
  stateFocusIsUseful,
  stateLaneY,
} from '@/components/chart/history/stateLayout';

const measure = (text: string, size: number) => text.length * size * 0.6;

test('wide and narrow layouts keep the documented lane proportions and gutters', () => {
  const wide = stateChartLayout(1_000, 40);
  expect(wide.narrow).toBe(false);
  expect(wide.plotTop).toBe(16);
  expect((wide.activeY - wide.plotTop) / wide.plotHeight).toBeCloseTo(0.3, 5);
  expect((wide.inactiveY - wide.plotTop) / wide.plotHeight).toBeCloseTo(0.7, 5);
  expect(wide.padLeft).toBeGreaterThanOrEqual(64);

  const narrow = stateChartLayout(320, 40);
  expect(narrow.narrow).toBe(true);
  expect((narrow.activeY - narrow.plotTop) / narrow.plotHeight).toBeCloseTo(0.34, 5);
  expect((narrow.inactiveY - narrow.plotTop) / narrow.plotHeight).toBeCloseTo(0.72, 5);
  expect(narrow.fontSize).toBe(10);
  expect(narrow.plotWidth).toBeGreaterThan(150);
});

test('the label gutter grows with longer state names so nothing clips', () => {
  const shortLabels = stateChartLayout(390, 24);
  const longLabels = stateChartLayout(390, 96);
  expect(longLabels.padLeft).toBeGreaterThan(shortLabels.padLeft);
  expect(longLabels.padLeft + longLabels.padRight).toBeLessThan(390);
});

test('degenerate widths still produce a usable plot', () => {
  const layout = stateChartLayout(0, 0);
  expect(layout.width).toBe(1);
  expect(layout.plotWidth).toBe(1);
  expect(stateChartX(0, 0, 10, layout)).toBe(layout.padLeft);
  expect(stateChartX(5, 10, 10, layout)).toBe(layout.padLeft);
});

test('lane lookup and x mapping follow the visible range', () => {
  const layout = stateChartLayout(1_000, 40);
  expect(stateLaneY(true, layout)).toBe(layout.activeY);
  expect(stateLaneY(false, layout)).toBe(layout.inactiveY);
  expect(stateChartX(0, 0, 100, layout)).toBe(layout.padLeft);
  expect(stateChartX(100, 0, 100, layout)).toBe(layout.width - layout.padRight);
  expect(stateChartX(50, 0, 100, layout)).toBeCloseTo(layout.padLeft + layout.plotWidth / 2, 5);
});

test('dense selection scales with plot width and keeps a minimum sparse budget', () => {
  expect(isDenseStateChart(20, 900)).toBe(false);
  expect(isDenseStateChart(45, 120)).toBe(true);
  expect(isDenseStateChart(250, 900)).toBe(true);
  expect(isDenseStateChart(40, 10_000)).toBe(false);
});

test('activity focus is only offered while the cluster is genuinely unreadable', () => {
  const day = 24 * 3_600_000;
  expect(stateFocusIsUseful(0, 4 * 60_000, 0, day, 900)).toBe(true);
  expect(stateFocusIsUseful(0, 6 * 3_600_000, 0, day, 900)).toBe(false);
  expect(stateFocusIsUseful(0, 4 * 60_000, 0, day, 0)).toBe(false);
  expect(stateFocusIsUseful(0, day, 0, day, 900)).toBe(false);
  expect(stateFocusIsUseful(0, 4 * 60_000, 0, 4 * 60_000, 900)).toBe(false);
  const boundary = (FOCUS_CLUSTER_MAX_PIXELS / 900) * day;
  expect(stateFocusIsUseful(0, boundary - 1, 0, day, 900)).toBe(true);
  expect(stateFocusIsUseful(0, boundary + 1, 0, day, 900)).toBe(false);
});

test('axis instants align to clock steps and to local days', () => {
  const base = new Date(2026, 9, 22, 15, 34, 12).getTime();
  const hourTicks = alignedAxisInstants(base, base + 3 * 3_600_000, 3_600_000);
  expect(hourTicks).toHaveLength(3);
  expect(hourTicks.every((tick) => new Date(tick).getMinutes() === 0 && new Date(tick).getSeconds() === 0)).toBe(true);
  expect(alignedAxisInstants(base, base, 3_600_000)).toEqual([]);

  const dayTicks = alignedAxisInstants(base, base + 3 * 86_400_000, 86_400_000);
  expect(dayTicks).toHaveLength(3);
  expect(dayTicks.every((tick) => new Date(tick).getHours() === 0)).toBe(true);
});

test('axis ticks use clock labels for short ranges and dates for multi-day ranges', () => {
  const now = new Date(2026, 9, 22, 15, 34, 12).getTime();
  const layout = stateChartLayout(900, 40);
  const shortTicks = stateChartTicks({ from: now, to: now + 3_600_000, width: layout.width, padLeft: layout.padLeft, padRight: layout.padRight, plotWidth: layout.plotWidth, fontSize: layout.fontSize, measure });
  expect(shortTicks.every((tick) => /\d{1,2}:\d{2}/.test(tick.label))).toBe(true);
  const longTicks = stateChartTicks({ from: now, to: now + 3 * 86_400_000, width: layout.width, padLeft: layout.padLeft, padRight: layout.padRight, plotWidth: layout.plotWidth, fontSize: layout.fontSize, measure });
  expect(longTicks.every((tick) => !/\d{1,2}:\d{2}/.test(tick.label))).toBe(true);
});

test('ticks always include both ends and never collide', () => {
  const from = new Date(2026, 9, 22, 15, 34, 12).getTime();
  const to = from + 24 * 3_600_000;
  for (const width of [320, 390, 768, 1280]) {
    const layout = stateChartLayout(width, 40);
    const ticks = stateChartTicks({ from, to, width: layout.width, padLeft: layout.padLeft, padRight: layout.padRight, plotWidth: layout.plotWidth, fontSize: layout.fontSize, measure });
    expect(ticks[0].timestamp).toBe(from);
    expect(ticks[ticks.length - 1].timestamp).toBe(to);
    let previousRight = Number.NEGATIVE_INFINITY;
    for (const tick of ticks) {
      const textWidth = measure(tick.label, layout.fontSize) + 4;
      const left = tick.anchor === 'start' ? tick.x : tick.anchor === 'end' ? tick.x - textWidth : tick.x - textWidth / 2;
      expect(left).toBeGreaterThanOrEqual(previousRight + 4);
      previousRight = left + textWidth;
    }
  }
});

test('a very short focus range keeps clock labels rather than collapsing to one', () => {
  const from = new Date(2026, 9, 22, 18, 54).getTime();
  const layout = stateChartLayout(390, 40);
  const ticks = stateChartTicks({ from, to: from + 8 * 60_000, width: layout.width, padLeft: layout.padLeft, padRight: layout.padRight, plotWidth: layout.plotWidth, fontSize: layout.fontSize, measure });
  expect(ticks.length).toBeGreaterThanOrEqual(2);
  expect(ticks[0].timestamp).toBe(from);
  expect(ticks[ticks.length - 1].timestamp).toBe(from + 8 * 60_000);
});
