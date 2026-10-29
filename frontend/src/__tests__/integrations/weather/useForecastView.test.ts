import { effectScope } from 'vue';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { useForecastView } from '@/integrations/weather/useForecastView';
import type { WeatherForecast } from '@/types/weather.type';

test('next 24 hours excludes the past and includes 24 future points', () => {
  const now = Date.parse('2026-10-28T10:30:00Z');
  const forecast: WeatherForecast = JSON.parse(
    readFileSync(resolve(__dirname, '../../fixtures/weather.json'), 'utf8')
  );
  forecast.evaluated_at = new Date(now).toISOString();
  forecast.timezone = 'UTC';
  forecast.daily = [];
  const point = forecast.hourly[0];
  forecast.hourly = Array.from({ length: 26 }, (_, index) => ({
    ...point, time: new Date(now + (index - 1) * 3600000).toISOString(),
  }));
  const scope = effectScope();
  const view = scope.run(() => useForecastView(() => forecast))!;
  expect(view.hourly.value).toHaveLength(24);
  expect(view.hourly.value[0].time).toBe(new Date(now + 3600000).toISOString());
  expect(view.hourly.value[23].time).toBe(new Date(now + 24 * 3600000).toISOString());
  scope.stop();
});
