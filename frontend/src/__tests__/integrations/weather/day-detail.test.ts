import { weatherDayDetail, weatherHourDetail } from '@/integrations/weather/day-detail';
import type { WeatherDay, WeatherForecast, WeatherPoint } from '@/types/integrations/weather';

function point(time: string, uv: number | null = null): WeatherPoint {
  return {
    time,
    condition: 'partly_cloudy',
    temperature: 14,
    feels_like: 12,
    wind: 18,
    gust: 25,
    direction: 210,
    humidity: 72,
    precipitation_probability: 20,
    precipitation: 0,
    uv,
  };
}

function day(overrides: Partial<WeatherDay> = {}): WeatherDay {
  return {
    date: '2026-10-28',
    condition: 'partly_cloudy',
    night_condition: 'clear',
    high: 16,
    low: 9,
    feels_like_high: 14,
    feels_like_low: 7,
    wind: 18,
    wind_gust: 29,
    wind_direction: 210,
    humidity: 75,
    pressure: 1013,
    visibility: 20000,
    precipitation_probability: 20,
    precipitation: 0,
    rain_probability: 15,
    snow_probability: 0,
    sunrise: '07:12',
    sunset: '17:48',
    uv_max: 4,
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
    timezone: 'UTC',
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

test('daily detail includes supplied stats and formatted hours', () => {
  const detail = weatherDayDetail(
    forecast([point('2026-10-28T10:00:00Z', 3), point('2026-10-29T10:00:00Z')]),
    day()
  );
  expect(detail.primary).toBe('16° / 9°');
  expect(detail.condition).toBe('partly_cloudy');
  expect(detail.conditionLabel).toBe('Partly cloudy');
  expect(detail.stats.map((stat) => stat.label)).toEqual([
    'Feels like high',
    'Feels like low',
    'Wind',
    'Gusts',
    'Wind direction',
    'Humidity',
    'Pressure',
    'Visibility',
    'Precipitation chance',
    'Rain chance',
    'Snow chance',
    'Precipitation',
    'Max UV',
    'Sunrise',
    'Sunset',
    'Night',
  ]);
  expect(detail.stats.find((stat) => stat.id === 'uv')?.value).toBe('4');
  expect(detail.stats.find((stat) => stat.id === 'pressure')?.value).toBe('1013 hPa');
  expect(detail.stats.find((stat) => stat.id === 'visibility')?.value).toBe('20 km');
  expect(detail.stats.find((stat) => stat.id === 'wind-direction')?.value).toBe('210°');
  expect(detail.hours).toHaveLength(1);
  expect(detail.hours[0].condition).toBe('partly_cloudy');
  expect(detail.hours[0].uv).toBe('UV 3');
  expect(detail.hourlyHorizonHours).toBe(34);
});

test('hour detail exposes the point values and formatted title', () => {
  const detail = weatherHourDetail(forecast([]), point('2026-10-28T10:00:00Z', 3));
  expect(detail.title).toBe('10:00');
  expect(detail.conditionLabel).toBe('Partly cloudy');
  expect(detail.temperature).toBe('14°C');
  expect(detail.stats.find((stat) => stat.id === 'rain-chance')?.value).toBe('20%');
  expect(detail.stats.find((stat) => stat.id === 'uv')?.value).toBe('3');
});

test('daily detail omits missing values and reports no hourly coverage', () => {
  const detail = weatherDayDetail(
    forecast([]),
    day({
      night_condition: null,
      feels_like_high: null,
      feels_like_low: null,
      wind: null,
      wind_gust: null,
      wind_direction: null,
      humidity: null,
      pressure: null,
      visibility: null,
      precipitation: null,
      rain_probability: null,
      snow_probability: null,
      uv_max: null,
      sunrise: null,
      sunset: null,
    })
  );
  expect(detail.stats.map((stat) => stat.id)).toEqual(['precipitation-chance']);
  expect(detail.hours).toEqual([]);
  expect(detail.hourlyHorizonHours).toBeNull();
});
