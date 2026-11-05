/*
 * Persisted zone activity derived from the hub's recorded presence entities. Raw positions are
 * not stored, so this reports occupancy/count history only; unknown time is kept unknown
 * instead of being counted as empty.
 */
import type { Device, Expose } from '@/types/device';
import type { HistoryRequest, HistorySeries } from '@/domain/history/types';
import type { SpatialZone } from '@/types/spatial';

export type ZoneHistoryEntities = { occupied?: string; count?: string };

/** Cells used by each zone timeline. */
export const ACTIVITY_BUCKETS = 48;

export type ZoneActivity = {
  zoneId: string;
  zoneName: string;
  sharePct: number | null;
  transitions: number;
  peakCount: number | null;
  lastOccupiedAt: number | null;
  partial: boolean;
  buckets: (number | null)[];
};

/** Resolves the derived zone entities for the workspace zones, by their `zoneId` attribute. */
export function zoneHistoryEntities(
  device: Device | undefined,
  zones: readonly SpatialZone[]
): Map<string, ZoneHistoryEntities> {
  const wanted = new Set(zones.map((zone) => zone.id));
  const result = new Map<string, ZoneHistoryEntities>();
  for (const entity of Object.values(device?.exposes ?? {})) {
    const zoneId = entity.attributes?.zoneId;
    if (typeof zoneId !== 'string' || !wanted.has(zoneId)) continue;
    const current = result.get(zoneId) ?? {};
    if (entity.name.endsWith('_occupied')) current.occupied = entity.name;
    else if (entity.name.endsWith('_count')) current.count = entity.name;
    result.set(zoneId, current);
  }
  return result;
}

/** Builds the history request for every mapped zone entity, or null when there is none. */
export function zoneActivityRequest(
  device: Device | undefined,
  zones: readonly SpatialZone[],
  from: number,
  to: number
): HistoryRequest | null {
  const entities = [...zoneHistoryEntities(device, zones).values()].flatMap((entry) =>
    [entry.occupied, entry.count].filter((name): name is string => !!name)
  );
  if (!entities.length) return null;
  return { deviceId: device!.id, exposes: [...new Set(entities)], from, to };
}

type BooleanSample = { at: number; value: boolean };

function booleanSamples(series: HistorySeries | undefined): BooleanSample[] {
  return (series?.data ?? [])
    .filter((point) => typeof point.value === 'boolean' && Number.isFinite(point.timestamp))
    .map((point) => ({ at: point.timestamp, value: point.value as boolean }));
}

/** Share of known time that was occupied inside [start, end); null when nothing is known. */
export function occupancyShare(samples: readonly BooleanSample[], start: number, end: number): {
  share: number | null;
  trueMs: number;
  knownMs: number;
} {
  if (end <= start) return { share: null, trueMs: 0, knownMs: 0 };
  let trueMs = 0;
  let knownMs = 0;
  for (let index = 0; index < samples.length; index++) {
    const sample = samples[index];
    const next = samples[index + 1]?.at ?? end;
    const intervalStart = Math.max(start, sample.at);
    const intervalEnd = Math.min(end, next);
    if (intervalEnd <= intervalStart) continue;
    knownMs += intervalEnd - intervalStart;
    if (sample.value) trueMs += intervalEnd - intervalStart;
  }
  return { share: knownMs > 0 ? (trueMs / knownMs) * 100 : null, trueMs, knownMs };
}

/** Computes per-zone activity and bucket timelines from returned series, never assuming values. */
export function buildZoneActivity(
  series: readonly HistorySeries[],
  zones: readonly SpatialZone[],
  entities: Map<string, ZoneHistoryEntities>,
  from: number,
  to: number,
  buckets = ACTIVITY_BUCKETS
): ZoneActivity[] {
  const byName = new Map(series.map((entry) => [entry.name, entry]));
  const bucketMs = (to - from) / buckets;
  return zones.map((zone) => {
    const mapped = entities.get(zone.id);
    const occupiedSeries = mapped?.occupied ? byName.get(mapped.occupied) : undefined;
    const samples = booleanSamples(occupiedSeries);
    const overall = occupancyShare(samples, from, to);
    let transitions = 0;
    for (let index = 1; index < samples.length; index++) {
      if (samples[index].value !== samples[index - 1].value) transitions++;
    }
    const countSeries = mapped?.count ? byName.get(mapped.count) : undefined;
    const counts = (countSeries?.data ?? [])
      .filter((point) => typeof point.value === 'number' && Number.isFinite(point.value))
      .map((point) => point.value as number);
    const bucketShares: (number | null)[] = [];
    for (let bucket = 0; bucket < buckets; bucket++) {
      const start = from + bucket * bucketMs;
      bucketShares.push(occupancyShare(samples, start, Math.min(to, start + bucketMs)).share);
    }
    const firstSample = samples[0]?.at;
    const lastTrue = samples.filter((sample) => sample.value).at(-1)?.at ?? null;
    return {
      zoneId: zone.id,
      zoneName: zone.name,
      sharePct: overall.share,
      transitions,
      peakCount: counts.length ? Math.max(...counts) : null,
      lastOccupiedAt: lastTrue,
      partial: !!occupiedSeries?.partial || firstSample === undefined || firstSample > from,
      buckets: bucketShares,
    };
  });
}
