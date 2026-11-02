import type { IconName } from '@/components/ui/icons';
import type { WeatherDay, WeatherForecast, WeatherPoint } from '@/types/integrations/weather';
import { weatherLabel } from '@/domain/integrations/weather';
import { calendarDateInZone } from '@/utils/date.utils';
import { formatQuantity } from '@/utils/units';

/** One labelled weather value; the icon is omitted when no semantic icon applies. */
export interface WeatherStat {
  id: string;
  label: string;
  value: string;
  icon?: IconName;
}

/** Semantic icon per weather quantity, all resolved through the shared icon registry. */
const WEATHER_ICONS = {
  temperature: 'weatherTemperature',
  rain: 'weatherRain',
  precipitation: 'weatherPrecipitation',
  humidity: 'weatherHumidity',
  wind: 'weatherWind',
  gust: 'weatherGust',
  direction: 'weatherDirection',
  sun: 'weatherSun',
  pressure: 'weatherPressure',
  visibility: 'weatherVisibility',
  sunrise: 'weatherSunrise',
  sunset: 'weatherSunset',
  snow: 'weatherSnow',
  night: 'weatherNight',
} as const satisfies Record<string, IconName>;

/** Nearest hourly point to an instant, used only as a per-value fallback. */
function nearestPoint(forecast: WeatherForecast, time: string): WeatherPoint | undefined {
  const target = Date.parse(time);
  let nearest: WeatherPoint | undefined;
  let nearestDistance = Infinity;
  for (const point of forecast.hourly) {
    const distance = Math.abs(Date.parse(point.time) - target);
    if (distance < nearestDistance) {
      nearest = point;
      nearestDistance = distance;
    }
  }
  return nearest;
}

/** Current point, preferring the projected observation then the latest elapsed hour. */
function currentPoint(forecast: WeatherForecast): WeatherPoint | null {
  if (forecast.current) return forecast.current;
  const now = Date.parse(forecast.evaluated_at);
  let elapsed: WeatherPoint | undefined;
  for (const point of forecast.hourly) {
    if (Date.parse(point.time) <= now) elapsed = point;
    else break;
  }
  return elapsed ?? forecast.hourly[0] ?? null;
}

/** Append observation/hour tiles shared by the current and hourly detail views. */
function addPointStats(stats: WeatherStat[], point: WeatherPoint, fallback?: WeatherPoint): void {
  const feelsLike = point.feels_like ?? fallback?.feels_like ?? null;
  if (feelsLike != null)
    stats.push({
      id: 'feels-like',
      label: 'Feels like',
      value: formatQuantity(feelsLike, 'temperature'),
      icon: WEATHER_ICONS.temperature,
    });
  const rain = point.precipitation_probability ?? fallback?.precipitation_probability ?? null;
  if (rain != null)
    stats.push({
      id: 'rain-chance',
      label: 'Rain chance',
      value: formatQuantity(rain, 'percent'),
      icon: WEATHER_ICONS.rain,
    });
  const precipitation = point.precipitation ?? fallback?.precipitation ?? null;
  if (precipitation != null)
    stats.push({
      id: 'precipitation',
      label: 'Precipitation',
      value: formatQuantity(precipitation, 'precipitation'),
      icon: WEATHER_ICONS.precipitation,
    });
  const humidity = point.humidity ?? fallback?.humidity ?? null;
  if (humidity != null)
    stats.push({
      id: 'humidity',
      label: 'Humidity',
      value: formatQuantity(humidity, 'percent'),
      icon: WEATHER_ICONS.humidity,
    });
  const wind = point.wind ?? fallback?.wind ?? null;
  if (wind != null)
    stats.push({ id: 'wind', label: 'Wind', value: formatQuantity(wind, 'speed'), icon: WEATHER_ICONS.wind });
  const gust = point.gust ?? fallback?.gust ?? null;
  if (gust != null)
    stats.push({ id: 'gust', label: 'Gusts', value: formatQuantity(gust, 'speed'), icon: WEATHER_ICONS.gust });
  const direction = point.direction ?? fallback?.direction ?? null;
  if (direction != null)
    stats.push({
      id: 'direction',
      label: 'Wind direction',
      value: formatQuantity(direction, 'direction'),
      icon: WEATHER_ICONS.direction,
    });
  const uv = point.uv ?? fallback?.uv ?? null;
  if (uv != null)
    stats.push({ id: 'uv', label: 'UV', value: formatQuantity(uv, 'uv'), icon: WEATHER_ICONS.sun });
}

/**
 * Build the current-conditions tiles from the observation, falling back per value to the nearest
 * hourly point. Missing values are omitted so the summary never renders a bare dash.
 */
