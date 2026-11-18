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
  withinDistanceRange,
  distanceRangeById,
  DISTANCE_RANGES,
  autoRadarSpan,
  radialRatio,
  ringsForSpan,
  stepRadarSpan,
  kindLabel,
  isUsefulFacet,
  RING_COUNT,
  DEFAULT_RADAR_SPAN,
  RADAR_SPANS_M,
  CLASS_PRESENTATION,
  classPresentation,
  classificationEvidence,
  isProvisional,
  learningProgress,
  shownByDefault,
} from '@/domain/observations/presentation';
import type {
  ObservationClassification,
  ObservationIdentityView,
  ObservationSnapshot,
  ObservationSourceDescriptor,
} from '@/types/observations.type';
import { ObservationClasses } from '@/types/observations.type';

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

  test('rounds to whole metres at every distance', () => {
    expect(roundedDistanceMetres(11.2)).toBe(11);
    expect(roundedDistanceMetres(13.4)).toBe(13);
    expect(roundedDistanceMetres(22)).toBe(22);
  });

  test('a small move never swings the displayed figure by more than the rounding', () => {
    // Coarse 5 m steps above 10 m were tried and rejected: a device moving 12 m -> 13 m changed the
    // card from "10 m" to "15 m" while its radar position barely moved. The label has to describe
    // the same value the picture shows.
    const a = roundedDistanceMetres(12) as number;
    const b = roundedDistanceMetres(13) as number;
    expect(Math.abs(b - a)).toBeLessThanOrEqual(1);
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
    expect(autoRadarSpan([0.4, 1.2])).toBe(2);
    expect(autoRadarSpan([2, 8])).toBe(10);
    expect(autoRadarSpan([4, 25])).toBe(30);
  });

  test('falls back to the widest preset when a reading exceeds all of them', () => {
    expect(autoRadarSpan([120])).toBe(50);
  });

  test('an empty or unmeasured set yields the default, not the narrowest', () => {
    // A narrow default would claim a one-metre range over an empty radar, which reads as a broken
    // scale rather than as "nothing here".
    expect(autoRadarSpan([])).toBe(DEFAULT_RADAR_SPAN);
    expect(autoRadarSpan([Number.NaN])).toBe(DEFAULT_RADAR_SPAN);
  });
});

describe('stepRadarSpan', () => {
  test('zooming in narrows the span', () => {
    // The direction the word means: a narrower span draws the same space larger.
    expect(stepRadarSpan(30, -1)).toBe(20);
    expect(stepRadarSpan(10, -1)).toBe(5);
  });

  test('zooming out widens the span', () => {
    expect(stepRadarSpan(10, 1)).toBe(20);
    expect(stepRadarSpan(3, 1)).toBe(5);
  });

  test('stops at the limits rather than wrapping', () => {
    const narrowest = RADAR_SPANS_M[0];
    const widest = RADAR_SPANS_M[RADAR_SPANS_M.length - 1];
    expect(stepRadarSpan(narrowest, -1)).toBe(narrowest);
    expect(stepRadarSpan(widest, 1)).toBe(widest);
  });

  test('an off-ladder span snaps to a preset before stepping', () => {
    // A span can end up off-ladder once it follows the data; stepping must stay predictable.
    expect(RADAR_SPANS_M).toContain(stepRadarSpan(7, 1));
    expect(RADAR_SPANS_M).toContain(stepRadarSpan(7, -1));
  });
});

