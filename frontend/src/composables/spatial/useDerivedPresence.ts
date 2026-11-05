import { computed } from 'vue';
import type { Device } from '@/types/device';
import {
  derivedPresenceDeviceId,
  derivedPresenceValues,
  derivedZoneSignals,
  type DerivedPresenceValues,
} from '@/domain/spatial/derived';
import { useHub } from '../useHub';

const EMPTY_VALUES: DerivedPresenceValues = { occupied: null, targetCount: null, lastZone: null, emptySeconds: null };

/**
 * Reads the derived presence device the hub publishes for a source (spec 041). The presence
 * tab uses it as the authority for zone occupancy and counts; when the hub has not derived
 * anything the caller falls back to the geometric preview and labels it as an overlay.
 */
export function useDerivedPresence(device: () => Device | undefined) {
  const { findDevice, statusOf, isConnected } = useHub();
  const derived = computed<Device | undefined>(() => {
    const source = device();
    return source ? findDevice(derivedPresenceDeviceId(source.id)) : undefined;
  });
  const online = computed(() => !!derived.value && isConnected.value && statusOf(derived.value) === 'online');
  const zones = computed(() => derivedZoneSignals(derived.value));
  const values = computed(() => (derived.value ? derivedPresenceValues(derived.value) : EMPTY_VALUES));
  return { derived, online, zones, values };
}
