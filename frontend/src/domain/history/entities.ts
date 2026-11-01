/*
 * History entity options use the same display labels as the rest of the UI.
 * Option values remain stable expose keys so history queries and preferences
 * are unchanged; only the human-readable label is shared.
 */
import type { Device } from '@/types/device';
import { exposeDisplayLabel } from '@/domain/exposes';
import { ExposeTypes } from '@/types/device.type';

/** One history entity selector option. */
export type HistoryEntityOption = { value: string; label: string };

const HISTORY_TYPES = new Set<string>([ExposeTypes.Numeric, ExposeTypes.Binary, ExposeTypes.Enum]);

/** Builds history selector options in display-label order. */
export function historyEntityOptions(device: Device | undefined): HistoryEntityOption[] {
  return Object.values(device?.exposes ?? {})
    .filter((expose) => HISTORY_TYPES.has(String(expose.type)))
    .map((expose) => ({ value: expose.name, label: exposeDisplayLabel(expose) }))
    .sort((left, right) => left.label.localeCompare(right.label) || left.value.localeCompare(right.value));
}