describe('ringsForSpan', () => {
  test('rings are evenly spaced on the canvas at every span', () => {
    // This is the property that was broken: the old implementation filtered an absolute ladder, so
    // gaps changed with zoom (23%, 23%, 13% at a 3 m span) and the scale read as broken.
    for (const span of [1, 3, 10, 30, 50]) {
      const rings = ringsForSpan(span);
      const gaps = rings.slice(1).map((ring, index) => ring.ratio - rings[index].ratio);
      for (const gap of gaps) {
        expect(gap).toBeCloseTo(1 / RING_COUNT, 10);
      }
    }
  });

  test('the outer ring is the span edge, at full radius', () => {
    for (const span of [1, 3, 10, 30, 50]) {
      const rings = ringsForSpan(span);
      const outer = rings[rings.length - 1];
      expect(outer.ratio).toBe(1);
      expect(outer.metres).toBe(span);
      expect(outer.outer).toBe(true);
    }
  });

  test('ring labels increase and stay inside the span', () => {
    const rings = ringsForSpan(30);
    for (let index = 1; index < rings.length; index++) {
      expect(rings[index].metres).toBeGreaterThan(rings[index - 1].metres);
    }
    for (const ring of rings) {
      expect(ring.metres).toBeLessThanOrEqual(30);
      expect(ring.metres).toBeGreaterThan(0);
    }
  });

  test('a ring label agrees with the projection it was drawn at', () => {
    // The ring and the blips go through the same projection, so a device at a ring's labelled
    // distance must land on that ring. This is the property that forbids coarse rounding: at a 3 m
    // span the 75% ring is truly at 1.52 m, and labelling it "2 m" would put a device at 2 m a
    // tenth of the radius away from the circle that claims to be 2 m.
    //
    // The tolerance is measured, not guessed: two significant digits holds it to under 0.02 of the
    // radius across every span.
    for (const span of [1, 2, 3, 5, 10, 20, 30, 50]) {
      for (const ring of ringsForSpan(span)) {
        expect(Math.abs(radialRatio(ring.metres, span) - ring.ratio)).toBeLessThan(0.02);
      }
    }
  });

  test('never draws a ring at the exact centre', () => {
    // Fraction 0 would sit on the centre marker, where the label is unreadable.
    for (const ring of ringsForSpan(30)) {
      expect(ring.ratio).toBeGreaterThan(0);
    }
  });
});

describe('kindLabel', () => {
  test('renders the wire values as words an operator reads', () => {
    // "adv" is jargon for what is simply Bluetooth, and it is on almost every row.
    expect(kindLabel('adv')).toBe('Bluetooth');
    expect(kindLabel('ibeacon')).toBe('iBeacon');
    expect(kindLabel('eddystone')).toBe('Eddystone');
  });

  test('passes an unknown kind through rather than hiding it', () => {
    // An unrecognised value means a newer backend; showing it keeps that visible.
    expect(kindLabel('wifi.probe')).toBe('wifi.probe');
    expect(kindLabel(undefined)).toBe('Unknown');
  });
});

