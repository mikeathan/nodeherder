import { markRaw, defineAsyncComponent } from 'vue';
const WasteResults = defineAsyncComponent(() => import('./Results.vue'));
import { wasteWidget, wasteIcon } from './widget';

import type {
  IntegrationPresenter,
  IntegrationSourcePresentation as HomeSourcePresentation,
  IntegrationResource,
  SourceRow,
} from '../types';
import type { ConnectionInstance } from '@/types/integrations.type';
import {
  presentWasteSchedule,
  wasteServiceDateLabel,
  type WasteSchedulePresentation,
  type WasteServiceCard,
} from '@/domain/waste-schedule';
import { connectionStatus } from './status';
type ResourceContext = Pick<HomeSourcePresentation, 'origin' | 'status' | 'checked' | 'warning' | 'error'>;

function resourceContext(instance: ConnectionInstance, schedule: WasteSchedulePresentation): ResourceContext {
  return {
    origin: instance.name,
    status: connectionStatus(instance, [schedule], true),
    checked: schedule.checkedLabel,
    warning: !instance.enabled || Boolean(instance.last_error) || schedule.freshness !== 'fresh',
    error: Boolean(instance.last_error),
  };
}

function collectionRow(service: WasteServiceCard): SourceRow {
  return {
    id: service.service_id,
    title: service.title,
    value: wasteServiceDateLabel(service),
    colour: service.bin_colour,
    icon: wasteIcon('waste_service', service.bin_colour),
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
    pickerLabel: context.origin + ' · complete schedule',
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
    selectable: false,
    value: wasteServiceDateLabel(service),
    colour: service.bin_colour,
    rows: [],
  };
}

function sources(instance: ConnectionInstance, resource: IntegrationResource): HomeSourcePresentation[] {
  if (resource.contract_version !== 'waste_schedule/v1') return [];
  const schedule = presentWasteSchedule(resource),
    context = resourceContext(instance, schedule);
  return [
    summaryPresentation(instance.id, schedule, context),
    ...schedule.cards.map((service) => servicePresentation(instance.id, schedule.resource_id, service, context)),
  ];
}
/** Waste owns its reference types, source summaries, icons and results component. */
export const wastePresenter: IntegrationPresenter = {
  setup: () => ({
    title: 'Set up bin collections',
    sectionTitle: 'Your address',
    defaultName: 'Bin collections',
    searchLabel: 'Postcode',
    searchButton: 'Find address',
    choiceLabel: 'Choose your address',
    choicePlaceholder: 'Select an address',
    searchMaxLength: 10,
    autocomplete: 'postal-code',
    notice:
      'The provider receives your postcode and selected address to retrieve collection dates. Refreshes are manual.',
  }),
  contractVersion: 'waste_schedule/v1',
  integrationType: 'waste_collection',
  widget: wasteWidget,
  sourceKinds: {
    waste_summary: { title: 'Bin collections', requiresService: false },
    waste_service: { title: 'Collection date', requiresService: true },
  },
  sources,
  icon: wasteIcon,
  results: markRaw(WasteResults),
};
