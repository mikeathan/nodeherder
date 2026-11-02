export type AirQualityContractVersion = 'air_quality/v1';
export type AirQualityIndexFamily = 'european' | 'us';
export type AirQualityCategory =
  | 'unknown'
  | 'good'
  | 'fair'
  | 'moderate'
  | 'poor'
  | 'very_poor'
  | 'extremely_poor'
  | 'unhealthy_sensitive'
  | 'unhealthy'
  | 'very_unhealthy'
  | 'hazardous';
export type AirQualityFreshness = 'fresh' | 'stale';

/** Pollen concentrations in grains per cubic metre; null preserves an absent species. */
export interface AirQualityPollen {
  alder: number | null;
  birch: number | null;
  grass: number | null;
  mugwort: number | null;
  olive: number | null;
  ragweed: number | null;
}

/** Pollutant concentrations in micrograms per cubic metre. */
export interface AirQualityValues {
  pm2_5: number | null;
  pm10: number | null;
  ozone: number | null;
  nitrogen_dioxide: number | null;
  sulphur_dioxide: number | null;
  carbon_monoxide: number | null;
}

/** One timestamped forecast hour of pollutant values. */
export interface AirQualityPoint extends AirQualityValues {
  time: string;
}

/** Validated air-quality snapshot owned by the existing integrations store. */
export interface AirQualityForecast {
  contract_version: AirQualityContractVersion;
  instance_id: string;
  resource_id: string;
  revision: number;
  location: string;
  timezone: string;
  latitude: number | null;
  longitude: number | null;
  index_family: AirQualityIndexFamily;
  index: number | null;
  category: AirQualityCategory;
  dominant: string;
  current: AirQualityValues | null;
  hourly: AirQualityPoint[];
  pollen: AirQualityPollen;
  attribution: string;
  captured_at: string;
  last_checked_at: string;
  evaluated_at: string;
  freshness: AirQualityFreshness;
}

/** Declarative provider series for air-quality fields. */
export interface AirQualitySeries {
  items: string;
  time: string;
  format: 'unix' | 'rfc3339';
  columns?: boolean;
  fields: Record<string, string>;
}

/** Air-quality-only mapping shared by Form/YAML ingress and backend validation. */
export interface AirQualityMapping {
  location: string;
  timezone: string;
  default_timezone: string;
  attribution: string;
  index_family: AirQualityIndexFamily;
  dominant?: string;
  current?: AirQualitySeries;
  hourly: AirQualitySeries;
}