describe('isUsefulFacet', () => {
  test('one value is not worth a control', () => {
    // With a single option both states match everything, so the filter cannot do anything. In an
    // ordinary household that is exactly the Kind facet, since every device is "adv".
    expect(isUsefulFacet(['adv'])).toBe(false);
    expect(isUsefulFacet([])).toBe(false);
  });

  test('two or more values are worth a control', () => {
    expect(isUsefulFacet(['adv', 'ibeacon'])).toBe(true);
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

describe('snapshotSummary', () => {
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

  test('summarises counts and arrivals without inventing detail', () => {
    expect(snapshotSummary(snapshot)).toContain('3 devices');
    expect(snapshotSummary(snapshot)).toContain('1 source');
    expect(snapshotSummary(snapshot)).toContain('2 new');
  });

  test('reports an explicit empty state', () => {
    expect(snapshotSummary(null)).toBe('No observations yet');
  });
});

/**
 * Class presentation.
 *
 * These tests exist because the wording is the feature: a class the operator cannot read or check
 * is a class they will not trust, and "learning" must never be shown as if it were an answer.
 */
describe('class presentation', () => {
  function classification(overrides: Partial<ObservationClassification> = {}): ObservationClassification {
    return {
      class: 'infrastructure',
      showByDefault: false,
      confidence: 0.7,
      hoursObserved: 30,
      advisory: true,
      ...overrides,
    };
  }

  test('every class has plain-language wording', () => {
    // No class may render as a raw identifier, which is what the operator complained about.
    for (const value of Object.values(ObservationClasses)) {
      const presentation = classPresentation(classification({ class: value }));
      expect(presentation.label).toBeTruthy();
      expect(presentation.label).not.toBe(value);
      expect(presentation.hint).toBeTruthy();
    }
  });

  test('an absent classification reads as learning, not as a class', () => {
    const presentation = classPresentation(undefined);
    expect(presentation.tone).toBe('learning');
    expect(presentation.label).toBe('Learning');
  });

  test('an unknown class falls back to learning rather than blank', () => {
    const presentation = classPresentation(classification({ class: 'something-new' }));
    expect(presentation.tone).toBe('learning');
  });

  test('infrastructure and neighbour are the demoted classes', () => {
    // The operator's complaint was a wall of MAC addresses; these two are what should collapse.
    expect(CLASS_PRESENTATION[ObservationClasses.Infrastructure].tone).toBe('installed');
    expect(CLASS_PRESENTATION[ObservationClasses.Neighbour].tone).toBe('outside');
  });

  test('provisional is true when the backend says so', () => {
    expect(isProvisional(classification({ provisional: true }))).toBe(true);
  });

  test('provisional is true for the pre-filter classes even without the flag', () => {
    // These values only exist before any history does, so treating them as settled would claim a
    // day-one guess was a verdict.
    expect(isProvisional(classification({ class: ObservationClasses.ProvisionalMobile }))).toBe(true);
    expect(isProvisional(classification({ class: ObservationClasses.ProvisionalNeighbour }))).toBe(true);
  });

  test('a settled class is not provisional', () => {
    expect(isProvisional(classification({ class: 'personal' }))).toBe(false);
  });

  test('an absent classification is provisional, because nothing is known', () => {
    expect(isProvisional(undefined)).toBe(true);
  });

  test('evidence prefers the backend reason', () => {
    const classificationWithReason = classification({ reason: 'Present 23 of 24 hours.' });
    expect(classificationEvidence(classificationWithReason)).toBe('Present 23 of 24 hours.');
  });

  test('evidence explains learning with its progress', () => {
    // A learning class with no explanation looks like a failure; naming the progress makes it read
    // as work in progress.
    const learning = classification({ class: 'learning', hoursObserved: 6, reason: undefined });
    expect(classificationEvidence(learning)).toContain('6 of 24');
  });

  test('learning progress is reported only while learning', () => {
    expect(learningProgress(classification({ class: 'learning', hoursObserved: 6 }))).toBe('6 of 24 hours');
    expect(learningProgress(classification({ class: 'personal', hoursObserved: 30 }))).toBeNull();
    expect(learningProgress(undefined)).toBeNull();
  });

  test('learning progress is capped at a full day', () => {
    expect(learningProgress(classification({ class: 'learning', hoursObserved: 40 }))).toBe('24 of 24 hours');
  });

  test('an identity is shown unless the backend says otherwise', () => {
    // Absence means an older backend, and showing too much is the safe failure.
    const shown = identity({ classification: classification({ showByDefault: true }) });
    const hidden = identity({ classification: classification({ showByDefault: false }) });
    const unknown = identity();

    expect(shownByDefault(shown as never)).toBe(true);
    expect(shownByDefault(hidden as never)).toBe(false);
    expect(shownByDefault(unknown as never)).toBe(true);
  });

  test('the demotion predicate identifies exactly the hidden set', () => {
    // The count in the footer is derived by filtering on this predicate, so the predicate is the
    // thing worth pinning rather than a second helper that duplicated it.
    const identities = [
      identity({ key: 'addr:aa:bb:cc:dd:ee:01', classification: classification({ showByDefault: true }) }),
      identity({ key: 'addr:aa:bb:cc:dd:ee:02', classification: classification({ showByDefault: false }) }),
      identity({ key: 'addr:aa:bb:cc:dd:ee:03', classification: classification({ showByDefault: false }) }),
    ];

    expect(identities.filter((entry) => !shownByDefault(entry))).toHaveLength(2);
  });
});

describe('autoRadarSpan', () => {
  test('a lone device is framed exactly, with no invented headroom', () => {
    // Headroom exists to leave room around a distribution; a single device has none, so a device
    // at 2 m belongs on a 2 m scale rather than a 3 m one.
    expect(autoRadarSpan([2])).toBe(2);
    expect(autoRadarSpan([45])).toBe(50);
  });

  test('a coherent set is covered completely', () => {
    // No tail to ignore, so nothing may be clipped — a two-device household must frame both.
    expect(autoRadarSpan([1, 2, 3, 4, 5])).toBe(5);
    expect(autoRadarSpan([1.9, 5.7])).toBe(10);
  });

  test('scales to the typical device, not to a distant tail', () => {
    // The bug this fixes: one distant device forced the scale wide and crushed everything nearby
    // into the middle. A real household reported 44 devices, 43 inside 12 m, with the scale set
    // to 50 m by a single reading.
    const household = [...Array(43)].map((_, index) => 1 + (index / 43) * 11);
    household.push(45);
    expect(autoRadarSpan(household)).toBe(10);
  });

  test('a wide spread does not let the far tail set the scale', () => {
    const spread = [0.5, 0.8, 1, 1, 1.2, 2, 2, 3, 15, 50];
    expect(autoRadarSpan(spread)).toBe(3);
  });

  test('ignores non-finite and non-positive readings', () => {
    // A zero or NaN distance is unmeasured, not "at the antenna".
    expect(autoRadarSpan([Number.NaN, 0, -3, 2, 3])).toBe(5);
  });

  test('an empty set falls back to the default rather than the narrowest', () => {
    expect(autoRadarSpan([])).toBe(DEFAULT_RADAR_SPAN);
  });
});

describe('autoRadarSpan uses every device, not only identifiable ones', () => {
  test('a single known device must not collapse the scale', () => {
    // The report: one identifiable device at 0.3 m and twelve rotating ones out to 25 m. Excluding
    // the rotating devices from the scale calculation collapsed the span to 1 m, so every far
    // device was clamped to the edge — stacked on one another, which read as a single dot — and
    // bucketed as "under 1 m" while its card correctly said 17 m.
    const all = [0.3, 0.4, 1, 2, 3, 5, 8, 13, 17, 25];
    const span = autoRadarSpan(all);

    // The scale must cover the bulk of the room rather than one nearby device. 1 m was the bug.
    expect(span).toBeGreaterThan(1);

    // Most devices are now inside the span rather than all of them piled on the edge. A long tail
    // is still drawn at the edge and reported, which is deliberate — the mistake was that *every*
    // device beyond the single near one was treated as tail.
    const inside = all.filter((metres) => radialRatio(metres, span) < 1);
    expect(inside.length).toBeGreaterThanOrEqual(7);
  });
});

describe('bearingForIdentity', () => {
  test('sequential keys land on unrelated bearings', () => {
    // The bug this guards: the previous hash barely avalanched, so sequential MAC addresses — which
    // is what one manufacturer's devices produce — landed on adjacent angles. Eight devices stacked
    // on a seven-degree arc, and their blips and labels piled on top of each other.
    const bearings = Array.from({ length: 8 }, (_, index) => bearingForIdentity(`addr:aa:bb:cc:dd:ef:0${index}`));
    const spread = Math.max(...bearings) - Math.min(...bearings);
    expect(spread).toBeGreaterThan(90);
  });

  test('the same key always gives the same bearing', () => {
    // Movement must read as radial drift, not teleporting between spokes.
    const key = 'ibeacon:2f234454-cf6d-4a0f-adf2-f4911ba9ffa6:1:2';
    expect(bearingForIdentity(key)).toBe(bearingForIdentity(key));
  });

  test('bearings stay inside the full circle', () => {
    for (let index = 0; index < 50; index += 1) {
      const bearing = bearingForIdentity(`fp:${index.toString(16)}`);
      expect(bearing).toBeGreaterThanOrEqual(0);
      expect(bearing).toBeLessThan(360);
    }
  });

  test('a realistic set covers the circle rather than clustering', () => {
    // Real keys, not toy ones: a 17-character MAC has a shape that matters. Sequential addresses
    // are what one manufacturer's devices produce, and they must not share a sector.
    const bearings = Array.from({ length: 10 }, (_, index) =>
      bearingForIdentity(`addr:aa:bb:cc:dd:ef:0${index}`),
    );
    const sectors = new Set(bearings.map((bearing) => Math.floor(bearing / 45)));
    expect(sectors.size).toBeGreaterThanOrEqual(6);
  });
});
