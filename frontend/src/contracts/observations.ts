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
  type ObservationDecision,
  type ObservationIdentityView,
  type ObservationMedium,
  type ObservationSnapshot,
  type ObservationSourceDescriptor,
  type ObservationSourceDiagnostics,
  type ObservationSourceReading,
  type ObservationTrend,
} from '@/types/observations.type';

const maxima: Record<'sources' | 'identities' | 'readings' | 'media' | 'facets' | 'text' | 'signal', number> = {
  sources: 64,
  identities: 1000,
  readings: 64,
  media: 8,
  facets: 16,
  text: 512,
  signal: 128,
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function boundedString(value: unknown, max = maxima.text): string | null {
  return typeof value === 'string' && value.length <= max ? value : null;
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
  // Trend fields are optional so a payload from an earlier backend still parses. An unknown
  // trend value is rejected rather than coerced, because it would become a false claim.
  const trend = parseTrend(raw.trend);
  if (trend === null) return null;
  const trendDeltaDb =
    raw.trendDeltaDb === undefined || raw.trendDeltaDb === null ? undefined : finiteNumber(raw.trendDeltaDb);
  if (raw.trendDeltaDb !== undefined && raw.trendDeltaDb !== null && trendDeltaDb === null) return null;
  const signal = parseSignal(raw.signal);
  if (signal === null) return null;
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
    sources,
    ...(trend ? { trend } : {}),
    ...(trendDeltaDb === null || trendDeltaDb === undefined ? {} : { trendDeltaDb }),
    ...(signal ? { signal } : {}),
  };
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
