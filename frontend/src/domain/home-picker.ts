import type { Device } from '@/types/device';
import type { HomeCard, HomeCardSource } from '@/types/home.type';
import type { HomeSourcePresentation } from './home-cards';
import { ExposeCategories } from '@/types/device.type';
import { exposeLabel } from './exposes';
import { homeSourceKey, newHomeId } from './home-layout';

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
}

/** List source selectors without changing devices or integration configuration. */
export function homePickerOptions(devices: readonly Device[]): HomePickerOption[] {
  const options = [{ value: 'integrations', label: 'Integration results' }];
  for (const device of devices) options.push({ value: `device:${device.id}`, label: device.friendly_name });
  return options;
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
    let label = item.title;
    if (item.reference.kind === 'waste_summary') label = `${label} · complete schedule`;
    choices.push(choice(item.reference, label, item.origin, existing));
  }
  return choices;
}

function deviceChoices(device: Device, existing: ReadonlySet<string>): HomePickerChoice[] {
  const choices: HomePickerChoice[] = [];
  for (const expose of Object.values(device.exposes)) {
    if (expose.category === ExposeCategories.Config) continue;
    const reference: HomeCardSource = { kind: 'device', device_id: device.id, expose: expose.name };
    choices.push(choice(reference, exposeLabel(expose.name), device.friendly_name, existing));
  }
  return choices;
}

/** Project selectable values for one source. Deduplication is local to the destination area;
 * another placement may legitimately show the same source elsewhere. Config exposes are excluded. */
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
  if (source) return 'This device reports no values that can be shown as tiles.';
  return '';
}
