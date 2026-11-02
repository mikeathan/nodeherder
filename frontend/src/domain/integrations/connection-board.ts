import { moveItem } from '../dashboard';
import type { ConnectionBoardEntry, ConnectionCatalog, ConnectionInstance } from '@/types/integrations';
import { connectionPresentation } from '@/integrations/registry';
import { integrationTypeTitle, matchesCatalogQuery } from './catalog';
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
  offset: number,
  typeOf?: (instanceId: string) => string
): ConnectionBoardEntry[] {
  const visible = entries.filter((entry) => !entry.hidden);
  const current = visible.findIndex((entry) => entry.instance_id === id);
  if (current < 0) return entries.map((entry) => ({ ...entry }));
  let target = current + offset;
  if (typeOf) {
    const type = typeOf(id);
    target = current;
    for (let index = current + offset; index >= 0 && index < visible.length; index += offset) {
      if (typeOf(visible[index].instance_id) === type) {
        target = index;
        break;
      }
    }
  }
  if (target < 0 || target >= visible.length || target === current)
    return entries.map((entry) => ({ ...entry }));
  const reordered = moveItem(visible, current, target);
  let index = 0;
  return entries.map((entry) => ({ ...(entry.hidden ? entry : reordered[index++]) }));
}

/** Project a widget through its registered capability without inspecting resource payloads. */
export function connectionWidget(instance: ConnectionInstance) {
  return connectionPresentation(instance);
}

export type ConnectionEntryStatus = 'ready' | 'attention' | 'disabled';

/** Coarse status used only by board filters; detailed status remains in the capability widget. */
export function connectionEntryStatus(instance: ConnectionInstance): ConnectionEntryStatus {
  if (!instance.enabled) return 'disabled';
  if (instance.last_error || instance.resources.some((resource) => resource.freshness === 'stale'))
    return 'attention';
  return 'ready';
}

/** Match a saved integration against the board search without decoding provider payloads. */
export function connectionMatchesQuery(
  instance: ConnectionInstance,
  query: string,
  catalog: ConnectionCatalog | null
): boolean {
  const typeId = instance.provider_key.split('/')[0];
  const summary = connectionWidget(instance);
  return matchesCatalogQuery(
    query,
    instance.name,
    instance.postcode,
    instance.provider_key,
    integrationTypeTitle(typeId, catalog),
    ...summary.rows.flatMap((row) => [row.name, row.date])
  );
}
