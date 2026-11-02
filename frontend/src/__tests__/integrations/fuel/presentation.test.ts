import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { fuelPresenter } from '@/integrations/fuel/presentation';
import type { FuelPrices } from '@/types/integrations/fuel';
import type { ConnectionInstance } from '@/types/integrations';

const example: FuelPrices = JSON.parse(
  readFileSync(resolve(__dirname, '../../fixtures/integrations/fuel.json'), 'utf8')
);

function instance(overrides: Partial<ConnectionInstance> = {}): ConnectionInstance {
  return {
    credential_slots: [],
    runs: [],
    id: 'instance',
    name: 'Fuel prices',
    provider_key: 'fuel_prices/fuel_finder',
    revision: 1,
    etag: 'etag',
    enabled: true,
    status: 'ready',
    postcode: '',
    credential_configured: false,
    resources: [],
    schedule: null,
    last_error: null,
    updated_at: '2026-10-29T14:02:00Z',
    ...overrides,
  };
}

test('the compact card leads with the nearest station, cheapest grade and distance', () => {
  const source = fuelPresenter.sources(instance(), example)[0];
  expect(source.value).toBe('Shell Abbeydale');
  expect(source.detail).toContain('E10 143p/l');
  expect(source.origin).toBe('Search area: Nether Edge');
  expect(source.status).toBe('1 station');
  expect(source.rows.map((row) => row.title)).toEqual(['Cheapest', 'Distance', 'Station']);
  expect(source.rows[0].value).toBe('E10 143p/l');
  expect(source.rows[2].value).toBe('S7 2QF');
  expect(source.rows.every((row) => row.icon)).toBe(true);
});

test('the compact card reports empty and stale states', () => {
  const empty: FuelPrices = structuredClone(example);
  empty.stations = [];
  const emptySource = fuelPresenter.sources(instance(), empty)[0];
  expect(emptySource.value).toBe('No nearby stations');
  expect(emptySource.rows).toEqual([]);
  empty.freshness = 'stale';
  expect(fuelPresenter.sources(instance(), empty)[0].status).toBe('Needs refreshing');
});
