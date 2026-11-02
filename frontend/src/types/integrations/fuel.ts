export type FuelPricesContractVersion = 'fuel_prices/v1';
export type FuelGrade = 'e10' | 'e5' | 'b7' | 'b10' | 'hvo' | 'other';
export type FuelFreshness = 'fresh' | 'stale';

/** Selected search location, required for every fuel snapshot. */
export interface FuelOrigin {
  label: string;
  latitude: number;
  longitude: number;
}

/** One grade's integer pence-per-litre price. */
export interface FuelPrice {
  grade: FuelGrade;
  label: string;
  pence_per_litre: number;
  updated_at: string | null;
}

/** One forecourt with optional coordinates and computed distance. */
export interface FuelStation {
  station_id: string;
  name: string;
  brand: string;
  address: string;
  postcode: string;
  latitude: number | null;
  longitude: number | null;
  distance_metres: number | null;
  prices: FuelPrice[];
}

/** Validated fuel-price snapshot owned by the existing integrations store. */
export interface FuelPrices {
  contract_version: FuelPricesContractVersion;
  instance_id: string;
  resource_id: string;
  revision: number;
  origin: FuelOrigin;
  radius_metres: number;
  currency: 'GBP';
  stations: FuelStation[];
  captured_at: string;
  last_checked_at: string;
  evaluated_at: string;
  freshness: FuelFreshness;
  attribution: string;
}

/** Declarative provider selectors for stations and their prices. */
export interface FuelPriceSeries {
  items: string;
  format: 'rfc3339' | 'unix';
  fields: Record<string, string>;
}
export interface FuelStationSeries {
  items: string;
  fields: Record<string, string>;
  prices: FuelPriceSeries;
}

/** Fuel-only mapping shared by Form/YAML ingress and backend validation. */
export interface FuelMapping {
  attribution: string;
  currency: 'GBP';
  radius_metres: number;
  origin_label?: string;
  stations: FuelStationSeries;
}
