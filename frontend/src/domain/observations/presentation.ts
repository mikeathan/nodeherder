/**
 * Presentation helpers for the wireless observation viewer (spec 064).
 *
 * Deliberately separate from `domain/spatial`: the spatial viewer is millimetre-Cartesian with
 * zones and bearings, while this view is range-only and bearing-less. Sharing the canvas would
 * corrupt both, so only the polar maths is reused.
 */
import type {
  ObservationDecision,
  ObservationIdentityView,
  ObservationSnapshot,
  ObservationSourceDescriptor,
  ObservationTrend,
} from '@/types/observations.type';
import { ObservationDecisions, ObservationTrends } from '@/types/observations.type';

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

/**
 * Render a BLE company identifier the way people recognise it, e.g. `0x004C` for Apple.
 * Returns an empty string when the identity carries no vendor, so the facet never invents one.
 */
export function vendorHex(vendor: number | undefined): string {
  if (typeof vendor !== 'number' || !Number.isFinite(vendor)) return '';
  return `0x${vendor.toString(16).toUpperCase().padStart(4, '0')}`;
}

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

/**
 * Explain why a source may show no data.
 *
 * Silence has several indistinguishable causes — unreachable host, no `bluetooth_proxy`
 * compiled in, firmware sending a different message type, or genuinely nothing in range. These
 * descriptions make the difference visible, which matters most on a first real flash.
 */
export function sourceStatus(source: ObservationSourceDescriptor): string {
  const diagnostics = source.diagnostics;
  if (!diagnostics) return 'Connected';
  if (diagnostics.advertisements > 0) {
    const dropped = diagnostics.dropped > 0 ? `, ${diagnostics.dropped} dropped` : '';
    return `Scanning · ${diagnostics.advertisements} advertisements${dropped}`;
  }
  return 'Connected, but no advertisements received. Check that this device runs bluetooth_proxy.';
}

/** Whether a source is connected yet has never delivered a packet, which signals a fault. */
export function sourceIsSilent(source: ObservationSourceDescriptor): boolean {
  const diagnostics = source.diagnostics;
  return !!diagnostics && diagnostics.advertisements === 0;
}

/**
 * Window in which an identity counts as "recently appeared".
 *
 * Novelty must come from arrival time, not from the triage decision. On a first run every
 * identity is untriaged, so filtering on `decision === 'unknown'` matches everything and the
 * control appears broken.
 */
export const RECENT_ARRIVAL_MS = 5 * 60 * 1000;

/**
 * Tolerated clock skew between the hub and the browser, in milliseconds.
 *
 * A reading slightly ahead of local time is normal; one far ahead is bad data and must not be
 * presented as a fresh arrival.
 */
export const ARRIVAL_SKEW_TOLERANCE_MS = 60 * 1000;

/** Whether an identity first appeared within the recent window, relative to now. */
export function isRecentArrival(
  identity: Pick<ObservationIdentityView, 'firstSeen'>,
  now: number = Date.now(),
  windowMs: number = RECENT_ARRIVAL_MS,
): boolean {
  const seen = Date.parse(identity.firstSeen);
  if (Number.isNaN(seen)) return false;
  const elapsed = now - seen;
  // Reject data from beyond the skew tolerance rather than showing an impossible arrival.
  if (elapsed < -ARRIVAL_SKEW_TOLERANCE_MS) return false;
  return elapsed <= windowMs;
}

/** Human group label for an identity, used to section the list view. */
export function identityGroup(identity: ObservationIdentityView): string {
  if (identity.kind === 'ibeacon') return 'iBeacon';
  if (identity.kind === 'eddystone') return 'Eddystone';
  const vendor = vendorHex(identity.vendor);
  if (vendor) return `Vendor ${vendor}`;
  return identity.stable ? 'Stable address' : 'Rotating address';
}

/**
 * Group identities into ordered sections for the list view.
 *
 * Sections are ordered by size so the noisiest category is first, and identities inside a
 * section keep the incoming order (most recently seen first). Groups with equal size fall back
 * to label order, so the output is deterministic.
 */
export function groupIdentities(
  identities: readonly ObservationIdentityView[],
): { group: string; identities: ObservationIdentityView[] }[] {
  const sections = new Map<string, ObservationIdentityView[]>();
  for (const identity of identities) {
    const group = identityGroup(identity);
    const bucket = sections.get(group);
    if (bucket) bucket.push(identity);
    else sections.set(group, [identity]);
  }
  return [...sections.entries()]
    .map(([group, members]) => ({ group, identities: members }))
    .sort((a, b) => b.identities.length - a.identities.length || a.group.localeCompare(b.group));
}

/**
 * Coarse proximity bands.
 *
 * A single anchor measures RSSI only, so the estimate carries 1-3 m of indoor error. Offering a
 * band alongside the figure keeps the UI from implying precision it cannot have.
 */
export const DistanceBands = {
  Near: 'near',
  Mid: 'mid',
  Far: 'far',
  Unknown: 'unknown',
} as const;

export type DistanceBand = (typeof DistanceBands)[keyof typeof DistanceBands];

/** Metres at which a reading stops counting as `near`. */
export const NEAR_BAND_METRES = 2;
/** Metres at which a reading starts counting as `far`. */
export const FAR_BAND_METRES = 6;

/** Classify a distance estimate into a band. An absent reading is unknown, never near. */
export function distanceBand(metres: number | null | undefined): DistanceBand {
  if (typeof metres !== 'number' || !Number.isFinite(metres)) return DistanceBands.Unknown;
  if (metres < NEAR_BAND_METRES) return DistanceBands.Near;
  if (metres < FAR_BAND_METRES) return DistanceBands.Mid;
  return DistanceBands.Far;
}

/** Human label for a band. */
export const DISTANCE_BAND_LABELS: Readonly<Record<DistanceBand, string>> = {
  near: 'near',
  mid: 'mid',
  far: 'far',
  unknown: 'unknown',
};

/**
 * Direction indicator for a trend.
 *
 * The arrow is paired with a word everywhere it renders: an arrow alone would be ambiguous, and
 * pointing up for "approaching" reads as "signal up" only to someone who already knows.
 */
export const TREND_PRESENTATION: Readonly<
  Record<ObservationTrend, { arrow: string; label: string }>
> = {
  approaching: { arrow: '↓', label: 'Approaching' },
  receding: { arrow: '↑', label: 'Receding' },
  steady: { arrow: '→', label: 'Steady' },
  unknown: { arrow: '', label: 'No trend yet' },
};

/** Presentation for a trend value, defaulting to unknown for an absent field. */
export function trendPresentation(trend: ObservationTrend | undefined) {
  return TREND_PRESENTATION[trend ?? ObservationTrends.Unknown] ?? TREND_PRESENTATION.unknown;
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
