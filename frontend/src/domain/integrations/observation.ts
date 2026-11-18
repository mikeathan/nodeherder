/**
 * Wireless observation as an ordinary integration.
 *
 * It used to be device-backed: the type appeared in the catalog whenever a capable device was
 * enrolled, was always present on the board, and could only be hidden with a browser-local
 * preference. That made it impossible to remove and impossible to disable, which is not how any
 * other integration behaves.
 *
 * It is now a real provider with a real instance, added and removed like the rest. This module
 * holds only what the frontend still needs to know about the type: its id, where the radar lives,
 * and which enrolled devices can supply it.
 */
import type { Device } from '@/types/device';
import { supportsObservation } from '@/domain/observations/capability';
import { RouteName } from '@/types/router';

/**
 * Type id for wireless observation.
 *
 * Must match the backend capability's integration type, because that is what the catalog groups
 * providers by. Kept as one constant so no other module compares a string literal.
 */
export const OBSERVATION_INTEGRATION_TYPE = 'wireless_observation';

/** Provider key for the single local provider of this type. */
export const OBSERVATION_PROVIDER_KEY = `${OBSERVATION_INTEGRATION_TYPE}/hub`;

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

/** The counts the census summary needs; a structural subset of the published resource. */
export interface ObservationCensusCounts {
  identities: number;
  hidden: number;
  sources: readonly { advertisements: number; decoded: number }[];
}

/**
 * Whether any source is hearing devices right now.
 *
 * Counters are lifetime totals for the current run, so a non-zero count means the source delivered.
 * A source with no counters is connected but has produced nothing.
 */
export function censusReporting(counts: ObservationCensusCounts): boolean {
  return counts.sources.some((source) => source.advertisements > 0);
}

/**
 * Whether advertisements are arriving but none is decoding into an identity.
 *
 * This is a distinct fault from silence, with a different fix: the proxy is talking, so the radio
 * and the link are fine, and the problem is in what it sends or how it is decoded. Collapsing it
 * into "no devices heard" sent the operator looking at the radio.
 */
export function censusUndecoded(counts: ObservationCensusCounts): boolean {
  return counts.identities === 0 && counts.sources.some((source) => source.advertisements > 0 && source.decoded === 0);
}

/**
 * The census headline, and the reason the list is not growing.
 *
 * Both live here rather than in the component because the wording is the feature, and because the
 * tense has to follow the data. Identities outlive their source by up to ten minutes — deliberately,
 * so a proxy restarting does not blank the view — which means a census can legitimately hold devices
 * while nothing is reporting. The page used to state both facts in the present tense, producing
 * "23 devices in range" directly above "No sensor is reporting": two claims about the same moment,
 * one of them false.
 *
 * So the headline says "heard earlier" when nothing is reporting, and the note explains what that
 * means. The two are computed together so they cannot drift into disagreeing again.
 */
export function censusSummary(counts: ObservationCensusCounts): { headline: string; note: string | null } {
  const { identities, hidden } = counts;
  const reporting = censusReporting(counts);
  const shown = identities - hidden;

  // The noun follows the number the phrase actually counts, which is `identities` in every branch.
  // The hidden branch used to hardcode the plural, so a single device rendered "0 of 1 devices
  // shown" directly above "1 device is hidden by the class filter".
  const noun = identities === 1 ? 'device' : 'devices';
  let headline: string;
  if (identities === 0) {
    headline = 'No devices heard';
  } else if (hidden > 0) {
    headline = `${shown} of ${identities} ${noun} ${reporting ? 'shown' : 'heard earlier'}`;
  } else {
    headline = `${identities} ${noun} ${reporting ? 'in range' : 'heard earlier'}`;
  }

  return { headline, note: censusNote(counts, reporting) };
}

/** Why the list is not growing, phrased for the state that actually applies. */
function censusNote(counts: ObservationCensusCounts, reporting: boolean): string | null {
  // Traffic arriving with nothing decoded is a fault even though the source is "reporting", so it is
  // checked before the success case rather than falling through to no note at all.
  if (censusUndecoded(counts)) {
    return 'Advertisements are arriving but none could be decoded into a device. Check that this sensor runs bluetooth_proxy with a working Bluetooth stack.';
  }
  if (reporting) return null;
  if (counts.sources.length === 0) {
    return counts.identities > 0
      ? 'No sensor is reporting now. These devices were heard before it stopped, and they drop off ten minutes after the last reading.'
      : 'No sensor is reporting.';
  }
  if (counts.sources.every((source) => source.advertisements === 0)) {
    return 'The sensor is connected but has not reported any advertisements yet. Confirm it runs bluetooth_proxy in its ESPHome configuration.';
  }
  return 'The sensor has not reported recently. These devices were heard earlier, and they drop off ten minutes after the last reading.';
}
