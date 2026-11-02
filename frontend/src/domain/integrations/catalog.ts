import type { IconName } from '@/components/ui/icons';
import type { ConnectionCatalog, ConnectionCatalogProvider } from '@/types/integrations';
import { formatScheduleInterval } from '@/domain/schedule';

const TYPE_ICONS: Record<string, IconName> = {
  waste_collection: 'schedule',
  weather: 'weatherSun',
  air_quality: 'weatherAir',
  alerts: 'warn',
  fuel_prices: 'fuel',
};

const TYPE_TITLES: Record<string, string> = {
  waste_collection: 'Bin collections',
  weather: 'Weather',
  air_quality: 'Air quality',
  alerts: 'Severe weather alerts',
  fuel_prices: 'Fuel prices',
};

/** Decorative icon owned by the integration type, with a safe generic fallback. */
export function integrationTypeIcon(typeId: string): IconName {
  return TYPE_ICONS[typeId] ?? 'connections';
}

const TYPE_ORDER = ['weather', 'air_quality', 'alerts', 'fuel_prices', 'waste_collection'];

/** Single catalog metadata registry for provider coverage; generic components stay branch-free. */
const PROVIDER_REGIONS: Record<string, string> = {
  'alerts/met_office': 'UK only',
  'alerts/nws': 'US only',
  'fuel_prices/fuel_finder': 'UK only',
  'weather/met_office': 'UK only',
};

/** Coverage label for provider selection; empty means no regional limit is declared. */
export function providerRegionLabel(provider: ConnectionCatalogProvider): string {
  return PROVIDER_REGIONS[provider.key] ?? '';
}

/** Stable reading order for configured entry groups. */
export function integrationTypeRank(typeId: string): number {
  const rank = TYPE_ORDER.indexOf(typeId);
  return rank === -1 ? TYPE_ORDER.length : rank;
}

/** Human label for a configured instance whose provider may no longer be installed. */
export function integrationTypeTitle(typeId: string, catalog?: ConnectionCatalog | null): string {
  return catalog?.types.find((type) => type.id === typeId)?.name ?? TYPE_TITLES[typeId] ?? 'Integration';
}

/** Providers belonging to one type, preserving catalog order. */
export function integrationTypeProviders(
  catalog: ConnectionCatalog | null,
  typeId: string
): ConnectionCatalogProvider[] {
  return catalog?.providers.filter((provider) => provider.integration_type === typeId) ?? [];
}

/** Readable configured-state summary for a type card. */
export function catalogStatusLabel(status: string | undefined): string {
  switch (status) {
    case 'ready':
      return 'Ready';
    case 'pending':
      return 'Waiting for first result';
    case 'disabled':
      return 'Disabled';
    case 'error':
      return 'Needs attention';
    default:
      return 'Not configured';
  }
}

/** Stable tone used by the type card status chip. */
export function catalogStatusTone(status: string | undefined): 'ready' | 'pending' | 'disabled' | 'error' | 'none' {
  switch (status) {
    case 'ready':
    case 'pending':
    case 'disabled':
    case 'error':
      return status;
    default:
      return 'none';
  }
}

/** Public credential requirement, never the credential value. */
export function providerCredentialLabel(provider: ConnectionCatalogProvider): string {
  return provider.requires_credential ? 'Credential required' : 'No credential';
}

/** Public refresh policy summary. */
export function providerRefreshLabel(provider: ConnectionCatalogProvider): string {
  const seconds = provider.refresh_default_seconds ?? 0;
  if (seconds <= 0) return 'Manual refresh';
  const minimum = provider.refresh_min_seconds ?? 0;
  const base = `Default ${formatScheduleInterval(seconds)}`;
  return minimum > 0 ? `${base} · min ${formatScheduleInterval(minimum)}` : base;
}

/** Case-insensitive discovery filter for type/provider names and descriptions. */
export function matchesCatalogQuery(query: string, ...values: string[]): boolean {
  const needle = query.trim().toLowerCase();
  return !needle || values.some((value) => value.toLowerCase().includes(needle));
}
