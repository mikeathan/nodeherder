import {
  catalogStatusLabel,
  catalogStatusTone,
  integrationTypeIcon,
  integrationTypeProviders,
  integrationTypeRank,
  integrationTypeTitle,
  matchesCatalogQuery,
  providerCredentialLabel,
  providerRegionLabel,
  providerRefreshLabel,
} from '@/domain/integrations/catalog';
import type { ConnectionCatalog } from '@/types/integrations';

const catalog: ConnectionCatalog = {
  available: true,
  availability_error: null,
  types: [
    { id: 'weather', name: 'Weather', description: 'Local forecasts', configured_count: 2, worst_status: 'error' },
    { id: 'fuel_prices', name: 'Fuel prices', description: 'Nearby forecourts' },
    { id: 'alerts', name: 'Severe weather alerts', description: 'Warnings' },
  ],
  providers: [
    {
      key: 'weather/open_meteo',
      id: 'open_meteo',
      name: 'Open-Meteo',
      integration_type: 'weather',
      origin: 'https://api.open-meteo.com',
      available: true,
      requires_credential: false,
      refresh_default_seconds: 1800,
      refresh_min_seconds: 900,
    },
  ],
};

test('resolve type metadata and hide discovery types without providers', () => {
  expect(integrationTypeIcon('weather')).toBe('weatherSun');
  expect(integrationTypeIcon('unknown')).toBe('connections');
  expect(integrationTypeTitle('air_quality')).toBe('Air quality');
  expect(integrationTypeTitle('custom', catalog)).toBe('Integration');
  expect(integrationTypeRank('weather')).toBeLessThan(integrationTypeRank('waste_collection'));
  expect(integrationTypeProviders(catalog, 'weather')).toHaveLength(1);
  expect(integrationTypeProviders(catalog, 'alerts')).toEqual([]);
});

test('present configured status and provider policy text', () => {
  expect(catalogStatusLabel('error')).toBe('Needs attention');
  expect(catalogStatusTone('pending')).toBe('pending');
  expect(providerCredentialLabel(catalog.providers[0])).toBe('No credential');
  expect(providerRefreshLabel(catalog.providers[0])).toBe('Default every 30 minutes · min every 15 minutes');
  expect(providerRegionLabel(catalog.providers[0])).toBe('');
  expect(providerRegionLabel({ ...catalog.providers[0], key: 'alerts/nws' })).toBe('US only');
  expect(matchesCatalogQuery('open', 'Weather', 'Open-Meteo')).toBe(true);
  expect(matchesCatalogQuery('missing', 'Weather', 'Open-Meteo')).toBe(false);
});
