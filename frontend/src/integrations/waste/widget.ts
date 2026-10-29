import { markRaw, defineAsyncComponent } from 'vue';
import type { ConnectionInstance } from '@/types/integrations.type';
import type {
  IntegrationWidget as ConnectionWidgetSummary,
  IntegrationWidgetRow as ConnectionWidgetRow,
} from '../types';
import {
  presentWasteSchedule,
  wasteServiceDateLabel,
  type WasteSchedulePresentation,
  type WasteServiceCard,
} from '@/domain/waste-schedule';
import { compareCollectionDates } from '@/domain/waste-calendar';
import { connectionStatus } from './status';
const MAX_WIDGET_ROWS = 3;
const Icon = markRaw(defineAsyncComponent(() => import('./SourceIcon.vue')));
/** Resolve waste summary/service icons without exposing them to generic cards. */
export function wasteIcon(kind: string, colour: string | null) {
  return { component: Icon, props: { kind, colour } };
}
/** Map one card without allowing date/order changes to become a new widget identity. */
function presentCollectionRow(card: WasteServiceCard, resourceId: string): ConnectionWidgetRow {
  return {
    id: `${resourceId}/${card.service_id}`,
    name: card.title,
    colour: card.bin_colour,
    nextDateISO: card.nextDateISO,
    date: wasteServiceDateLabel(card),
    icon: wasteIcon('waste_service', card.bin_colour),
  };
}

/** Merging already-ordered resources still requires one overall order before row limiting. */
function collectionRows(resources: readonly WasteSchedulePresentation[]): ConnectionWidgetRow[] {
  return resources
    .flatMap((resource) => resource.cards.map((card) => presentCollectionRow(card, resource.resource_id)))
    .sort((left, right) => compareCollectionDates(left.nextDateISO, right.nextDateISO));
}

/**
 * Compose a compact widget from accepted connection/resource state.
 * @returns At most three chronological rows plus lifecycle feedback; never starts provider work
 * or infers recurring dates. Saved objects and the user's connection layout remain untouched.
 */
export function wasteWidget(instance: ConnectionInstance): ConnectionWidgetSummary {
  const isWaste = instance.provider_key.startsWith('waste_collection/');
  const resources = instance.resources
    .filter((r) => r.contract_version === 'waste_schedule/v1')
    .map(presentWasteSchedule);
  const rows = collectionRows(resources);
  return {
    rows: rows.slice(0, MAX_WIDGET_ROWS),
    remaining: Math.max(0, rows.length - MAX_WIDGET_ROWS),
    status: connectionStatus(instance, resources, isWaste),
    checked: resources[0]?.checkedLabel ?? null,
    remainingLabel: 'more collections',
    icon: wasteIcon('waste_summary', null),
  };
}
