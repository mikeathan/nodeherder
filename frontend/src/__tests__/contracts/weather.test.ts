import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parseWeatherForecast } from '@/contracts/weather';
const example = JSON.parse(
  readFileSync(resolve(__dirname, '../fixtures/weather.json'), 'utf8')
);
test('accepts ordered hourly and daily forecasts while preserving missing values', () => {
  const f = structuredClone(example);
  f.current.temperature = null;
  expect(parseWeatherForecast(f).current?.temperature).toBeNull();
  expect(parseWeatherForecast(f).daily).toHaveLength(15);
});
test.each(['invalid timezone', 'duplicate hour', 'invalid probability', 'invalid day', 'invalid range'])(
  'rejects %s at ingress',
  (name) => {
    const f = structuredClone(example);
    if (name === 'invalid timezone') f.timezone = 'invalid';
    if (name === 'duplicate hour') f.hourly[1].time = f.hourly[0].time;
    if (name === 'invalid probability') f.current.precipitation_probability = 101;
    if (name === 'invalid day') f.daily[0].date = '2026-02-30';
    if (name === 'invalid range') f.daily[0].low = 30;
    expect(() => parseWeatherForecast(f)).toThrow();
  }
);
