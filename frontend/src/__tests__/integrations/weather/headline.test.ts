import { weatherHeadline, weatherHeadlineValue } from '@/integrations/weather/headline';
import type { WeatherDay, WeatherForecast, WeatherPoint } from '@/types/weather.type';

function point(time: string, temperature: number | null = 14, condition = 'partly_cloudy'): WeatherPoint {
  return {
    time,
    condition,
    temperature,
    feels_like: 12,
    wind: 18,
    gust: 25,
    direction: 210,
    humidity: 72,
    precipitation_probability: 20,
    precipitation: 0,
  };
}

const day = (date: string, high: number | null = 16): WeatherDay => ({
  date,
  condition: 'partly_cloudy',
  high,
  low: 9,
  wind: 18,
  precipitation_probability: 20,
  precipitation: 0,
});

function forecast(overrides: Partial<WeatherForecast> = {}): WeatherForecast {
  return {
    contract_version: 'weather_forecast/v1',
    instance_id: 'instance',
    resource_id: 'resource',
    revision: 1,
    location: 'Bristol',
    timezone: 'UTC',
    attribution: 'Test',
    captured_at: '2026-10-28T12:00:00Z',
    last_checked_at: '2026-10-28T12:00:00Z',
    evaluated_at: '2026-10-28T12:00:00Z',
    freshness: 'fresh',
    current: point('2026-10-28T12:00:00Z'),
    hourly: [point('2026-10-28T12:00:00Z'), point('2026-10-28T13:00:00Z')],
    daily: [day('2026-10-28')],
    ...overrides,
  };
}

test('prefers a projected current observation', () => {
  const headline = weatherHeadline(forecast());
  expect(headline).toMatchObject({ source: 'observation', label: 'Now', temperature: 14 });
  expect(weatherHeadlineValue(headline)).toBe('14° · Partly cloudy');
});

test('falls back to the latest elapsed hour when current has expired', () => {
  const headline = weatherHeadline(forecast({ current: null, evaluated_at: '2026-10-28T12:30:00Z' }));
  expect(headline).toMatchObject({ source: 'hourly', at: '2026-10-28T12:00:00Z', label: 'Observed 12:00' });
  expect(weatherHeadlineValue(headline)).toBe('14° · Partly cloudy · Observed 12:00');
});

test('uses the next hour when no hour has elapsed', () => {
  const headline = weatherHeadline(
    forecast({ current: null, hourly: [point('2026-10-28T13:00:00Z')], evaluated_at: '2026-10-28T12:00:00Z' })
  );
  expect(headline).toMatchObject({ source: 'hourly', label: 'Next 13:00' });
  expect(weatherHeadlineValue(headline)).toBe('14° · Partly cloudy · Next 13:00');
});

test('falls back to the daily summary only when no current or hourly data exists', () => {
  const headline = weatherHeadline(forecast({ current: null, hourly: [] }));
  expect(headline).toMatchObject({ source: 'daily', label: 'Today', temperature: 16 });
  expect(weatherHeadlineValue(headline)).toBe('16° · Partly cloudy · Today');
});

test('labels a non-today daily fallback with its date', () => {
  const headline = weatherHeadline(
    forecast({ current: null, hourly: [], daily: [day('2026-10-29')], evaluated_at: '2026-10-28T12:00:00Z' })
  );
  expect(headline?.label).toBe('2026-10-29');
});

test('returns no headline when all weather data is empty', () => {
  const empty = forecast({ current: null, hourly: [], daily: [] });
  expect(weatherHeadline(empty)).toBeNull();
  expect(weatherHeadlineValue(null)).toBe('No saved forecast');
});

test('never emits Unavailable while a forecast point exists', () => {
  const cases = [
    forecast({ current: null, evaluated_at: '2026-10-28T12:30:00Z' }),
    forecast({ current: null, hourly: [point('2026-10-28T13:00:00Z')] }),
    forecast({ current: null, hourly: [] }),
  ];
  for (const value of cases) expect(weatherHeadlineValue(weatherHeadline(value))).not.toContain('Unavailable');
});

test('keeps a condition when temperature is missing', () => {
  const headline = weatherHeadline(forecast({ current: point('2026-10-28T12:00:00Z', null, 'rain') }));
  expect(weatherHeadlineValue(headline)).toBe('Rain');
});
