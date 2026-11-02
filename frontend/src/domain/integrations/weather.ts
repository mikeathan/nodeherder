import { calendarDateInZone } from '@/utils/date.utils';
import type { WeatherForecast, WeatherPoint } from '@/types/integrations/weather';

/** Maximum hourly rows rendered for one daily detail; bounds a provider's long horizon. */
export const WEATHER_DAY_HOUR_LIMIT = 24;
const MILLISECONDS_PER_HOUR = 3_600_000;

/** Present normalized conditions as readable text; unrecognized or absent data remains explicit. */
export function weatherLabel(value: string | undefined): string {
  return value ? value.replaceAll('_', ' ').replace(/^./, (c) => c.toUpperCase()) : 'Unavailable';
}

/**
 * Select forecast hours whose local calendar date matches `date`. Provider order is retained
 * and the result is capped so a long horizon cannot grow the detail view without bound.
 */
export function hoursForDate(forecast: WeatherForecast, date: string): WeatherPoint[] {
  return forecast.hourly
    .filter((point) => calendarDateInZone(point.time, forecast.timezone) === date)
    .slice(0, WEATHER_DAY_HOUR_LIMIT);
}

/**
 * Report how many hours of hourly data remain from the evaluation time. Returns null when the
 * provider supplied no future hours, so the UI does not invent a horizon.
 */
export function weatherHourlyHorizonHours(forecast: WeatherForecast): number | null {
  const last = forecast.hourly[forecast.hourly.length - 1];
  if (!last) return null;
  const lastTime = Date.parse(last.time);
  const evaluated = Date.parse(forecast.evaluated_at);
  if (!Number.isFinite(lastTime) || !Number.isFinite(evaluated) || lastTime <= evaluated) return null;
  return Math.ceil((lastTime - evaluated) / MILLISECONDS_PER_HOUR);
}
