import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { connectionWidget, moveBoardEntry, reconcileBoard } from '@/domain/connection-board';
import type { ConnectionInstance, WasteSchedule } from '@/types/integrations.type';
const examples = JSON.parse(
  readFileSync(
    resolve(__dirname, '../../../../specs/032-external-integrations/contracts/waste_schedule.v1.examples.json'),
    'utf8'
  )
);
function instance(): Omit<ConnectionInstance, 'resources'> & { resources: WasteSchedule[] } {
  return {
    id: 'one',
    name: 'Bins',
    provider_key: 'waste_collection/whenisbins',
    revision: 1,
    etag: 'etag',
    enabled: true,
    status: 'ready',
    postcode: 'XX1 1XX',
    credential_slots: [],
    credential_configured: false,
    runs: [],
    resources: [structuredClone(examples.exact)],
    last_error: null,
    updated_at: '2026-10-23T12:00:00Z',
  };
}
test('reconciles missing/deleted/duplicate members without changing the accepted input', () => {
  const entries = [
    { instance_id: 'b', hidden: true },
    { instance_id: 'gone', hidden: false },
    { instance_id: 'b', hidden: false },
  ];
  expect(reconcileBoard(entries, ['a', 'b'])).toEqual([
    { instance_id: 'b', hidden: true },
    { instance_id: 'a', hidden: false },
  ]);
  expect(entries).toHaveLength(3);
});
test('movement uses visible order and preserves hidden placements', () => {
  const entries = [
    { instance_id: 'a', hidden: false },
    { instance_id: 'hidden', hidden: true },
    { instance_id: 'b', hidden: false },
  ];
  expect(moveBoardEntry(entries, 'a', 1).map((entry) => entry.instance_id)).toEqual(['b', 'hidden', 'a']);
  expect(moveBoardEntry(entries, 'a', -1)).toEqual(entries);
  expect(entries[0].instance_id).toBe('a');
});
test('summary uses UK projected dates and distinguishes empty, exhausted and unavailable', () => {
  const item = instance();
  const resource = item.resources[0];
  resource.evaluated_at = '2026-10-24T23:30:00Z';
  resource.services[0].dates = ['2026-10-24', '2026-10-25'];
  expect(connectionWidget(item).rows[0].date).toBe('Sun 25 Oct');
  resource.services[0].dates = ['2026-10-24'];
  expect(connectionWidget(item).rows[0].date).toBe('No upcoming date');
  resource.services[0].date_state = 'confirmed_empty';
  resource.services[0].dates = [];
  expect(connectionWidget(item).rows[0].date).toBe('None scheduled');
  resource.services[0].date_state = 'unavailable';
  expect(connectionWidget(item).rows[0].date).toBe('Dates unavailable');
});
test('errors, disabled and pending are distinct, and unknown types have generic fallback', () => {
  const item = instance();
  item.resources = [];
  expect(connectionWidget(item).status).toBe('Waiting for collection dates');
  item.operation_id = 'operation';
  expect(connectionWidget(item).status).toBe('Checking for updates');
  item.last_error = { code: 'offline', message: 'Provider unavailable' };
  expect(connectionWidget(item).status).toBe('Provider unavailable');
  item.enabled = false;
  expect(connectionWidget(item).status).toBe('Disabled');
  item.enabled = true;
  item.operation_id = undefined;
  item.last_error = null;
  item.provider_key = 'future/example';
  expect(connectionWidget(item)).toMatchObject({ status: 'Waiting for information', rows: [] });
});

test('widget orders all resources chronologically before limiting to three services', () => {
  const item = instance();
  const resource = item.resources[0];
  resource.evaluated_at = '2026-10-26T12:00:00Z';
  const original = resource.services[0];
  resource.services = [
    { ...original, service_id: 'blue', dates: ['2026-11-17'] },
    { ...original, service_id: 'undated', date_state: 'unavailable', dates: [] },
    { ...original, service_id: 'black', dates: ['2026-10-27'] },
  ];
  item.resources.push({
    ...structuredClone(resource),
    resource_id: 'other-resource',
    services: [
      { ...original, service_id: 'brown', dates: ['2026-11-03'] },
      { ...original, service_id: 'later', dates: ['2027-01-01'] },
    ],
  });
  const before = structuredClone(item);
  const summary = connectionWidget(item);
  expect(summary.rows.map((row) => row.id.split('/').pop())).toEqual(['black', 'brown', 'blue']);
  expect(summary.remaining).toBe(2);
  expect(item).toEqual(before);
});
