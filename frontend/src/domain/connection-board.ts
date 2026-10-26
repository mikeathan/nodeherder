import { moveItem } from './dashboard';
import { compareCollectionDates } from './waste-calendar';
import { presentWasteSchedule } from './waste-schedule';
import type { WasteSchedulePresentation, WasteServiceCard } from './waste-schedule';
import type { ConnectionBoardEntry, ConnectionInstance, WasteService } from '@/types/integrations.type';

/** One display row, identified by resource/service IDs rather than its chronological position. */
export interface ConnectionWidgetRow {
  id: string;
  name: string;
  colour: string | null;
  nextDateISO: string | null;
  date: string;
}

/** Compact connection view model; it contains display data and no transport/lifecycle ownership. */
export interface ConnectionWidgetSummary {
  rows: ConnectionWidgetRow[];
  remaining: number;
  status: string;
  checked: string | null;
  isWaste: boolean;
}

const MAX_WIDGET_ROWS = 3;
const missingDateLabels: Readonly<Record<WasteService['date_state'], string>> = {
  known: 'No upcoming date',
  confirmed_empty: 'None scheduled',
  unavailable: 'Dates unavailable',
};

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

/** Keep confirmed-empty, unavailable and exhausted horizons distinct when there is no next date. */
function rowDateLabel(card: WasteServiceCard): string {
  if (card.nextDateShort !== null) return card.nextDateShort;
  return missingDateLabels[card.date_state];
}

/** Map one card without allowing date/order changes to become a new widget identity. */
function presentCollectionRow(card: WasteServiceCard, resourceId: string): ConnectionWidgetRow {
  return {
    id: `${resourceId}/${card.service_id}`,
    name: card.title,
    colour: card.bin_colour,
    nextDateISO: card.nextDateISO,
    date: rowDateLabel(card),
  };
}

/** Merging already-ordered resources still requires one overall order before row limiting. */
function collectionRows(resources: readonly WasteSchedulePresentation[]): ConnectionWidgetRow[] {
  return resources
    .flatMap((resource) => resource.cards.map((card) => presentCollectionRow(card, resource.resource_id)))
    .sort((left, right) => compareCollectionDates(left.nextDateISO, right.nextDateISO));
}

/** Health/lifecycle messages take precedence over an otherwise valid saved schedule. */
function widgetStatus(
  instance: ConnectionInstance,
  resources: readonly WasteSchedulePresentation[],
  isWaste: boolean
): string {
  if (!instance.enabled) return 'Disabled';
  if (instance.last_error) return instance.last_error.message;
  if (instance.operation_id) return 'Checking for updates';
  if (resources.some((resource) => resource.freshness === 'stale')) return 'Needs refreshing';
  if (resources.length) return resources[0].statusLabel;
  return isWaste ? 'Waiting for collection dates' : 'Waiting for information';
}

/**
 * Compose a compact widget from accepted connection/resource state.
 * @returns At most three chronological rows plus lifecycle feedback; never starts provider work
 * or infers recurring dates. Saved objects and the user's connection layout remain untouched.
 */
export function connectionWidget(instance: ConnectionInstance): ConnectionWidgetSummary {
  const isWaste = instance.provider_key.startsWith('waste_collection/');
  const resources = instance.resources.map(presentWasteSchedule);
  const rows = collectionRows(resources);
  return {
    rows: rows.slice(0, MAX_WIDGET_ROWS),
    remaining: Math.max(0, rows.length - MAX_WIDGET_ROWS),
    status: widgetStatus(instance, resources, isWaste),
    checked: resources[0]?.checkedLabel ?? null,
    isWaste,
  };
}
