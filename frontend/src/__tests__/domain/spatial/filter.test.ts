import { createSpatialFilter, SPATIAL_PREDICT_MS, SPATIAL_REACQUIRE_MM } from '@/domain/spatial/filter';
import type { SpatialSmoothing } from '@/domain/spatial/smoothing';

const STEP_MS = 250;

function feed(
  filter: ReturnType<typeof createSpatialFilter>,
  xs: readonly number[],
  smoothing: SpatialSmoothing = 'medium',
  startAt = 0
) {
  return xs.map((x, index) => filter.apply({ x, y: 0, at: startAt + index * STEP_MS }, smoothing));
}

function meanStep(points: readonly { x: number }[]): number {
  let total = 0;
  for (let i = 1; i < points.length; i++) total += Math.abs(points[i].x - points[i - 1].x);
  return total / Math.max(1, points.length - 1);
}

test('smoothing level applies immediately, including on existing filter state', () => {
  const noise = [30, -30, 30, -30, 30, -30, 30, -30, 30];
  const medium = createSpatialFilter();
  const raw = createSpatialFilter();
  feed(medium, noise, 'medium');
  feed(raw, noise, 'off');

  const mediumOut = medium.apply({ x: 60, y: 0, at: noise.length * STEP_MS }, 'medium');
  const rawOut = raw.apply({ x: 60, y: 0, at: noise.length * STEP_MS }, 'off');
  expect(Math.abs(rawOut.x - 60)).toBeLessThan(0.01);
  expect(Math.abs(mediumOut.x - rawOut.x)).toBeGreaterThan(1);

  // The same filter that produced raw output must smooth as soon as the level changes.
  const switched = createSpatialFilter();
  feed(switched, noise, 'off');
  const highOut = switched.apply({ x: 60, y: 0, at: noise.length * STEP_MS }, 'high');
  expect(highOut.x).toBeLessThan(60);
});

test('reduces stationary jitter without moving the dot', () => {
  // Deterministic noise around a fixed spot, the pattern a still person produces.
  const noise = [18, -22, 12, -30, 25, -14, 8, -26, 20, -10];
  const xs = Array.from({ length: 60 }, (_, index) => noise[index % noise.length]);
  const input = xs.map((x) => ({ x }));
  const filter = createSpatialFilter();
  const output = feed(filter, xs);

  const settledIn = meanStep(input.slice(20));
  const settledOut = meanStep(output.slice(20));
  expect(settledOut).toBeLessThanOrEqual(settledIn * 0.5);
  expect(Math.abs(output.at(-1)!.x)).toBeLessThanOrEqual(35);
});

test('keeps up with a walking target and never jumps', () => {
  const xs = Array.from({ length: 60 }, (_, index) => index * 250);
  const filter = createSpatialFilter();
  const output = feed(filter, xs);

  const lastInput = xs.at(-1)!;
  const lastOutput = output.at(-1)!;
  expect(Math.abs(lastInput - lastOutput.x)).toBeLessThanOrEqual(150);
  for (let i = 1; i < output.length; i++) {
    expect(Math.abs(output[i].x - output[i - 1].x)).toBeLessThanOrEqual(500);
  }
});

test('rejects an implausible jump and re-acquires when two samples agree', () => {
  const filter = createSpatialFilter();
  feed(filter, [0, 0, 0]);
  const rejected = filter.apply({ x: 5000, y: 0, at: 4 * STEP_MS }, 'medium');
  expect(rejected.state).toBe('predicted');
  expect(rejected.x).toBeLessThanOrEqual(5);

  const reacquired = filter.apply({ x: 5000 + SPATIAL_REACQUIRE_MM / 2, y: 0, at: 5 * STEP_MS }, 'medium');
  expect(reacquired.state).toBe('live');
  expect(reacquired.x).toBeGreaterThan(4000);
});

test('accepts a sustained disagreement instead of holding forever', () => {
  const filter = createSpatialFilter();
  feed(filter, [0, 0]);
  const at = 2 * STEP_MS;
  let output = filter.apply({ x: 5000, y: 0, at }, 'medium');
  expect(output.state).toBe('predicted');
  const samples = Math.ceil(SPATIAL_PREDICT_MS / STEP_MS) + 1;
  for (let index = 1; index <= samples; index++) {
    output = filter.apply({ x: 5000 + index * 400, y: 0, at: at + index * STEP_MS }, 'medium');
  }
  expect(output.state).toBe('live');
  expect(output.x).toBeGreaterThan(5000);
});

test('restarts after an update gap instead of dragging across it', () => {
  const filter = createSpatialFilter();
  feed(filter, [0, 0, 0]);
  const afterGap = filter.apply({ x: 1000, y: 0, at: 3 * STEP_MS + 2000 }, 'medium');
  expect(afterGap.state).toBe('live');
  expect(afterGap.x).toBeCloseTo(1000, 0);
});

test('reset forgets position and prediction state', () => {
  const filter = createSpatialFilter();
  feed(filter, [0, 0, 0]);
  expect(filter.apply({ x: 5000, y: 0, at: 4 * STEP_MS }, 'medium').state).toBe('predicted');
  filter.reset();
  const afterReset = filter.apply({ x: 5000, y: 0, at: 5 * STEP_MS }, 'medium');
  expect(afterReset.state).toBe('live');
  expect(afterReset.x).toBeCloseTo(5000, 0);
});

test('off still rejects teleports and passes accepted motion through', () => {
  const filter = createSpatialFilter();
  feed(filter, [0, 0, 0]);
  expect(filter.apply({ x: 4000, y: 0, at: 4 * STEP_MS }, 'off').state).toBe('predicted');
  const step = filter.apply({ x: 250, y: 0, at: 5 * STEP_MS }, 'off');
  expect(step.state).toBe('live');
});
