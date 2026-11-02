import { markRaw, defineAsyncComponent } from 'vue';
const WeatherResults = defineAsyncComponent(() => import('./Results.vue'));
const SourceIcon = defineAsyncComponent(() => import('./SourceIcon.vue'));
import type { IntegrationPresenter, IntegrationResource, IntegrationSourcePresentation, SourceIcon, SourceRow } from '../types';
import type { IconName } from '@/components/ui/icons';
import type { WeatherForecast } from '@/types/integrations/weather';
import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations';
import { weatherHelp } from './help';
import { formatQuantity } from '@/utils/units';
import { serviceHomeURL } from '@/domain/integrations/service-links';
import { weatherHeadline, weatherHeadlineValue } from './headline';
import { calendarDateInZone, formatLocalDateTime } from '@/utils/date.utils';

const MetricIcon = markRaw(defineAsyncComponent(() => import('@/components/ui/UiIcon.vue')));

/** Build a capability-owned metric icon for the compact Home card. */
function metricIcon(name: IconName): SourceIcon {
  return { component: MetricIcon, props: { name } };
}

/** Compact card metrics: today's range, next-hour rain chance and daily maximum UV. */
function metricRows(forecast: WeatherForecast): SourceRow[] {
  const today = calendarDateInZone(forecast.evaluated_at, forecast.timezone);
  const day = forecast.daily.find((value) => value.date === today);
  const now = Date.parse(forecast.evaluated_at);
  const upcoming = forecast.hourly.find(
    (point) => Date.parse(point.time) > now && Date.parse(point.time) <= now + 3600000
  );
  const rows: SourceRow[] = [];
  if (day?.high != null || day?.low != null) {
    const temperatures: string[] = [];
    if (day?.high != null) temperatures.push(`H ${formatQuantity(day.high, 'temperature', { short: true })}`);
    if (day?.low != null) temperatures.push(`L ${formatQuantity(day.low, 'temperature', { short: true })}`);
    rows.push({
      id: 'range',
      title: 'High / Low',
      value: temperatures.join(' · '),
      colour: null,
      icon: metricIcon('weatherTemperature'),
    });
  }
  if (forecast.freshness === 'fresh' && upcoming?.precipitation_probability != null)
    rows.push({
      id: 'rain',
      title: 'Rain chance',
      value: formatQuantity(upcoming.precipitation_probability, 'percent'),
      colour: null,
      icon: metricIcon('weatherRain'),
    });
  if (day?.uv_max != null)
    rows.push({
      id: 'uv',
      title: 'UV max',
      value: formatQuantity(day.uv_max, 'uv'),
      colour: null,
      icon: metricIcon('weatherSun'),
    });
  return rows;
}

function sources(instance: ConnectionInstance, resource: IntegrationResource): IntegrationSourcePresentation[] {
  if (resource.contract_version !== 'weather_forecast/v1') return [];
  const headline = weatherHeadline(resource);
  let status = 'Forecast';
  if (!headline) status = 'No saved forecast';
  else if (resource.freshness === 'stale') status = 'Needs refreshing';
  if (!instance.enabled) status = 'Disabled';
  if (instance.last_error) status = instance.last_error.message;
  return [
    {
      reference: { kind: 'weather', instance_id: instance.id, resource_id: resource.resource_id },
      title: 'Weather',
      icon: { component: markRaw(SourceIcon), props: { condition: headline?.condition } },
      value: weatherHeadlineValue(headline),
      origin: resource.location || instance.name,
      status,
      checked: formatLocalDateTime(resource.captured_at),
      colour: null,
      warning: !instance.enabled || !!instance.last_error || resource.freshness === 'stale' || !headline,
      error: !!instance.last_error,
      rows: metricRows(resource),
    },
  ];
}
/** Weather owns its reference type, summaries, icon and forecast results component. */
export const weatherPresenter: IntegrationPresenter = {
  setup: weatherSetup,
  contractVersion: 'weather_forecast/v1',
  integrationType: 'weather',
  help: weatherHelp,
  widget: weatherWidget,
  sourceKinds: { weather: { title: 'Weather', requiresService: false } },
  sources,
  icon: () => ({ component: markRaw(SourceIcon), props: {} }),
  results: markRaw(WeatherResults),
};

function weatherWidget(instance: ConnectionInstance) {
  const source = instance.resources.flatMap((resource) => sources(instance, resource))[0];
  return {
    remaining: 0,
    remainingLabel: 'more results',
    status: source?.status ?? 'Waiting for weather',
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

function weatherSetup(provider: ProviderDefinition) {
  const keys: Record<string, string> = {
    met_office: 'https://datahub.metoffice.gov.uk/docs/getting-started',
    weatherapi: 'https://www.weatherapi.com/signup.aspx',
  };
  const notices = [
    'Choose a location to send to the selected weather provider. Refreshes run manually or on the saved schedule.',
  ];
  const services = [];
  if (provider.setup.search.base_url)
    services.push({ label: 'Location search', url: serviceHomeURL(provider.setup.search.base_url) });
  if (provider.setup.search.postcode)
    services.push({ label: 'UK postcode search', url: serviceHomeURL(provider.setup.search.postcode.base_url) });
  if (provider.id === 'met_office') notices.push('UK locations only. Save the DataHub API key before searching.');
  if (provider.id === 'open_meteo') notices.push('Free hosted access is for non-commercial use.');
  return {
    title: 'Set up weather',
    sectionTitle: 'Your location',
    defaultName: 'Weather',
    searchLabel: provider.id === 'met_office' ? 'UK town or postcode' : 'Town or postcode',
    searchPlaceholder: provider.id === 'met_office' ? 'For example Sheffield or S8 9ED' : 'For example Sheffield',
    noResults:
      provider.id === 'met_office'
        ? 'No UK location was found. Met Office weather covers the United Kingdom; try a UK town or postcode.'
        : 'No location was found. Try a different town or postcode.',
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
