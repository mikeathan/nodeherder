import type { HomeCard, HomeCardSource, HomeIntegrationCard } from '@/types/home.type';
import type { ConnectionInstance } from '@/types/integrations.type';
import {
  presentWasteSchedule,
  wasteServiceDateLabel,
  type WasteSchedulePresentation,
  type WasteServiceCard,
} from './waste-schedule';
import { connectionStatus } from './connection-board';
import { homeSourceKey } from './home-layout';
import { exposeLabel } from './exposes';

export interface HomeCollectionRow {
  id: string;
  title: string;
  value: string;
  colour: string | null;
}

/** Derived source presentation shared by Home, panel and picker; it owns no placement or transport. */
export interface HomeSourcePresentation {
  reference: HomeCardSource;
  title: string;
  value: string;
  origin: string;
  status: string;
  checked: string;
  colour: string | null;
  warning: boolean;
  rows: HomeCollectionRow[];
}

type ResourceContext = Pick<HomeSourcePresentation, 'origin' | 'status' | 'checked' | 'warning'>;

function resourceContext(instance: ConnectionInstance, schedule: WasteSchedulePresentation): ResourceContext {
  return {
    origin: instance.name,
    status: connectionStatus(instance, [schedule], true),
    checked: schedule.checkedLabel,
    warning: !instance.enabled || Boolean(instance.last_error) || schedule.freshness !== 'fresh',
  };
}

function collectionRow(service: WasteServiceCard): HomeCollectionRow {
  return {
    id: service.service_id,
    title: service.title,
    value: wasteServiceDateLabel(service),
    colour: service.bin_colour,
  };
}

function summaryPresentation(
  instanceId: string,
  schedule: WasteSchedulePresentation,
  context: ResourceContext
): HomeSourcePresentation {
  return {
    ...context,
    reference: { kind: 'waste_summary', instance_id: instanceId, resource_id: schedule.resource_id },
    title: context.origin,
    value: 'Collection schedule',
    colour: null,
    rows: schedule.cards.map(collectionRow),
  };
}

function servicePresentation(
  instanceId: string,
  resourceId: string,
  service: WasteServiceCard,
  context: ResourceContext
): HomeSourcePresentation {
  return {
    ...context,
    reference: {
      kind: 'waste_service',
      instance_id: instanceId,
      resource_id: resourceId,
      service_id: service.service_id,
    },
    title: service.title,
    value: wasteServiceDateLabel(service),
    colour: service.bin_colour,
    rows: [],
  };
}

/** Index each accepted resource once. Summary/services reuse the existing calendar and lifecycle
 * projections; this function neither changes sources nor stores current values in placements. */
export function homeIntegrationSources(instances: readonly ConnectionInstance[]): Map<string, HomeSourcePresentation> {
  const sources = new Map<string, HomeSourcePresentation>();
  for (const instance of instances) {
    for (const resource of instance.resources) {
      if (resource.contract_version !== 'waste_schedule/v1') continue;
      const schedule = presentWasteSchedule(resource);
      const context = resourceContext(instance, schedule);
      const summary = summaryPresentation(instance.id, schedule, context);
      sources.set(homeSourceKey(summary.reference), summary);
      for (const service of schedule.cards) {
        const item = servicePresentation(instance.id, schedule.resource_id, service, context);
        sources.set(homeSourceKey(item.reference), item);
      }
    }
  }
  return sources;
}

/** Resolve a stable source reference, preserving an explicit placeholder if it disappears. Never
 * substitute another service based on dates, labels or current display order. */
export function homeIntegrationCard(
  card: HomeIntegrationCard,
  sources: ReadonlyMap<string, HomeSourcePresentation>
): HomeSourcePresentation {
  const accepted = sources.get(homeSourceKey(card));
  if (accepted) return accepted;
  return {
    reference: card,
    title: card.kind === 'waste_summary' ? 'Bin collections' : 'Collection date',
    value: 'Unavailable',
    origin: 'Integration result unavailable',
    status: 'The source or service is no longer available.',
    checked: '',
    colour: null,
    warning: true,
    rows: [],
  };
}

/** Label a placement from its current source presentation; missing placements stay explicit. */
export function homeCardTitle(
  card: HomeCard | undefined,
  sources: ReadonlyMap<string, HomeSourcePresentation>
): string {
  if (!card) return 'Card unavailable';
  if (card.kind === 'device') return exposeLabel(card.expose);
  return homeIntegrationCard(card, sources).title;
}
