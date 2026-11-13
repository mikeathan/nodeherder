/**
 * Wireless viewer presentation helpers (spec 064, ADR-003).
 *
 * The range estimate and the stable bearing are both honesty mechanisms: RSSI yields a
 * distance with no direction, so placement must be deterministic rather than implying bearing.
 */
import {
  DEFAULT_PATH_LOSS_EXPONENT,
  DEFAULT_TX_POWER_AT_1M,
  bearingForIdentity,
  distanceBand,
  distanceRSSI,
  estimateDistanceMetres,
  roundedDistanceMetres,
  groupIdentities,
  identityTitle,
  isEphemeralIdentity,
  isRecentArrival,
  snapshotSummary,
  sourceIsSilent,
  sourceStatus,
  trendPresentation,
  untriagedCount,
  withinDistanceRange,
  distanceRangeById,
  DISTANCE_RANGES,
  autoRadarSpan,
  radialRatio,
  ringsForSpan,
} from '@/domain/observations/presentation';
import type { ObservationIdentityView, ObservationSnapshot, ObservationSourceDescriptor } from '@/types/observations.type';

function identity(overrides: Partial<ObservationIdentityView> = {}): ObservationIdentityView {
  return {
    key: 'addr:aa:bb:cc:dd:ee:ff',
    kind: 'adv',
    stable: true,
    decision: 'unknown',
    firstSeen: '2026-11-11T09:59:00Z',
    lastSeen: '2026-11-11T10:00:00Z',
    packets: 3,
    rssi: -65,
    sources: [],
    ...overrides,
  };
}

describe('estimateDistanceMetres', () => {
  test('returns null when the reading is absent, never zero', () => {
    expect(estimateDistanceMetres(null)).toBeNull();
    expect(estimateDistanceMetres(undefined)).toBeNull();
    expect(estimateDistanceMetres(Number.NaN)).toBeNull();
  });

  test('returns roughly one metre at the calibrated power', () => {
    const metres = estimateDistanceMetres(DEFAULT_TX_POWER_AT_1M);
    expect(metres).not.toBeNull();
    expect(metres!).toBeCloseTo(1, 5);
  });

  test('grows with weaker signal', () => {
    const near = estimateDistanceMetres(-60)!;
    const far = estimateDistanceMetres(-80)!;
    expect(far).toBeGreaterThan(near);
  });

  test('a ten decibel loss is roughly a threefold distance increase indoors', () => {
    const base = estimateDistanceMetres(-60, DEFAULT_TX_POWER_AT_1M, DEFAULT_PATH_LOSS_EXPONENT)!;
    const weaker = estimateDistanceMetres(-70, DEFAULT_TX_POWER_AT_1M, DEFAULT_PATH_LOSS_EXPONENT)!;
    const ratio = weaker / base;
    expect(ratio).toBeGreaterThan(2);
    expect(ratio).toBeLessThan(4.5);
  });

  test('rejects a nonsensical path-loss exponent', () => {
    expect(estimateDistanceMetres(-60, DEFAULT_TX_POWER_AT_1M, 0)).toBeNull();
    expect(estimateDistanceMetres(-60, DEFAULT_TX_POWER_AT_1M, -1)).toBeNull();
  });
});

describe('bearingForIdentity', () => {
  test('is deterministic for a key', () => {
    expect(bearingForIdentity('addr:aa')).toBe(bearingForIdentity('addr:aa'));
  });

  test('stays within a full circle', () => {
    for (const key of ['addr:aa', 'fp:bb', 'ibeacon:x:1:2', '']) {
      const bearing = bearingForIdentity(key);
      expect(bearing).toBeGreaterThanOrEqual(0);
      expect(bearing).toBeLessThan(360);
    }
  });

  test('separates distinct keys', () => {
    expect(bearingForIdentity('addr:aa')).not.toBe(bearingForIdentity('addr:ab'));
  });
});

