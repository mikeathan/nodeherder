import { currentWeatherStats, dailyWeatherStats, hourlyWeatherStats } from '@/integrations/weather/stats';
import type { WeatherDay, WeatherForecast, WeatherPoint } from '@/types/integrations/weather';

function point(time: string, overrides: Partial<WeatherPoint> = {}): WeatherPoint {
  return {
    time,
    condition: 'cloudy',
    temperature: 10,
    feels_like: 7,
    wind: 12,
    gust: 26,
    direction: 30,
    humidity: 82,
    precipitation_probability: 6,
    precipitation: 0,
    uv: 0,
    ...overrides,
  };
}

function day(overrides: Partial<WeatherDay> = {}): WeatherDay {
  return {
    date: '2026-11-03',
    condition: 'cloudy',
    night_condition: 'clear',
    high: 11,
    low: 8,
    feels_like_high: 7,
    feels_like_low: 5,
    wind: 17,
    wind_gust: 37,
    wind_direction: 30,
    humidity: 82,
    pressure: 1013,
    visibility: 20000,
    precipitation_probability: 17,
    precipitation: 0.6,
    rain_probability: 15,
    snow_probability: 0,
    sunrise: '07:08',
    sunset: '16:30',
    uv_max: 0.7,
    ...overrides,
  };
}

function forecast(overrides: Partial<WeatherForecast> = {}): WeatherForecast {
  return {
    contract_version: 'weather_forecast/v1',
    instance_id: 'instance',
    resource_id: 'resource',
    revision: 1,
    location: 'Bristol',
    timezone: 'UTC',
    attribution: 'Test',
    captured_at: '2026-11-03T12:00:00Z',
    last_checked_at: '2026-11-03T12:00:00Z',
    evaluated_at: '2026-11-03T12:00:00Z',
    freshness: 'fresh',
    current: point('2026-11-03T12:00:00Z'),
    hourly: [point('2026-11-03T12:00:00Z')],
    daily: [day()],
    ...overrides,
  };
}

test('current stats fall back per value to the nearest hourly point', () => {
  const resource = forecast({
    current: point('2026-11-03T12:00:00Z', { precipitation_probability: null, humidity: null }),
    hourly: [point('2026-11-03T12:00:00Z', { precipitation_probability: 20, humidity: 82 })],
  });
  const stats = currentWeatherStats(resource);
  expect(stats.find((stat) => stat.id === 'rain-chance')?.value).toBe('20%');
  expect(stats.find((stat) => stat.id === 'humidity')?.value).toBe('82%');
  expect(stats.some((stat) => stat.value === '—')).toBe(false);
});

test('current stats omit unavailable values instead of rendering a dash', () => {
  const stats = currentWeatherStats(
    forecast({
      current: point('2026-11-03T12:00:00Z', {
        feels_like: null,
        wind: null,
        gust: null,
        direction: null,
        humidity: null,
        precipitation_probability: null,
        precipitation: null,
        uv: null,
      }),
      hourly: [],
      daily: [],
    })
  );
  expect(stats).toEqual([]);
});

test('current stats include today’s daily context', () => {
  const stats = currentWeatherStats(forecast());
  expect(stats.find((stat) => stat.id === 'today-high')?.value).toBe('11°C');
  expect(stats.find((stat) => stat.id === 'today-low')?.value).toBe('8°C');
  expect(stats.find((stat) => stat.id === 'today-uv')?.value).toBe('0.7');
  expect(stats.find((stat) => stat.id === 'sunrise')?.value).toBe('07:08');
  expect(stats.find((stat) => stat.id === 'sunset')?.value).toBe('16:30');
});

test('hourly stats expose the point values with icons', () => {
  const stats = hourlyWeatherStats(point('2026-11-03T14:00:00Z'));
  expect(stats.find((stat) => stat.id === 'feels-like')?.value).toBe('7°C');
  expect(stats.find((stat) => stat.id === 'rain-chance')?.value).toBe('6%');
  expect(stats.find((stat) => stat.id === 'humidity')?.value).toBe('82%');
  expect(stats.find((stat) => stat.id === 'wind')?.icon).toBe('weatherWind');
  expect(stats.find((stat) => stat.id === 'uv')?.value).toBe('0');
});

test('daily stats use the primary high/low and carry icons', () => {
  const stats = dailyWeatherStats(day());
  expect(stats.some((stat) => stat.id === 'high' || stat.id === 'low')).toBe(false);
  expect(stats.find((stat) => stat.id === 'uv')?.icon).toBe('weatherSun');
  expect(stats.find((stat) => stat.id === 'sunrise')?.icon).toBe('weatherSunrise');
  expect(stats.find((stat) => stat.id === 'night')?.value).toBe('Clear');
});
