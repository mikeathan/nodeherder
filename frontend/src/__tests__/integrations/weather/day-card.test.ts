import { weatherDayCard } from '@/integrations/weather/day-card';
import type { WeatherDay } from '@/types/integrations/weather';

function day(overrides: Partial<WeatherDay> = {}): WeatherDay {
  return {
    date: '2026-11-01',
    condition: 'drizzle',
    night_condition: 'cloudy',
    high: 11,
    low: 5,
    feels_like_high: 9,
    feels_like_low: 3,
    wind: 15,
    wind_gust: 28,
    wind_direction: 180,
    humidity: 80,
    pressure: 1010,
    visibility: 18000,
    precipitation_probability: 14,
    precipitation: 0.2,
    rain_probability: 14,
    snow_probability: 0,
    sunrise: '07:04',
    sunset: '16:33',
    uv_max: 0.5,
    ...overrides,
  };
}

test('builds the compact daily card with icon-led metrics', () => {
  const card = weatherDayCard(day());
  expect(card.title).toBe('Sun 1 Nov');
  expect(card.primary).toBe('11° / 5°');
  expect(card.conditionLabel).toBe('Drizzle');
  expect(card.metrics).toEqual([
    { id: 'rain', label: 'Rain', value: '14%', icon: 'weatherRain' },
    { id: 'wind', label: 'Wind', value: '15 km/h', icon: 'weatherWind' },
    { id: 'uv', label: 'UV', value: '0.5', icon: 'weatherSun' },
  ]);
});

test('omits unavailable metrics from the compact card', () => {
  const card = weatherDayCard(day({ precipitation_probability: null, wind: null, uv_max: null }));
  expect(card.metrics).toEqual([]);
});
