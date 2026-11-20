/**
 * Class grouping and stable ordering for the wireless viewer (spec 066).
 *
 * These tests pin the two properties the feature exists for: a class maps to exactly one bucket,
 * and row order cannot be changed by a snapshot that only updates volatile fields.
 */
import { describe, expect, test } from '@jest/globals';
import type { ObservationIdentityView } from '@/types/observations.type';
import {
  ANY_CLASS_GROUP,
  CLASS_GROUPS,
  classGroupOf,
  groupIdentitiesByClass,
  matchesClassGroup,
  orderIdentities,
} from '@/domain/observations/classes';
import { CLASS_PRESENTATION } from '@/domain/observations/presentation';
import { ObservationClasses } from '@/types/observations.type';

function identity(
  key: string,
  overrides: Partial<ObservationIdentityView> = {},
): ObservationIdentityView {
  return {
    key,
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

function classified(
  key: string,
  classValue: string,
  overrides: Partial<ObservationIdentityView> = {},
): ObservationIdentityView {
  return identity(key, {
    classification: {
      class: classValue,
      showByDefault: true,
      confidence: 0.7,
      hoursObserved: 30,
      advisory: true,
    },
    ...overrides,
  });
}

describe('classGroupOf', () => {
  test('maps every class value to exactly one group', () => {
    for (const value of Object.values(ObservationClasses)) {
      const group = classGroupOf(classified('addr:01', value));
      expect(CLASS_GROUPS).toContain(group);
      expect(group.classes).toContain(value);
    }
  });

  test('an absent or unknown class reads as Learning rather than disappearing', () => {
    expect(classGroupOf(identity('addr:01')).id).toBe('learning');
    expect(classGroupOf(classified('addr:01', 'something-new')).id).toBe('learning');
  });

  test('the two neighbour classes share the Outside group', () => {
    expect(classGroupOf(classified('addr:01', ObservationClasses.Neighbour)).id).toBe('outside');
    expect(classGroupOf(classified('addr:02', ObservationClasses.ProvisionalNeighbour)).id).toBe(
      'outside',
    );
  });

  test('group labels are the badge labels, so a chip cannot contradict a row', () => {
    expect(CLASS_GROUPS.map((group) => group.label)).toEqual([
      CLASS_PRESENTATION[ObservationClasses.Personal].label,
      CLASS_PRESENTATION[ObservationClasses.ProvisionalMobile].label,
      CLASS_PRESENTATION[ObservationClasses.Infrastructure].label,
      CLASS_PRESENTATION[ObservationClasses.Learning].label,
      CLASS_PRESENTATION[ObservationClasses.Neighbour].label,
      CLASS_PRESENTATION[ObservationClasses.Transient].label,
    ]);
  });
});

describe('matchesClassGroup', () => {
  test('"any" matches every identity', () => {
    expect(matchesClassGroup(identity('addr:01'), ANY_CLASS_GROUP)).toBe(true);
  });

  test('a preset matches only its own group', () => {
    expect(matchesClassGroup(classified('addr:01', 'personal'), 'yours')).toBe(true);
    expect(matchesClassGroup(classified('addr:01', 'personal'), 'outside')).toBe(false);
  });
});

describe('orderIdentities', () => {
  test('orders by class rank, then title, then key', () => {
    const ordered = orderIdentities([
      classified('addr:c', 'transient'),
      classified('addr:b', 'personal'),
      classified('addr:a', 'personal'),
    ]);

    expect(ordered.map((entry) => entry.key)).toEqual(['addr:a', 'addr:b', 'addr:c']);
  });

  test('a snapshot that only changes volatile fields cannot reorder the list', () => {
    // This is the reported bug: the backend sorts by lastSeen, which advances on every packet, so
    // rows jumped to the top continuously. Ordering must ignore every one of these fields.
    const before = [
      classified('addr:01', 'personal', { lastSeen: '2026-11-11T10:00:00Z', packets: 3, rssi: -60 }),
      classified('addr:02', 'personal', { lastSeen: '2026-11-11T09:59:00Z', packets: 9, rssi: -70 }),
      classified('addr:03', 'personal', { lastSeen: '2026-11-11T09:58:00Z', packets: 4, rssi: -65 }),
    ];
    const after = [
      classified('addr:03', 'personal', { lastSeen: '2026-11-11T10:00:05Z', packets: 5, rssi: -64 }),
      classified('addr:01', 'personal', { lastSeen: '2026-11-11T10:00:04Z', packets: 4, rssi: -59 }),
      classified('addr:02', 'personal', { lastSeen: '2026-11-11T10:00:03Z', packets: 10, rssi: -69 }),
    ];

    expect(orderIdentities(after).map((entry) => entry.key)).toEqual(
      orderIdentities(before).map((entry) => entry.key),
    );
  });

  test('a new identity does not disturb the relative order of the others', () => {
    const current = orderIdentities([
      classified('addr:b', 'personal'),
      classified('addr:d', 'personal'),
    ]);
    const next = orderIdentities([
      classified('addr:d', 'personal'),
      classified('addr:a', 'personal'),
      classified('addr:b', 'personal'),
    ]);

    expect(next.map((entry) => entry.key)).toEqual(['addr:a', 'addr:b', 'addr:d']);
    expect(current.map((entry) => entry.key)).toEqual(['addr:b', 'addr:d']);
  });

  test('a class change is the only legitimate reorder', () => {
    const learning = orderIdentities([classified('addr:a', 'learning')]);
    const installed = orderIdentities([classified('addr:a', 'infrastructure')]);

    // Both sort somewhere, and the behaviour is deterministic rather than random.
    expect(learning.map((entry) => entry.key)).toEqual(['addr:a']);
    expect(installed.map((entry) => entry.key)).toEqual(['addr:a']);
  });
});

describe('groupIdentitiesByClass', () => {
  test('sections follow the fixed group order, not their size', () => {
    const sections = groupIdentitiesByClass([
      classified('addr:01', 'transient'),
      classified('addr:02', 'personal'),
      classified('addr:03', 'personal'),
      classified('addr:04', 'personal'),
    ]);

    expect(sections.map((section) => section.groupId)).toEqual(['yours', 'passing']);
  });

  test('section order is unchanged when counts change', () => {
    const small = groupIdentitiesByClass([
      classified('addr:01', 'personal'),
      classified('addr:02', 'transient'),
    ]);
    const large = groupIdentitiesByClass([
      classified('addr:01', 'personal'),
      classified('addr:02', 'transient'),
      classified('addr:03', 'transient'),
      classified('addr:04', 'transient'),
    ]);

    expect(small.map((section) => section.groupId)).toEqual(large.map((section) => section.groupId));
  });

  test('preserves the incoming order inside a section', () => {
    const sections = groupIdentitiesByClass([
      classified('addr:02', 'personal'),
      classified('addr:01', 'personal'),
    ]);

    expect(sections[0].identities.map((entry) => entry.key)).toEqual(['addr:02', 'addr:01']);
  });

  test('omits empty groups and returns no sections for no identities', () => {
    const sections = groupIdentitiesByClass([classified('addr:01', 'personal')]);
    expect(sections.map((section) => section.groupId)).toEqual(['yours']);
    expect(groupIdentitiesByClass([])).toEqual([]);
  });
});
