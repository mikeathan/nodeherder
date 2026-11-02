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
import type { WeatherAlerts } from '@/types/integrations/weather-alerts';
import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations';
import { alertsHelp } from './help';
import {
  activeAlertCountLabel,
  alertLevelLabel,
  alertSeverityLabel,
} from '@/domain/integrations/weather-alerts';
import { serviceHomeURL } from '@/domain/integrations/service-links';
import { formatLocalDateTime } from '@/utils/date.utils';

function metricIcon(name: IconName): SourceIconModel {
  return { component: MetricIcon, props: { name } };
}

/** Compact rows lead with severity, affected area and the first expiry. */
function metricRows(alerts: WeatherAlerts): SourceRow[] {
  const first = alerts.warnings[0];
  if (!first) return [];
  return [
    {
      id: 'severity',
      title: 'Severity',
      value: `${alertSeverityLabel(first.severity)} · ${alertLevelLabel(first.level)}`,
      colour: null,
      icon: metricIcon('warn'),
    },
    {
      id: 'area',
      title: 'Area',
      value: first.areas[0] ?? 'All areas',
      colour: null,
      icon: metricIcon('pin'),
    },
    {
      id: 'until',
      title: 'Until',
      value: formatLocalDateTime(first.valid_to),
      colour: null,
      icon: metricIcon('clock'),
    },
  ];
}

function sources(instance: ConnectionInstance, resource: IntegrationResource): IntegrationSourcePresentation[] {
  if (resource.contract_version !== 'weather_alerts/v1') return [];
  const first = resource.warnings[0];
  let status = activeAlertCountLabel(resource.warnings.length);
  if (resource.freshness === 'stale') status = 'Needs refreshing';
  if (!instance.enabled) status = 'Disabled';
  if (instance.last_error) status = instance.last_error.message;
  return [
    {
      reference: { kind: 'weather_alerts', instance_id: instance.id, resource_id: resource.resource_id },
      title: 'Alerts',
      icon: { component: markRaw(SourceIcon), props: {} },
      value: first?.headline ?? 'No active warnings',
      detail: first
        ? `${alertSeverityLabel(first.severity)} · ${alertLevelLabel(first.level)} · ${first.type}`
        : 'Provider confirms no active warnings',
      origin: instance.name,
      status,
      checked: formatLocalDateTime(resource.last_checked_at),
      colour: null,
      warning: !instance.enabled || !!instance.last_error || resource.freshness === 'stale' || !!first,
      error: !!instance.last_error,
      rows: metricRows(resource),
    },
  ];
}

function alertsWidget(instance: ConnectionInstance) {
  const source = instance.resources.flatMap((resource) => sources(instance, resource))[0];
  return {
    remaining: 0,
    remainingLabel: 'more results',
    status: source?.status ?? 'Waiting for alerts',
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

function alertsSetup(provider: ProviderDefinition) {
  const keys: Record<string, string> = {
    met_office: 'https://datahub.metoffice.gov.uk/docs/getting-started',
    nws: 'https://www.weather.gov/documentation/services-web-api',
  };
  const nws = provider.id === 'nws';
  const metOffice = provider.id === 'met_office';
  const notices = [
    'Choose a location to send to the selected alert provider. Warnings run manually or on the saved schedule.',
  ];
  const services = [];
  if (provider.setup.search.base_url)
    services.push({ label: 'Location search', url: serviceHomeURL(provider.setup.search.base_url) });
  if (provider.setup.search.postcode)
    services.push({ label: 'UK postcode search', url: serviceHomeURL(provider.setup.search.postcode.base_url) });
  if (metOffice) notices.push('UK warnings only. Save the DataHub API key before searching.');
  if (nws) notices.push('US warnings only. Enter a US town or ZIP code; UK postcodes do not return results.');
  return {
    title: 'Set up weather alerts',
    sectionTitle: 'Your area',
    defaultName: 'Weather alerts',
    searchLabel: nws ? 'US town or ZIP code' : metOffice ? 'UK town or postcode' : 'Town or postcode',
    searchPlaceholder: nws
      ? 'For example New York or 10001'
      : metOffice
        ? 'For example Sheffield or S8 9ED'
        : undefined,
    noResults: nws
      ? 'No US location was found. National Weather Service covers the United States only; UK postcodes do not return results.'
      : metOffice
        ? 'No UK location was found. Try a UK town or postcode.'
        : undefined,
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

/** Alerts own their reference, summaries, icons and results component. */
export const alertsPresenter: IntegrationPresenter = {
  setup: alertsSetup,
  contractVersion: 'weather_alerts/v1',
  integrationType: 'alerts',
  help: alertsHelp,
  widget: alertsWidget,
  sourceKinds: { weather_alerts: { title: 'Weather alerts', requiresService: false } },
  sources,
  icon: () => ({ component: markRaw(SourceIcon), props: {} }),
  results: markRaw(Results),
};
