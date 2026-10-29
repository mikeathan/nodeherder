import type { WeatherForecast } from '@/types/weather.type';
function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Invalid weather data');
  return value as Record<string, unknown>;
}
function text(value: unknown) {
  if (typeof value !== 'string' || !value || value.length > 512) throw new Error('Invalid weather text');
  return value;
}
function time(value: unknown) {
  const s = text(value);
  if (!s.endsWith('Z') || !Number.isFinite(Date.parse(s))) throw new Error('Invalid weather timestamp');
  return s;
}
function quantity(value: unknown, min: number, max: number) {
  if (value === null) return;
  if (typeof value !== 'number' || !Number.isFinite(value) || value < min || value > max)
    throw new Error('Invalid weather quantity');
}
const conditions = [
  'unknown',
  'clear',
  'partly_cloudy',
  'cloudy',
  'fog',
  'drizzle',
  'rain',
  'snow',
  'showers',
  'thunderstorm',
];
function condition(value: unknown) {
  if (!conditions.includes(text(value))) throw new Error('Invalid weather condition');
}
function point(value: unknown) {
  const v = record(value);
  time(v.time);
  condition(v.condition);
  for (const key of ['temperature', 'feels_like']) quantity(v[key], -150, 100);
  for (const key of ['wind', 'gust']) quantity(v[key], 0, 600);
  quantity(v.direction, 0, 360);
  quantity(v.humidity, 0, 100);
  quantity(v.precipitation_probability, 0, 100);
  quantity(v.precipitation, 0, 10000);
}
function validateHeader(forecast: Record<string, unknown>) {
  if (forecast.contract_version !== 'weather_forecast/v1') throw new Error('Invalid weather version');
  for (const field of ['instance_id', 'resource_id', 'location', 'attribution']) text(forecast[field]);
  if (typeof forecast.revision !== 'number' || !Number.isSafeInteger(forecast.revision) || forecast.revision < 1)
    throw new Error('Invalid weather revision');
  new Intl.DateTimeFormat('en', { timeZone: text(forecast.timezone) });
  for (const field of ['captured_at', 'last_checked_at', 'evaluated_at']) time(forecast[field]);
  if (forecast.freshness !== 'fresh' && forecast.freshness !== 'stale') throw new Error('Invalid weather freshness');
}

function forecastList(value: unknown, maximum: number): unknown[] {
  if (!Array.isArray(value) || value.length > maximum) throw new Error('Invalid weather horizon');
  return value;
}

function validateHours(hours: unknown[]) {
  let previous = 0;
  for (const value of hours) {
    point(value);
    const instant = Date.parse(String(record(value).time));
    if (instant <= previous) throw new Error('Unordered weather hours');
    previous = instant;
  }
}

function calendarDate(value: unknown): string {
  const date = text(value);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) throw new Error('Invalid forecast date');
  if (!Number.isFinite(Date.parse(date))) throw new Error('Invalid forecast date');
  if (new Date(date).toISOString().slice(0, 10) !== date) throw new Error('Invalid forecast date');
  return date;
}

function validateDay(value: unknown): string {
  const day = record(value);
  const date = calendarDate(day.date);
  condition(day.condition);
  quantity(day.high, -150, 100);
  quantity(day.low, -150, 100);
  if (typeof day.high === 'number' && typeof day.low === 'number' && day.low > day.high)
    throw new Error('Invalid temperature range');
  quantity(day.wind, 0, 600);
  quantity(day.precipitation_probability, 0, 100);
  quantity(day.precipitation, 0, 10000);
  return date;
}

function validateDays(days: unknown[]) {
  let previous = '';
  for (const day of days) {
    const date = validateDay(day);
    if (date <= previous) throw new Error('Unordered weather days');
    previous = date;
  }
}

/** Validate provider-independent API weather before it enters shared state.
 * Checks UTC instants, the location timezone, ordered horizons and nullable quantities;
 * malformed data throws and cannot replace the accepted resource.
 */
export function parseWeatherForecast(value: unknown): WeatherForecast {
  const forecast = record(value);
  validateHeader(forecast);
  if (forecast.current !== null) point(forecast.current);
  const hours = forecastList(forecast.hourly, 384);
  const days = forecastList(forecast.daily, 16);
  if (hours.length + days.length === 0) throw new Error('Empty weather horizon');
  validateHours(hours);
  validateDays(days);
  return value as WeatherForecast;
}
