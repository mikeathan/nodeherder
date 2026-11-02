import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parseWeatherAlerts } from '@/contracts/integrations/weather-alerts';

const example = JSON.parse(readFileSync(resolve(__dirname, '../../fixtures/integrations/alerts.json'), 'utf8'));

test('accepts a bounded alert snapshot and preserves the provider geometry', () => {
  const parsed = parseWeatherAlerts(structuredClone(example));
  expect(parsed.warnings).toHaveLength(1);
  expect(parsed.warnings[0].severity).toBe('severe');
  expect(parsed.warnings[0].polygon?.type).toBe('MultiPolygon');
});

test.each([
  'status mismatch',
  'duplicate id',
  'missing headline',
  'invalid coordinate',
  'inverted window',
])('rejects %s at ingress', (name) => {
  const value = structuredClone(example);
  if (name === 'status mismatch') value.status = 'none';
  if (name === 'duplicate id') value.warnings.push(structuredClone(value.warnings[0]));
  if (name === 'missing headline') value.warnings[0].headline = '';
  if (name === 'invalid coordinate') value.warnings[0].polygon.coordinates[0][0][0][0] = 200;
  if (name === 'inverted window') value.warnings[0].valid_to = value.warnings[0].valid_from;
  expect(() => parseWeatherAlerts(value)).toThrow();
});
