import type { IconName } from '@/components/ui/icons';
import { weatherLabel } from '@/domain/integrations/weather';
import type { WeatherDay } from '@/types/integrations/weather';
import { formatShortCalendarDate } from '@/utils/date.utils';
import { formatQuantity } from '@/utils/units';

/** One compact metric shown on a daily forecast card. */
export interface WeatherDayCardMetric {
  id: string;
  label: string;
  value: string;
  icon: IconName;
}

/** Compact daily-card presentation; full detail remains in the selected-day panel. */
export interface WeatherDayCardModel {
  title: string;
  primary: string;
  condition: string;
  conditionLabel: string;
  metrics: WeatherDayCardMetric[];
}

/**
 * Build the compact daily card: date, high/low, condition and the three most scannable metrics.
 * Sunrise, sunset, feels-like and other daily values stay in the selected-day detail.
 */
export function weatherDayCard(day: WeatherDay): WeatherDayCardModel {
  const primaryParts: string[] = [];
  if (day.high != null) primaryParts.push(formatQuantity(day.high, 'temperature', { short: true }));
  if (day.low != null) primaryParts.push(formatQuantity(day.low, 'temperature', { short: true }));
  const metrics: WeatherDayCardMetric[] = [];
  if (day.precipitation_probability != null)
    metrics.push({
      id: 'rain',
      label: 'Rain',
      value: formatQuantity(day.precipitation_probability, 'percent'),
      icon: 'weatherRain',
    });
  if (day.wind != null)
    metrics.push({ id: 'wind', label: 'Wind', value: formatQuantity(day.wind, 'speed'), icon: 'weatherWind' });
  if (day.uv_max != null)
    metrics.push({ id: 'uv', label: 'UV', value: formatQuantity(day.uv_max, 'uv'), icon: 'weatherSun' });
  return {
    title: formatShortCalendarDate(day.date),
    primary: primaryParts.join(' / '),
    condition: day.condition,
    conditionLabel: weatherLabel(day.condition),
    metrics,
  };
}
