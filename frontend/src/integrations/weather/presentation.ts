import { markRaw, defineAsyncComponent } from 'vue';
const WeatherResults = defineAsyncComponent(() => import('./Results.vue'));
const SourceIcon = defineAsyncComponent(() => import('./SourceIcon.vue'));
import type { IntegrationPresenter, IntegrationResource, IntegrationSourcePresentation } from '../types';
import type { WeatherForecast } from '@/types/weather.type';
import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations.type';
import { weatherValue, weatherLabel } from '@/domain/weather';
import { calendarDateInZone, formatLocalDateTime } from '@/utils/date.utils';
function sources(instance: ConnectionInstance, resource: IntegrationResource): IntegrationSourcePresentation[] {
  if (resource.contract_version !== 'weather_forecast/v1') return [];
  let status = 'Forecast';
  if (resource.freshness === 'stale') status = 'Needs refreshing';
  if (!instance.enabled) status = 'Disabled';
  if (instance.last_error) status = instance.last_error.message;
  return [
    {
      reference: { kind: 'weather', instance_id: instance.id, resource_id: resource.resource_id },
      title: 'Weather',
      icon: { component: markRaw(SourceIcon), props: { condition: resource.current?.condition } },
      detail: forecastSummary(resource),
      value: weatherValue(resource.current?.temperature, '°') + ' · ' + weatherLabel(resource.current?.condition),
      origin: instance.name + ' · ' + resource.location,
      status,
      checked: formatLocalDateTime(resource.captured_at),
      colour: null,
      warning: !instance.enabled || !!instance.last_error || resource.freshness === 'stale',
      error: !!instance.last_error,
      rows: [],
    },
  ];
}
/** Weather owns its reference type, summaries, icon and forecast results component. */
export const weatherPresenter: IntegrationPresenter = {
  setup: weatherSetup,
  contractVersion: 'weather_forecast/v1',
  integrationType: 'weather',
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
  const notices = ['Choose a location to send to the selected weather provider. Refreshes are manual.'];
  if (provider.setup.search.base_url) notices.push(`Place search uses ${provider.setup.search.base_url}.`);
  if (provider.setup.search.postcode) notices.push(`UK postcode search uses ${provider.setup.search.postcode.base_url}.`);
  if (provider.id === 'met_office') notices.push('UK locations only.');
  if (provider.id === 'open_meteo') notices.push('Free hosted access is for non-commercial use.');
  return {
    title: 'Set up weather',
    sectionTitle: 'Your location',
    defaultName: 'Weather',
    searchLabel: 'Town or postcode',
    searchButton: 'Find location',
    choiceLabel: 'Choose your location',
    choicePlaceholder: 'Select a location',
    searchMaxLength: 120,
    autocomplete: 'off',
    notice: notices.join(' '),
    keyURL: keys[provider.id],
  };
}

function forecastSummary(forecast: WeatherForecast): string {
  const today = calendarDateInZone(forecast.evaluated_at, forecast.timezone);
  const day = forecast.daily.find((value) => value.date === today);
  const details: string[] = [];
  if (day?.high != null) details.push(`H ${weatherValue(day.high, '°')}`);
  if (day?.low != null) details.push(`L ${weatherValue(day.low, '°')}`);
  const now = Date.parse(forecast.evaluated_at);
  const upcoming = forecast.hourly.find(
    (point) => Date.parse(point.time) > now && Date.parse(point.time) <= now + 3600000
  );
  if (upcoming?.precipitation_probability != null)
    details.push(`Rain chance ${weatherValue(upcoming.precipitation_probability, '%')}`);
  return details.join(' · ');
}
