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
  estimateDistanceMetres,
  groupIdentities,
  identityTitle,
  isEphemeralIdentity,
  isRecentArrival,
  snapshotSummary,
  sourceIsSilent,
  sourceStatus,
  untriagedCount,
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
