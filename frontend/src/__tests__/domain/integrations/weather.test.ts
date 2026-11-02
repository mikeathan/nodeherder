import { hoursForDate, weatherHourlyHorizonHours, WEATHER_DAY_HOUR_LIMIT } from '@/domain/integrations/weather';
import type { WeatherDay, WeatherForecast, WeatherPoint } from '@/types/integrations/weather';

function point(time: string): WeatherPoint {
  return {
    time,
    condition: 'clear',
    temperature: 14,
    feels_like: 12,
    wind: 10,
    gust: 15,
    direction: 180,
    humidity: 70,
    precipitation_probability: 10,
    precipitation: 0,
    uv: null,
  };
}

function day(overrides: Partial<WeatherDay> = {}): WeatherDay {
  return {
    date: '2026-10-28',
    condition: 'clear',
    night_condition: null,
    high: 16,
    low: 9,
    feels_like_high: null,
    feels_like_low: null,
    wind: 10,
    wind_gust: null,
    wind_direction: null,
    humidity: null,
    pressure: null,
    visibility: null,
    precipitation_probability: 10,
    precipitation: 0,
    rain_probability: null,
    snow_probability: null,
    sunrise: null,
    sunset: null,
    uv_max: null,
    ...overrides,
  };
}

function forecast(hourly: WeatherPoint[]): WeatherForecast {
  return {
    contract_version: 'weather_forecast/v1',
    instance_id: 'instance',
    resource_id: 'resource',
    revision: 1,
    location: 'Bristol',
    timezone: 'Europe/London',
    attribution: 'Test',
    captured_at: '2026-10-28T00:00:00Z',
    last_checked_at: '2026-10-28T00:00:00Z',
    evaluated_at: '2026-10-28T00:00:00Z',
    freshness: 'fresh',
    current: null,
    hourly,
    daily: [day()],
  };
}

test('hoursForDate keeps only the matching local calendar date', () => {
  const result = hoursForDate(
    forecast([point('2026-10-28T10:00:00Z'), point('2026-10-29T10:00:00Z')]),
    '2026-10-28'
  );
  expect(result.map((hour) => hour.time)).toEqual(['2026-10-28T10:00:00Z']);
});

test('hoursForDate bounds a long series', () => {
  const hourly = Array.from({ length: WEATHER_DAY_HOUR_LIMIT + 6 }, (_, index) =>
    point(new Date(Date.UTC(2026, 9, 28, 0, index * 30)).toISOString())
  );
  expect(hoursForDate(forecast(hourly), '2026-10-28')).toHaveLength(WEATHER_DAY_HOUR_LIMIT);
});

test('weatherHourlyHorizonHours reports remaining future coverage', () => {
  expect(
    weatherHourlyHorizonHours(
      forecast([point('2026-10-28T10:00:00Z'), point('2026-10-29T10:00:00Z')])
    )
  ).toBe(34);
  expect(weatherHourlyHorizonHours(forecast([]))).toBeNull();
});
