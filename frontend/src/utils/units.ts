/**
 * Single owner for numeric values rendered with a unit.
 *
 * Components and domain mappers pass a quantity kind, never a unit string. That keeps
 * rounding, spacing and missing-value copy consistent and prevents literals such as
 * `'°C'`, `'%'` or `' km/h'` leaking across call sites.
 */
export type QuantityKind =
  | 'temperature'
  | 'percent'
  | 'speed'
  | 'precipitation'
  | 'pressure'
  | 'power'
  | 'energy'
  | 'illuminance'
  | 'uv'
  | 'aqi'
  | 'pollutant'
  | 'distance';

/** Formatting overrides; every field has a documented default for the quantity. */
export interface QuantityOptions {
  /** Use the card-friendly symbol, for example `14°` instead of `14°C`. */
  short?: boolean;
  /** Force a space (`true`) or no space (`false`) before the unit. */
  spaced?: boolean;
  /** Decimal places; defaults to the quantity's precision. */
  digits?: number;
  /** Text for null, undefined or non-finite values. Defaults to an em dash. */
  missing?: string;
}

interface UnitSpec {
  unit: string;
  shortUnit?: string;
  digits: number;
  spaced?: boolean;
}

const QUANTITIES: Record<Exclude<QuantityKind, 'distance'>, UnitSpec> = {
  temperature: { unit: '°C', shortUnit: '°', digits: 0 },
  percent: { unit: '%', digits: 0 },
  speed: { unit: 'km/h', digits: 0, spaced: true },
  precipitation: { unit: 'mm', digits: 1, spaced: true },
  pressure: { unit: 'hPa', digits: 0, spaced: true },
  power: { unit: 'W', digits: 1, spaced: true },
  energy: { unit: 'kWh', digits: 2, spaced: true },
  illuminance: { unit: 'lx', digits: 0, spaced: true },
  uv: { unit: '', digits: 1 },
  aqi: { unit: '', digits: 0 },
  pollutant: { unit: 'µg/m³', digits: 0, spaced: true },
};

/**
 * Format a known physical quantity with its canonical unit.
 *
 * Null and non-finite values return `options.missing` (an em dash by default) so a
 * missing reading never becomes zero. Distance is expressed in metres and converted
 * to metres or kilometres automatically.
 */
export function formatQuantity(
  value: number | null | undefined,
  kind: QuantityKind,
  options: QuantityOptions = {}
): string {
  if (value == null || !Number.isFinite(value)) return options.missing ?? '—';
  if (kind === 'distance') return formatDistance(value, options);
  const spec = QUANTITIES[kind];
  const unit = options.short && spec.shortUnit !== undefined ? spec.shortUnit : spec.unit;
  const text = formatNumber(value, options.digits ?? spec.digits);
  if (!unit) return text;
  const spaced = options.spaced ?? spec.spaced ?? false;
  return `${text}${spaced ? ' ' : ''}${unit}`;
}

/**
 * Format a value whose unit comes from data rather than a known quantity, such as a
 * Zigbee expose or a provider field. Prefer `formatQuantity` when the quantity is known.
 */
export function formatWithUnit(
  value: number | null | undefined,
  unit: string,
  options: QuantityOptions = {}
): string {
  if (value == null || !Number.isFinite(value)) return options.missing ?? '—';
  const text = formatNumber(value, options.digits ?? 1);
  if (!unit) return text;
  const spaced =
    options.spaced === true || (options.spaced !== false && !/^[%°]/.test(unit));
  return `${text}${spaced ? ' ' : ''}${unit}`;
}

/** Format metres as metres below 1 km, otherwise as kilometres. */
export function formatDistance(
  metres: number | null | undefined,
  options: QuantityOptions = {}
): string {
  if (metres == null || !Number.isFinite(metres)) return options.missing ?? '—';
  if (Math.abs(metres) < 1000) return `${formatNumber(metres, options.digits ?? 0)} m`;
  return `${formatNumber(metres / 1000, options.digits ?? 1)} km`;
}

function formatNumber(value: number, digits: number): string {
  // Match the existing device formatter's rounding so migrating call sites cannot change values.
  const rounded = Number(value.toFixed(digits));
  return Object.is(rounded, -0) ? '0' : String(rounded);
}
