import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { homeIntegrationCard, homeIntegrationSources } from '@/domain/home-cards';
import { homeSourceKey } from '@/domain/home-layout';
import type { ConnectionInstance } from '@/types/integrations.type';
import type { HomeCard } from '@/types/home.type';

const example = JSON.parse(
  readFileSync(
    resolve(__dirname, '../../../../specs/032-external-integrations/contracts/waste_schedule.v1.examples.json'),
    'utf8'
  )
).exact;
const instance = (): ConnectionInstance => ({
  id: 'one',
  name: 'Bins',
  provider_key: 'waste_collection/whenisbins',
  revision: 1,
  etag: 'one',
  enabled: true,
  status: 'ready',
  postcode: 'XX1 1XX',
  credential_slots: [],
  credential_configured: false,
  runs: [],
  resources: [structuredClone(example)],
  last_error: null,
  updated_at: '2026-10-26T08:00:00Z',
});

test('all services and summary share accepted projection with stable source identity', () => {
  const item = instance();
  const before = structuredClone(item);
  const catalog = homeIntegrationSources([item]);
  expect(catalog.size).toBe(item.resources[0].services.length + 1);
  const entry = [...catalog.values()].find((item) => item.reference.kind === 'waste_service')!;
  const key = homeSourceKey(entry.reference);
  item.resources[0].services.reverse();
  expect(homeIntegrationSources([item]).get(key)?.title).toBe(entry.title);
  expect(before.resources[0].services).toHaveLength(item.resources[0].services.length);
});

test('missing/empty/unavailable states cannot silently become another service or guessed date', () => {
  const item = instance();
  const service = item.resources[0].services[0];
  service.date_state = 'unavailable';
  service.dates = [];
  const reference: HomeCard = {
    id: 'placement',
    kind: 'waste_service',
    panel: true,
    instance_id: item.id,
    resource_id: item.resources[0].resource_id,
    service_id: service.service_id,
  };
  expect(homeIntegrationCard(reference, homeIntegrationSources([item])).value).toBe('Dates unavailable');
  service.date_state = 'confirmed_empty';
  expect(homeIntegrationCard(reference, homeIntegrationSources([item])).value).toBe('None scheduled');
  const missing = { ...reference, service_id: 'removed' };
  expect(homeIntegrationCard(missing, homeIntegrationSources([item]))).toMatchObject({
    value: 'Unavailable',
    warning: true,
  });
  item.enabled = false;
  expect(homeIntegrationCard(reference, homeIntegrationSources([item])).status).toBe('Disabled');
});
