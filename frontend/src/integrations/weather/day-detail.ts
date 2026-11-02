import { hoursForDate, weatherHourlyHorizonHours, weatherLabel } from '@/domain/integrations/weather';
import { createZonedTimeFormatter, formatInstant, formatShortCalendarDate } from '@/utils/date.utils';
import { formatQuantity } from '@/utils/units';
import type { WeatherDay, WeatherForecast, WeatherPoint } from '@/types/integrations/weather';
import { dailyWeatherStats, hourlyWeatherStats, type WeatherStat } from './stats';

/** One hourly row in the daily detail, already formatted for display. */
export interface WeatherDayHour {
  id: string;
  time: string;
  /** Raw normalized condition for the shared weather icon. */
  condition: string;
  temperature: string | null;
  precipitation: string | null;
  uv: string | null;
}

/** Presentation model for one daily entry, bounded by the domain hour limit. */
export interface WeatherDayDetail {
  title: string;
  /** Raw condition code for the shared weather icon. */
  condition: string;
  conditionLabel: string;
  /** High / low pair shown as the panel's primary value. */
  primary: string;
  stats: WeatherStat[];
  hours: WeatherDayHour[];
  /** Future hourly coverage in hours; null when the provider supplied none. */
  hourlyHorizonHours: number | null;
}

/** Inline detail for one hourly card. */
export interface WeatherHourDetail {
  id: string;
  title: string;
  /** Raw condition code for the shared weather icon. */
  condition: string;
  conditionLabel: string;
  temperature: string | null;
  stats: WeatherStat[];
}

/**
 * Build the bounded presentation for one daily entry. Optional values are omitted rather than
 * rendered as zero; the hourly list comes from `hoursForDate`, which retains provider order.
 */
export function weatherDayDetail(forecast: WeatherForecast, day: WeatherDay): WeatherDayDetail {
  const stats = dailyWeatherStats(day);
  const primaryParts: string[] = [];
  if (day.high != null) primaryParts.push(formatQuantity(day.high, 'temperature', { short: true }));
  if (day.low != null) primaryParts.push(formatQuantity(day.low, 'temperature', { short: true }));

  const hours = weatherHourCards(forecast, hoursForDate(forecast, day.date));

  return {
    title: formatShortCalendarDate(day.date),
    condition: day.condition,
    conditionLabel: weatherLabel(day.condition),
    primary: primaryParts.join(' / '),
    stats,
    hours,
    hourlyHorizonHours: weatherHourlyHorizonHours(forecast),
  };
}

/** Format the hourly cards shared by the next-24-hours and selected-day strips. */
export function weatherHourCards(forecast: WeatherForecast, points: WeatherPoint[]): WeatherDayHour[] {
  const formatter = createZonedTimeFormatter(forecast.timezone);
  return points.map((point) => ({
    id: point.time,
    time: formatInstant(point.time, formatter),
    condition: point.condition,
    temperature: point.temperature == null ? null : formatQuantity(point.temperature, 'temperature', { short: true }),
    precipitation:
      point.precipitation_probability == null ? null : formatQuantity(point.precipitation_probability, 'percent'),
    uv: point.uv == null ? null : `UV ${formatQuantity(point.uv, 'uv')}`,
  }));
}

/** Build the inline detail for one hourly card. */
export function weatherHourDetail(forecast: WeatherForecast, point: WeatherPoint): WeatherHourDetail {
  const formatter = createZonedTimeFormatter(forecast.timezone);
  return {
    id: point.time,
    title: formatInstant(point.time, formatter),
    condition: point.condition,
    conditionLabel: weatherLabel(point.condition),
    temperature: point.temperature == null ? null : formatQuantity(point.temperature, 'temperature'),
    stats: hourlyWeatherStats(point),
  };
}
