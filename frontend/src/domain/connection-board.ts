import { moveItem } from './dashboard';
import type { ConnectionBoardEntry, ConnectionInstance } from '@/types/integrations.type';
import { connectionPresentation } from '@/integrations/registry';
export type {
  IntegrationWidget as ConnectionWidgetSummary,
  IntegrationWidgetRow as ConnectionWidgetRow,
} from '@/integrations/types';
/**
 * Reconcile layout membership without losing saved order or hiding new connections.
 * @param entries Accepted placements, possibly read just before an instance was added/deleted.
 * @param ids Current live instance IDs.
 * @returns Owned placements; neither input is changed and duplicate references are removed.
 */
export function reconcileBoard(
  entries: readonly ConnectionBoardEntry[],
  ids: readonly string[]
): ConnectionBoardEntry[] {
  const live = new Set(ids);
  const seen = new Set<string>();
  const result = entries
    .filter((entry) => {
      if (!live.has(entry.instance_id) || seen.has(entry.instance_id)) return false;
      seen.add(entry.instance_id);
      return true;
    })
    .map((entry) => ({ ...entry }));
  for (const id of ids) if (!seen.has(id)) result.push({ instance_id: id, hidden: false });
  return result;
}

/**
 * Move a visible widget while preserving hidden placements in the full membership.
 * @param offset Relative movement in visible reading order; an out-of-range move is ignored.
 * @returns A new array of owned entries suitable for an unsaved layout draft.
 */
export function moveBoardEntry(
  entries: readonly ConnectionBoardEntry[],
  id: string,
  offset: number
): ConnectionBoardEntry[] {
  const visible = entries.filter((entry) => !entry.hidden);
  const current = visible.findIndex((entry) => entry.instance_id === id);
  const target = current + offset;
  if (current < 0 || target < 0 || target >= visible.length) return entries.map((entry) => ({ ...entry }));
  const reordered = moveItem(visible, current, target);
  let index = 0;
  return entries.map((entry) => ({ ...(entry.hidden ? entry : reordered[index++]) }));
}

/** Project a widget through its registered capability without inspecting resource payloads. */
export function connectionWidget(instance: ConnectionInstance) {
  return connectionPresentation(instance);
}
