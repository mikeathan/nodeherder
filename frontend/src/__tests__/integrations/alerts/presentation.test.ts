import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { alertsPresenter } from '@/integrations/alerts/presentation';
import type { WeatherAlerts } from '@/types/integrations/weather-alerts';
import type { ConnectionInstance } from '@/types/integrations';

const example: WeatherAlerts = JSON.parse(
  readFileSync(resolve(__dirname, '../../fixtures/integrations/alerts.json'), 'utf8')
);

function instance(overrides: Partial<ConnectionInstance> = {}): ConnectionInstance {
  return {
    credential_slots: [],
    runs: [],
    id: 'instance',
    name: 'Weather alerts',
    provider_key: 'alerts/met_office',
    revision: 1,
    etag: 'etag',
    enabled: true,
    status: 'ready',
    postcode: '',
    credential_configured: false,
    resources: [],
    schedule: null,
    last_error: null,
    updated_at: '2026-10-29T15:41:00Z',
    ...overrides,
  };
}

test('the compact card leads with the warning severity, area and expiry', () => {
  const source = alertsPresenter.sources(instance(), example)[0];
  expect(source.value).toBe('Strong winds may cause disruption on Thursday');
  expect(source.detail).toBe('Severe · Amber · Wind');
  expect(source.status).toBe('1 active warning');
  expect(source.rows.map((row) => row.title)).toEqual(['Severity', 'Area', 'Until']);
  expect(source.rows[0].value).toBe('Severe · Amber');
  expect(source.rows[1].value).toBe('Yorkshire & Humber');
  expect(source.rows.every((row) => row.icon)).toBe(true);
});

test('the compact card reports a confirmed clear state and stale data', () => {
  const clear: WeatherAlerts = structuredClone(example);
  clear.status = 'none';
  clear.warnings = [];
  const source = alertsPresenter.sources(instance(), clear)[0];
  expect(source.value).toBe('No active warnings');
  expect(source.rows).toEqual([]);
  expect(source.warning).toBe(false);
  clear.freshness = 'stale';
  expect(alertsPresenter.sources(instance(), clear)[0].status).toBe('Needs refreshing');
});
