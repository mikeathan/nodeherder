/*
 * One grouping rule shared by every place that lists a device's values (Home tile picker,
 * History selector). The group comes from the value's meaning, not its raw key: ESPHome
 * entities carry a device_class and Zigbee2MQTT ones use canonical names, and
 * exposeSemanticName() already folds both into the same vocabulary.
 */
import type { Expose } from '@/types/device';
import { ExposeAccessModes, ExposeCategories } from '@/types/device.type';
import { exposeKind, exposeSemanticName, type ExposeKind } from './exposes';

/** A value's group: its colour family, plus a catch-all for diagnostic values with no known family. */
export type ExposeGroup = Exclude<ExposeKind, 'misc'> | 'other' | 'diagnostic';

/** Display order and wording of the groups. */
export const EXPOSE_GROUPS: ReadonlyArray<{ id: ExposeGroup; label: string }> = [
  { id: 'motion', label: 'Presence & motion' },
  { id: 'light', label: 'Light level' },
  { id: 'temp', label: 'Temperature' },
  { id: 'humidity', label: 'Humidity & moisture' },
  { id: 'air', label: 'Air quality' },
  { id: 'contact', label: 'Doors & windows' },
  { id: 'alarm', label: 'Alarms' },
  { id: 'switch', label: 'Switches & lights' },
  { id: 'power', label: 'Power' },
  { id: 'energy', label: 'Energy' },
  { id: 'battery', label: 'Battery' },
  { id: 'signal', label: 'Signal' },
  { id: 'event', label: 'Events' },
  { id: 'diagnostic', label: 'Diagnostics' },
  { id: 'other', label: 'Other' },
];

/** Groups that are noise for a casual user and stay out of the default view. */
export const MINOR_GROUPS: ReadonlySet<ExposeGroup> = new Set<ExposeGroup>(['other', 'diagnostic']);

export function exposeGroupLabel(group: ExposeGroup): string {
  return EXPOSE_GROUPS.find((item) => item.id === group)?.label ?? 'Other';
}

export function exposeGroup(expose: Expose): ExposeGroup {
  const kind = exposeKind(exposeSemanticName(expose));
  if (kind !== 'misc') return kind;
  return expose.category === ExposeCategories.Diagnostic ? 'diagnostic' : 'other';
}

/** A value a person can look at: not configuration and not write-only (buttons, identify). */
export function isObservableExpose(expose: Expose): boolean {
  return expose.category !== ExposeCategories.Config && expose.access_mode !== ExposeAccessModes.Write;
}

/** Sort rank of a group so lists always read in the same order. */
export function exposeGroupRank(group: ExposeGroup): number {
  return EXPOSE_GROUPS.findIndex((item) => item.id === group);
}
