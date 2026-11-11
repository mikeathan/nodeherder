/**
 * Runtime validation for the observations snapshot (spec 064, AC-11).
 *
 * A malformed payload must reject the whole snapshot: applying part of one would present
 * mixed-generation data. Unknown facets are ignored rather than rejected so a newer backend
 * can add them without breaking this browser.
 */
import { parseObservationSnapshot } from '@/contracts/observations';

function validSnapshot(): Record<string, unknown> {
  return {
    at: '2026-11-11T10:15:22.480Z',
    truncated: false,
    sources: [{ id: 'source-1', name: 'Hall proxy', media: ['ble.adv'], facets: ['kind', 'vendor'] }],
    identities: [
      {
        key: 'ibeacon:2f234454-cf6d-4a0f-adf2-f4911ba9ffa6:1:2',
        kind: 'ibeacon',
        vendor: 76,
        stable: true,
        label: 'Front door tag',
        decision: 'known',
        lastSeen: '2026-11-11T10:15:20.000Z',
        packets: 42,
        rssi: -62,
        sources: [{ id: 'source-1', rssi: -62 }],
      },
    ],
    arrivals: 1,
  };
}

describe('parseObservationSnapshot', () => {
  test('accepts a well-formed snapshot', () => {
    const parsed = parseObservationSnapshot(validSnapshot());
    expect(parsed).not.toBeNull();
    expect(parsed?.identities).toHaveLength(1);
    expect(parsed?.identities[0].key).toBe('ibeacon:2f234454-cf6d-4a0f-adf2-f4911ba9ffa6:1:2');
    expect(parsed?.sources[0].facets).toEqual(['kind', 'vendor']);
  });

  test('accepts an empty but valid snapshot', () => {
    const parsed = parseObservationSnapshot({
      at: '2026-11-11T10:15:22.480Z',
      truncated: false,
      sources: [],
      identities: [],
      arrivals: 0,
    });
    expect(parsed).not.toBeNull();
    expect(parsed?.identities).toEqual([]);
  });

  test('rejects non-object payloads', () => {
    for (const payload of [null, undefined, 'text', 42, []]) {
      expect(parseObservationSnapshot(payload)).toBeNull();
    }
  });

  test('rejects a missing or unparseable timestamp', () => {
    const missing = validSnapshot();
    delete missing.at;
    expect(parseObservationSnapshot(missing)).toBeNull();
    expect(parseObservationSnapshot({ ...validSnapshot(), at: 'not-a-date' })).toBeNull();
  });

  test('rejects an unknown decision rather than defaulting it', () => {
    const payload = validSnapshot();
    (payload.identities as Record<string, unknown>[])[0].decision = 'maybe';
    expect(parseObservationSnapshot(payload)).toBeNull();
  });

  test('rejects a non-boolean stable flag', () => {
    const payload = validSnapshot();
    (payload.identities as Record<string, unknown>[])[0].stable = 'yes';
    expect(parseObservationSnapshot(payload)).toBeNull();
  });

  test('rejects an identity with an invalid lastSeen', () => {
    const payload = validSnapshot();
    (payload.identities as Record<string, unknown>[])[0].lastSeen = 'yesterday';
    expect(parseObservationSnapshot(payload)).toBeNull();
  });

  test('rejects an undeclared medium rather than guessing', () => {
    const payload = validSnapshot();
    (payload.sources as Record<string, unknown>[])[0].media = ['infrared'];
    expect(parseObservationSnapshot(payload)).toBeNull();
  });

  test('rejects oversized collections', () => {
    const payload = validSnapshot();
    payload.sources = Array.from({ length: 65 }, (_, index) => ({
      id: `source-${index}`,
      name: 'proxy',
      media: ['ble.adv'],
      facets: [],
    }));
    expect(parseObservationSnapshot(payload)).toBeNull();
  });

  test('keeps an absent RSSI distinct from zero', () => {
    const payload = validSnapshot();
    delete (payload.identities as Record<string, unknown>[])[0].rssi;
    const parsed = parseObservationSnapshot(payload);
    expect(parsed?.identities[0].rssi).toBeNull();
    expect(parsed?.identities[0].rssi).not.toBe(0);
  });

  test('tolerates an unknown facet name', () => {
    const payload = validSnapshot();
    (payload.sources as Record<string, unknown>[])[0].facets = ['kind', 'futureFacet'];
    expect(parseObservationSnapshot(payload)).not.toBeNull();
  });
});
