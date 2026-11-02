import { defineAsyncComponent, markRaw } from 'vue';
const Results = defineAsyncComponent(() => import('./Results.vue'));
const SourceIcon = defineAsyncComponent(() => import('./SourceIcon.vue'));
const MetricIcon = markRaw(defineAsyncComponent(() => import('@/components/ui/UiIcon.vue')));
import type {
  IntegrationPresenter,
  IntegrationResource,
  IntegrationSourcePresentation,
  SourceIcon as SourceIconModel,
  SourceRow,
} from '../types';
import type { IconName } from '@/components/ui/icons';
import type { FuelPrices } from '@/types/integrations/fuel';
import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations';
import { fuelHelp } from './help';
import { cheapestFuelPrice, fuelPriceValue } from '@/domain/integrations/fuel';
import { serviceHomeURL } from '@/domain/integrations/service-links';
import { formatLocalDateTime } from '@/utils/date.utils';
import { formatQuantity } from '@/utils/units';

function metricIcon(name: IconName): SourceIconModel {
  return { component: MetricIcon, props: { name } };
}

/** Compact rows lead with the nearest station, its cheapest grade and distance. */
function metricRows(fuel: FuelPrices): SourceRow[] {
  const first = fuel.stations[0];
  if (!first) return [];
  const cheapest = cheapestFuelPrice(first);
  const rows: SourceRow[] = [
    {
      id: 'cheapest',
      title: 'Cheapest',
      value: cheapest ? `${cheapest.label} ${fuelPriceValue(cheapest)}` : 'No prices',
      colour: null,
      icon: metricIcon('fuel'),
    },
  ];
  if (first.distance_metres != null)
    rows.push({
      id: 'distance',
      title: 'Distance',
      value: formatQuantity(first.distance_metres, 'distance'),
      colour: null,
      icon: metricIcon('pin'),
    });
  const stationLocation = first.postcode || first.address;
  if (stationLocation)
    rows.push({
      id: 'station',
      title: 'Station',
      value: stationLocation,
      colour: null,
      icon: metricIcon('pinOutline'),
    });
  return rows;
}

function sources(instance: ConnectionInstance, resource: IntegrationResource): IntegrationSourcePresentation[] {
  if (resource.contract_version !== 'fuel_prices/v1') return [];
  const first = resource.stations[0];
  const cheapest = first ? cheapestFuelPrice(first) : null;
  let status = resource.stations.length
    ? resource.stations.length === 1
      ? '1 station'
      : `${resource.stations.length} stations`
    : 'No nearby stations';
  if (resource.freshness === 'stale') status = 'Needs refreshing';
  if (!instance.enabled) status = 'Disabled';
  if (instance.last_error) status = instance.last_error.message;
  return [
    {
      reference: { kind: 'fuel_prices', instance_id: instance.id, resource_id: resource.resource_id },
      title: 'Fuel prices',
      icon: { component: markRaw(SourceIcon), props: {} },
      value: first?.name ?? 'No nearby stations',
      detail: first
        ? cheapest
          ? `${cheapest.label} ${fuelPriceValue(cheapest)}${first.distance_metres != null ? ` · ${formatQuantity(first.distance_metres, 'distance')}` : ''}`
          : 'No prices returned'
        : 'Provider returned no stations in this radius',
      origin: `Search area: ${resource.origin.label}`,
      status,
      checked: formatLocalDateTime(resource.last_checked_at),
      colour: null,
      warning: !instance.enabled || !!instance.last_error || resource.freshness === 'stale',
      error: !!instance.last_error,
      rows: metricRows(resource),
    },
  ];
}

function fuelWidget(instance: ConnectionInstance) {
  const source = instance.resources.flatMap((resource) => sources(instance, resource))[0];
  return {
    remaining: 0,
    remainingLabel: 'more results',
    status: source?.status ?? 'Waiting for fuel prices',
    checked: source?.checked ?? null,
    icon: source?.icon ?? { component: markRaw(SourceIcon), props: {} },
    rows: source
      ? [
          {
            id: source.reference.resource_id,
            name: source.title,
            colour: null,
            nextDateISO: null,
            date: source.value,
          },
        ]
      : [],
  };
}

function fuelSetup(provider: ProviderDefinition) {
  const keys: Record<string, string> = {
    fuel_finder: 'https://www.developer.fuel-finder.service.gov.uk/fuel-finder/get-started-ifr/onelogin',
  };
  const notices = [
    'Choose a location to send to the selected fuel provider. Prices are refreshed manually or on the saved schedule.',
  ];
  const services = [];
  if (provider.setup.search.base_url)
    services.push({ label: 'Location search', url: serviceHomeURL(provider.setup.search.base_url) });
  if (provider.setup.search.postcode)
    services.push({ label: 'UK postcode search', url: serviceHomeURL(provider.setup.search.postcode.base_url) });
  if (provider.auth.mode === 'oauth2_client_credentials')
    notices.push(
      'Save the Fuel Finder client ID and secret before searching; they are sent only to its token endpoint. The first collection can take a couple of minutes.'
    );
  return {
    title: 'Set up fuel prices',
    sectionTitle: 'Your area',
    defaultName: 'Fuel prices',
    searchLabel: 'UK town or postcode',
    searchPlaceholder: 'For example Sheffield or S8 9ED',
    noResults: 'No UK location was found. Fuel Finder covers the United Kingdom; try a UK town or postcode.',
    searchButton: 'Find location',
    choiceLabel: 'Choose your location',
    choicePlaceholder: 'Select a location',
    searchMaxLength: 120,
    autocomplete: 'off',
    notice: notices.join(' '),
    keyURL: keys[provider.id],
    providerURL: serviceHomeURL(provider.base_url),
    services,
  };
}

/** Fuel owns its reference, summaries, icons and results component. */
export const fuelPresenter: IntegrationPresenter = {
  setup: fuelSetup,
  contractVersion: 'fuel_prices/v1',
  integrationType: 'fuel_prices',
  help: fuelHelp,
  widget: fuelWidget,
  sourceKinds: { fuel_prices: { title: 'Fuel prices', requiresService: false } },
  sources,
  icon: () => ({ component: markRaw(SourceIcon), props: {} }),
  results: markRaw(Results),
};