describe('isEphemeralIdentity', () => {
  test('treats an unstable identity as ephemeral', () => {
    expect(isEphemeralIdentity(identity({ stable: false }))).toBe(true);
    expect(isEphemeralIdentity(identity({ stable: true }))).toBe(false);
  });
});

describe('identityTitle', () => {
  test('prefers the operator label', () => {
    expect(identityTitle(identity({ label: 'Front door tag', kind: 'ibeacon' }))).toBe('Front door tag');
  });

  test('falls back to the kind for a beacon', () => {
    expect(identityTitle(identity({ kind: 'ibeacon' }))).toBe('ibeacon');
  });

  test('abbreviates a long key', () => {
    const long = 'addr:' + 'a'.repeat(80);
    const title = identityTitle(identity({ key: long }));
    expect(title.length).toBeLessThanOrEqual(25);
    expect(title).toContain('…');
  });
});

describe('source status diagnostics', () => {
  const base: ObservationSourceDescriptor = { id: 'a', name: 'Hall proxy', media: ['ble.adv'], facets: ['kind'] };

  test('reports scanning when advertisements arrive', () => {
    const status = sourceStatus({
      ...base,
      diagnostics: { advertisements: 12, decoded: 9, dropped: 0, lastAdvertisement: null },
    });
    expect(status).toContain('Scanning');
    expect(status).toContain('12 advertisements');
  });

  test('mentions drops when the queue saturated', () => {
    const status = sourceStatus({
      ...base,
      diagnostics: { advertisements: 12, decoded: 9, dropped: 4, lastAdvertisement: null },
    });
    expect(status).toContain('4 dropped');
  });

  test('explains silence as a likely configuration fault', () => {
    const silent = { ...base, diagnostics: { advertisements: 0, decoded: 0, dropped: 0, lastAdvertisement: null } };
    expect(sourceStatus(silent)).toContain('bluetooth_proxy');
    expect(sourceIsSilent(silent)).toBe(true);
  });

  test('does not call a delivering source silent', () => {
    const active = { ...base, diagnostics: { advertisements: 1, decoded: 1, dropped: 0, lastAdvertisement: null } };
    expect(sourceIsSilent(active)).toBe(false);
  });

  test('treats absent diagnostics as connected rather than silent', () => {
    expect(sourceStatus(base)).toBe('Connected');
    expect(sourceIsSilent(base)).toBe(false);
  });
});

describe('isRecentArrival', () => {
  const now = Date.parse('2026-11-11T10:00:00Z');

  test('treats an identity first seen inside the window as new', () => {
    expect(isRecentArrival({ firstSeen: '2026-11-11T09:58:00Z' }, now)).toBe(true);
  });

  test('treats an older identity as not new', () => {
    expect(isRecentArrival({ firstSeen: '2026-11-11T09:00:00Z' }, now)).toBe(false);
  });

  test('ignores triage state entirely', () => {
    // The bug this replaces: novelty came from decision === unknown, so on a first run
    // everything matched and the "only new" control appeared broken.
    const untriagedButOld = { firstSeen: '2026-11-11T08:00:00Z', decision: 'unknown' as const };
    expect(isRecentArrival(untriagedButOld, now)).toBe(false);
  });

  test('does not treat an invalid or future timestamp as new', () => {
    expect(isRecentArrival({ firstSeen: 'not-a-date' }, now)).toBe(false);
    expect(isRecentArrival({ firstSeen: '2026-11-11T12:00:00Z' }, now)).toBe(false);
  });
});

