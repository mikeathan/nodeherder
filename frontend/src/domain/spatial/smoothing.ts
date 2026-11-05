/*
 * User-facing smoothing levels. Raw One Euro parameters are deliberately not exposed:
 * three named levels keep one source of truth for the filter, the selector and the help
 * text, and prevent a tuning regress into per-room guesswork.
 */

/** Smoothing levels offered in the spatial view. */
export type SpatialSmoothing = 'off' | 'low' | 'medium' | 'high';

/** One Euro parameters for a smoothing level. */
export type SpatialSmoothingPreset = { minCutoff: number; beta: number };

/** Default level: enough calm to stop shimmer without visible lag on a slow walk. */
export const SPATIAL_DEFAULT_SMOOTHING: SpatialSmoothing = 'medium';

/**
 * One Euro parameters per level. Lower `minCutoff` smooths a stationary dot more; higher
 * `beta` lets a moving dot follow sooner. Shared by every spatial surface.
 */
export const SPATIAL_SMOOTHING_PRESETS: Readonly<Record<Exclude<SpatialSmoothing, 'off'>, SpatialSmoothingPreset>> = {
  low: { minCutoff: 1.7, beta: 0.02 },
  medium: { minCutoff: 1, beta: 0.01 },
  high: { minCutoff: 0.6, beta: 0.005 },
};

/** Selector label per smoothing level. */
export const SPATIAL_SMOOTHING_LABELS: Readonly<Record<SpatialSmoothing, string>> = {
  off: 'Off (raw)',
  low: 'Light',
  medium: 'Balanced',
  high: 'Smooth',
};

/** Selector options in display order. */
export const SPATIAL_SMOOTHING_OPTIONS: readonly { value: SpatialSmoothing; label: string }[] = (
  ['off', 'low', 'medium', 'high'] as const
).map((value) => ({ value, label: SPATIAL_SMOOTHING_LABELS[value] }));
