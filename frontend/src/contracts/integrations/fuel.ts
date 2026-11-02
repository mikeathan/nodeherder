import type { FuelGrade, FuelPrices, FuelPrice, FuelStation } from '@/types/integrations/fuel';

function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Invalid fuel-price data');
  return value as Record<string, unknown>;
}
function text(value: unknown, max: number): string {
  if (typeof value !== 'string' || value.length > max) throw new Error('Invalid fuel-price text');
  return value;
}
function requiredText(value: unknown, max: number): string {
  const result = text(value, max);
  if (!result) throw new Error('Missing fuel-price text');
  return result;
}
function timestamp(value: unknown): string {
  const result = requiredText(value, 64);
  if (!result.endsWith('Z') || !Number.isFinite(Date.parse(result))) throw new Error('Invalid fuel-price timestamp');
  return result;
}
function optionalTimestamp(value: unknown): string | null {
  return value === null ? null : timestamp(value);
}
function integer(value: unknown, min: number, max: number): number {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < min || value > max)
    throw new Error('Invalid fuel-price quantity');
  return value;
}
function nullableCoordinate(value: unknown, limit: number): number | null {
  if (value === null) return null;
  if (typeof value !== 'number' || !Number.isFinite(value) || value < -limit || value > limit)
    throw new Error('Invalid fuel coordinate');
  return value;
}
function grade(value: unknown): FuelGrade {
  const result = text(value, 16);
  if (!['e10', 'e5', 'b7', 'b10', 'hvo', 'other'].includes(result)) throw new Error('Invalid fuel grade');
  return result as FuelGrade;
}
function price(value: unknown): FuelPrice {
  const v = record(value);
  return {
    grade: grade(v.grade),
    label: requiredText(v.label, 40),
    pence_per_litre: integer(v.pence_per_litre, 1, 1000),
    updated_at: optionalTimestamp(v.updated_at),
  };
}
function prices(value: unknown): FuelPrice[] {
  if (!Array.isArray(value) || value.length > 6) throw new Error('Invalid fuel-price list');
  const seen = new Set<FuelGrade>();
  return value.map((entry) => {
    const parsed = price(entry);
    if (seen.has(parsed.grade)) throw new Error('Duplicate fuel grade');
    seen.add(parsed.grade);
    return parsed;
  });
}
function station(value: unknown): FuelStation {
  const v = record(value);
  const latitude = nullableCoordinate(v.latitude, 90);
  const longitude = nullableCoordinate(v.longitude, 180);
  if ((latitude === null) !== (longitude === null)) throw new Error('Invalid station coordinates');
  return {
    station_id: requiredText(v.station_id, 128),
    name: requiredText(v.name, 160),
    brand: text(v.brand, 80),
    address: text(v.address, 240),
    postcode: text(v.postcode, 16),
    latitude,
    longitude,
    distance_metres: v.distance_metres === null ? null : integer(v.distance_metres, 0, Number.MAX_SAFE_INTEGER),
    prices: prices(v.prices),
  };
}

/**
 * Validate a provider-independent fuel-price resource before it enters shared state.
 * Prices remain integer pence; zero/fractional values are rejected rather than rounded.
 * @throws Error for malformed data; TypeScript types alone do not validate API input.
 */
export function parseFuelPrices(value: unknown): FuelPrices {
  const v = record(value);
  if (v.contract_version !== 'fuel_prices/v1') throw new Error('Invalid fuel-price version');
  for (const field of ['instance_id', 'resource_id', 'attribution']) requiredText(v[field], field === 'attribution' ? 512 : 128);
  if (typeof v.revision !== 'number' || !Number.isSafeInteger(v.revision) || v.revision < 1)
    throw new Error('Invalid fuel-price revision');
  const origin = record(v.origin);
  requiredText(origin.label, 256);
  nullableCoordinate(origin.latitude, 90);
  nullableCoordinate(origin.longitude, 180);
  if (typeof origin.latitude !== 'number' || typeof origin.longitude !== 'number')
    throw new Error('Invalid fuel origin');
  integer(v.radius_metres, 100, 50000);
  if (v.currency !== 'GBP') throw new Error('Unsupported fuel currency');
  for (const field of ['captured_at', 'last_checked_at', 'evaluated_at']) timestamp(v[field]);
  if (v.freshness !== 'fresh' && v.freshness !== 'stale') throw new Error('Invalid fuel freshness');
  if (!Array.isArray(v.stations) || v.stations.length > 100) throw new Error('Invalid station list');
  const seen = new Set<string>();
  v.stations.forEach((entry) => {
    const parsed = station(entry);
    if (seen.has(parsed.station_id)) throw new Error('Duplicate fuel station');
    seen.add(parsed.station_id);
  });
  return value as FuelPrices;
}
