/**
 * Wireless observation capability predicates (spec 064, AC-09, AC-10).
 *
 * The point of these tests is that a new source kind needs no UI change: capability is read
 * from a declared field, never from a provider name.
 */
import { OBSERVATION_CAPABILITY, supportsObservation } from '@/domain/observations/capability';
import type { Device } from '@/types/device';

function device(overrides: Partial<Device> = {}): Device {
  return {
    id: 'device-1',
    friendly_name: 'Hall proxy',
    description: '',
    connection_type: 'esphome',
    power_source: 'mains',
    last_seen: '2026-11-11T10:00:00Z',
    availability: 'online',
    exposes: {},
    properties: {},
    ...overrides,
  } as Device;
}

describe('supportsObservation', () => {
  test('accepts a device that declares the capability', () => {
    expect(supportsObservation(device({ capabilities: [OBSERVATION_CAPABILITY] }))).toBe(true);
  });

  test('rejects a device with no declared capabilities', () => {
    expect(supportsObservation(device())).toBe(false);
  });

  test('rejects a device declaring unrelated capabilities', () => {
    expect(supportsObservation(device({ capabilities: ['something.else'] }))).toBe(false);
  });

  test('handles absent devices', () => {
    expect(supportsObservation(undefined)).toBe(false);
    expect(supportsObservation(null)).toBe(false);
  });

  test('is provider-agnostic: a future provider declaring the capability is accepted', () => {
    // A WiFi-probe or USB source needs no code change here, which is the whole point.
    expect(supportsObservation(device({ provider: 'wifi-probe', capabilities: [OBSERVATION_CAPABILITY] }))).toBe(true);
    expect(supportsObservation(device({ provider: 'esphome' }))).toBe(false);
  });
});
