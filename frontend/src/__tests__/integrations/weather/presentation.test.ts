import { weatherPresenter } from '@/integrations/weather/presentation';
import type { ConnectionInstance } from '@/types/integrations';
import type { WeatherDay, WeatherForecast, WeatherPoint } from '@/types/integrations/weather';

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
    uv: null,
  };
}

const day = (date: string): WeatherDay => ({
  date,
  condition: 'partly_cloudy',
  night_condition: null,
  high: 16,
  low: 9,
  feels_like_high: null,
  feels_like_low: null,
  wind: 18,
  wind_gust: null,
  wind_direction: null,
  humidity: null,
  pressure: null,
  visibility: null,
  precipitation_probability: 20,
  precipitation: 0,
  rain_probability: null,
  snow_probability: null,
  sunrise: null,
  sunset: null,
  uv_max: null,
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
    captured_at: '2026-10-28T17:59:48Z',
    last_checked_at: '2026-10-28T17:59:48Z',
    evaluated_at: '2026-10-28T17:59:48Z',
    freshness: 'fresh',
    current: point('2026-10-28T17:00:00Z'),
    hourly: [point('2026-10-28T17:00:00Z'), point('2026-10-28T18:00:00Z')],
    daily: [day('2026-10-28')],
    ...overrides,
  };
}

function instance(overrides: Partial<ConnectionInstance> = {}): ConnectionInstance {
  return {
    credential_slots: [],
    runs: [],
    id: 'instance',
    name: 'Weather',
    provider_key: 'weather/met_office',
    revision: 1,
    etag: 'etag',
    enabled: true,
    status: 'ready',
    postcode: '',
    credential_configured: false,
    resources: [],
    schedule: null,
    last_error: null,
    updated_at: '2026-10-28T17:59:48Z',
    ...overrides,
  };
}

test('an expired current observation still renders a real headline', () => {
  const resource = forecast({ current: null, evaluated_at: '2026-10-28T18:00:01Z' });
  const source = weatherPresenter.sources(instance(), resource)[0];
  expect(source.value).toBe('14° · Partly cloudy · Observed 18:00');
  expect(source.value).not.toContain('Unavailable');
  expect(source.status).toBe('Forecast');
  expect(source.warning).toBe(false);
});

test('a stale snapshot says it needs refreshing without hiding the value', () => {
  const resource = forecast({ current: null, freshness: 'stale', evaluated_at: '2026-10-28T19:00:00Z' });
  const source = weatherPresenter.sources(instance(), resource)[0];
  expect(source.value).toContain('Observed 18:00');
  expect(source.status).toBe('Needs refreshing');
  expect(source.warning).toBe(true);
});

test('an empty snapshot reports no saved forecast instead of a sentinel pair', () => {
  const resource = forecast({ current: null, hourly: [], daily: [] });
  const source = weatherPresenter.sources(instance(), resource)[0];
  expect(source.value).toBe('No saved forecast');
  expect(source.status).toBe('No saved forecast');
  expect(source.warning).toBe(true);
});

test('the headline condition drives the card icon when current is absent', () => {
  const resource = forecast({ current: null, evaluated_at: '2026-10-28T18:00:01Z' });
  const source = weatherPresenter.sources(instance(), resource)[0];
  expect(source.icon?.props.condition).toBe('partly_cloudy');
});

test('an observation keeps the plain value and its condition', () => {
  const source = weatherPresenter.sources(instance(), forecast())[0];
  expect(source.value).toBe('14° · Partly cloudy');
  expect(source.status).toBe('Forecast');
});

test('the compact card exposes icon-led range, rain and UV metrics', () => {
  const resource = forecast();
  resource.daily[0].uv_max = 4;
  const source = weatherPresenter.sources(instance(), resource)[0];
  expect(source.origin).toBe('Bristol');
  expect(source.rows.map((row) => row.title)).toEqual(['High / Low', 'Rain chance', 'UV max']);
  expect(source.rows[0].value).toBe('H 16° · L 9°');
  expect(source.rows[1].value).toBe('20%');
  expect(source.rows[2].value).toBe('4');
  expect(source.rows.every((row) => row.icon)).toBe(true);
});

test('the compact card omits metrics the provider did not return', () => {
  const resource = forecast({
    hourly: [],
    daily: [day('2026-10-28')],
  });
  resource.daily[0].high = null;
  resource.daily[0].low = null;
  resource.daily[0].uv_max = null;
  expect(weatherPresenter.sources(instance(), resource)[0].rows).toEqual([]);
});
