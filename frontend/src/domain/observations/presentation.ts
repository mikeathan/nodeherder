/**
 * Presentation helpers for the wireless observation viewer (spec 064).
 *
 * Deliberately separate from `domain/spatial`: the spatial viewer is millimetre-Cartesian with
 * zones and bearings, while this view is range-only and bearing-less. Sharing the canvas would
 * corrupt both, so only the polar maths is reused.
 */
import type { ObservationDecision, ObservationIdentityView, ObservationSnapshot } from '@/types/observations.type';
import { ObservationDecisions } from '@/types/observations.type';

/** Log range rings in metres. RSSI error is multiplicative, so linear rings would mislead. */
export const RANGE_RINGS_M = [1, 3, 10, 30] as const;

/** Calibrated transmit power at one metre, in dBm, used for the range estimate. */
export const DEFAULT_TX_POWER_AT_1M = -59;

/** Indoor path-loss exponent. 2 is free space; 2.5-3.5 is typical indoors. */
export const DEFAULT_PATH_LOSS_EXPONENT = 2.5;

/** Tolerance applied to per-packet RSSI before it is treated as a real change. */
export const RSSI_SMOOTHING_DB = 4;

/**
 * Estimate distance from RSSI using log-distance path loss.
 *
 * Returns null when the reading is absent, because "not observed" must never render as zero
 * metres. The result is a coarse range band, not a measurement: indoor error is metres.
 */
export function estimateDistanceMetres(
  rssi: number | null | undefined,
  txPowerAt1m: number = DEFAULT_TX_POWER_AT_1M,
  pathLossExponent: number = DEFAULT_PATH_LOSS_EXPONENT,
): number | null {
  if (typeof rssi !== 'number' || !Number.isFinite(rssi)) return null;
  if (!Number.isFinite(txPowerAt1m) || !Number.isFinite(pathLossExponent) || pathLossExponent <= 0) return null;
  return 10 ** ((txPowerAt1m - rssi) / (10 * pathLossExponent));
}

/**
 * Place an identity on a fixed spoke derived from its key.
 *
 * A single source measures range, not bearing, so the angle is a stable hash rather than a
 * claim about direction. Using the key means one identity always sits on the same spoke, so
 * movement reads as radial drift instead of teleporting.
 */
export function bearingForIdentity(key: string): number {
  let hash = 0;
  for (let index = 0; index < key.length; index += 1) {
    hash = (hash * 31 + key.charCodeAt(index)) | 0;
  }
  return Math.abs(hash) % 360;
}

/** Whether an identity is a rotating-address fingerprint rather than a stable device. */
export function isEphemeralIdentity(identity: ObservationIdentityView): boolean {
  return !identity.stable;
}

/** Human label for a triage decision. */
export const DECISION_LABELS: Record<ObservationDecision, string> = {
  [ObservationDecisions.Unknown]: 'Unknown',
  [ObservationDecisions.Known]: 'Known',
  [ObservationDecisions.Ignored]: 'Ignored',
};

/** Display name for an identity, preferring the operator label, then the advertised name. */
export function identityTitle(identity: ObservationIdentityView): string {
  if (identity.label) return identity.label;
  if (identity.kind && identity.kind !== 'adv') return identity.kind;
  return identity.key.length > 24 ? `${identity.key.slice(0, 24)}…` : identity.key;
}

/** Count of identities that are either new or ephemeral, used in the summary line. */
export function untriagedCount(snapshot: ObservationSnapshot | null): number {
  if (!snapshot) return 0;
  return snapshot.identities.filter((identity) => identity.decision === ObservationDecisions.Unknown).length;
}

/** Human summary of a snapshot for the header, avoiding invented precision. */
export function snapshotSummary(snapshot: ObservationSnapshot | null): string {
  if (!snapshot) return 'No observations yet';
  const identities = snapshot.identities.length;
  const sources = snapshot.sources.length;
  const arrivals = snapshot.arrivals;
  const base = `${identities} ${identities === 1 ? 'device' : 'devices'} from ${sources} ${sources === 1 ? 'source' : 'sources'}`;
  if (arrivals > 0) return `${base} · ${arrivals} new`;
  return base;
}
