/*
 * Resolution of the hub-derived presence device (spec 041). The hub owns zone occupancy and
 * counts for automations; the browser reads those entities so the screen cannot disagree with
 * the automation. Nothing here computes presence — it only locates and normalises values that
 * arrived as ordinary device exposes.
 */
import type { Device, Expose } from '@/types/device';

/** Derived device identifiers mirror the hub's `presence-<source id>` convention. */
const derivedPrefix = 'presence-';

/** Global derived entity names; zone entities are discovered by their `zoneId` attribute. */
const derivedEntity = {
  occupied: 'occupied',
  targetCount: 'target_count',
  lastZone: 'last_zone',
  emptySeconds: 'empty_seconds',
} as const;

/** Zone signal resolved from derived entities; null means the value was absent or unusable. */
export type DerivedZoneSignal = { occupied: boolean | null; count: number | null; dwell: number | null };

/** Global derived presence values for one source device. */
export type DerivedPresenceValues = {
  occupied: boolean | null;
  targetCount: number | null;
  lastZone: string | null;
  emptySeconds: number | null;
};

/** Stable identifier of the derived device for one source device. */
export function derivedPresenceDeviceId(sourceId: string): string {
  return `${derivedPrefix}${sourceId}`;
}

function entityList(device: Device | undefined): Expose[] {
  return device ? Object.values(device.exposes ?? {}) : [];
}

function booleanValue(entity: Expose | undefined): boolean | null {
  return entity && typeof entity.data === 'boolean' ? entity.data : null;
}

function numberValue(entity: Expose | undefined): number | null {
  return entity && typeof entity.data === 'number' && Number.isFinite(entity.data) ? entity.data : null;
}

function textValue(entity: Expose | undefined): string | null {
  if (!entity || typeof entity.data !== 'string' || !entity.data.trim()) return null;
  return entity.data;
}

/** Resolves every zone's derived occupancy, count and dwell, keyed by zone id. */
export function derivedZoneSignals(device: Device | undefined): Map<string, DerivedZoneSignal> {
  const signals = new Map<string, DerivedZoneSignal>();
  for (const entity of entityList(device)) {
    const zoneId = entity.attributes?.zoneId;
    if (typeof zoneId !== 'string' || !zoneId) continue;
    const current = signals.get(zoneId) ?? { occupied: null, count: null, dwell: null };
    if (entity.name.endsWith('_occupied')) current.occupied = booleanValue(entity);
    else if (entity.name.endsWith('_count')) current.count = numberValue(entity);
    else if (entity.name.endsWith('_dwell_seconds')) current.dwell = numberValue(entity);
    signals.set(zoneId, current);
  }
  return signals;
}

/** Resolves the derived device's global presence values. */
export function derivedPresenceValues(device: Device | undefined): DerivedPresenceValues {
  const byName = new Map(entityList(device).map((entity) => [entity.name, entity]));
  return {
    occupied: booleanValue(byName.get(derivedEntity.occupied)),
    targetCount: numberValue(byName.get(derivedEntity.targetCount)),
    lastZone: textValue(byName.get(derivedEntity.lastZone)),
    emptySeconds: numberValue(byName.get(derivedEntity.emptySeconds)),
  };
}
