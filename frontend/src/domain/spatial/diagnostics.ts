/*
 * Session diagnostics over recorded raw frames. These are measurements, not firmware
 * guarantees and never a diagnosis of cause; unknown or sparse data stays labelled.
 */
import type { ReplayFrame } from '@/domain/spatial/replay';
import { SPATIAL_MERGE_MM } from '@/domain/spatial/tracks';

/** Slots with fewer samples are not rated. */
export const DIAGNOSTIC_MIN_SAMPLES = 10;
/** Positional spread considered stable, in millimetres (heuristic). */
export const DIAGNOSTIC_STABLE_MM = 40;
/** Spread above this is noisy (heuristic). */
export const DIAGNOSTIC_UNSETTLED_MM = 100;

export type SlotStability = 'insufficient' | 'stable' | 'unsettled' | 'noisy';

export type SlotDiagnostics = {
  slot: number;
  samples: number;
  meanIntervalMs: number | null;
  sigmaX: number;
  sigmaY: number;
  movementMm: number;
  stability: SlotStability;
};

export type DiagnosticRecording = { frames: number; from: number; to: number; durationMs: number };

function standardDeviation(values: readonly number[]): number {
  if (!values.length) return 0;
  const mean = values.reduce((sum, value) => sum + value, 0) / values.length;
  return Math.sqrt(values.reduce((sum, value) => sum + (value - mean) ** 2, 0) / values.length);
}

/** Recording extent of the session buffer, for the view's context line. */
export function diagnosticRecording(frames: readonly ReplayFrame[]): DiagnosticRecording {
  if (!frames.length) return { frames: 0, from: 0, to: 0, durationMs: 0 };
  const from = frames[0].at;
  const to = frames[frames.length - 1].at;
  return { frames: frames.length, from, to, durationMs: Math.max(0, to - from) };
}

/** Per-slot stability statistics from raw recorded frames. */
export function slotDiagnostics(frames: readonly ReplayFrame[]): SlotDiagnostics[] {
  const bySlot = new Map<number, { x: number; y: number; at: number }[]>();
  for (const frame of frames) {
    for (const sample of frame.samples) {
      const list = bySlot.get(sample.slot) ?? [];
      list.push({ x: sample.x, y: sample.y, at: frame.at });
      bySlot.set(sample.slot, list);
    }
  }
  return [...bySlot.entries()]
    .sort(([left], [right]) => left - right)
    .map(([slot, samples]) => {
      const intervals = samples.slice(1).map((sample, index) => sample.at - samples[index].at).filter((value) => value > 0);
      let movementMm = 0;
      for (let index = 1; index < samples.length; index++) {
        movementMm += Math.hypot(samples[index].x - samples[index - 1].x, samples[index].y - samples[index - 1].y);
      }
      const sigmaX = standardDeviation(samples.map((sample) => sample.x));
      const sigmaY = standardDeviation(samples.map((sample) => sample.y));
      const spread = Math.max(sigmaX, sigmaY);
      const stability: SlotStability =
        samples.length < DIAGNOSTIC_MIN_SAMPLES
          ? 'insufficient'
          : spread <= DIAGNOSTIC_STABLE_MM
            ? 'stable'
            : spread <= DIAGNOSTIC_UNSETTLED_MM
              ? 'unsettled'
              : 'noisy';
      return {
        slot,
        samples: samples.length,
        meanIntervalMs:
          intervals.length > 0 ? intervals.reduce((sum, value) => sum + value, 0) / intervals.length : null,
        sigmaX,
        sigmaY,
        movementMm,
        stability,
      };
    });
}

export type OverlapSummary = { frames: number; closestMm: number | null; pair: [number, number] | null };

/** Counts frames where two concurrently active slots are closer than the merge distance. */
export function overlapSummary(frames: readonly ReplayFrame[], threshold = SPATIAL_MERGE_MM): OverlapSummary {
  let overlapped = 0;
  let closestMm: number | null = null;
  let pair: [number, number] | null = null;
  for (const frame of frames) {
    let frameHasPair = false;
    for (let left = 0; left < frame.samples.length; left++) {
      for (let right = left + 1; right < frame.samples.length; right++) {
        const distance = Math.hypot(
          frame.samples[left].x - frame.samples[right].x,
          frame.samples[left].y - frame.samples[right].y
        );
        if (distance >= threshold) continue;
        frameHasPair = true;
        if (closestMm === null || distance < closestMm) {
          closestMm = distance;
          pair = [frame.samples[left].slot, frame.samples[right].slot];
        }
      }
    }
    if (frameHasPair) overlapped++;
  }
  return { frames: overlapped, closestMm, pair };
}
