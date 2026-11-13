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

  test('accepts an identity without trend fields so an older backend still parses', () => {
    const payload = validSnapshot();
    const parsed = parseObservationSnapshot(payload);
    expect(parsed).not.toBeNull();
    expect(parsed?.identities[0].trend).toBeUndefined();
  });

  test('parses trend, delta and the signal history', () => {
    const payload = validSnapshot();
    Object.assign((payload.identities as Record<string, unknown>[])[0], {
      trend: 'approaching',
      trendDeltaDb: 4.5,
      signal: [-70, -68, -65],
    });
    const parsed = parseObservationSnapshot(payload);
    expect(parsed?.identities[0].trend).toBe('approaching');
    expect(parsed?.identities[0].trendDeltaDb).toBe(4.5);
    expect(parsed?.identities[0].signal).toEqual([-70, -68, -65]);
  });

  test('rejects an unrecognised trend rather than coercing it', () => {
    // A coerced value would become a false claim about movement.
    const payload = validSnapshot();
    (payload.identities as Record<string, unknown>[])[0].trend = 'getting-closer';
    expect(parseObservationSnapshot(payload)).toBeNull();
  });

  test('rejects a malformed signal history', () => {
    for (const signal of ['abc', [1, 'two'], { a: 1 }]) {
      const payload = validSnapshot();
      (payload.identities as Record<string, unknown>[])[0].signal = signal;
      expect(parseObservationSnapshot(payload)).toBeNull();
    }
  });

  test('rejects an oversized signal history', () => {
    const payload = validSnapshot();
    (payload.identities as Record<string, unknown>[])[0].signal = Array.from({ length: 129 }, () => -60);
    expect(parseObservationSnapshot(payload)).toBeNull();
  });

  test('tolerates an unknown facet name', () => {
    const payload = validSnapshot();
    (payload.sources as Record<string, unknown>[])[0].facets = ['kind', 'futureFacet'];
    expect(parseObservationSnapshot(payload)).not.toBeNull();
  });

  test('accepts a source without diagnostics so an older backend still parses', () => {
    const payload = validSnapshot();
    delete (payload.sources as Record<string, unknown>[])[0].diagnostics;
    const parsed = parseObservationSnapshot(payload);
    expect(parsed).not.toBeNull();
    expect(parsed?.sources[0].diagnostics).toBeUndefined();
  });

  test('parses delivery counters when present', () => {
    const payload = validSnapshot();
    (payload.sources as Record<string, unknown>[])[0].diagnostics = {
      advertisements: 120,
      decoded: 96,
      dropped: 3,
      lastAdvertisement: '2026-11-11T10:15:21.000Z',
    };
    const parsed = parseObservationSnapshot(payload);
    expect(parsed?.sources[0].diagnostics).toEqual({
      advertisements: 120,
      decoded: 96,
      dropped: 3,
      lastAdvertisement: '2026-11-11T10:15:21.000Z',
    });
  });

  test('rejects negative or non-numeric counters', () => {
    for (const diagnostics of [
      { advertisements: -1, decoded: 0, dropped: 0 },
      { advertisements: 'many', decoded: 0, dropped: 0 },
      { advertisements: 0, decoded: 0 },
    ]) {
      const payload = validSnapshot();
      (payload.sources as Record<string, unknown>[])[0].diagnostics = diagnostics;
      expect(parseObservationSnapshot(payload)).toBeNull();
    }
  });
});

/**
 * Classification is optional in the contract but strict when present.
 *
 * The distinction that matters: absent means the backend did not classify, which is normal for an
 * older backend, while an unrecognised class means the payload disagrees with this build. The first
 * is allowed through, the second is refused so the mismatch is visible rather than rendered as a
 * blank badge.
 */
function withClassification(classification: unknown): Record<string, unknown> {
  const snapshot = validSnapshot();
  const identity = (snapshot.identities as Record<string, unknown>[])[0];
  identity.classification = classification;
  return snapshot;
}

