import {
  censusReporting,
  censusSummary,
  observationDevices,
  OBSERVATION_INTEGRATION_TYPE,
  OBSERVATION_PROVIDER_KEY,
} from '@/domain/integrations/observation';
import { OBSERVATION_CAPABILITY } from '@/domain/observations/capability';
import type { Device } from '@/types/device';

/**
 * These cover the one thing this module now owns: which devices can supply observation, and the
 * stable ordering the card and the picker both rely on.
 *
 * The lifecycle itself — add, enable, disable, remove — is not testable here. It is an instance
 * lifecycle now, shared with every other integration, so it is covered by the backend tests that own
 * the runtime rather than duplicated as a frontend mock.
 */
function device(overrides: Partial<Device> & { friendly_name: string }): Device {
  return {
    device_id: overrides.friendly_name,
    name: overrides.friendly_name,
    ...overrides,
  } as Device;
}

describe('observation integration identity', () => {
  it('uses the type id the backend publishes, since the catalog groups providers by it', () => {
    expect(OBSERVATION_INTEGRATION_TYPE).toBe('wireless_observation');
    expect(OBSERVATION_PROVIDER_KEY).toBe('wireless_observation/hub');
  });
});

describe('observationDevices', () => {
  it('returns nothing for a missing or empty device list', () => {
    expect(observationDevices(undefined)).toEqual([]);
    expect(observationDevices([])).toEqual([]);
  });

  it('sorts capable devices by name and drops the rest', () => {
    const devices = [
      device({ friendly_name: 'zulu proxy', capabilities: [OBSERVATION_CAPABILITY] }),
      device({ friendly_name: 'alpha proxy', capabilities: [OBSERVATION_CAPABILITY] }),
      device({ friendly_name: 'plain sensor' }),
    ];
    expect(observationDevices(devices).map((entry) => entry.friendly_name)).toEqual([
      'alpha proxy',
      'zulu proxy',
    ]);
  });
});

/**
 * The census summary.
 *
 * These exist because the page shipped a self-contradiction: "23 devices in range" directly above
 * "No sensor is reporting". Identities outlive their source by up to ten minutes by design, so both
 * halves were individually true of different moments while reading as claims about the same one.
 * The wording is the feature, so it is tested rather than reviewed.
 */
const reporting = [{ advertisements: 134, decoded: 130 }];
const neverDelivered = [{ advertisements: 0, decoded: 0 }];

describe('censusSummary', () => {
  it('states devices in range when a source is reporting', () => {
    const { headline, note } = censusSummary({ identities: 23, hidden: 0, sources: reporting });
    expect(headline).toBe('23 devices in range');
    expect(note).toBeNull();
  });

  it('never claims devices are in range when nothing is reporting', () => {
    // The exact reported defect: devices present, source gone. The count survives because it is
    // still true of the recent past, but the claim has to be in the past tense.
    const { headline, note } = censusSummary({ identities: 23, hidden: 0, sources: [] });
    expect(headline).toBe('23 devices heard earlier');
    expect(headline).not.toContain('in range');
    expect(note).toContain('No sensor is reporting now');
  });

  it('distinguishes a connected-but-silent sensor from an absent one', () => {
    // Both have no devices, and the fixes are different: one is a configuration problem, the other
    // is nothing enrolled.
    const silent = censusSummary({ identities: 0, hidden: 0, sources: neverDelivered });
    expect(silent.headline).toBe('No devices heard');
    expect(silent.note).toContain('bluetooth_proxy');

    const none = censusSummary({ identities: 0, hidden: 0, sources: [] });
    expect(none.note).toBe('No sensor is reporting.');
  });

  it('keeps the past tense for hidden devices too', () => {
    const { headline } = censusSummary({ identities: 23, hidden: 3, sources: [] });
    expect(headline).toBe('20 of 23 devices heard earlier');
  });

  it('uses the singular for one device', () => {
    expect(censusSummary({ identities: 1, hidden: 0, sources: reporting }).headline).toBe('1 device in range');
  });

  it('keeps the singular when the only device is hidden', () => {
    // The hidden branch used to hardcode "devices", so this rendered "0 of 1 devices shown" right
    // above "1 device is hidden by the class filter".
    const { headline } = censusSummary({ identities: 1, hidden: 1, sources: reporting });
    expect(headline).toBe('0 of 1 device shown');
  });

  it('explains traffic that arrives but does not decode', () => {
    // Distinct from silence with a different fix: the radio and link are fine, the decoding is not.
    // Falling through to no note left the page saying "No devices heard" with nothing else.
    const { headline, note } = censusSummary({
      identities: 0,
      hidden: 0,
      sources: [{ advertisements: 50, decoded: 0 }],
    });
    expect(headline).toBe('No devices heard');
    expect(note).toContain('none could be decoded');
  });
});

describe('censusReporting', () => {
  it('is true only when a source has delivered', () => {
    expect(censusReporting({ identities: 0, hidden: 0, sources: reporting })).toBe(true);
    expect(censusReporting({ identities: 5, hidden: 0, sources: neverDelivered })).toBe(false);
    expect(censusReporting({ identities: 5, hidden: 0, sources: [] })).toBe(false);
  });
});
