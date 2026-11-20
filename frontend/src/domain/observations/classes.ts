/**
 * Class grouping and stable ordering for the wireless viewer (spec 066).
 *
 * One ordered table drives the Class filter presets, the row order and the List-view sections,
 * so those three surfaces cannot disagree about what a class means or where it sorts. Ordering
 * deliberately ignores every volatile field: a snapshot that only changes `lastSeen`, RSSI or
 * packet counts must not move a row under the operator's pointer.
 */
import type { ObservationIdentityView } from '@/types/observations.type';
import { ObservationClasses } from '@/types/observations.type';
import { classLabel, identityTitle } from './presentation';

/** Filter sentinel meaning "no class restriction". */
export const ANY_CLASS_GROUP = 'any';

/** A bucket of advisory classes used by the filter, the sort and the list sections. */
export type ObservationClassGroup =
  | 'yours'
  | 'someone'
  | 'installed'
  | 'learning'
  | 'outside'
  | 'passing';

/** Selected class preset: one group, or every group. */
export type ObservationClassFilter = ObservationClassGroup | typeof ANY_CLASS_GROUP;

export type ClassGroup = {
  id: ObservationClassGroup;
  /** Operator-facing label, taken from the badge copy so a chip and a row agree. */
  label: string;
  /** Class values in this group. */
  classes: readonly string[];
};

/**
 * Groups in display order.
 *
 * The two neighbour values share a group because both answer "probably not in my home" and both
 * are hidden by default; every other class gets a chip that matches its badge wording exactly.
 */
const YOURS: ClassGroup = {
  id: 'yours',
  label: classLabel(ObservationClasses.Personal),
  classes: [ObservationClasses.Personal],
};

const SOMEONE: ClassGroup = {
  id: 'someone',
  label: classLabel(ObservationClasses.ProvisionalMobile),
  classes: [ObservationClasses.ProvisionalMobile],
};

const INSTALLED: ClassGroup = {
  id: 'installed',
  label: classLabel(ObservationClasses.Infrastructure),
  classes: [ObservationClasses.Infrastructure],
};

const LEARNING: ClassGroup = {
  id: 'learning',
  label: classLabel(ObservationClasses.Learning),
  classes: [ObservationClasses.Learning],
};

const OUTSIDE: ClassGroup = {
  id: 'outside',
  label: classLabel(ObservationClasses.Neighbour),
  classes: [ObservationClasses.Neighbour, ObservationClasses.ProvisionalNeighbour],
};

const PASSING: ClassGroup = {
  id: 'passing',
  label: classLabel(ObservationClasses.Transient),
  classes: [ObservationClasses.Transient],
};

export const CLASS_GROUPS: readonly ClassGroup[] = [
  YOURS,
  SOMEONE,
  INSTALLED,
  LEARNING,
  OUTSIDE,
  PASSING,
];

const GROUP_BY_CLASS = new Map<string, ClassGroup>();
for (const group of CLASS_GROUPS) {
  for (const value of group.classes) GROUP_BY_CLASS.set(value, group);
}

/** Rank of each group, so ordering does not search the array per comparison. */
const RANK_BY_GROUP = new Map<ObservationClassGroup, number>(
  CLASS_GROUPS.map((group, index) => [group.id, index]),
);

/** Group for an identity; an absent or newer-than-this-build class reads as Learning. */
export function classGroupOf(identity: ObservationIdentityView): ClassGroup {
  const value = identity.classification?.class;
  return (value ? GROUP_BY_CLASS.get(value) : undefined) ?? LEARNING;
}

/** Whether an identity belongs to the selected class preset. */
export function matchesClassGroup(
  identity: ObservationIdentityView,
  filter: ObservationClassFilter,
): boolean {
  return filter === ANY_CLASS_GROUP || classGroupOf(identity).id === filter;
}

/** Group rank for ordering; Learning's rank is the fallback for an unknown class. */
function rankOf(identity: ObservationIdentityView): number {
  return RANK_BY_GROUP.get(classGroupOf(identity).id) ?? RANK_BY_GROUP.get('learning') ?? 0;
}

/** Locale-independent text comparison, so order is identical on every machine. */
function compareText(left: string, right: string): number {
  const a = left.toLowerCase();
  const b = right.toLowerCase();
  if (a < b) return -1;
  if (a > b) return 1;
  return 0;
}

/**
 * Order identities for display.
 *
 * Class rank first, then display title, then the raw key. All three are stable under a snapshot
 * update, so the only legitimate reorder is a new identity or a class change.
 */
export function orderIdentities(
  identities: readonly ObservationIdentityView[],
): ObservationIdentityView[] {
  return [...identities].sort((a, b) => {
    const byRank = rankOf(a) - rankOf(b);
    if (byRank !== 0) return byRank;
    const byTitle = compareText(identityTitle(a), identityTitle(b));
    if (byTitle !== 0) return byTitle;
    return compareText(a.key, b.key);
  });
}

/** One List-view section. */
export type IdentitySection = {
  group: string;
  groupId: ObservationClassGroup;
  identities: ObservationIdentityView[];
};

/**
 * Section an already-ordered list by class group, in `CLASS_GROUPS` order.
 *
 * Sections keep the caller's order rather than sorting members or sections, so a changing count
 * cannot reshuffle the list. Empty groups are omitted.
 */
export function groupIdentitiesByClass(
  identities: readonly ObservationIdentityView[],
): IdentitySection[] {
  const buckets = new Map<ObservationClassGroup, ObservationIdentityView[]>();
  for (const identity of identities) {
    const group = classGroupOf(identity).id;
    const bucket = buckets.get(group);
    if (bucket) bucket.push(identity);
    else buckets.set(group, [identity]);
  }

  return CLASS_GROUPS.flatMap((group) => {
    const members = buckets.get(group.id);
    return members && members.length > 0
      ? [{ group: group.label, groupId: group.id, identities: members }]
      : [];
  });
}
