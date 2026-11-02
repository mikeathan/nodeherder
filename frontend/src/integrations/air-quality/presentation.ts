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
import type { AirQualityForecast } from '@/types/integrations/air-quality';
import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations';
import { airQualityHelp } from './help';
import {
  airQualityCategoryLabel,
  airQualityIndexFamilyLabel,
  dominantPollen,
} from '@/domain/integrations/air-quality';
import { formatQuantity } from '@/utils/units';
import { serviceHomeURL } from '@/domain/integrations/service-links';
import { formatLocalDateTime } from '@/utils/date.utils';

/** Build a capability-owned metric icon for the compact Home card. */
function metricIcon(name: IconName): SourceIconModel {
  return { component: MetricIcon, props: { name } };
}

/** Current values, falling back to the first forecast hour for compact presentation. */
function currentValues(forecast: AirQualityForecast) {
  return forecast.current ?? forecast.hourly[0] ?? null;
}

/** Compact capability-owned rows for the Home/Panel card. */
function metricRows(forecast: AirQualityForecast): SourceRow[] {
  const current = currentValues(forecast);
  if (!current) return [];
  const rows: SourceRow[] = [];
  if (current.pm2_5 != null)
    rows.push({
      id: 'pm25',
      title: 'PM2.5',
      value: formatQuantity(current.pm2_5, 'pollutant'),
      colour: null,
      icon: metricIcon('weatherParticle'),
    });
  if (current.pm10 != null)
    rows.push({
      id: 'pm10',
      title: 'PM10',
      value: formatQuantity(current.pm10, 'pollutant'),
      colour: null,
      icon: metricIcon('weatherParticle'),
    });
  const pollen = dominantPollen(forecast.pollen);
  if (pollen && pollen.value > 0)
    rows.push({
      id: 'pollen',
      title: 'Pollen',
      value: `${pollen.label} ${formatQuantity(pollen.value, 'pollen')}`,
      colour: null,
      icon: metricIcon('weatherPollen'),
    });
  return rows;
}

function sources(instance: ConnectionInstance, resource: IntegrationResource): IntegrationSourcePresentation[] {
  if (resource.contract_version !== 'air_quality/v1') return [];
  const current = currentValues(resource);
  const valueParts = [airQualityCategoryLabel(resource.category)];
  if (resource.index != null)
    valueParts.push(`${airQualityIndexFamilyLabel(resource.index_family)} ${formatQuantity(resource.index, 'aqi')}`);
  let status = 'Air quality';
  if (!current) status = 'No saved air quality';
  else if (resource.freshness === 'stale') status = 'Needs refreshing';
  if (!instance.enabled) status = 'Disabled';
  if (instance.last_error) status = instance.last_error.message;
  return [
    {
      reference: { kind: 'air_quality', instance_id: instance.id, resource_id: resource.resource_id },
      title: 'Air quality',
      icon: { component: markRaw(SourceIcon), props: { category: resource.category } },
      value: valueParts.join(' · '),
      detail: `Captured ${formatLocalDateTime(resource.captured_at)}`,
      origin: resource.location,
      status,
      checked: formatLocalDateTime(resource.last_checked_at),
      colour: null,
      warning:
        !instance.enabled ||
        !!instance.last_error ||
        resource.freshness === 'stale' ||
        !current ||
        resource.category === 'unknown',
      error: !!instance.last_error,
      rows: metricRows(resource),
    },
  ];
}

function airQualityWidget(instance: ConnectionInstance) {
  const source = instance.resources.flatMap((resource) => sources(instance, resource))[0];
  return {
    remaining: 0,
    remainingLabel: 'more results',
    status: source?.status ?? 'Waiting for air quality',
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

function airQualitySetup(provider: ProviderDefinition) {
  const notices = [
    'Choose a location to send to Open-Meteo Air Quality. Data by CAMS. Refreshes run manually or on the saved schedule.',
  ];
  const services = [];
  if (provider.setup.search.base_url)
    services.push({ label: 'Location search', url: serviceHomeURL(provider.setup.search.base_url) });
  if (provider.setup.search.postcode)
    services.push({ label: 'UK postcode search', url: serviceHomeURL(provider.setup.search.postcode.base_url) });
  return {
    title: 'Set up air quality',
    sectionTitle: 'Your location',
    defaultName: 'Air quality',
    searchLabel: 'Town or postcode',
    searchPlaceholder: 'For example Sheffield or S8 9ED',
    noResults: 'No location was found. Try a nearby town or postcode.',
    searchButton: 'Find location',
    choiceLabel: 'Choose your location',
    choicePlaceholder: 'Select a location',
    searchMaxLength: 120,
    autocomplete: 'off',
    notice: notices.join(' '),
    providerURL: serviceHomeURL(provider.base_url),
    services,
  };
}

/** Air quality owns its reference, summaries, icons and results component. */
export const airQualityPresenter: IntegrationPresenter = {
  setup: airQualitySetup,
  contractVersion: 'air_quality/v1',
  integrationType: 'air_quality',
  help: airQualityHelp,
  widget: airQualityWidget,
  sourceKinds: { air_quality: { title: 'Air quality', requiresService: false } },
  sources,
  icon: () => ({ component: markRaw(SourceIcon), props: {} }),
  results: markRaw(Results),
};
