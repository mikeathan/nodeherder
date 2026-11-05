/*
 * Adaptive position filter for one radar slot. Radar coordinates are noisy by a few
 * centimetres at several hertz, so an unfiltered dot shimmers and appears to wander while a
 * person stands still. This module applies a One Euro filter (Casiez, Roussel & Vogel, CHI
 * 2012) with a stationary dead band, a plausibility gate for implausible jumps, and a
 * bounded prediction window, and reports a jitter estimate so the UI can show evidence
 * instead of claiming accuracy.
 */
import { SPATIAL_SMOOTHING_PRESETS } from './smoothing';
import type { SpatialSmoothing } from './smoothing';

/** Derivative smoothing cutoff in hertz for the One Euro filter. */
export const SPATIAL_FILTER_D_CUTOFF_HZ = 1;
/** Movement below this distance is treated as jitter while the target is stationary. */
export const SPATIAL_DEADBAND_MM = 35;
/** Movement faster than this is treated as an implausible jump, not as motion. */
export const SPATIAL_MAX_SPEED_MM_S = 3000;
/** A sample gap longer than this restarts the filter instead of dragging across it. */
export const SPATIAL_GAP_MS = 600;
/** Longest time a rejected jump keeps its last known position marked as predicted. */
export const SPATIAL_PREDICT_MS = 1000;
/** Two rejected samples closer than this are treated as one re-acquired target. */
export const SPATIAL_REACQUIRE_MM = 400;
/** Speed below which the position counts as stationary for the dead band. */
export const SPATIAL_STATIONARY_SPEED_MM_S = 50;
/** Speed below which sample-to-sample movement contributes to the jitter estimate. */
export const SPATIAL_JITTER_SPEED_MM_S = 150;
/** Smoothing applied to the jitter variance estimate. */
const JITTER_SMOOTHING = 0.85;

/** One raw sample in millimetres with its frame time. */
export type SpatialFilterInput = { x: number; y: number; at: number };

/** Filtered position, velocity in mm/s, and how the sample was accepted. */
export type SpatialFilterOutput = {
  x: number;
  y: number;
  vx: number;
  vy: number;
  /** `predicted` means the last accepted position is shown because a jump was rejected. */
  state: 'live' | 'predicted';
  /** Root-mean-square sample movement while stationary, in millimetres. */
  jitterMm: number;
};

/** Stateful filter for one slot; reset it when the slot clears or the view changes source. */
export type SpatialFilter = {
  apply: (input: SpatialFilterInput, smoothing: SpatialSmoothing) => SpatialFilterOutput;
  reset: () => void;
};

type Axis = { raw: number; smoothed: number; derivative: number };

function createAxis(value: number): Axis {
  return { raw: value, smoothed: value, derivative: 0 };
}

function alpha(cutoff: number, dt: number): number {
  const timeConstant = 1 / (2 * Math.PI * cutoff);
  return 1 / (1 + timeConstant / dt);
}

function smoothAxis(axis: Axis, value: number, dt: number, minCutoff: number, beta: number): void {
  const rawDerivative = (value - axis.raw) / dt;
  axis.raw = value;
  axis.derivative += alpha(SPATIAL_FILTER_D_CUTOFF_HZ, dt) * (rawDerivative - axis.derivative);
  const cutoff = minCutoff + beta * Math.abs(axis.derivative);
  axis.smoothed += alpha(cutoff, dt) * (value - axis.smoothed);
}

/**
 * Creates a filter for one target slot. Smoothing is passed to each `apply` call so a level
 * change takes effect on the next sample without discarding position history. `off` keeps the
 * position unfiltered but still applies the plausibility gate, because a teleport must never be
 * shown as motion.
 */
