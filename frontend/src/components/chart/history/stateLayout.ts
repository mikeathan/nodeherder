/*
 * Pure geometry and axis policy for the binary state history chart. Kept out of the
 * component so lane positions, tick collision and the activity-focus thresholds can be
 * unit tested without mounting Vue or touching the DOM.
 */

export type StateChartLayout = {
  width: number;
  narrow: boolean;
  fontSize: number;
  height: number;
  padLeft: number;
  padRight: number;
  plotTop: number;
  plotBottom: number;
  plotWidth: number;
  plotHeight: number;
  activeY: number;
  inactiveY: number;
};

import { formatHistoryAxisLabel } from '@/domain/history/format';

export type StateAxisTick = {
  timestamp: number;
  x: number;
  label: string;
  anchor: 'start' | 'middle' | 'end';
};

export type StateAxisInput = {
  from: number;
  to: number;
  width: number;
  padLeft: number;
  padRight: number;
  plotWidth: number;
  fontSize: number;
  /** Injected so this module stays DOM-free; the component measures with a canvas. */
  measure: (text: string, fontSize: number) => number;
};

/** Below this width the chart switches to narrower gutters, lanes and labels. */
export const NARROW_CHART_WIDTH = 560;
/** Sparse mode is kept for at least this many intervals regardless of width. */
export const DENSE_MIN_SEGMENTS = 40;
/** More than one transition per this many pixels becomes an unreadable comb. */
export const DENSE_PIXELS_PER_SEGMENT = 4;
/** A focus action is only useful while the activity cluster stays this narrow. */
export const FOCUS_CLUSTER_MAX_PIXELS = 50;

const LANES = {
  narrow: { active: 0.34, inactive: 0.72 },
  wide: { active: 0.3, inactive: 0.7 },
};

const AXIS_STEPS = [
  30_000, 60_000, 2 * 60_000, 5 * 60_000, 10 * 60_000, 15 * 60_000, 30 * 60_000,
  60 * 60_000, 2 * 3_600_000, 3 * 3_600_000, 4 * 3_600_000, 6 * 3_600_000,
  8 * 3_600_000, 12 * 3_600_000, 24 * 3_600_000, 2 * 86_400_000,
  3 * 86_400_000, 7 * 86_400_000, 14 * 86_400_000, 30 * 86_400_000,
];

export function stateChartLayout(width: number, labelWidth: number): StateChartLayout {
  const measured = Math.max(1, width);
  const narrow = measured < NARROW_CHART_WIDTH;
  const fontSize = narrow ? 10 : 11;
  const padLeft = Math.max(narrow ? 56 : 64, Math.ceil(labelWidth + 28));
  const padRight = narrow ? 10 : 16;
  const height = narrow ? 164 : 200;
  const plotTop = narrow ? 12 : 16;
  const plotBottom = height - (narrow ? 30 : 34);
  const plotWidth = Math.max(1, measured - padLeft - padRight);
  const plotHeight = plotBottom - plotTop;
  const lanes = narrow ? LANES.narrow : LANES.wide;
  return {
    width: measured,
    narrow,
    fontSize,
    height,
    padLeft,
    padRight,
    plotTop,
    plotBottom,
    plotWidth,
    plotHeight,
    activeY: plotTop + plotHeight * lanes.active,
    inactiveY: plotTop + plotHeight * lanes.inactive,
  };
}

export function stateChartX(timestamp: number, from: number, to: number, layout: StateChartLayout): number {
  const span = to - from;
  if (span <= 0) return layout.padLeft;
  return layout.padLeft + ((timestamp - from) / span) * layout.plotWidth;
}

export function stateLaneY(active: boolean, layout: StateChartLayout): number {
  return active ? layout.activeY : layout.inactiveY;
}

export function isDenseStateChart(segmentCount: number, plotWidth: number): boolean {
  return segmentCount > Math.max(DENSE_MIN_SEGMENTS, plotWidth / DENSE_PIXELS_PER_SEGMENT);
}

/**
 * Offer the activity focus only when the cluster is narrower than the readable
 * threshold and would actually zoom in. `plotWidth <= 0` means "not measured yet".
 */
export function stateFocusIsUseful(
  clusterFrom: number,
  clusterTo: number,
  from: number,
  to: number,
  plotWidth: number
): boolean {
  const span = to - from;
  const cluster = clusterTo - clusterFrom;
  if (plotWidth <= 0 || span <= 0 || cluster <= 0 || cluster >= span) return false;
  return (cluster / span) * plotWidth < FOCUS_CLUSTER_MAX_PIXELS;
}

/** Fixed-width steps keep labels on clock/hour/day boundaries instead of raw offsets. */
export function alignedAxisInstants(from: number, to: number, step: number): number[] {
  if (step <= 0 || to <= from) return [];
  if (step < 86_400_000) {
    const ticks: number[] = [];
    for (let tick = Math.floor(from / step + 1) * step; tick < to; tick += step) ticks.push(tick);
    return ticks;
  }
  const days = Math.round(step / 86_400_000);
  const ticks: number[] = [];
  const cursor = new Date(from);
  cursor.setHours(0, 0, 0, 0);
  cursor.setDate(cursor.getDate() + days);
  while (cursor.getTime() < to) {
    if (cursor.getTime() > from) ticks.push(cursor.getTime());
    cursor.setDate(cursor.getDate() + days);
  }
  return ticks;
}

/**
 * Choose the finest step whose labels fit without colliding. Both ends are always
 * labelled, so the visible range stays readable even on a very narrow phone plot.
 */
export function stateChartTicks(input: StateAxisInput): StateAxisTick[] {
  const { from, to, width, padLeft, padRight, plotWidth, fontSize, measure } = input;
  if (to <= from) return [];
  const range = to - from;
  const maxLabels = Math.min(8, Math.max(2, Math.floor(plotWidth / 70)));
  const build = (timestamps: number[]): StateAxisTick[] =>
    timestamps.map((timestamp, index) => ({
      timestamp,
      x: padLeft + ((timestamp - from) / range) * plotWidth,
      label: formatHistoryAxisLabel(timestamp, range, index === 0 || index === timestamps.length - 1),
      anchor: index === 0 ? 'start' as const : index === timestamps.length - 1 ? 'end' as const : 'middle' as const,
    }));

  for (const step of AXIS_STEPS) {
    const values = [from, ...alignedAxisInstants(from, to, step), to];
    if (values.length > maxLabels) continue;
    const candidate = build(values);
    let previousRight = Number.NEGATIVE_INFINITY;
    const fits = candidate.every((tick) => {
      const tickWidth = measure(tick.label, fontSize) + 4;
      const left = tick.anchor === 'start' ? tick.x : tick.anchor === 'end' ? tick.x - tickWidth : tick.x - tickWidth / 2;
      const right = left + tickWidth;
      if (left < previousRight + 4) return false;
      previousRight = right;
      return true;
    });
    if (fits) return candidate;
  }
  return build([from, to]).map((tick, index) => ({
    ...tick,
    x: index === 0 ? padLeft : width - padRight,
  }));
}