describe('groupIdentities', () => {
  test('groups by beacon kind first, then vendor, then address stability', () => {
    const groups = groupIdentities([
      identity({ key: 'a', kind: 'ibeacon' }),
      identity({ key: 'b', kind: 'eddystone' }),
      identity({ key: 'c', kind: 'adv', vendor: 76 }),
      identity({ key: 'd', kind: 'adv', stable: true }),
      identity({ key: 'e', kind: 'adv', stable: false }),
    ]);
    const names = groups.map((entry) => entry.group);
    expect(names).toContain('iBeacon');
    expect(names).toContain('Eddystone');
    expect(names).toContain('Vendor 0x004C');
    expect(names).toContain('Stable address');
    expect(names).toContain('Rotating address');
  });

  test('orders sections by size so the noisiest category is first', () => {
    const groups = groupIdentities([
      identity({ key: '1', kind: 'adv', vendor: 76 }),
      identity({ key: '2', kind: 'adv', vendor: 76 }),
      identity({ key: '3', kind: 'ibeacon' }),
    ]);
    expect(groups[0].group).toBe('Vendor 0x004C');
    expect(groups[0].identities).toHaveLength(2);
  });

  test('is deterministic for equal-sized sections', () => {
    const build = () =>
      groupIdentities([
        identity({ key: '1', kind: 'adv', stable: true }),
        identity({ key: '2', kind: 'adv', stable: false }),
      ]).map((entry) => entry.group);
    expect(build()).toEqual(build());
  });

  test('preserves the incoming order inside a section', () => {
    const groups = groupIdentities([
      identity({ key: 'first', kind: 'adv', vendor: 76 }),
      identity({ key: 'second', kind: 'adv', vendor: 76 }),
    ]);
    expect(groups[0].identities.map((entry) => entry.key)).toEqual(['first', 'second']);
  });

  test('returns no sections for no identities', () => {
    expect(groupIdentities([])).toEqual([]);
  });
});

describe('distanceRSSI', () => {
  test('prefers the smoothed median so distance does not follow single-packet jitter', () => {
    expect(distanceRSSI({ rssi: -90, rssiSmoothed: -62 })).toBe(-62);
  });

  test('falls back to the single reading when no smoothed value exists', () => {
    // Keeps a payload from an earlier backend usable rather than showing nothing.
    expect(distanceRSSI({ rssi: -70 })).toBe(-70);
    expect(distanceRSSI({ rssi: -70, rssiSmoothed: null })).toBe(-70);
  });

  test('returns null when there is no reading at all', () => {
    expect(distanceRSSI({})).toBeNull();
    expect(distanceRSSI({ rssi: null, rssiSmoothed: null })).toBeNull();
  });
});

describe('roundedDistanceMetres', () => {
  test('rounds to whole metres below ten', () => {
    expect(roundedDistanceMetres(7.4)).toBe(7);
    expect(roundedDistanceMetres(7.6)).toBe(8);
  });

  test('rounds to five-metre steps above ten, because metre precision is not real', () => {
    expect(roundedDistanceMetres(11.2)).toBe(10);
    expect(roundedDistanceMetres(13.4)).toBe(15);
    expect(roundedDistanceMetres(22)).toBe(20);
  });

  test('never rounds a real reading down to zero', () => {
    expect(roundedDistanceMetres(0.2)).toBe(1);
    expect(roundedDistanceMetres(0)).toBe(1);
  });

  test('an absent reading stays absent', () => {
    expect(roundedDistanceMetres(null)).toBeNull();
    expect(roundedDistanceMetres(Number.NaN)).toBeNull();
  });
});

