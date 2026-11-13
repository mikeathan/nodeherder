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

/**
 * Log range rings in metres.
 *
 * RSSI error is multiplicative, so rings are logarithmic. The set reaches below a metre because
 * a nearby device is the common case and was previously indistinguishable from the origin.
 */
export const RANGE_RINGS_M = [0.5, 1, 3, 10, 30] as const;

/**
 * Smallest distance the radial projection will place, in metres.
 *
 * The log projection needs a positive floor: `log10(1) == 0` collapses anything at or below 1 m
 * onto the exact centre, so a device half a metre away drew the same dot as one at the antenna.
 * Flooring slightly below the smallest ring keeps sub-metre readings distinguishable.
 */
export const MIN_PROJECTED_METRES = 0.2;

/**
 * Fraction of the canvas radius used by the outermost ring, leaving room for labels.
 * Shared by the projection and the ring drawing so the two cannot disagree.
 */
export const RADAR_RADIUS_FRACTION = 0.9;

/**
 * Radial position of a distance on the canvas, as a fraction of the canvas radius.
 *
 * Returns a value in (0, 1]: 1 is the outermost ring. Monotonic, so a closer device is never
 * drawn further out than a distant one.
 */
export function radialRatio(
  metres: number,
  maxRing: number = RANGE_RINGS_M[RANGE_RINGS_M.length - 1],
): number {
  if (!Number.isFinite(metres) || !Number.isFinite(maxRing) || maxRing <= MIN_PROJECTED_METRES) return 0;
  const clamped = Math.min(Math.max(metres, MIN_PROJECTED_METRES), maxRing);
  const ratio = Math.log10(clamped / MIN_PROJECTED_METRES) / Math.log10(maxRing / MIN_PROJECTED_METRES);
  return Math.min(Math.max(ratio, 0.02), 1);
}

/** Span presets for the radar scale, in metres. */
export const RADAR_SPANS_M = [3, 10, 30] as const;

export type RadarSpan = (typeof RADAR_SPANS_M)[number];

/**
 * Choose a span from the data, rounding up to the smallest preset that contains it.
 *
 * Scaling to the data is what makes sub-metre bands legible: on a 30 m scale every nearby device
 * collapses into the middle few percent of the canvas.
 */
export function autoRadarSpan(metres: readonly number[]): RadarSpan {
  const largest = metres.reduce((max, value) => (Number.isFinite(value) && value > max ? value : max), 0);
  return RADAR_SPANS_M.find((span) => largest <= span) ?? RADAR_SPANS_M[RADAR_SPANS_M.length - 1];
}

/** Rings to draw for a span: every standard ring inside it, plus the span edge itself. */
export function ringsForSpan(span: number): number[] {
  const within = RANGE_RINGS_M.filter((ring) => ring < span);
  return [...within, span];
}

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
 * Signal strength to base a distance on.
 *
 * The smoothed median is preferred because a single packet swings several dB, which through the
 * log-distance curve becomes a multi-metre jump. Falling back to the raw reading keeps an older
 * payload usable rather than showing nothing.
 */
export function distanceRSSI(identity: Pick<ObservationIdentityView, 'rssi' | 'rssiSmoothed'>): number | null {
  if (typeof identity.rssiSmoothed === 'number' && Number.isFinite(identity.rssiSmoothed)) {
    return identity.rssiSmoothed;
  }
  if (typeof identity.rssi === 'number' && Number.isFinite(identity.rssi)) return identity.rssi;
  return null;
}

/**
 * Round a distance to something the estimate can actually support.
 *
 * Below 10 m the value is whole metres; above that, 5 m steps. Sub-metre precision is not real
 * when indoor error is 1-3 m, and showing it invites false confidence. Never rounds to zero: a
 * positive estimate under a metre reads `1 m`.
 */
export function roundedDistanceMetres(metres: number | null | undefined): number | null {
  if (typeof metres !== 'number' || !Number.isFinite(metres)) return null;
  if (metres < 10) return Math.max(1, Math.round(metres));
  return Math.round(metres / 5) * 5;
}

/**
 * Distance filter ranges.
 *
 * Ranges rather than open-ended limits: "within 20 m" covers nearly every device, so it cannot
 * isolate anything. A band answers "which of these is in the next room", and the boundaries are
 * deliberately small because beyond a few metres the estimate is too coarse to subdivide.
 *
 * Bounds are half-open `[min, max)`, so neighbouring ranges never both match one reading. A null
 * bound is open-ended, making the set exhaustive with no overlap.
 */
export type DistanceRange = {
  /** Stable identifier, used for selection state and test assertions. */
  id: string;
  /** Inclusive lower bound in metres, or null for no lower bound. */
  min: number | null;
  /** Exclusive upper bound in metres, or null for no upper bound. */
  max: number | null;
  label: string;
};

/**
 * Ranges offered as a segmented control, smallest first.
 *
 * Resolution is finer near the antenna because that is where the estimate is most trustworthy and
 * where the interesting question lives ("is this on my desk or across the room"). Above 10 m the
 * log-distance error is several metres, so further subdivision would imply precision that does
 * not exist.
 */
export const DISTANCE_RANGES: readonly DistanceRange[] = [
  { id: 'any', min: null, max: null, label: 'Any' },
  { id: 'under05', min: null, max: 0.5, label: '< 0.5 m' },
  { id: '05to1', min: 0.5, max: 1, label: '0.5–1 m' },
  { id: '1to2', min: 1, max: 2, label: '1–2 m' },
  { id: '2to5', min: 2, max: 5, label: '2–5 m' },
  { id: '5to10', min: 5, max: 10, label: '5–10 m' },
  { id: 'over10', min: 10, max: null, label: '≥ 10 m' },
];

/** The unfiltered range; every other range narrows to a band. */
export const ANY_DISTANCE_RANGE: DistanceRange = DISTANCE_RANGES[0];

/** Resolve a range by id, falling back to "any" so an unknown id cannot hide every device. */
export function distanceRangeById(id: string): DistanceRange {
  return DISTANCE_RANGES.find((range) => range.id === id) ?? ANY_DISTANCE_RANGE;
}

/**
 * Whether an identity falls inside a range.
 *
 * An identity with no usable reading belongs to no measured band: it is excluded from every
 * range except "any", rather than being assumed near. The distance comes from the smoothed median
 * so the filter agrees with the figure shown in Focus.
 */
export function withinDistanceRange(
  identity: Pick<ObservationIdentityView, 'rssi' | 'rssiSmoothed'>,
  range: DistanceRange,
): boolean {
  if (range.min === null && range.max === null) return true;
  const metres = estimateDistanceMetres(distanceRSSI(identity));
  if (metres === null) return false;
  if (range.min !== null && metres < range.min) return false;
  if (range.max !== null && metres >= range.max) return false;
  return true;
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
