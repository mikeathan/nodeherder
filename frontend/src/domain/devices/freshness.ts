/*
 * Reporting freshness is a display-only fact derived from `last_seen`.
 *
 * It is deliberately separate from availability (`deviceStatus`): a quiet device is still
 * "online" for command gating and automation routing. This module never mutates device
 * state and has no transport/UI dependencies, so it can be unit-tested and reused by the
 * tile, list, device page and panel surfaces.
 */
import type { Device } from '@/types/device';
import { isBatteryPowered } from './index';

const HOUR = 60 * 60 * 1000;

/** How long a device may stay silent before the UI calls it quiet or stale. */
export type FreshnessPolicy = {
  /** Age at which the UI shows an amber “not reporting” hint. */
  quietAfterMs: number;
  /** Age at which the UI marks values as last known/stale. */
  staleAfterMs: number;
};

/** Mains-powered devices are expected to report regularly. */
export const MAINS_FRESHNESS: FreshnessPolicy = { quietAfterMs: 2 * HOUR, staleAfterMs: 6 * HOUR };

/**
 * Battery/change-only devices may legitimately stay silent for a day or more, so they get a
 * much longer window. A false “quiet” is only an amber hint, but it still erodes trust.
 */
export const BATTERY_FRESHNESS: FreshnessPolicy = { quietAfterMs: 24 * HOUR, staleAfterMs: 48 * HOUR };

/** Freshness bands used by presentation. `unknown` means the timestamp is missing/invalid. */
export type ReportFreshness = 'fresh' | 'quiet' | 'stale' | 'unknown';

/** Policy for one device based on its reported power source. */
export function freshnessPolicy(device: Pick<Device, 'power_source'> | undefined): FreshnessPolicy {
  return device && isBatteryPowered(device) ? BATTERY_FRESHNESS : MAINS_FRESHNESS;
}

/** Milliseconds since the last report, or null when the timestamp cannot be parsed. */
export function reportAgeMs(lastSeen: string | null | undefined, now: number): number | null {
  if (!lastSeen) return null;
  const at = Date.parse(lastSeen);
  if (!Number.isFinite(at)) return null;
  // A future timestamp is clock skew, not a reason to hide the value.
  return Math.max(0, now - at);
}

/** Classifies reporting age for presentation. Never throws on malformed input. */
export function reportFreshness(
  lastSeen: string | null | undefined,
  now: number,
  policy: FreshnessPolicy = MAINS_FRESHNESS
): ReportFreshness {
  const age = reportAgeMs(lastSeen, now);
  if (age === null) return 'unknown';
  if (age >= policy.staleAfterMs) return 'stale';
  if (age >= policy.quietAfterMs) return 'quiet';
  return 'fresh';
}
