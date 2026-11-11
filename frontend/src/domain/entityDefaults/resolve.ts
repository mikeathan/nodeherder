import type { Expose } from '@/types/device';
import { ExposeAttributeKeys } from '@/types/device.type';
import { exposeRange, numericStep } from '@/domain/controls';
import { controlOf, exposeDisplayLabel, exposePresets, isWritable } from '@/domain/exposes';
import { normalizeEntityKey } from '@/domain/entityKey';
import { ENTITY_DEFAULTS_BY_OBJECT_ID } from './catalog';
import type { EntityDefaultTarget, EntityDefaultValue } from './types';

function applicable(expose: Expose, value: EntityDefaultValue): boolean {
  if (!isWritable(expose)) return false;
  switch (controlOf(expose)) {
    case 'switch':
      return typeof value === 'boolean';
    case 'slider':
    case 'number': {
      if (typeof value !== 'number' || !Number.isFinite(value)) return false;
      const range = exposeRange(expose);
      if (!range) return true;
      const steps = (value - range.min) / numericStep(expose);
      return value >= range.min && value <= range.max && Math.abs(steps - Math.round(steps)) < 1e-6;
    }
    case 'choice':
      return typeof value === 'string' && exposePresets(expose).some((preset) => preset.value === value);
    case 'text':
      return typeof value === 'string';
    default:
      return false;
  }
}

/**
 * Resolves the documented default for an expose.
 *
 * Matching uses `object_id`, then the normalized label for older records. A value that no
 * longer fits the current control type, range/step or select options is ignored rather than
 * coerced, so a firmware change cannot make reset send an invalid command.
 */
export function resolveEntityDefault(expose: Expose): EntityDefaultTarget | null {
  const objectID = normalizeEntityKey(expose.attributes?.[ExposeAttributeKeys.ObjectID]);
  const label = normalizeEntityKey(exposeDisplayLabel(expose));
  const entry =
    (objectID ? ENTITY_DEFAULTS_BY_OBJECT_ID[objectID] : undefined) ??
    (label ? ENTITY_DEFAULTS_BY_OBJECT_ID[label] : undefined);
  if (!entry || !applicable(expose, entry.value)) return null;
  return { expose, value: entry.value, source: entry.source };
}

/** All resettable exposes in the supplied order. */
export function entityDefaultTargets(entities: readonly Expose[]): EntityDefaultTarget[] {
  const targets: EntityDefaultTarget[] = [];
  for (const expose of entities) {
    const target = resolveEntityDefault(expose);
    if (target) targets.push(target);
  }
  return targets;
}

/** Confirmation copy naming the affected settings without claiming the reset succeeded. */
export function resetDefaultsMessage(targets: readonly EntityDefaultTarget[]): string {
  const labels = targets.map((target) => exposeDisplayLabel(target.expose));
  const preview =
    labels.length <= 3 ? labels.join(', ') : `${labels.slice(0, 3).join(', ')} and ${labels.length - 3} more`;
  const noun = labels.length === 1 ? 'setting' : 'settings';
  return `Reset ${labels.length} documented ${noun} to their defaults: ${preview}. Values update only after the device confirms each change.`;
}
