import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parseWasteSchedule, parseSetup } from '@/contracts/integrations';
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
