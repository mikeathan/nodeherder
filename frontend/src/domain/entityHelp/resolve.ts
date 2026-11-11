import type { Expose } from '@/types/device';
import { ExposeAttributeKeys } from '@/types/device.type';
import { exposeDisplayLabel, showsExposeDescription } from '@/domain/exposes';
import { normalizeEntityKey } from '@/domain/entityKey';
import { ENTITY_HELP_BY_DEVICE_CLASS, ENTITY_HELP_BY_OBJECT_ID, type EntityHelpEntry } from './catalog';
import { EntityHelpKinds, type EntityHelp, type EntityHelpResolver } from './types';

function lookup(key: string, entries: Readonly<Record<string, EntityHelpEntry>>): EntityHelpEntry | undefined {
  return key ? entries[key] : undefined;
}

/**
 * Resolves help from curated data: `object_id`, then the normalized label, then the
 * generic `device_class`. The label fallback keeps older ESPHome records working.
 */
export function catalogEntityHelp(expose: Expose): EntityHelp | null {
  const objectID = normalizeEntityKey(expose.attributes?.[ExposeAttributeKeys.ObjectID]);
  const label = normalizeEntityKey(exposeDisplayLabel(expose));
  const deviceClass = normalizeEntityKey(expose.attributes?.[ExposeAttributeKeys.DeviceClass]);
  const entry =
    lookup(objectID, ENTITY_HELP_BY_OBJECT_ID) ??
    lookup(label, ENTITY_HELP_BY_OBJECT_ID) ??
    lookup(deviceClass, ENTITY_HELP_BY_DEVICE_CLASS);
  return entry ? { text: entry.text, kind: EntityHelpKinds.Catalog, source: entry.source } : null;
}

/**
 * Returns a description supplied by the provider itself (for example Zigbee2MQTT).
 * A description that merely repeats the display label is not help.
 */
export function providerEntityHelp(expose: Expose): EntityHelp | null {
  const text = expose.description?.trim();
  if (!text || !showsExposeDescription(expose, true)) return null;
  return { text, kind: EntityHelpKinds.Provider };
}

/**
 * Ordered help sources; the first non-null answer wins.
 *
 * Append a source here to extend resolution without touching consumers.
 */
export const ENTITY_HELP_RESOLVERS: readonly EntityHelpResolver[] = [providerEntityHelp, catalogEntityHelp];

/** Resolves the best available explanation for an expose, or `null` when none is known. */
export function resolveEntityHelp(expose: Expose): EntityHelp | null {
  for (const resolve of ENTITY_HELP_RESOLVERS) {
    const help = resolve(expose);
    if (help) return help;
  }
  return null;
}