export function currentWeatherStats(forecast: WeatherForecast): WeatherStat[] {
  const current = currentPoint(forecast);
  if (!current) return [];
  const fallback = nearestPoint(forecast, current.time);
  const stats: WeatherStat[] = [];
  addPointStats(stats, current, fallback);
  const today = calendarDateInZone(forecast.evaluated_at, forecast.timezone);
  const day = forecast.daily.find((entry) => entry.date === today);
  if (day) {
    if (day.high != null)
      stats.push({
        id: 'today-high',
        label: 'Today high',
        value: formatQuantity(day.high, 'temperature'),
        icon: WEATHER_ICONS.temperature,
      });
    if (day.low != null)
      stats.push({
        id: 'today-low',
        label: 'Today low',
        value: formatQuantity(day.low, 'temperature'),
        icon: WEATHER_ICONS.temperature,
      });
    if (day.uv_max != null)
      stats.push({
        id: 'today-uv',
        label: 'Max UV today',
        value: formatQuantity(day.uv_max, 'uv'),
        icon: WEATHER_ICONS.sun,
      });
    if (day.sunrise != null)
      stats.push({ id: 'sunrise', label: 'Sunrise', value: day.sunrise, icon: WEATHER_ICONS.sunrise });
    if (day.sunset != null)
      stats.push({ id: 'sunset', label: 'Sunset', value: day.sunset, icon: WEATHER_ICONS.sunset });
  }
  return stats;
}

/** Build the detail tiles for one hourly forecast point. */
export function hourlyWeatherStats(point: WeatherPoint): WeatherStat[] {
  const stats: WeatherStat[] = [];
  addPointStats(stats, point);
  return stats;
}

/**
 * Build the selected-day tiles. High and low are the panel's primary value, so they are not
 * repeated; every other supplied daily value is included with its semantic icon.
 */
export function dailyWeatherStats(day: WeatherDay): WeatherStat[] {
  const stats: WeatherStat[] = [];
  if (day.feels_like_high != null)
    stats.push({
      id: 'feels-like-high',
      label: 'Feels like high',
      value: formatQuantity(day.feels_like_high, 'temperature'),
      icon: WEATHER_ICONS.temperature,
    });
  if (day.feels_like_low != null)
    stats.push({
      id: 'feels-like-low',
      label: 'Feels like low',
      value: formatQuantity(day.feels_like_low, 'temperature'),
      icon: WEATHER_ICONS.temperature,
    });
  if (day.wind != null)
    stats.push({ id: 'wind', label: 'Wind', value: formatQuantity(day.wind, 'speed'), icon: WEATHER_ICONS.wind });
  if (day.wind_gust != null)
    stats.push({ id: 'wind-gust', label: 'Gusts', value: formatQuantity(day.wind_gust, 'speed'), icon: WEATHER_ICONS.gust });
  if (day.wind_direction != null)
    stats.push({
      id: 'wind-direction',
      label: 'Wind direction',
      value: formatQuantity(day.wind_direction, 'direction'),
      icon: WEATHER_ICONS.direction,
    });
  if (day.humidity != null)
    stats.push({ id: 'humidity', label: 'Humidity', value: formatQuantity(day.humidity, 'percent'), icon: WEATHER_ICONS.humidity });
  if (day.pressure != null)
    stats.push({ id: 'pressure', label: 'Pressure', value: formatQuantity(day.pressure, 'pressure'), icon: WEATHER_ICONS.pressure });
  if (day.visibility != null)
    stats.push({
      id: 'visibility',
      label: 'Visibility',
      value: formatQuantity(day.visibility, 'distance'),
      icon: WEATHER_ICONS.visibility,
    });
  if (day.precipitation_probability != null)
    stats.push({
      id: 'precipitation-chance',
      label: 'Precipitation chance',
      value: formatQuantity(day.precipitation_probability, 'percent'),
      icon: WEATHER_ICONS.rain,
    });
  if (day.rain_probability != null)
    stats.push({
      id: 'rain-chance',
      label: 'Rain chance',
      value: formatQuantity(day.rain_probability, 'percent'),
      icon: WEATHER_ICONS.rain,
    });
  if (day.snow_probability != null)
    stats.push({
      id: 'snow-chance',
      label: 'Snow chance',
      value: formatQuantity(day.snow_probability, 'percent'),
      icon: WEATHER_ICONS.snow,
    });
  if (day.precipitation != null)
    stats.push({
      id: 'precipitation',
      label: 'Precipitation',
      value: formatQuantity(day.precipitation, 'precipitation'),
      icon: WEATHER_ICONS.precipitation,
    });
  if (day.uv_max != null)
    stats.push({ id: 'uv', label: 'Max UV', value: formatQuantity(day.uv_max, 'uv'), icon: WEATHER_ICONS.sun });
  if (day.sunrise != null)
    stats.push({ id: 'sunrise', label: 'Sunrise', value: day.sunrise, icon: WEATHER_ICONS.sunrise });
  if (day.sunset != null)
    stats.push({ id: 'sunset', label: 'Sunset', value: day.sunset, icon: WEATHER_ICONS.sunset });
  if (day.night_condition != null)
    stats.push({
      id: 'night',
      label: 'Night',
      value: weatherLabel(day.night_condition),
      icon: WEATHER_ICONS.night,
    });
  return stats;
}