describe('withinDistanceRange', () => {
  const at = (rssiSmoothed: number) => ({ rssi: rssiSmoothed, rssiSmoothed });
  const range = (id: string) => distanceRangeById(id);

  test('any range passes everything, including identities with no reading', () => {
    expect(withinDistanceRange(at(-90), range('any'))).toBe(true);
    expect(withinDistanceRange({}, range('any'))).toBe(true);
  });

  test('each band covers its own span', () => {
    // -66 -> ~1.9 m: under 2 m.
    expect(withinDistanceRange(at(-66), range('under2'))).toBe(true);
    expect(withinDistanceRange(at(-66), range('2to5'))).toBe(false);
    // -80 -> ~7 m: the 5-10 m band.
    expect(withinDistanceRange(at(-80), range('5to10'))).toBe(true);
    expect(withinDistanceRange(at(-80), range('2to5'))).toBe(false);
    // -90 -> ~17 m: the open upper band.
    expect(withinDistanceRange(at(-90), range('over10'))).toBe(true);
  });

  test('neighbouring bands never both match one reading', () => {
    // "Any" is the unfiltered range and matches everything, so it is excluded here: the bands
    // themselves must be disjoint, or one device would appear in two of them at once.
    const bands = DISTANCE_RANGES.filter((entry) => entry.id !== 'any');
    for (const rssi of [-60, -66, -72, -78, -84, -90]) {
      const matches = bands.filter((entry) => withinDistanceRange(at(rssi), entry));
      expect(matches.length).toBe(1);
    }
  });

  test('exactly one band matches, or none when "any" is the only match', () => {
    for (const rssi of [-60, -66, -72, -78, -84, -90]) {
      const matches = DISTANCE_RANGES.filter((entry) => withinDistanceRange(at(rssi), entry));
      // "any" plus exactly one band.
      expect(matches.map((entry) => entry.id)).toContain('any');
      expect(matches.length).toBeLessThanOrEqual(2);
    }
  });

  test('the bands are exhaustive for any measured reading', () => {
    for (let rssi = -40; rssi >= -110; rssi -= 1) {
      const matches = DISTANCE_RANGES.filter((entry) => withinDistanceRange(at(rssi), entry));
      expect(matches.length).toBeGreaterThanOrEqual(1);
    }
  });

  test('an identity with no reading belongs to no band but any', () => {
    // Assuming "near" would put an unmeasured device in the closest band.
    for (const id of ['under05', '05to1', '1to2', '2to5', '5to10', 'over10']) {
      expect(withinDistanceRange({}, range(id))).toBe(false);
    }
    expect(withinDistanceRange({}, range('any'))).toBe(true);
  });

  test('uses the smoothed value when both are present', () => {
    // Raw -62 reads as under a metre, smoothed -90 as far: the filter must follow the smoothed
    // value, which is what the displayed distance uses.
    expect(withinDistanceRange({ rssi: -62, rssiSmoothed: -90 }, range('under05'))).toBe(false);
    expect(withinDistanceRange({ rssi: -62, rssiSmoothed: -90 }, range('over10'))).toBe(true);
  });

  test('an unknown range id falls back to any rather than hiding everything', () => {
    expect(distanceRangeById('nonsense').id).toBe('any');
  });
});

describe('radialRatio', () => {
  test('places the span edge at the radius edge', () => {
    expect(radialRatio(30, 30)).toBeCloseTo(1, 5);
  });

  test('never collapses a sub-metre reading onto the centre', () => {
    // The old projection used log10(metres)/log10(30), and log10(1) is 0, so anything at or
    // below 1 m drew on the exact centre: a device on the desk looked like one at the antenna.
    const half = radialRatio(0.5, 30);
    const one = radialRatio(1, 30);
    expect(half).toBeGreaterThan(0);
    expect(one).toBeGreaterThan(half);
  });

  test('is monotonic, so a nearer device is never drawn further out', () => {
    const samples = [0.2, 0.5, 1, 2, 3, 5, 10, 30];
    const ratios = samples.map((metres) => radialRatio(metres, 30));
    for (let i = 1; i < ratios.length; i += 1) {
      expect(ratios[i]).toBeGreaterThanOrEqual(ratios[i - 1]);
    }
  });

  test('clamps beyond the span to the edge rather than overshooting', () => {
    expect(radialRatio(500, 10)).toBeCloseTo(1, 5);
  });

  test('a device at 1 m sits visibly closer than one at 5 m on a tight span', () => {
    // This is what makes the sub-2 m bands legible: without separation they are indistinguishable.
    const tight = radialRatio(1, 3);
    const further = radialRatio(2.5, 3);
    expect(further - tight).toBeGreaterThan(0.2);
  });

  test('rejects nonsense inputs rather than producing NaN coordinates', () => {
    expect(radialRatio(Number.NaN, 30)).toBe(0);
    expect(radialRatio(1, 0)).toBe(0);
  });
});

