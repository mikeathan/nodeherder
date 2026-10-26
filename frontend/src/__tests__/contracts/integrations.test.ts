import { WASTE_SCHEDULE_TIMEZONE } from '@/config/integrations';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parseWasteSchedule, parseSetup, parseBoard } from '@/contracts/integrations';
const examples = JSON.parse(
  readFileSync(
    resolve(__dirname, '../../../../specs/032-external-integrations/contracts/waste_schedule.v1.examples.json'),
    'utf8'
  )
);
test.each(['exact', 'partial', 'confirmed_empty'])('accepts canonical %s fixture', (key) => {
  expect(parseWasteSchedule(examples[key]).contract_version).toBe('waste_schedule/v1');
});
test.each(['provisional', 'unavailable'])('rejects %s as durable resource ingress', (key) => {
  expect(() => parseWasteSchedule(examples[key])).toThrow();
});
test('rejects invalid calendar dates and unsupported contracts', () => {
  const value = structuredClone(examples.exact);
  value.services[0].dates = ['2026-02-30'];
  expect(() => parseWasteSchedule(value)).toThrow();
  value.contract_version = 'waste_schedule/v9';
  expect(() => parseWasteSchedule(value)).toThrow();
});
test('rejects missing setup identity and secret-like unknown shapes', () => {
  expect(() => parseSetup({ status: 'pending', choices: [] })).toThrow();
});

test('board ingress rejects ambiguous membership and visibility', () => {
  expect(parseBoard({ etag: 'revision', entries: [] })).toEqual({ etag: 'revision', entries: [] });
  for (const entries of [
    [{ instance_id: 'a' }],
    [{ instance_id: 'a', hidden: null }],
    [
      { instance_id: 'a', hidden: false },
      { instance_id: 'a', hidden: true },
    ],
    [{ instance_id: '', hidden: false }],
  ])
    expect(() => parseBoard({ etag: 'revision', entries })).toThrow();
});

test('waste source timezone stays explicit and cannot become an arbitrary app timezone', () => {
  const schedule = structuredClone(examples.exact);
  expect(parseWasteSchedule(schedule).timezone).toBe(WASTE_SCHEDULE_TIMEZONE);
  schedule.timezone = 'America/New_York';
  expect(() => parseWasteSchedule(schedule)).toThrow('Unsupported Integrations value');
});
