import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parseAirQualityForecast } from '@/contracts/integrations/air-quality';

const example = JSON.parse(readFileSync(resolve(__dirname, '../../fixtures/integrations/airquality.json'), 'utf8'));

test('accepts a bounded air-quality snapshot and preserves null values', () => {
  const parsed = parseAirQualityForecast(structuredClone(example));
  expect(parsed.pollen.olive).toBeNull();
  expect(parsed.hourly).toHaveLength(2);
});

test.each([
  'invalid family',
  'invalid category',
  'negative pollutant',
  'single coordinate',
  'unordered hours',
  'empty horizon',
])('rejects %s at ingress', (name) => {
  const value = structuredClone(example);
  if (name === 'invalid family') value.index_family = 'martian';
  if (name === 'invalid category') value.category = 'apocalyptic';
  if (name === 'negative pollutant') value.current.pm2_5 = -1;
  if (name === 'single coordinate') value.longitude = null;
  if (name === 'unordered hours') value.hourly[1].time = value.hourly[0].time;
  if (name === 'empty horizon') {
    value.current = null;
    value.hourly = [];
  }
  expect(() => parseAirQualityForecast(value)).toThrow();
});