export function createSpatialFilter(): SpatialFilter {
  let axes: { x: Axis; y: Axis } | null = null;
  let lastInput: SpatialFilterInput | null = null;
  let displayed: { x: number; y: number } | null = null;
  let pendingReacquire: { x: number; y: number } | null = null;
  let predictedSince = 0;
  let jitterVariance = 0;

  function reset(): void {
    axes = null;
    lastInput = null;
    displayed = null;
    pendingReacquire = null;
    predictedSince = 0;
    jitterVariance = 0;
  }

  function jitterMm(): number {
    return Math.sqrt(jitterVariance);
  }

  /** Accepts a sample as the new reference, optionally restarting the smoothing state. */
  function accept(input: SpatialFilterInput, restart: boolean): SpatialFilterOutput {
    if (restart || !axes) axes = { x: createAxis(input.x), y: createAxis(input.y) };
    lastInput = input;
    displayed = { x: axes.x.smoothed, y: axes.y.smoothed };
    pendingReacquire = null;
    predictedSince = 0;
    return { x: displayed.x, y: displayed.y, vx: axes.x.derivative, vy: axes.y.derivative, state: 'live', jitterMm: jitterMm() };
  }

  function predicted(): SpatialFilterOutput {
    const point = displayed ?? { x: lastInput?.x ?? 0, y: lastInput?.y ?? 0 };
    return { x: point.x, y: point.y, vx: 0, vy: 0, state: 'predicted', jitterMm: jitterMm() };
  }

  return {
    apply(input, smoothing) {
      const preset = smoothing === 'off' ? null : SPATIAL_SMOOTHING_PRESETS[smoothing];
      if (!Number.isFinite(input.x) || !Number.isFinite(input.y)) return predicted();
      if (!lastInput) return accept(input, true);

      const dt = (input.at - lastInput.at) / 1000;
      if (!(dt > 0)) {
        // A repeated timestamp carries no new information; keep the current presentation.
        return displayed ? { ...displayed, vx: 0, vy: 0, state: 'live', jitterMm: jitterMm() } : accept(input, true);
      }

      const distance = Math.hypot(input.x - lastInput.x, input.y - lastInput.y);
      const speed = distance / dt;

      if (distance > 0 && speed > SPATIAL_MAX_SPEED_MM_S) {
        if (pendingReacquire && Math.hypot(input.x - pendingReacquire.x, input.y - pendingReacquire.y) <= SPATIAL_REACQUIRE_MM) {
          // Two consecutive samples agree on the new position: the target really moved.
          return accept(input, true);
        }
        pendingReacquire = { x: input.x, y: input.y };
        if (!predictedSince) predictedSince = input.at;
        if (input.at - predictedSince >= SPATIAL_PREDICT_MS) {
          // A sustained disagreement is the new truth; holding forever would be a lie.
          return accept(input, true);
        }
        return predicted();
      }
      pendingReacquire = null;
      predictedSince = 0;

      if (speed < SPATIAL_JITTER_SPEED_MM_S) {
        jitterVariance = JITTER_SMOOTHING * jitterVariance + (1 - JITTER_SMOOTHING) * distance * distance;
      }

      if (dt * 1000 > SPATIAL_GAP_MS || !axes) {
        // Do not drag smoothing across an update gap; the jitter estimate starts over too.
        jitterVariance = 0;
        return accept(input, true);
      }

      if (preset) {
        smoothAxis(axes.x, input.x, dt, preset.minCutoff, preset.beta);
        smoothAxis(axes.y, input.y, dt, preset.minCutoff, preset.beta);
      } else {
        axes.x.raw = axes.x.smoothed = input.x;
        axes.y.raw = axes.y.smoothed = input.y;
        axes.x.derivative = (input.x - (displayed?.x ?? input.x)) / dt;
        axes.y.derivative = (input.y - (displayed?.y ?? input.y)) / dt;
      }
      lastInput = input;

      const speedNow = Math.hypot(axes.x.derivative, axes.y.derivative);
      let x = axes.x.smoothed;
      let y = axes.y.smoothed;
      if (preset && displayed && speedNow < SPATIAL_STATIONARY_SPEED_MM_S && Math.hypot(x - displayed.x, y - displayed.y) < SPATIAL_DEADBAND_MM) {
        // Inside the dead band the reading is indistinguishable from jitter: hold the dot.
        x = displayed.x;
        y = displayed.y;
      }
      displayed = { x, y };
      return { x, y, vx: axes.x.derivative, vy: axes.y.derivative, state: 'live', jitterMm: jitterMm() };
    },
    reset,
  };
}
