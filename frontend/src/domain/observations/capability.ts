/**
 * Wireless observation capability predicates (spec 064, FR-07).
 *
 * These read a declared capability rather than a provider name, matching the existing
 * `supportsSpatialMapping` precedent. That is what keeps a future WiFi-probe or USB source
 * from requiring changes to navigation, the viewer or its filters.
 */
import type { Device } from '@/types/device';

/** Capability a device declares when it can observe at least one radio medium. */
export const OBSERVATION_CAPABILITY = 'wireless.observe';

/**
 * Whether a device declares the observation capability.
 *
 * An absent `capabilities` array means "not declared", which is treated as unsupported. This
 * is deliberately not a provider check: provider names are what made the ESPHome-specific
 * gates a maintenance problem.
 */
export function supportsObservation(device: Device | undefined | null): boolean {
  if (!device) return false;
  return (device.capabilities ?? []).includes(OBSERVATION_CAPABILITY);
}

/** Whether any enrolled device offers an observation source, which gates the nav item. */
export function hasObservationSources(devices: readonly Device[] | undefined): boolean {
  if (!devices || devices.length === 0) return false;
  return devices.some((device) => supportsObservation(device));
}
