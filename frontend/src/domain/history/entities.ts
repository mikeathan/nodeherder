/*
 * History entity options use the same display labels as the rest of the UI.
 * Option values remain stable expose keys so history queries and preferences
 * are unchanged; only the human-readable label and the grouping are shared.
 */
import type { Device } from '@/types/device';
import { exposeDisplayLabel } from '@/domain/exposes';
import { exposeGroup, exposeGroupLabel, exposeGroupRank, isObservableExpose } from '@/domain/exposeGroups';
import { ExposeTypes } from '@/types/device.type';

/** One history entity selector option; `group` names the section it belongs to. */
export type HistoryEntityOption = { value: string; label: string; group: string };

const HISTORY_TYPES = new Set<string>([ExposeTypes.Numeric, ExposeTypes.Binary, ExposeTypes.Enum]);

/** Builds history selector options ordered by group, then display label. Configuration and
 * write-only values have no history worth charting and are left out. */
export function historyEntityOptions(device: Device | undefined): HistoryEntityOption[] {
  return Object.values(device?.exposes ?? {})
    .filter((expose) => HISTORY_TYPES.has(String(expose.type)) && isObservableExpose(expose))
    .map((expose) => ({ expose, group: exposeGroup(expose), label: exposeDisplayLabel(expose) }))
    .sort(
      (a, b) =>
        exposeGroupRank(a.group) - exposeGroupRank(b.group) ||
        a.label.localeCompare(b.label) ||
        a.expose.name.localeCompare(b.expose.name)
    )
    .map(({ expose, group, label }) => ({ value: expose.name, label, group: exposeGroupLabel(group) }));
}
