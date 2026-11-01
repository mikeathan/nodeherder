/*
 * Generic device identity presentation. The icon is inferred from the device's
 * primary measurement capabilities, never from a model name or a single
 * auxiliary entity such as a status LED.
 */
import type { Device, Expose } from '@/types/device';
import { ExposeAttributeKeys, ExposeCategories } from '@/types/device.type';
import { exposeSemanticName } from '@/domain/exposes';

/** Stable semantic icon families; the UI registry owns the actual glyphs. */
export const DeviceIconKinds = {
  Motion: 'motion',
  Smoke: 'smoke',
  Contact: 'contact',
  Light: 'light',
  Power: 'power',
  Remote: 'remote',
  Climate: 'climate',
  Air: 'air',
  Generic: 'generic',
} as const;
export type DeviceIconKind = (typeof DeviceIconKinds)[keyof typeof DeviceIconKinds];

/** Semantic names that indicate presence/radar tracking. */
const MOTION_NAMES = ['presence', 'occupancy', 'motion', 'target_distance'] as const;
/** Semantic names that indicate climate sensing. */
const CLIMATE_NAMES = ['temperature', 'humidity'] as const;
/** Semantic names that indicate air-quality sensing. */
const AIR_NAMES = ['pm25', 'pm10', 'pm1', 'voc', 'co2'] as const;

/** Measurement entities describe identity better than config/diagnostic support entities. */
function identityEntities(device: Device): Expose[] {
  const entities = Object.values(device.exposes ?? {});
  const measurements = entities.filter((entity) => entity.category === ExposeCategories.Measurement);
  return measurements.length ? measurements : entities;
}

/** Normalizes a discovered display label into a semantic lookup token. */
function labelToken(expose: Expose): string {
  const label = expose.attributes?.[ExposeAttributeKeys.Label] ?? expose.description;
  return typeof label === 'string' ? label.trim().toLowerCase().replace(/[^a-z0-9]+/g, '_') : '';
}

/** Collects semantic names from device classes and discovered labels. */
function semanticNames(entities: readonly Expose[]): Set<string> {
  const names = new Set<string>();
  for (const entity of entities) {
    names.add(exposeSemanticName(entity));
    const token = labelToken(entity);
    if (token) names.add(token);
  }
  return names;
}

/** Selects a device icon family from discovered capabilities, with presence ahead of auxiliary lights. */
export function deviceIconKind(device: Device): DeviceIconKind {
  const entities = identityEntities(device);
  const names = semanticNames(entities);
  const has = (name: string) => names.has(name);
  const hasAny = (candidates: readonly string[]) => candidates.some(has);

  if (hasAny(MOTION_NAMES)) return DeviceIconKinds.Motion;
  if (has('smoke') || has('alarm')) return DeviceIconKinds.Smoke;
  if (has('contact')) return DeviceIconKinds.Contact;
  const isLight =
    entities.some((entity) => entity.attributes?.[ExposeAttributeKeys.EntityKind] === 'light') ||
    has('brightness') ||
    has('color_temp');
  if (isLight) return DeviceIconKinds.Light;
  if (has('state')) return DeviceIconKinds.Power;
  if ([...names].some((name) => name.startsWith('action'))) return DeviceIconKinds.Remote;
  if (hasAny(CLIMATE_NAMES)) return DeviceIconKinds.Climate;
  if (hasAny(AIR_NAMES)) return DeviceIconKinds.Air;
  return DeviceIconKinds.Generic;
}
