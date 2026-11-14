/**
 * Presentation helpers for the wireless observation viewer (spec 064).
 *
 * Deliberately separate from `domain/spatial`: the spatial viewer is millimetre-Cartesian with
 * zones and bearings, while this view is range-only and bearing-less. Sharing the canvas would
 * corrupt both, so only the polar maths is reused.
 */
import type {
  ObservationClassification,
  ObservationDecision,
  ObservationIdentityView,
  ObservationSnapshot,
  ObservationSourceDescriptor,
  ObservationTrend,
} from '@/types/observations.type';
import {
  ObservationClasses,
  ObservationDecisions,
  ObservationTrends,
} from '@/types/observations.type';

/**
 * Number of range rings drawn, including the outer edge.
 *
 * Rings are evenly spaced in canvas space on purpose. The projection is logarithmic, so evenly
 * spaced circles are the visually honest choice: a log scale drawn with an absolute ladder
 * (0.5, 1, 3, 10 m) produces circles at uneven radii whose gaps change with every zoom level,
 * which reads as a broken scale rather than a log one. Even spacing makes a doubling of distance
 * occupy a constant amount of screen, which is what a log scale is for.
 */
export const RING_COUNT = 4;

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
export function radialRatio(metres: number, maxRing: number = DEFAULT_RADAR_SPAN): number {
  if (!Number.isFinite(metres) || !Number.isFinite(maxRing) || maxRing <= MIN_PROJECTED_METRES) return 0;
  const clamped = Math.min(Math.max(metres, MIN_PROJECTED_METRES), maxRing);
  const ratio = Math.log10(clamped / MIN_PROJECTED_METRES) / Math.log10(maxRing / MIN_PROJECTED_METRES);
  return Math.min(Math.max(ratio, 0.02), 1);
}

/**
 * Span presets for the radar scale, in metres, widest-last.
 *
 * These are the zoom stops rather than the only values: the viewer steps through them with
 * zoom in/out, so they are ordered and the step is a doubling. A doubling is chosen because the
 * projection is logarithmic, so each stop reveals the same amount of new detail.
 */
export const RADAR_SPANS_M = [1, 2, 3, 5, 10, 20, 30, 50] as const;

export type RadarSpan = (typeof RADAR_SPANS_M)[number];

/** Default span, used when no data determines one. */
export const DEFAULT_RADAR_SPAN: RadarSpan = 30;

/**
 * Next span when zooming one step, or the current one at the limit.
 *
 * `direction` is -1 to zoom in (narrower span, more detail) and 1 to zoom out. Zooming in must
 * shrink the span: a narrower span draws the same space larger, which is the direction the word
 * means and the one an operator expects from a plus/minus pair.
 */
export function stepRadarSpan(current: number, direction: -1 | 1): RadarSpan {
  const presets = RADAR_SPANS_M;
  // Snap an off-ladder span to the nearest preset first, so a value from a future data-driven
  // span still steps predictably rather than jumping to an end.
  let index = presets.findIndex((span) => span >= current);
  if (index < 0) index = presets.length - 1;
  const next = index + direction;
  if (next < 0) return presets[0];
  if (next >= presets.length) return presets[presets.length - 1];
  return presets[next];
}

/**
 * How many times the median device distance the Auto span covers.
 *
 * Scaling to the furthest reading sounds right and looks wrong: RSSI estimates are noisy and a
 * single distant device — a neighbour's meter, a car outside — forces the whole scale wide and
 * crushes everything the operator cares about into the middle. A percentile helps but is still
 * pulled by the tail: with devices from 1 m to 50 m the 90th percentile landed on 50 m and the
 * near cluster occupied about a tenth of the radius.
 *
 * Scaling to a multiple of the *median* frames the typical device whatever the spread. The
 * multiple was tuned against realistic sets rather than picked: at 1.2x the median device sat too
 * near the edge, at 2.5x a tight 1-5 m household was pushed to a 10 m scale it did not need.
 */
