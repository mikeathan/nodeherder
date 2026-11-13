/**
 * Runtime validation for the `observations` WebSocket event (spec 064, FE-02).
 *
 * The payload crosses a trust boundary, so it is checked before it reaches application
 * state. A malformed snapshot rejects the whole payload: a partially applied snapshot would
 * present mixed-generation data. Unknown facet names are ignored rather than rejected, so a
 * newer backend can add facets without breaking an older browser.
 */
import {
  ObservationDecisions,
  ObservationMedia,
  ObservationTrends,
  ObservationClasses,
  type ObservationClassification,
  type ObservationDecision,
  type ObservationIdentityView,
  type ObservationMedium,
  type ObservationSnapshot,
  type ObservationSourceDescriptor,
  type ObservationSourceDiagnostics,
  type ObservationSourceReading,
  type ObservationTrend,
} from '@/types/observations.type';

const maxima: Record<
  'sources' | 'identities' | 'readings' | 'media' | 'facets' | 'text' | 'signal' | 'reason',
  number
> = {
  sources: 64,
  identities: 1000,
  readings: 64,
  media: 8,
  facets: 16,
  text: 512,
  signal: 128,
  // Long enough for a full sentence from the classifier, short enough that a malformed or hostile
  // payload cannot bloat a snapshot.
  reason: 240,
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function boundedString(value: unknown, max = maxima.text): string | null {
  return typeof value === 'string' && value.length <= max ? value : null;
}

/**
 * Truncate explanatory text rather than dropping it.
 *
 * A reason is the only thing that makes a class arguable, so losing it entirely because the backend
 * wrote an unusually long sentence would remove the explanation the operator needs. Truncation
 * degrades the sentence; rejection removes it.
 */
function truncatedString(value: unknown, max: number): string | null {
  if (typeof value !== 'string' || value.length === 0) return null;
  return value.length <= max ? value : value.slice(0, max);
}

function finiteNumber(value: unknown): number | null {
  return typeof value === 'number' && Number.isFinite(value) ? value : null;
}

/** Parse optional delivery counters. Absent counters are valid, so an older backend parses. */
function parseDiagnostics(raw: unknown): ObservationSourceDiagnostics | undefined | null {
  if (raw === undefined || raw === null) return undefined;
  if (!isRecord(raw)) return null;
  const advertisements = finiteNumber(raw.advertisements);
  const decoded = finiteNumber(raw.decoded);
  const dropped = finiteNumber(raw.dropped);
  if (advertisements === null || decoded === null || dropped === null) return null;
  if (advertisements < 0 || decoded < 0 || dropped < 0) return null;
  const lastAdvertisement =
    raw.lastAdvertisement === undefined || raw.lastAdvertisement === null
      ? null
      : boundedString(raw.lastAdvertisement, 64);
  if (raw.lastAdvertisement !== undefined && raw.lastAdvertisement !== null && lastAdvertisement === null) return null;
  return { advertisements, decoded, dropped, lastAdvertisement };
}

/** Parse one source descriptor, or null when it is unusable. */
function parseSource(raw: unknown): ObservationSourceDescriptor | null {
  if (!isRecord(raw)) return null;
  const id = boundedString(raw.id);
  const name = boundedString(raw.name);
  if (!id || name === null) return null;
  const media = parseMedia(raw.media);
  const facets = parseStringList(raw.facets, maxima.facets);
  if (!media || !facets) return null;
  const diagnostics = parseDiagnostics(raw.diagnostics);
  if (diagnostics === null) return null;
  return { id, name, media, facets, ...(diagnostics ? { diagnostics } : {}) };
}

/** Parse a media list, rejecting values outside the declared closed set. */
function parseMedia(raw: unknown): ObservationMedium[] | null {
  if (!Array.isArray(raw) || raw.length > maxima.media) return null;
  const known = new Set<string>(Object.values(ObservationMedia));
  const media: ObservationMedium[] = [];
  for (const value of raw) {
    if (typeof value !== 'string' || !known.has(value)) return null;
    media.push(value as ObservationMedium);
  }
  return media;
}

/** Parse a bounded string list. */
function parseStringList(raw: unknown, max: number): string[] | null {
  if (!Array.isArray(raw) || raw.length > max) return null;
  const values: string[] = [];
  for (const value of raw) {
    const text = boundedString(value, 64);
    if (text === null) return null;
    values.push(text);
  }
  return values;
}

/** Parse per-source readings. */
function parseReadings(raw: unknown): ObservationSourceReading[] | null {
  if (!Array.isArray(raw) || raw.length > maxima.readings) return null;
  const readings: ObservationSourceReading[] = [];
  for (const entry of raw) {
    if (!isRecord(entry)) return null;
    const id = boundedString(entry.id);
    if (!id) return null;
    const rssi = entry.rssi === undefined || entry.rssi === null ? null : finiteNumber(entry.rssi);
    if (entry.rssi !== undefined && entry.rssi !== null && rssi === null) return null;
    readings.push({ id, rssi });
  }
  return readings;
}

function parseDecision(raw: unknown): ObservationDecision | null {
  const known = new Set<string>(Object.values(ObservationDecisions));
  return typeof raw === 'string' && known.has(raw) ? (raw as ObservationDecision) : null;
}

/** Parse one identity, or null when it is unusable. */
function parseIdentity(raw: unknown): ObservationIdentityView | null {
  if (!isRecord(raw)) return null;
  const key = boundedString(raw.key);
  const kind = boundedString(raw.kind, 64);
  const lastSeen = boundedString(raw.lastSeen, 64);
  // firstSeen is optional so a payload from an earlier backend still parses; callers fall back
  // to lastSeen rather than inventing a timestamp.
  const firstSeen = raw.firstSeen === undefined || raw.firstSeen === null ? null : boundedString(raw.firstSeen, 64);
  const decision = parseDecision(raw.decision);
  const sources = parseReadings(raw.sources);
  const packets = finiteNumber(raw.packets);
  if (!key || kind === null || !lastSeen || !decision || !sources || packets === null) return null;
  if (typeof raw.stable !== 'boolean' || Number.isNaN(Date.parse(lastSeen))) return null;
  if (raw.firstSeen !== undefined && raw.firstSeen !== null) {
    if (firstSeen === null || Number.isNaN(Date.parse(firstSeen))) return null;
  }
  const vendor = raw.vendor === undefined ? undefined : finiteNumber(raw.vendor);
  if (raw.vendor !== undefined && vendor === null) return null;
  const rssi = raw.rssi === undefined || raw.rssi === null ? null : finiteNumber(raw.rssi);
  if (raw.rssi !== undefined && raw.rssi !== null && rssi === null) return null;
  // Optional so an older backend payload still parses; callers fall back to the single reading.
  const rssiSmoothed =
    raw.rssiSmoothed === undefined || raw.rssiSmoothed === null ? null : finiteNumber(raw.rssiSmoothed);
  if (raw.rssiSmoothed !== undefined && raw.rssiSmoothed !== null && rssiSmoothed === null) return null;
  // Trend fields are optional so a payload from an earlier backend still parses. An unknown
  // trend value is rejected rather than coerced, because it would become a false claim.
  const trend = parseTrend(raw.trend);
  if (trend === null) return null;
  const trendDeltaDb =
    raw.trendDeltaDb === undefined || raw.trendDeltaDb === null ? undefined : finiteNumber(raw.trendDeltaDb);
  if (raw.trendDeltaDb !== undefined && raw.trendDeltaDb !== null && trendDeltaDb === null) return null;
  const signal = parseSignal(raw.signal);
  if (signal === null) return null;
  // Classification is optional: an older backend simply does not send it, and the viewer then
  // falls back to whatever it can derive locally. An unrecognised class is rejected rather than
  // passed through, because rendering an unknown class would produce a blank badge with no
  // explanation — worse than showing nothing.
  const classification = parseClassification(raw.classification);
  if (classification === null) return null;
  return {
    key,
    kind,
    vendor: vendor === null ? undefined : vendor,
    stable: raw.stable,
    label: boundedString(raw.label, 120) ?? undefined,
    decision,
    firstSeen: firstSeen ?? lastSeen,
    lastSeen,
    packets,
    rssi,
    rssiSmoothed,
    sources,
    ...(trend ? { trend } : {}),
    ...(trendDeltaDb === null || trendDeltaDb === undefined ? {} : { trendDeltaDb }),
    ...(signal ? { signal } : {}),
    ...(classification ? { classification } : {}),
  };
}

/**
 * Parse an advisory classification.
 *
 * Absent is allowed. An unrecognised class is rejected: the alternative is a row whose class the
 * UI cannot name, which reads as a bug. A class the UI does not know yet is a signal that the
 * backend is newer than the frontend, and refusing it keeps the failure visible rather than silent.
 */
function parseClassification(raw: unknown): ObservationClassification | undefined | null {
  if (raw === undefined || raw === null) return undefined;
  if (!isRecord(raw)) return null;

  const known = new Set<string>(Object.values(ObservationClasses));
  const value = boundedString(raw.class, 64);
  if (!value || !known.has(value)) return null;

  const confidence = finiteNumber(raw.confidence);
  const hoursObserved = finiteNumber(raw.hoursObserved);
  if (confidence === null || hoursObserved === null) return null;
  if (typeof raw.showByDefault !== 'boolean') return null;

  return {
    class: value,
    showByDefault: raw.showByDefault,
    confidence: clampUnit(confidence),
    hoursObserved: Math.max(0, Math.round(hoursObserved)),
    ...(raw.provisional === true ? { provisional: true } : {}),
    ...(raw.advisory === true ? { advisory: true } : {}),
    ...(boundedString(raw.candidate, 64) ? { candidate: boundedString(raw.candidate, 64)! } : {}),
    ...(raw.changePending === true ? { changePending: true } : {}),
    ...(finiteNumber(raw.daysConsidered) !== null
      ? { daysConsidered: Math.max(0, Math.round(finiteNumber(raw.daysConsidered)!)) }
      : {}),
    ...(truncatedString(raw.reason, maxima.reason) ? { reason: truncatedString(raw.reason, maxima.reason)! } : {}),
  };
}

/** Bound a value into [0, 1]; a confidence outside that range is a contract violation. */
function clampUnit(value: number): number {
  return value < 0 ? 0 : value > 1 ? 1 : value;
}

/** Parse a trend value. Absent is allowed; an unrecognised value is rejected. */
function parseTrend(raw: unknown): ObservationTrend | undefined | null {
  if (raw === undefined || raw === null) return undefined;
  const known = new Set<string>(Object.values(ObservationTrends));
  return typeof raw === 'string' && known.has(raw) ? (raw as ObservationTrend) : null;
}

/** Parse the bounded signal history used for the sparkline. */
function parseSignal(raw: unknown): number[] | undefined | null {
  if (raw === undefined || raw === null) return undefined;
  if (!Array.isArray(raw) || raw.length > maxima.signal) return null;
  const samples: number[] = [];
  for (const value of raw) {
    const sample = finiteNumber(value);
    if (sample === null) return null;
    samples.push(sample);
  }
  return samples;
}

/**
 * Validate one `observations` payload.
 *
 * Returns null when the payload cannot be trusted, so the caller leaves state untouched
 * rather than applying a partial snapshot.
 */
export function parseObservationSnapshot(payload: unknown): ObservationSnapshot | null {
  if (!isRecord(payload)) return null;
  const at = boundedString(payload.at, 64);
  if (!at || Number.isNaN(Date.parse(at))) return null;
  if (typeof payload.truncated !== 'boolean') return null;
  const arrivals = finiteNumber(payload.arrivals);
  if (arrivals === null || arrivals < 0) return null;
  if (!Array.isArray(payload.sources) || payload.sources.length > maxima.sources) return null;
  if (!Array.isArray(payload.identities) || payload.identities.length > maxima.identities) return null;

  const sources: ObservationSourceDescriptor[] = [];
  for (const raw of payload.sources) {
    const source = parseSource(raw);
    if (!source) return null;
    sources.push(source);
  }
  const identities: ObservationIdentityView[] = [];
  for (const raw of payload.identities) {
    const identity = parseIdentity(raw);
    if (!identity) return null;
    identities.push(identity);
  }
  return { at, truncated: payload.truncated, sources, identities, arrivals };
}
