import { parseCatalog, parseInstance } from '@/contracts/integrations';
import { parseObservationCensus } from '@/contracts/integrations/observation';

/**
 * The census parser exists because `parseResource` dispatches on contract version and previously
 * fell through to the waste parser for anything it did not recognise. That failure was silent in the
 * UI — the instance was created and the card never appeared — so it is pinned here.
 */
const census = {
  contract_version: 'observation_census/v1',
  resource_id: 'resource-1',
  revision: 1,
  identities: 3,
  hidden: 1,
  sources: [{ id: 'ble-1', name: 'Hallway', advertisements: 12, decoded: 9, dropped: 3 }],
  observed_at: '2026-11-13T12:00:00Z',
};

describe('parseObservationCensus', () => {
  it('accepts a census the hub publishes', () => {
    const parsed = parseObservationCensus(census);
    expect(parsed.identities).toBe(3);
    expect(parsed.hidden).toBe(1);
    expect(parsed.sources[0].advertisements).toBe(12);
  });

  it('is reachable through the shared resource dispatch, not the waste fallback', () => {
    // This is the regression: the instance response carries the census, and parsing it through the
    // waste validator threw, which the page reported as "This action could not be completed".
    const instance = parseInstance({
      sources: [],
      credential_slots: [],
      runs: [],
      schedule: null,
      id: 'i-1',
      name: 'Enrolled sensors',
      provider_key: 'wireless_observation/hub',
      revision: 2,
      etag: '"instance:i-1:2"',
      enabled: true,
      status: 'ready',
      postcode: '',
      credential_configured: false,
      resources: [census],
      last_error: null,
      updated_at: '2026-11-13T12:00:00Z',
    });
    expect(instance.resources[0].contract_version).toBe('observation_census/v1');
  });

  it('rejects a census that hides more devices than it has', () => {
    // A negative shown-count would render as "-2 of 3" beside a radar showing three blips.
    expect(() => parseObservationCensus({ ...census, identities: 1, hidden: 3 })).toThrow();
  });

  it('rejects negative counters rather than rendering them', () => {
    expect(() => parseObservationCensus({ ...census, sources: [{ ...census.sources[0], dropped: -1 }] })).toThrow();
  });

  it('keeps a local provider identifiable in the catalog', () => {
    // The frontend decides "Add" versus "Set up" from this, so losing it would send a local provider
    // through a wizard with nothing in it.
    const catalog = parseCatalog({
      available: true,
      availability_error: null,
      types: [],
      providers: [
        {
          key: 'wireless_observation/hub',
          id: 'hub',
          name: 'Enrolled sensors',
          integration_type: 'wireless_observation',
          origin: 'local',
          available: true,
          requires_credential: false,
        },
      ],
    });
    expect(catalog.providers[0].origin).toBe('local');
  });

  it('carries the server\u2019s answer for whether an integration is continuous', () => {
    // Without this field the page offered an address change and a refresh schedule to a provider
    // that has neither, and saving the schedule could only fail. Validated at ingress so a
    // non-boolean cannot silently become "scheduled".
    const base = {
      sources: [],
      credential_slots: [],
      runs: [],
      schedule: null,
      id: 'i-1',
      name: 'Enrolled sensors',
      provider_key: 'wireless_observation/hub',
      revision: 2,
      etag: '"instance:i-1:2"',
      enabled: true,
      status: 'ready',
      postcode: '',
      credential_configured: false,
      resources: [census],
      last_error: null,
      updated_at: '2026-11-13T12:00:00Z',
    };
    expect(parseInstance({ ...base, continuous: true }).continuous).toBe(true);
    // An older backend has no field, and every provider it knew about was scheduled.
    expect(parseInstance(base).continuous).toBeUndefined();
    expect(() => parseInstance({ ...base, continuous: 'yes' })).toThrow();
  });
});