export const AUTO_SPAN_MEDIAN_MULTIPLE = 1.5;

/**
 * How far beyond the median a reading can be before it counts as a tail rather than the data.
 *
 * With no tail the scale must cover everything, or a two-device household would clip one of them.
 * Past this ratio the far devices are drawn at the edge and counted, which is the documented
 * behaviour for anything beyond the span.
 */
export const AUTO_SPAN_TAIL_RATIO = 2;

/**
 * Choose a span from the data, framed on the typical device rather than the extremes.
 *
 * Scaling to the data is what makes sub-metre bands legible: on a 30 m scale every nearby device
 * collapses into the middle few percent of the canvas. With no readings the default span is used
 * rather than the narrowest, so an empty radar does not claim a one-metre range.
 */
export function autoRadarSpan(metres: readonly number[]): RadarSpan {
  const finite = metres.filter((value) => Number.isFinite(value) && value > 0).sort((a, b) => a - b);
  if (finite.length === 0) return DEFAULT_RADAR_SPAN;

  // A single device has no distribution, so headroom would only inflate the scale: a device at
  // exactly 2 m belongs on a 2 m scale, not a 3 m one.
  if (finite.length === 1) {
    return RADAR_SPANS_M.find((span) => finite[0] <= span) ?? RADAR_SPANS_M[RADAR_SPANS_M.length - 1];
  }

  const median = medianOf(finite);
  const outermost = finite[finite.length - 1];
  // A coherent set has no meaningful tail, so it is covered completely — otherwise a two-device
  // household would clip one of its devices for no reason.
  const reference =
    outermost <= median * AUTO_SPAN_TAIL_RATIO
      ? Math.max(median * AUTO_SPAN_MEDIAN_MULTIPLE, outermost)
      : median * AUTO_SPAN_MEDIAN_MULTIPLE;

  return RADAR_SPANS_M.find((span) => reference <= span) ?? RADAR_SPANS_M[RADAR_SPANS_M.length - 1];
}

/**
 * Median of a sorted slice.
 *
 * The median rather than the mean: one device at 50 m among forty at 2 m moves a mean to about
 * 3.2 m but leaves the median at 2 m, and it is the typical device the scale should frame.
 */
function medianOf(sorted: readonly number[]): number {
  const middle = Math.floor(sorted.length / 2);
  if (sorted.length % 2 === 1) return sorted[middle];
  return (sorted[middle - 1] + sorted[middle]) / 2;
}

/**
 * One range ring: where to draw it, and what to call it.
 *
 * Both are carried because they must not be derived from each other. The ratio is the exact
 * position and is what makes the circles evenly spaced; the metres value is that position
 * expressed as a distance and is what the label shows. Rounding the metres for a tidy label must
 * never move the circle, which is why the ratio is stored rather than recomputed from the label.
 */
export type RadarRing = {
  /** Position as a fraction of the canvas radius, in (0, 1]. */
  ratio: number;
  /** Distance this ring represents, rounded for display. */
  metres: number;
  /** True for the outermost ring, which is the span edge. */
  outer: boolean;
};

/**
 * Rings for a span, innermost first.
 *
 * Circles are placed at equal fractions of the canvas radius and labelled by *inverting* the
 * projection at each position. Equal spacing is what a log scale should look like: each ring
 * multiplies the distance by the same factor, so a doubling of range is a constant distance on
 * screen at every zoom level.
 *
 * The previous implementation filtered a fixed absolute ladder (0.5, 1, 3, 10 m) instead, which
 * put circles at uneven radii whose gaps changed with every zoom — 23%, 23%, 13% at a 3 m span.
 * That reads as a broken scale, and it was one: an absolute ladder wearing a log projection.
 *
 * The ring at fraction 0 is skipped because it would sit on the centre marker, where its label is
 * unreadable and its circle indistinguishable from the origin.
 */
