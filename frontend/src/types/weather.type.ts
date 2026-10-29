export type WeatherContractVersion = 'weather_forecast/v1';
export type WeatherFreshness = 'fresh' | 'stale';

/** Provider-independent weather quantities; null preserves unavailable data. */
export interface WeatherValues {
  temperature: number | null;
  feels_like: number | null;
  wind: number | null;
  gust: number | null;
  direction: number | null;
  humidity: number | null;
  precipitation_probability: number | null;
  precipitation: number | null;
}
/** A forecast valid at a UTC instant, formatted in its resource timezone. */
export interface WeatherPoint extends WeatherValues {
  time: string;
  condition: string;
}
/** Daily summary for a local date, retaining optional provider fields. */
export interface WeatherDay {
  date: string;
  condition: string;
  high: number | null;
  low: number | null;
  wind: number | null;
  precipitation_probability: number | null;
  precipitation: number | null;
}
/** Validated integration snapshot, owned by the existing integrations store. */
export interface WeatherForecast {
  contract_version: WeatherContractVersion;
  instance_id: string;
  resource_id: string;
  revision: number;
  location: string;
  timezone: string;
  attribution: string;
  captured_at: string;
  last_checked_at: string;
  evaluated_at: string;
  freshness: WeatherFreshness;
  current: WeatherPoint | null;
  hourly: WeatherPoint[];
  daily: WeatherDay[];
}
/** Declarative provider series; selectors are JSON pointers, never executable expressions. */
export interface WeatherSeries {
  offset?: string;
  items: string;
  parent?: string;
  time: string;
  format: string;
  columns?: boolean;
  fields: Record<string, string>;
  scales?: Record<string, number>;
}
/** Weather-only mapping shared by Form/YAML ingress and backend capability validation. */
export interface WeatherMapping {
  location: string;
  timezone: string;
  default_timezone: string;
  attribution: string;
  conditions: Record<string, string>;
  current?: WeatherSeries;
  hourly: WeatherSeries;
  daily: WeatherSeries;
}
