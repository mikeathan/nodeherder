/*
 * Spatial frame ingress and conversion. Frames arrive from the hub as untrusted external
 * data, so they are validated here before any state is committed, and converted once into
 * millimetre samples that the rest of the spatial domain can use without units.
 */
import type { SpatialFrame, SpatialFrameAxis, SpatialFrameTarget } from '@/types/spatial';
import { LENGTH_TO_MM } from './index';

/** Highest slot number the hub can report; extra entries are ignored rather than trusted. */
export const SPATIAL_MAX_FRAME_TARGETS = 3;

/**
 * One validated slot reading, converted to millimetres where an axis was usable. Timing is
 * deliberately absent: consumers age frames by browser arrival time, not by the hub clock.
 */
export type SpatialFrameSample = {
  /** Mapping row number, 1–3. A slot, not an identity. */
  slot: number;
  active: boolean;
  /** Millimetres, or null when the axis was absent, non-finite or in an unknown unit. */
  x: number | null;
  y: number | null;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

/** Converts one declared axis to millimetres; unknown units and non-finite values are unusable. */
export function axisMillimetres(axis: SpatialFrameAxis | undefined): number | null {
  if (!axis || !Number.isFinite(axis.value)) return null;
  const multiplier = LENGTH_TO_MM[axis.unit.trim().toLowerCase()];
  return multiplier ? axis.value * multiplier : null;
}

function parseAxis(value: unknown): SpatialFrameAxis | undefined {
  if (!isRecord(value)) return undefined;
  const { value: rawValue, unit } = value;
  if (typeof rawValue !== 'number' || !Number.isFinite(rawValue)) return undefined;
  if (typeof unit !== 'string' || !LENGTH_TO_MM[unit.trim().toLowerCase()]) return undefined;
  return { value: rawValue, unit };
}

function parseTarget(value: unknown): SpatialFrameTarget | null {
  if (!isRecord(value)) return null;
  const { id, active } = value;
  if (typeof id !== 'number' || !Number.isInteger(id) || id < 1 || id > SPATIAL_MAX_FRAME_TARGETS) return null;
  if (typeof active !== 'boolean') return null;
  const target: SpatialFrameTarget = { id, active };
  // Axes are optional by contract: a cleared slot reports the flag without coordinates.
  const x = parseAxis(value.x);
  const y = parseAxis(value.y);
  if (x) target.x = x;
  if (y) target.y = y;
  return target;
}

/**
 * Validates one `spatialTargets` payload. Returns null for a malformed envelope so the store
 * can ignore it, and drops unusable targets or axes instead of committing them. Repeated
 * slot numbers keep the first entry, so a consumer never sees two readings for one slot.
 */
export function parseSpatialFrame(payload: unknown): SpatialFrame | null {
  if (!isRecord(payload)) return null;
  const { id, at, targets } = payload;
  if (typeof id !== 'string' || !id) return null;
  if (typeof at !== 'string' || Number.isNaN(Date.parse(at))) return null;
  if (!Array.isArray(targets)) return null;
  const seen = new Set<number>();
  const parsed: SpatialFrameTarget[] = [];
  for (const entry of targets) {
    if (parsed.length >= SPATIAL_MAX_FRAME_TARGETS) break;
    const target = parseTarget(entry);
    if (!target || seen.has(target.id)) continue;
    seen.add(target.id);
    parsed.push(target);
  }
  return { id, at, targets: parsed };
}

/**
 * Converts a validated frame into millimetre samples, one per reported slot. Axes that are
 * missing or in an unknown unit become null; the caller decides whether a slot is drawable.
 */
export function frameSamples(frame: SpatialFrame): SpatialFrameSample[] {
  return frame.targets.map((target) => ({
    slot: target.id,
    active: target.active,
    x: axisMillimetres(target.x),
    y: axisMillimetres(target.y),
  }));
}