/**
 * Rings whose labels can be drawn without colliding, innermost first.
 *
 * Greedy placement, working outward: keep a label only if it clears the last kept one by
 * MIN_LABEL_GAP. Dropping the inner rings rather than the outer ones is deliberate — the centre
 * is where labels crowd, and the outer distances are the ones an operator reads to judge range.
 *
 * A ring that loses its label still draws its circle. The circle is the information; the label is
 * an aid, so a crowded scale degrades to fewer numbers rather than to overlapping text.
 *
 * This is exact rather than iterative: because the rings are evenly spaced by construction, one
 * pass in order is sufficient and there is no arrangement to converge on.
 */
export function labelableRings(rings: readonly RadarRing[], minGap = MIN_LABEL_GAP): RadarRing[] {
  // Sort inward-to-outward first rather than depending on the caller's order. The radar draws
  // rings outermost-first so labels paint over circles, and assuming that order here silently
  // reduced four labels to one — the walk started at the innermost ring and then rejected
  // everything outside it. Ordering the input explicitly makes the function correct for either.
  const ordered = [...rings].sort((a, b) => a.ratio - b.ratio);

  // Walk outward from the edge so the outermost ring is always kept: it carries the span, which
  // is the number the operator needs to read the scale.
  const kept: RadarRing[] = [];
  let lastRatio = Number.POSITIVE_INFINITY;
  for (let index = ordered.length - 1; index >= 0; index--) {
    const ring = ordered[index];
    if (lastRatio - ring.ratio < minGap && kept.length > 0) continue;
    kept.push(ring);
    lastRatio = ring.ratio;
  }
  return kept.reverse();
}

export function ringsForSpan(span: number): RadarRing[] {
  if (!Number.isFinite(span) || span <= MIN_PROJECTED_METRES) {
    return [{ ratio: 1, metres: span, outer: true }];
  }

  const rings: RadarRing[] = [];
  for (let index = 1; index < RING_COUNT; index++) {
    const ratio = index / RING_COUNT;
    // Invert radialRatio: the projection is log-linear, so this is its exact inverse.
    const metres = MIN_PROJECTED_METRES * Math.pow(span / MIN_PROJECTED_METRES, ratio);
    rings.push({ ratio, metres: roundRing(metres), outer: false });
  }
  rings.push({ ratio: 1, metres: span, outer: true });
  return rings;
}

/**
 * Minimum gap between two ring labels, as a fraction of the canvas radius.
 *
 * Labels sit on the vertical axis above each circle. At a narrow zoom the circles crowd toward
 * the centre, so labelling every one would overlap the text into an unreadable smear. Twelve
 * percent of the radius is roughly two lines of the label's own height at any span.
 */
export const MIN_LABEL_GAP = 0.12;

/**
 * Round a ring distance to two significant digits.
 *
 * The precision is chosen by measurement, not taste. A ring is drawn at its exact ratio, so the
 * label has to describe that position closely enough that a device at the labelled distance lands
 * on the ring. One or two decimal places failed: at a 3 m span the ring at 75% is truly at 1.52 m,
 * and labelling it "2 m" put a device at 2 m a tenth of the radius away from it — the ring would
 * have been lying about where 2 m is.
 *
 * Two significant digits holds the disagreement to about 1% of the radius at every span tested,
 * while one significant digit did not. Labels therefore read as 0.77 m or 2.4 m rather than always
 * landing on round numbers, which is the correct trade: a slightly awkward number is honest, and a
 * tidy number drawn in the wrong place is not.
 */
