import type { Device } from '@/types/device';
import type { HomeCard, HomeCardSource } from '@/types/home';
import type { HomeSourcePresentation } from './cards';
import type { Expose } from '@/types/device';
import { exposeDisplayLabel } from '../exposes';
import {
  EXPOSE_GROUPS,
  MINOR_GROUPS,
  exposeGroup,
  isObservableExpose,
  type ExposeGroup,
} from '../exposeGroups';
import { homeSourceKey, newHomeId } from './layout';
import { supportsSpatialMapping } from '@/domain/spatial';

export interface HomePickerOption {
  value: string;
  label: string;
}
export interface HomePickerChoice {
  key: string;
  label: string;
  detail: string;
  reference: HomeCardSource;
  alreadyShown: boolean;
  /** Meaning of a device value; absent for integration results. */
  group?: ExposeGroup;
}

/** List source selectors without changing devices or integration configuration. */
export function homePickerOptions(devices: readonly Device[]): HomePickerOption[] {
  const options = [{ value: 'integrations', label: 'Integration results' }];
  for (const device of devices) options.push({ value: `device:${device.id}`, label: device.friendly_name });
  return options;
}

/** Selector identity is what a person can choose, not the reading behind it. */
function samePickerOptions(previous: readonly HomePickerOption[] | undefined, next: readonly HomePickerOption[]): boolean {
  return (
    !!previous &&
    previous.length === next.length &&
    next.every((option, index) => option.value === previous[index].value && option.label === previous[index].label)
  );
}

/** Reuse the previous source selectors while the choosable devices are unchanged. Every
 * device report replaces the device list, and rebuilding the native menu can drop an
 * open or newly committed choice. */
export function stableHomePickerOptions(
  previous: readonly HomePickerOption[] | undefined,
  devices: readonly Device[]
): HomePickerOption[] {
  const next = homePickerOptions(devices);
  return samePickerOptions(previous, next) ? (previous as HomePickerOption[]) : next;
}

function choice(
  reference: HomeCardSource,
  label: string,
  detail: string,
  existing: ReadonlySet<string>
): HomePickerChoice {
  const key = homeSourceKey(reference);
  return { key, label, detail, reference, alreadyShown: existing.has(key) };
}

function integrationChoices(
  sources: ReadonlyMap<string, HomeSourcePresentation>,
  existing: ReadonlySet<string>
): HomePickerChoice[] {
  const choices: HomePickerChoice[] = [];
  for (const item of sources.values()) {
    if (item.selectable === false) continue;
    const label = item.pickerLabel ?? item.title;
    choices.push(choice(item.reference, label, item.origin, existing));
  }
  return choices;
}

/** A tile needs something to show: the device has reported a value. */
function hasReading(expose: Expose): boolean {
  return expose.data !== null && expose.data !== undefined && expose.data !== '';
}

function deviceChoices(device: Device, existing: ReadonlySet<string>): HomePickerChoice[] {
  const choices: HomePickerChoice[] = [];
  // Whole-device presence is offered only when the device can carry a spatial mapping; it
  // reads the hub-derived values, so no single expose is required.
  if (supportsSpatialMapping(device)) {
    choices.push(choice({ kind: 'presence', device_id: device.id }, 'Presence', device.friendly_name, existing));
  }
  for (const expose of Object.values(device.exposes)) {
    if (!isObservableExpose(expose) || !hasReading(expose)) continue;
    const reference: HomeCardSource = { kind: 'device', device_id: device.id, expose: expose.name };
    choices.push({
      ...choice(reference, exposeDisplayLabel(expose), device.friendly_name, existing),
      group: exposeGroup(expose),
    });
  }
  return choices;
}

/** Project selectable values for one source. Deduplication is local to the destination area;
 * another placement may legitimately show the same source elsewhere. Config, write-only and not-yet-reporting exposes are excluded. */
export function homePickerChoices(
  source: string | null,
  devices: readonly Device[],
  sources: ReadonlyMap<string, HomeSourcePresentation>,
  existingCards: readonly HomeCard[]
): HomePickerChoice[] {
  const existing = new Set(existingCards.map(homeSourceKey));
  if (source === 'integrations') return integrationChoices(sources, existing);
  if (!source?.startsWith('device:')) return [];
  const deviceId = source.slice('device:'.length);
  const device = devices.find((device) => device.id === deviceId);
  return device ? deviceChoices(device, existing) : [];
}

/** Filter ids for the value list: the everyday values, every value, or one group. */
export const HOME_FILTER_MAIN = 'main';
export const HOME_FILTER_ALL = 'all';

/** Filter options for the groups a device actually has, in a fixed order. */
export function homePickerFilters(choices: readonly HomePickerChoice[]): HomePickerOption[] {
  const present = new Set(choices.map((item) => item.group));
  const groups = EXPOSE_GROUPS.filter((item) => present.has(item.id)).map((item) => ({
    value: item.id as string,
    label: item.label,
  }));
  return [{ value: HOME_FILTER_MAIN, label: 'Sensors and controls' }, ...groups, { value: HOME_FILTER_ALL, label: 'Everything' }];
}

/** Reuse the previous value filters while the groups stay the same, so a live report
 * does not rebuild an open native filter menu. */
export function stableHomePickerFilters(
  previous: readonly HomePickerOption[] | undefined,
  choices: readonly HomePickerChoice[]
): HomePickerOption[] {
  const next = homePickerFilters(choices);
  return samePickerOptions(previous, next) ? (previous as HomePickerOption[]) : next;
}

/** Narrow device choices by filter; integration results (no group) are never filtered. */
export function filterHomeChoices(choices: readonly HomePickerChoice[], filter: string): HomePickerChoice[] {
  return choices.filter((item) => {
    if (!item.group || filter === HOME_FILTER_ALL) return true;
    if (filter === HOME_FILTER_MAIN) return !MINOR_GROUPS.has(item.group);
    return item.group === filter;
  });
}

/** Create independent placement IDs for selected, eligible choices; source references stay owned copies. */
export function selectedHomeCards(choices: readonly HomePickerChoice[], selectedKeys: readonly string[]): HomeCard[] {
  const selected = new Set(selectedKeys);
  const cards: HomeCard[] = [];
  for (const item of choices) {
    if (selected.has(item.key) && !item.alreadyShown) cards.push({ ...item.reference, id: newHomeId(), panel: true });
  }
  return cards;
}

/** Explain an empty source selection without conflating absent results and device configuration. */
export function homePickerEmptyMessage(source: string | null): string {
  if (source === 'integrations')
    return 'No saved integration results are available. Configure an integration on the Integrations page first.';
  if (source) return 'This device has no readings to show yet. Values appear here once the device has reported them.';
  return '';
}
