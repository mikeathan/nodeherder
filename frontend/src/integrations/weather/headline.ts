import type { WeatherForecast, WeatherPoint } from '@/types/weather.type';
import { weatherLabel } from '@/domain/weather';
import { calendarDateInZone, createZonedTimeFormatter, formatInstant } from '@/utils/date.utils';
import { formatQuantity } from '@/utils/units';

/** Where the near-now value came from. Missing data is never presented as an observation. */
export type WeatherHeadlineSource = 'observation' | 'hourly' | 'daily';

/** The best available near-now value, with its source and the time it describes. */
export interface WeatherHeadline {
  temperature: number | null;
  condition: string;
  source: WeatherHeadlineSource;
  at: string;
  label: string;
}

/**
 * Choose the near-now value without inventing data.
 *
 * Order: a projected current observation, the latest elapsed hourly point, the next
 * hourly point, then today's daily summary. Decisions use the server's `evaluated_at`
 * rather than the browser clock, so every client agrees and automatic refresh needs no
 * presentation change: a new snapshot simply moves the same selector forward.
 */
export function weatherHeadline(forecast: WeatherForecast): WeatherHeadline | null {
  if (forecast.current) {
    return {
      temperature: forecast.current.temperature,
      condition: forecast.current.condition,
      source: 'observation',
      at: forecast.current.time,
      label: 'Now',
    };
  }

  const now = Date.parse(forecast.evaluated_at);
  let elapsed: WeatherPoint | undefined;
  let upcoming: WeatherPoint | undefined;
  for (const point of forecast.hourly) {
    if (Date.parse(point.time) <= now) elapsed = point;
    else {
      upcoming = point;
      break;
    }
  }
  const formatter = createZonedTimeFormatter(forecast.timezone);
  if (elapsed) return headlineOf(elapsed, `Observed ${formatInstant(elapsed.time, formatter)}`);
  if (upcoming) return headlineOf(upcoming, `Next ${formatInstant(upcoming.time, formatter)}`);

  const today = calendarDateInZone(forecast.evaluated_at, forecast.timezone);
  const day = forecast.daily.find((entry) => entry.date === today) ?? forecast.daily[0];
  if (day) {
    return {
      temperature: day.high,
      condition: day.condition,
      source: 'daily',
      at: day.date,
      label: day.date === today ? 'Today' : day.date,
    };
  }
  return null;
}

/** Compose a single card-ready value; never join two "Unavailable" sentinels. */
export function weatherHeadlineValue(headline: WeatherHeadline | null): string {
  if (!headline) return 'No saved forecast';
  const parts: string[] = [];
  if (headline.temperature != null)
    parts.push(formatQuantity(headline.temperature, 'temperature', { short: true }));
  if (headline.condition) parts.push(weatherLabel(headline.condition));
  if (headline.source !== 'observation') parts.push(headline.label);
  return parts.join(' · ') || 'No saved forecast';
}

function headlineOf(point: WeatherPoint, label: string): WeatherHeadline {
  return {
    temperature: point.temperature,
    condition: point.condition,
    source: 'hourly',
    at: point.time,
    label,
  };
}