function roundRing(metres: number): number {
  if (!Number.isFinite(metres) || metres <= 0) return 0;
  const magnitude = Math.pow(10, Math.ceil(Math.log10(metres)) - 2);
  // Round the quotient first, then scale. Scaling an untidy quotient produces binary-float
  // artefacts — 3.7614 became "3.8000000000000003 m" on the ring label — which read as a bug.
  const scaled = Math.round(metres / magnitude) * magnitude;
  // Two significant digits, measured rather than guessed. A tidier one-decimal label was tried and
  // rejected: at a 3 m span the ring truly at 1.52 m would read "2 m", and a device actually at
  // 2 m is a tenth of the radius away from that circle — the label would lie about where 2 m is.
  // Two digits hold the disagreement under 1% of the radius at every span, so the ring is where
  // its label says it is. "0.53 m" is awkward; a tidy number in the wrong place is worse.
  return metres < 10 ? Math.round(scaled * 100) / 100 : Math.round(scaled);
}

/** Calibrated transmit power at one metre, in dBm, used for the range estimate. */
export const DEFAULT_TX_POWER_AT_1M = -59;

/** Indoor path-loss exponent. 2 is free space; 2.5-3.5 is typical indoors. */
export const DEFAULT_PATH_LOSS_EXPONENT = 2.5;

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

/**
 * Render a BLE company identifier the way people recognise it, e.g. `0x004C` for Apple.
 * Returns an empty string when the identity carries no vendor, so the facet never invents one.
 */
export function vendorHex(vendor: number | undefined): string {
  if (typeof vendor !== 'number' || !Number.isFinite(vendor)) return '';
  return `0x${vendor.toString(16).toUpperCase().padStart(4, '0')}`;
}

/**
 * Readable label for an observation kind.
 *
 * The wire values are `adv`, `ibeacon` and `eddystone`. Only the last two are meaningful jargon —
 * `adv` means nothing to an operator and is uninformative anyway, since it is what every ordinary
 * device sends. It is therefore rendered as "Bluetooth", which says what the medium is rather than
 * naming a packet type the operator did not ask about.
 *
 * An unknown kind is returned as-is rather than hidden: a value this build does not recognise is a
 * newer backend, and showing it keeps that visible.
 */
export function kindLabel(kind: string | undefined): string {
  switch (kind) {
    case 'adv':
      return 'Bluetooth';
    case 'ibeacon':
      return 'iBeacon';
    case 'eddystone':
      return 'Eddystone';
    default:
      return kind ?? 'Unknown';
  }
}

/**
 * Whether a facet is worth rendering as a control.
 *
 * A filter with one value is a control that cannot do anything: it either matches everything
 * (off) or matches everything (on). `adv` alone is exactly that state for an ordinary household,
 * which is why the Kind facet was an inert row of buttons.
 *
 * The facet is not removed from the model, only from the UI: the day a real beacon appears there
 * are two values and the control becomes useful. This is a presentation decision, not a change to
 * what the backend declares.
 */
export function isUsefulFacet(values: readonly string[]): boolean {
  return values.length > 1;
}

