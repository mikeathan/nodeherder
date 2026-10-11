/*
 * Home dashboard groups (spec 007 US-02, US-06 / FR-09).
 * Groups are stored by name in a map; `order` (optional) gives their position.
 * Groups without `order` follow ordered ones, by name (AC-21).
 */
import { DashboardGroup, DashboardGroups } from '@/types/settings.type';

export type GroupEntity = { deviceId: string; expose: string };

const byName = (a: DashboardGroup, b: DashboardGroup) => a.name.localeCompare(b.name, undefined, { sensitivity: 'base' });

export function sortGroups(groups: DashboardGroups): DashboardGroup[] {
  return Object.values(groups).sort((a, b) => {
    const oa = typeof a.order === 'number' ? a.order : null;
    const ob = typeof b.order === 'number' ? b.order : null;
    if (oa !== null && ob !== null) return oa - ob || byName(a, b);
    if (oa !== null) return -1;
    if (ob !== null) return 1;
    return byName(a, b);
  });
}

/** Returns a new array with the item at `from` moved to `to` (clamped). */
export function moveItem<T>(list: readonly T[], from: number, to: number): T[] {
  const next = [...list];
  if (from < 0 || from >= next.length) return next;
  const target = Math.max(0, Math.min(next.length - 1, to));
  const [item] = next.splice(from, 1);
  next.splice(target, 0, item);
  return next;
}

/**
 * Assigns positions 0…n-1 to the given visual order and returns only the groups whose
 * `order` changed, as new objects ready to save (one saveDashboardGroup each, AC-19).
 */
export function reorderedGroups(ordered: readonly DashboardGroup[]): DashboardGroup[] {
  return ordered.flatMap((g, index) => (g.order === index ? [] : [{ ...g, order: index }]));
}

/** Next free position for a newly created group. */
export function nextOrder(groups: DashboardGroups): number {
  return Object.values(groups).reduce((max, g) => (typeof g.order === 'number' && g.order >= max ? g.order + 1 : max), 0);
}

/** Tiles of a group in stored order (device insertion order, then expose order). */
export function groupEntities(group: DashboardGroup): GroupEntity[] {
  return Object.values(group.deviceGroup ?? {}).flatMap((dg) => dg.exposes.map((expose) => ({ deviceId: dg.deviceId, expose })));
}

export function withoutEntity(group: DashboardGroup, deviceId: string, expose: string): DashboardGroup {
  const deviceGroup = { ...group.deviceGroup };
  const entry = deviceGroup[deviceId];
  if (!entry) return group;
  const exposes = entry.exposes.filter((x) => x !== expose);
  if (exposes.length) deviceGroup[deviceId] = { ...entry, exposes };
  else delete deviceGroup[deviceId];
  return { ...group, deviceGroup };
}

export function withEntities(group: DashboardGroup, deviceId: string, exposes: string[]): DashboardGroup {
  const current = group.deviceGroup[deviceId]?.exposes ?? [];
  const merged = [...current, ...exposes.filter((x) => !current.includes(x))];
  return { ...group, deviceGroup: { ...group.deviceGroup, [deviceId]: { deviceId, exposes: merged } } };
}

/** Plain-language validation for create/rename (replaces alert() calls). */
export function validateGroupName(name: string, groups: DashboardGroups, current?: string): string | null {
  const trimmed = name.trim();
  if (!trimmed) return 'Enter a name for the area.';
  if (trimmed.length > 60) return 'Use 60 characters or fewer.';
  if (trimmed !== current && Object.keys(groups).some((k) => k.toLowerCase() === trimmed.toLowerCase())) return 'An area with this name already exists.';
  return null;
}
