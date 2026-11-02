import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { airQualityPresenter } from '@/integrations/air-quality/presentation';
import type { AirQualityForecast } from '@/types/integrations/air-quality';
import type { ConnectionInstance } from '@/types/integrations';

const example: AirQualityForecast = JSON.parse(
  readFileSync(resolve(__dirname, '../../fixtures/integrations/airquality.json'), 'utf8')
);

function instance(overrides: Partial<ConnectionInstance> = {}): ConnectionInstance {
  return {
    credential_slots: [],
    runs: [],
    id: 'instance',
    name: 'Air quality',
    provider_key: 'air_quality/open_meteo',
    revision: 1,
    etag: 'etag',
    enabled: true,
    status: 'ready',
    postcode: '',
    credential_configured: false,
    resources: [],
    schedule: null,
    last_error: null,
    updated_at: '2026-11-03T12:00:00Z',
    ...overrides,
  };
}

test('the compact card exposes category, index and icon-led pollutant rows', () => {
  const source = airQualityPresenter.sources(instance(), example)[0];
  expect(source.value).toBe('Fair · European AQI 25');
  expect(source.origin).toBe('Sheffield');
  expect(source.rows.map((row) => row.title)).toEqual(['PM2.5', 'PM10', 'Pollen']);
  expect(source.rows[0].value).toBe('12 µg/m³');
  expect(source.rows[1].value).toBe('20 µg/m³');
  expect(source.rows[2].value).toBe('Grass 3 grains/m³');
  expect(source.rows.every((row) => row.icon)).toBe(true);
});

test('the compact card reports stale data and omits missing values', () => {
  const stale: AirQualityForecast = structuredClone(example);
  stale.freshness = 'stale';
  stale.current = null;
  stale.hourly[0].pm2_5 = null;
  stale.hourly[0].pm10 = null;
  stale.pollen = {
    alder: null,
    birch: null,
    grass: null,
    mugwort: null,
    olive: null,
    ragweed: null,
  };
  const source = airQualityPresenter.sources(instance(), stale)[0];
  expect(source.status).toBe('Needs refreshing');
  expect(source.warning).toBe(true);
  expect(source.rows).toEqual([]);
});