/** Display name for an identity, preferring the operator label, then the advertised name. */
export function identityTitle(identity: ObservationIdentityView): string {
  if (identity.label) return identity.label;
  if (identity.kind && identity.kind !== 'adv') return identity.kind;
  return identity.key.length > 24 ? `${identity.key.slice(0, 24)}…` : identity.key;
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
  // Whole metres throughout. Coarse 5 m steps above 10 m were tried and rejected: they made the
  // displayed figure disagree with the device's position on the radar by up to 2 m, and a 1 m move
  // across a step boundary swung the number by 3. The radar plots the raw distance, so the card has
  // to describe the same value — a label that contradicts the picture beside it is worse than one
  // with a digit the estimate cannot really support.
  //
  // Never rounds to zero: a positive estimate under a metre reads `1 m` rather than claiming the
  // device is at the antenna.
  return Math.max(1, Math.round(metres));
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

/**
 * Plain-language presentation for each class.
 *
 * The words matter more than the badge colour. "Infrastructure" is jargon an operator has to
 * translate; "Installed" says the same thing in the vocabulary of their house. The hint carries the
 * reasoning, because a class the operator cannot check is a class they cannot trust.
 *
 * Every entry also states whether the class is hidden by default, so the "N hidden" footer can
 * explain itself without repeating the rule.
 */
export const CLASS_PRESENTATION: Readonly<
  Record<string, { label: string; hint: string; tone: 'installed' | 'yours' | 'outside' | 'passing' | 'learning' }>
> = {
  [ObservationClasses.Infrastructure]: {
    label: 'Installed',
    hint: 'Always here and never moves, like a hub or a TV on standby.',
    tone: 'installed',
  },
  [ObservationClasses.Personal]: {
    label: 'Yours',
    hint: 'Comes and goes and its signal changes, like a phone or a watch.',
    tone: 'yours',
  },
  [ObservationClasses.Neighbour]: {
    label: 'Outside',
    hint: 'Always much weaker than your own devices, so probably a neighbour.',
    tone: 'outside',
  },
  [ObservationClasses.Transient]: {
    label: 'Passing',
    hint: 'Only heard briefly, so probably someone walking or driving past.',
    tone: 'passing',
  },
  [ObservationClasses.Learning]: {
    label: 'Learning',
    hint: 'Not enough history yet to say what this is.',
    tone: 'learning',
  },
  [ObservationClasses.ProvisionalMobile]: {
    label: 'Someone’s device',
    hint: 'Its address rotates, so it belongs to a person rather than the building.',
    tone: 'yours',
  },
  [ObservationClasses.ProvisionalNeighbour]: {
    label: 'Probably outside',
    hint: 'Never heard strongly, so probably beyond the room.',
    tone: 'outside',
  },
};

/** Presentation for a class, falling back to learning rather than showing nothing. */
export function classPresentation(classification: ObservationClassification | undefined) {
  if (!classification) return CLASS_PRESENTATION[ObservationClasses.Learning];
  return CLASS_PRESENTATION[classification.class] ?? CLASS_PRESENTATION[ObservationClasses.Learning];
}

/**
 * Whether a class is a guess rather than a verdict.
 *
 * Two things make a class provisional: the backend says so, or the class is one of the pre-filter
 * values, which only ever exist before any history does. Both are rendered differently, because a
 * guess presented as a fact is how an operator learns to distrust the whole feature.
 */
export function isProvisional(classification: ObservationClassification | undefined): boolean {
  if (!classification) return true;
  if (classification.provisional === true) return true;
  return (
    classification.class === ObservationClasses.ProvisionalMobile ||
    classification.class === ObservationClasses.ProvisionalNeighbour
  );
}

/**
 * One line describing the evidence behind a class.
 *
 * It prefers the backend's own wording, because that reason was written with the numbers in hand.
 * A locally composed fallback exists so a class from an older backend is still explainable.
 */
export function classificationEvidence(classification: ObservationClassification | undefined): string {
  if (!classification) return '';
  if (classification.reason) return classification.reason;
  if (classification.class === ObservationClasses.Learning) {
    return `Learning — ${classification.hoursObserved} of 24 hours observed.`;
  }
  return '';
}

/**
 * The learning progress line, shown instead of a bare class while evidence accumulates.
 *
 * A slow classifier that shows nothing looks broken; showing how far it has to go is the difference
 * between "still working" and "gave up".
 */
export function learningProgress(classification: ObservationClassification | undefined): string | null {
  if (!classification) return null;
  if (classification.class !== ObservationClasses.Learning) return null;
  return `${Math.min(classification.hoursObserved, 24)} of 24 hours`;
}

/**
 * Whether an identity belongs in the default view.
 *
 * The backend decides this and sends it, so the rule lives in one place. When the field is absent —
 * an older backend — everything is shown rather than guessed at, because hiding a device the
 * operator wants is the worse failure.
 */
export function shownByDefault(identity: ObservationIdentityView): boolean {
  return identity.classification?.showByDefault !== false;
}

