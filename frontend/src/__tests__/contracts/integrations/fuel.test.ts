import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parseFuelPrices } from '@/contracts/integrations/fuel';

const example = JSON.parse(readFileSync(resolve(__dirname, '../../fixtures/integrations/fuel.json'), 'utf8'));

test('accepts a bounded fuel snapshot with integer pence prices', () => {
  const parsed = parseFuelPrices(structuredClone(example));
  expect(parsed.stations).toHaveLength(1);
  expect(parsed.stations[0].prices[0].pence_per_litre).toBe(143);
  expect(parsed.origin.label).toBe('Nether Edge');
});

test.each(['currency', 'fractional price', 'zero price', 'duplicate station', 'single coordinate'])(
  'rejects %s at ingress',
  (name) => {
    const value = structuredClone(example);
    if (name === 'currency') value.currency = 'EUR';
    if (name === 'fractional price') value.stations[0].prices[0].pence_per_litre = 142.9;
    if (name === 'zero price') value.stations[0].prices[0].pence_per_litre = 0;
    if (name === 'duplicate station') value.stations.push(structuredClone(value.stations[0]));
    if (name === 'single coordinate') value.stations[0].longitude = null;
    expect(() => parseFuelPrices(value)).toThrow();
  }
);
