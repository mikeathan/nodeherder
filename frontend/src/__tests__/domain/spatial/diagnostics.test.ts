import {
  DIAGNOSTIC_MIN_SAMPLES,
  diagnosticRecording,
  overlapSummary,
  slotDiagnostics,
} from '@/domain/spatial/diagnostics';
import type { ReplayFrame } from '@/domain/spatial/replay';

const frame = (at: number, samples: { slot: number; x: number; y: number }[]): ReplayFrame => ({ at, samples });

test('per-slot statistics describe spread, cadence and movement', () => {
  const frames: ReplayFrame[] = Array.from({ length: 20 }, (_, index) =>
    frame(index * 250, [{ slot: 1, x: index % 2 === 0 ? 990 : 1010, y: 2000 }])
  );
  const [slot] = slotDiagnostics(frames);
  expect(slot.samples).toBe(20);
  expect(slot.meanIntervalMs).toBe(250);
  expect(slot.sigmaX).toBeCloseTo(10, 6);
  expect(slot.sigmaY).toBe(0);
  expect(slot.stability).toBe('stable');
  expect(slot.movementMm).toBeCloseTo(19 * 20, 6);
});

test('sparse slots are not rated and noisy spread is labelled', () => {
  const sparse = slotDiagnostics([frame(0, [{ slot: 2, x: 0, y: 0 }])]);
  expect(sparse[0].stability).toBe('insufficient');
  expect(sparse[0].meanIntervalMs).toBeNull();
  const noisyFrames = Array.from({ length: DIAGNOSTIC_MIN_SAMPLES }, (_, index) =>
    frame(index * 100, [{ slot: 3, x: index * 200, y: 0 }])
  );
  expect(slotDiagnostics(noisyFrames)[0].stability).toBe('noisy');
});

test('recording extent and overlap evidence are reported without claiming a cause', () => {
  const frames = [
    frame(1000, [{ slot: 1, x: 0, y: 1000 }, { slot: 2, x: 100, y: 1000 }]),
    frame(2000, [{ slot: 1, x: 0, y: 1000 }, { slot: 2, x: 800, y: 1000 }]),
    frame(3000, [{ slot: 1, x: 0, y: 1000 }]),
  ];
  expect(diagnosticRecording(frames)).toEqual({ frames: 3, from: 1000, to: 3000, durationMs: 2000 });
  expect(overlapSummary(frames)).toEqual({ frames: 1, closestMm: 100, pair: [1, 2] });
  expect(overlapSummary([])).toEqual({ frames: 0, closestMm: null, pair: null });
});