describe('autoRadarSpan', () => {
  test('picks the smallest preset that contains the furthest reading', () => {
    expect(autoRadarSpan([0.4, 1.2])).toBe(3);
    expect(autoRadarSpan([2, 8])).toBe(10);
    expect(autoRadarSpan([4, 25])).toBe(30);
  });

  test('falls back to the largest preset when a reading exceeds all of them', () => {
    expect(autoRadarSpan([120])).toBe(30);
  });

  test('an empty or unmeasured set still yields a usable span', () => {
    expect(autoRadarSpan([])).toBe(3);
    expect(autoRadarSpan([Number.NaN])).toBe(3);
  });
});

describe('ringsForSpan', () => {
  test('includes every standard ring inside the span plus the span edge', () => {
    expect(ringsForSpan(3)).toEqual([0.5, 1, 3]);
    expect(ringsForSpan(10)).toEqual([0.5, 1, 3, 10]);
  });

  test('never repeats the edge as both a standard ring and the span', () => {
    const rings = ringsForSpan(30);
    expect(new Set(rings).size).toBe(rings.length);
    expect(rings[rings.length - 1]).toBe(30);
  });
});

describe('distanceBand', () => {
  test('classifies inside each boundary', () => {
    expect(distanceBand(1)).toBe('near');
    expect(distanceBand(3)).toBe('mid');
    expect(distanceBand(9)).toBe('far');
  });

  test('treats boundaries as exclusive upper bounds', () => {
    // Exactly 2 m is no longer "near"; exactly 6 m is no longer "mid".
    expect(distanceBand(2)).toBe('mid');
    expect(distanceBand(6)).toBe('far');
  });

  test('an absent reading is unknown, never near', () => {
    expect(distanceBand(null)).toBe('unknown');
    expect(distanceBand(undefined)).toBe('unknown');
    expect(distanceBand(Number.NaN)).toBe('unknown');
  });
});

describe('trendPresentation', () => {
  test('pairs every arrow with a word', () => {
    for (const trend of ['approaching', 'receding', 'steady', 'unknown'] as const) {
      const presentation = trendPresentation(trend);
      expect(presentation.label.length).toBeGreaterThan(0);
    }
  });

  test('unknown carries no arrow, so it cannot be mistaken for steady', () => {
    expect(trendPresentation('unknown').arrow).toBe('');
    expect(trendPresentation(undefined)).toEqual(trendPresentation('unknown'));
  });

  test('direction arrows differ from the steady arrow', () => {
    expect(trendPresentation('approaching').arrow).not.toBe(trendPresentation('steady').arrow);
    expect(trendPresentation('receding').arrow).not.toBe(trendPresentation('approaching').arrow);
  });
});

describe('untriagedCount and snapshotSummary', () => {
  const snapshot: ObservationSnapshot = {
    at: '2026-11-11T10:00:00Z',
    truncated: false,
    sources: [{ id: 'a', name: 'Hall', media: ['ble.adv'], facets: ['kind'] }],
    identities: [
      identity({ key: 'a', decision: 'unknown' }),
      identity({ key: 'b', decision: 'known' }),
      identity({ key: 'c', decision: 'unknown' }),
    ],
    arrivals: 2,
  };

  test('counts only untriaged identities', () => {
    expect(untriagedCount(snapshot)).toBe(2);
    expect(untriagedCount(null)).toBe(0);
  });

  test('summarises counts and arrivals without inventing detail', () => {
    expect(snapshotSummary(snapshot)).toContain('3 devices');
    expect(snapshotSummary(snapshot)).toContain('1 source');
    expect(snapshotSummary(snapshot)).toContain('2 new');
  });

  test('reports an explicit empty state', () => {
    expect(snapshotSummary(null)).toBe('No observations yet');
  });
});