function validClassification(): Record<string, unknown> {
  return {
    class: 'infrastructure',
    showByDefault: false,
    confidence: 0.72,
    hoursObserved: 26,
    daysConsidered: 3,
    reason: 'Present almost all day, including overnight, and barely moved.',
    advisory: true,
  };
}

describe('classification contract', () => {
  test('accepts a well-formed classification', () => {
    const parsed = parseObservationSnapshot(withClassification(validClassification()));
    expect(parsed).not.toBeNull();
    const classification = parsed?.identities[0].classification;
    expect(classification?.class).toBe('infrastructure');
    expect(classification?.showByDefault).toBe(false);
    expect(classification?.hoursObserved).toBe(26);
    expect(classification?.reason).toContain('Present almost all day');
  });

  test('accepts a snapshot with no classification at all', () => {
    // An older backend sends none. The viewer must render the identity anyway, because refusing the
    // whole snapshot would blank the page on a mixed-version deployment.
    const parsed = parseObservationSnapshot(validSnapshot());
    expect(parsed).not.toBeNull();
    expect(parsed?.identities[0].classification).toBeUndefined();
  });

  test('accepts every known class', () => {
    for (const value of ['infrastructure', 'personal', 'neighbour', 'transient', 'learning', 'mobile-unidentifiable', 'likely-neighbour']) {
      const classification = { ...validClassification(), class: value };
      expect(parseObservationSnapshot(withClassification(classification))).not.toBeNull();
    }
  });

  test('rejects an unrecognised class', () => {
    // A class this build cannot name would render as a blank badge, which reads as a bug. Refusing
    // the identity keeps the mismatch visible.
    expect(parseObservationSnapshot(withClassification({ ...validClassification(), class: 'appliance' }))).toBeNull();
  });

  test('rejects a classification without the required fields', () => {
    for (const missing of ['class', 'showByDefault', 'confidence', 'hoursObserved']) {
      const classification = validClassification();
      delete classification[missing];
      expect(parseObservationSnapshot(withClassification(classification))).toBeNull();
    }
  });

  test('rejects a non-boolean showByDefault', () => {
    // This field decides visibility, so a coerced value would silently change what the operator
    // sees. String "false" is famously truthy; rejecting is the only safe reading.
    expect(parseObservationSnapshot(withClassification({ ...validClassification(), showByDefault: 'false' }))).toBeNull();
  });

  test('rejects a non-object classification', () => {
    for (const bad of ['infrastructure', 42, [], true]) {
      expect(parseObservationSnapshot(withClassification(bad))).toBeNull();
    }
  });

  test('clamps a confidence outside the unit range', () => {
    // Confidence is only used for wording intensity, so clamping is honest: the class itself is
    // unaffected, and rejecting the whole snapshot over a display number would be disproportionate.
    const high = parseObservationSnapshot(withClassification({ ...validClassification(), confidence: 4 }));
    expect(high?.identities[0].classification?.confidence).toBe(1);

    const low = parseObservationSnapshot(withClassification({ ...validClassification(), confidence: -2 }));
    expect(low?.identities[0].classification?.confidence).toBe(0);
  });

  test('keeps the provisional and pending-change flags when present', () => {
    const parsed = parseObservationSnapshot(
      withClassification({
        ...validClassification(),
        class: 'mobile-unidentifiable',
        provisional: true,
        candidate: 'personal',
        changePending: true,
      }),
    );
    const classification = parsed?.identities[0].classification;
    expect(classification?.provisional).toBe(true);
    expect(classification?.candidate).toBe('personal');
    expect(classification?.changePending).toBe(true);
  });

  test('bounds an oversized reason', () => {
    const parsed = parseObservationSnapshot(
      withClassification({ ...validClassification(), reason: 'x'.repeat(400) }),
    );
    expect(parsed?.identities[0].classification?.reason?.length).toBeLessThanOrEqual(240);
  });
});
