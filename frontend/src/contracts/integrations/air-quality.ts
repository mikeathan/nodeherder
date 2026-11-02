import type {
  AirQualityCategory,
  AirQualityForecast,
  AirQualityIndexFamily,
  AirQualityPoint,
  AirQualityPollen,
  AirQualityValues,
} from '@/types/integrations/air-quality';

function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Invalid air-quality data');
  return value as Record<string, unknown>;
}
function text(value: unknown): string {
  if (typeof value !== 'string' || !value || value.length > 512) throw new Error('Invalid air-quality text');
  return value;
}
function timestamp(value: unknown): string {
  const result = text(value);
  if (!result.endsWith('Z') || !Number.isFinite(Date.parse(result))) throw new Error('Invalid air-quality timestamp');
  return result;
}
function number(value: unknown, min: number, max: number): number {
  if (typeof value !== 'number' || !Number.isFinite(value) || value < min || value > max)
    throw new Error('Invalid air-quality quantity');
  return value;
}
function nullableQuantity(value: unknown, min: number, max: number): number | null {
  return value === null ? null : number(value, min, max);
}
function nullableIndex(value: unknown): number | null {
  return value === null ? null : number(value, 0, 1000);
}
function nullableCoordinate(value: unknown, limit: number): number | null {
  return value === null ? null : number(value, -limit, limit);
}
function category(value: unknown): AirQualityCategory {
  const result = text(value);
  if (
    ![
      'unknown',
      'good',
      'fair',
      'moderate',
      'poor',
      'very_poor',
      'extremely_poor',
      'unhealthy_sensitive',
      'unhealthy',
      'very_unhealthy',
      'hazardous',
    ].includes(result)
  )
    throw new Error('Invalid air-quality category');
  return result as AirQualityCategory;
}
function family(value: unknown): AirQualityIndexFamily {
  const result = text(value);
  if (result !== 'european' && result !== 'us') throw new Error('Invalid air-quality index family');
  return result;
}
function values(value: unknown): AirQualityValues {
  const v = record(value);
  return {
    pm2_5: nullableQuantity(v.pm2_5, 0, 10000),
    pm10: nullableQuantity(v.pm10, 0, 10000),
    ozone: nullableQuantity(v.ozone, 0, 10000),
    nitrogen_dioxide: nullableQuantity(v.nitrogen_dioxide, 0, 10000),
    sulphur_dioxide: nullableQuantity(v.sulphur_dioxide, 0, 10000),
    carbon_monoxide: nullableQuantity(v.carbon_monoxide, 0, 10000),
  };
}
function pollen(value: unknown): AirQualityPollen {
  const v = record(value);
  return {
    alder: nullableQuantity(v.alder, 0, 10000),
    birch: nullableQuantity(v.birch, 0, 10000),
    grass: nullableQuantity(v.grass, 0, 10000),
    mugwort: nullableQuantity(v.mugwort, 0, 10000),
    olive: nullableQuantity(v.olive, 0, 10000),
    ragweed: nullableQuantity(v.ragweed, 0, 10000),
  };
}
function point(value: unknown): AirQualityPoint {
  const v = record(value);
  return { time: timestamp(v.time), ...values(v) };
}
function orderedPoints(value: unknown): AirQualityPoint[] {
  if (!Array.isArray(value) || value.length > 168) throw new Error('Invalid air-quality horizon');
  let previous = 0;
  return value.map((entry) => {
    const result = point(entry);
    const instant = Date.parse(result.time);
    if (instant <= previous) throw new Error('Unordered air-quality hours');
    previous = instant;
    return result;
  });
}

/**
 * Validate a provider-independent air-quality resource before it enters shared state.
 * Checks bounded concentrations/index, the index family, coordinates, pollen and ordered hours.
 * @throws Error for malformed or unsupported data; TypeScript types alone do not validate API input.
 */
export function parseAirQualityForecast(value: unknown): AirQualityForecast {
  const v = record(value);
  if (v.contract_version !== 'air_quality/v1') throw new Error('Invalid air-quality version');
  for (const field of ['instance_id', 'resource_id', 'location', 'attribution']) text(v[field]);
  if (typeof v.revision !== 'number' || !Number.isSafeInteger(v.revision) || v.revision < 1)
    throw new Error('Invalid air-quality revision');
  new Intl.DateTimeFormat('en', { timeZone: text(v.timezone) });
  for (const field of ['captured_at', 'last_checked_at', 'evaluated_at']) timestamp(v[field]);
  if (v.freshness !== 'fresh' && v.freshness !== 'stale') throw new Error('Invalid air-quality freshness');
  family(v.index_family);
  nullableIndex(v.index);
  category(v.category);
  if (typeof v.dominant !== 'string' || v.dominant.length > 64) throw new Error('Invalid air-quality dominant pollutant');
  const latitude = nullableCoordinate(v.latitude, 90);
  const longitude = nullableCoordinate(v.longitude, 180);
  if ((latitude === null) !== (longitude === null)) throw new Error('Invalid air-quality coordinates');
  const current = v.current === null ? null : values(v.current);
  const hourly = orderedPoints(v.hourly);
  pollen(v.pollen);
  if (!current && !hourly.length) throw new Error('Empty air-quality horizon');
  return value as AirQualityForecast;
}
