import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { homeIntegrationCard, homeIntegrationSources } from '@/domain/home-cards';
import { homeSourceKey } from '@/domain/home-layout';
import { homePickerChoices } from '@/domain/home-picker';
import type { ConnectionInstance, WasteSchedule } from '@/types/integrations.type';
import type { HomeCard } from '@/types/home.type';

const example = JSON.parse(
  readFileSync(
    resolve(__dirname, '../../../../specs/032-external-integrations/contracts/waste_schedule.v1.examples.json'),
    'utf8'
  )
).exact;
const instance = (): Omit<ConnectionInstance, 'resources'> & { resources: WasteSchedule[] } => ({
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

test('weather sources keep resource identity and explicit unavailable placements', () => {
  const weather = JSON.parse(
    readFileSync(resolve(__dirname, '../fixtures/weather.json'), 'utf8')
  );
  const bins = instance();
  const item: ConnectionInstance = { ...bins, provider_key: 'weather/open_meteo', resources: [weather] };
  const sources = homeIntegrationSources([item]);
  expect(sources.size).toBe(1);
  const entry = [...sources.values()][0];
  expect(entry.reference.kind).toBe('weather');
  expect(entry.value).toContain('14°');
  expect(entry.icon?.props.condition).toBe(weather.current.condition);
  const card: HomeCard = { ...entry.reference, id: 'weather-card', panel: true };
  expect(homeIntegrationCard(card, new Map()).value).toBe('Unavailable');
  item.last_error = { code: 'credential_required', message: 'This provider requires a credential.' };
  expect([...homeIntegrationSources([item]).values()][0]).toMatchObject({
    error: true,
    status: 'This provider requires a credential.',
  });
});


test('bin picker offers the full schedule while existing service cards remain available', () => {
  const sources = homeIntegrationSources([instance()]);
  const choices = homePickerChoices('integrations', [], sources, []);
  expect(choices.map((choice) => choice.reference.kind)).toEqual(['waste_summary']);
  const service = [...sources.values()].find((source) => source.reference.kind === 'waste_service')!;
  const card: HomeCard = { ...service.reference, id: 'existing', panel: true };
  expect(homeIntegrationCard(card, sources).title).toBe(service.title);
});
