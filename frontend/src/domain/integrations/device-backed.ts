/**
 * Integration types whose availability comes from enrolled devices rather than a remote provider.
 *
 * Most integration types describe a service the hub fetches from: weather, bin collections, fuel
 * prices. Wireless observation is different — the hub is not fetching anything, it is reading a
 * radio through a device you already own. Giving it a provider definition would mean inventing a
 * remote endpoint and a fetch loop that duplicate the observation pipeline.
 *
 * So the type is real and appears in the catalog with the others, but it is *assembled here* from
 * the device list instead of from the backend's provider repository. Two reasons this belongs on
 * the frontend:
 *
 *  1. The composition layer is already the only place that names installed adapters
 *     (`integrations/registry.ts`), so this follows the existing pattern rather than inventing one.
 *  2. It keeps the backend integration runtime, which is a pure persisted-instance read model, from
 *     depending on live device state. A remote-service catalog should not fail because a device is
 *     offline.
 *
 * The trade-off: a consumer reading the raw backend catalog will not see this type. That is
 * acceptable because the catalog is a presentation surface — nothing dispatches work by reading it.
 */
import type { ConnectionCatalog, ConnectionCatalogType } from '@/types/integrations';
import type { Device } from '@/types/device';
import { supportsObservation } from '@/domain/observations/capability';
import { RouteName } from '@/types/router';

/**
 * Type id for wireless observation.
 *
 * Not exported: it is compared only inside this module, and the entry card and home tile reach the
 * feature through `OBSERVATION_ROUTE` instead. An exported id invites a second place to compare it,
 * which is how provider-name gating crept into this codebase before.
 */
const OBSERVATION_INTEGRATION_TYPE = 'wireless_observation';

/** Where the radar lives, so every entry point agrees without repeating the route. */
export const OBSERVATION_ROUTE = RouteName.Wireless;

/**
 * Devices that can act as an observation source, in a stable order.
 *
 * Sorting by name rather than by device id, because the card names them and the operator reads
 * those names. `localeCompare` is used through `undefined` locale so the order does not change with
 * the browser's language — a list that reorders between machines is a bug, not a preference.
 */
export function observationDevices(devices: readonly Device[] | undefined): Device[] {
  if (!devices || devices.length === 0) return [];
  return devices
    .filter((device) => supportsObservation(device))
    .sort((left, right) => left.friendly_name.localeCompare(right.friendly_name, undefined));
}

/**
 * The device-backed type, present only while a capable device is enrolled.
 *
 * Returning `null` rather than an unavailable entry is deliberate: an integration card that exists
 * to say "you have no such device" is noise in a list where every other card is something the
 * operator chose to add. When the device is enrolled the card appears on its own.
 */
export function observationCatalogType(
  devices: readonly Device[] | undefined,
): ConnectionCatalogType | null {
  const sources = observationDevices(devices);
  if (sources.length === 0) return null;

  return {
    id: OBSERVATION_INTEGRATION_TYPE,
    name: 'Wireless observation',
    description:
      'Read nearby Bluetooth devices through an enrolled sensor. Shows distance and movement, and classifies what it hears.',
    configured_count: sources.length,
    // Device-backed types have no refresh lifecycle, so they are always ready once a source exists.
    // There is nothing to fetch, retry or schedule.
    worst_status: 'ready',
  };
}

/**
 * Merge device-backed types into a catalog without touching the backend's.
 *
 * The result is a new object rather than a mutated one: the argument may be a store getter's value,
 * and mutating it would change what every other consumer sees.
 */
export function withDeviceBackedTypes(
  catalog: ConnectionCatalog | null,
  devices: readonly Device[] | undefined,
): ConnectionCatalog | null {
  if (!catalog) return catalog;
  const type = observationCatalogType(devices);
  if (!type) return catalog;
  // A backend that already describes this type wins, so this cannot duplicate an entry once the
  // catalog gains real support for it.
  if (catalog.types.some((entry) => entry.id === type.id)) return catalog;
  return { ...catalog, types: [...catalog.types, type] };
}
