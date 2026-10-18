import type { Expose } from '@/types/device';
export type NumericRange = { min: number; max: number };

/** Native bounds are shared by device rows, full-screen controls and display conversions. */
export function exposeRange(expose: Expose): NumericRange | null {
  const attributes = expose.attributes ?? {};
  const max = Number(attributes.max ?? attributes.value_max);
  if (!Number.isFinite(max)) return null;
  const min = Number(attributes.min ?? attributes.value_min ?? 0);
  if (!Number.isFinite(min) || max <= min) return null;
  return { min, max };
}

export function numericStep(expose: Expose): number {
  const a = expose.attributes ?? {};
  const step = Number(a.step ?? a.value_step ?? 1);
  return Number.isFinite(step) && step > 0 ? step : 1;
}
export function nativeNumber(expose: Expose, value: number): number | null {
  if (!Number.isFinite(value)) return null;
  const range = exposeRange(expose);
  const min = range?.min ?? 0;
  const step = numericStep(expose);
  const rounded = Number((min + Math.round((value - min) / step) * step).toFixed(8));
  if (!range) return rounded;
  return Math.min(range.max, Math.max(range.min, rounded));
}
export function isMired(expose: Expose): boolean {
  return expose.name === 'color_temp' && String(expose.unit).toLowerCase() !== 'k';
}
/** Missing bounds allow a number field; contradictory bounds must disable writes. */
export function invalidRange(expose: Expose): boolean {
  const a = expose.attributes ?? {};
  const max = a.max ?? a.value_max;
  return max !== undefined && max !== null && !exposeRange(expose);
}
export function sliderValue(expose: Expose, value: number): number {
  const range = exposeRange(expose);
  if (range && isMired(expose)) return range.min + range.max - value;
  return value;
}

/** Visual position of a native value, clamped so unknown/out-of-range data cannot escape. */
export function rangePercent(value: number, min: number, max: number): number {
  if (!Number.isFinite(value) || !Number.isFinite(min) || !Number.isFinite(max) || max <= min) return 0;
  return Math.max(0, Math.min(100, ((value - min) / (max - min)) * 100));
}

/** Display conversion keeps full sample precision; formatting is a separate concern. */
export function displayNumber(expose: Expose, value: number): number {
  if (expose.name === 'brightness') return (value / (exposeRange(expose)?.max || 254)) * 100;
  if (isMired(expose) && value > 0) return 1e6 / value;
  return value;
}
