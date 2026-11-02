import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parseWeatherForecast } from '@/contracts/integrations/weather';
const example = JSON.parse(
  readFileSync(resolve(__dirname, '../../fixtures/integrations/weather.json'), 'utf8')
);
test('accepts ordered hourly and daily forecasts while preserving missing values', () => {
  const f = structuredClone(example);
  f.current.temperature = null;
  expect(parseWeatherForecast(f).current?.temperature).toBeNull();
  expect(parseWeatherForecast(f).daily).toHaveLength(15);
});
test('accepts additive daily signals and older snapshots without them', () => {
  const older = structuredClone(example);
  expect(parseWeatherForecast(older).current?.uv).toBeUndefined();
  const f = structuredClone(example);
  f.current.uv = 5;
  f.daily[0].uv_max = 4;
  f.daily[0].sunrise = '07:12';
  f.daily[0].sunset = '17:48';
  f.daily[0].feels_like_high = 14;
  f.daily[0].feels_like_low = 7;
  f.daily[0].wind_gust = 29;
  f.daily[0].wind_direction = 210;
  f.daily[0].humidity = 75;
  f.daily[0].pressure = 1013;
  f.daily[0].visibility = 20000;
  f.daily[0].rain_probability = 15;
  f.daily[0].snow_probability = 0;
  f.daily[0].night_condition = 'clear';
  const parsed = parseWeatherForecast(f);
  expect(parsed.current?.uv).toBe(5);
  expect(parsed.daily[0].uv_max).toBe(4);
  expect(parsed.daily[0].sunrise).toBe('07:12');
  expect(parsed.daily[0].pressure).toBe(1013);
  expect(parsed.daily[0].night_condition).toBe('clear');
});
test.each([
  'invalid timezone',
  'duplicate hour',
  'invalid probability',
  'invalid day',
  'invalid range',
  'invalid uv',
  'invalid sunrise',
  'invalid sunset order',
  'invalid feels like',
  'invalid humidity',
  'invalid pressure',
  'invalid visibility',
  'invalid snow probability',
  'invalid night condition',
])('rejects %s at ingress', (name) => {
  const f = structuredClone(example);
  if (name === 'invalid timezone') f.timezone = 'invalid';
  if (name === 'duplicate hour') f.hourly[1].time = f.hourly[0].time;
  if (name === 'invalid probability') f.current.precipitation_probability = 101;
  if (name === 'invalid day') f.daily[0].date = '2026-02-30';
  if (name === 'invalid range') f.daily[0].low = 30;
  if (name === 'invalid uv') f.current.uv = 21;
  if (name === 'invalid sunrise') f.daily[0].sunrise = '7:05';
  if (name === 'invalid sunset order') {
    f.daily[0].sunrise = '19:00';
    f.daily[0].sunset = '06:00';
  }
  if (name === 'invalid feels like') f.daily[0].feels_like_high = 101;
  if (name === 'invalid humidity') f.daily[0].humidity = 101;
  if (name === 'invalid pressure') f.daily[0].pressure = 1200;
  if (name === 'invalid visibility') f.daily[0].visibility = -1;
  if (name === 'invalid snow probability') f.daily[0].snow_probability = 101;
  if (name === 'invalid night condition') f.daily[0].night_condition = 'stormy';
  expect(() => parseWeatherForecast(f)).toThrow();
});
