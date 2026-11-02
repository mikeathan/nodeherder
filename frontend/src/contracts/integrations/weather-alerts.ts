import type {
  AlertLevel,
  AlertPolygon,
  AlertPolygonType,
  AlertSeverity,
  AlertStatus,
  WeatherAlerts,
  WeatherWarning,
} from '@/types/integrations/weather-alerts';

function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Invalid weather-alert data');
  return value as Record<string, unknown>;
}
function text(value: unknown, max = 512): string {
  if (typeof value !== 'string' || value.length > max) throw new Error('Invalid weather-alert text');
  return value;
}
function requiredText(value: unknown, max: number): string {
  const result = text(value, max);
  if (!result) throw new Error('Missing weather-alert text');
  return result;
}
function optionalText(value: unknown, max: number): string {
  return value === null ? '' : text(value, max);
}
function timestamp(value: unknown): string {
  const result = requiredText(value, 64);
  if (!result.endsWith('Z') || !Number.isFinite(Date.parse(result))) throw new Error('Invalid weather-alert timestamp');
  return result;
}
function optionalTimestamp(value: unknown): string | null {
  return value === null ? null : timestamp(value);
}
function status(value: unknown): AlertStatus {
  const result = text(value);
  if (result !== 'active' && result !== 'none') throw new Error('Invalid weather-alert status');
  return result;
}
function severity(value: unknown): AlertSeverity {
  const result = text(value);
  if (!['minor', 'moderate', 'severe', 'extreme', 'unknown'].includes(result))
    throw new Error('Invalid weather-alert severity');
  return result as AlertSeverity;
}
function level(value: unknown): AlertLevel {
  const result = text(value);
  if (!['yellow', 'amber', 'red', 'unknown'].includes(result)) throw new Error('Invalid weather-alert level');
  return result as AlertLevel;
}
function areas(value: unknown): string[] {
  if (!Array.isArray(value) || value.length > 50) throw new Error('Invalid weather-alert areas');
  return value.map((area) => requiredText(area, 120));
}
function ring(value: unknown): number {
  if (!Array.isArray(value) || value.length < 4) throw new Error('Invalid alert polygon ring');
  for (const position of value) {
    if (!Array.isArray(position) || position.length !== 2) throw new Error('Invalid alert position');
    const [longitude, latitude] = position;
    if (
      typeof longitude !== 'number' ||
      typeof latitude !== 'number' ||
      !Number.isFinite(longitude) ||
      !Number.isFinite(latitude) ||
      longitude < -180 ||
      longitude > 180 ||
      latitude < -90 ||
      latitude > 90
    )
      throw new Error('Invalid alert coordinate');
  }
  return value.length;
}
function polygon(value: unknown): { polygon: AlertPolygon | null; vertices: number } {
  if (value === null) return { polygon: null, vertices: 0 };
  const v = record(value);
  const type = requiredText(v.type, 32) as AlertPolygonType;
  if (type !== 'Polygon' && type !== 'MultiPolygon') throw new Error('Invalid alert polygon type');
  if (!Array.isArray(v.coordinates)) throw new Error('Invalid alert polygon coordinates');
  let vertices = 0;
  if (type === 'Polygon') {
    for (const item of v.coordinates) vertices += ring(item);
  } else {
    for (const item of v.coordinates) {
      if (!Array.isArray(item)) throw new Error('Invalid alert multipolygon');
      for (const area of item) vertices += ring(area);
    }
  }
  if (vertices > 2000) throw new Error('Alert polygon exceeds its bound');
  return {
    polygon: { type, coordinates: v.coordinates as AlertPolygon['coordinates'] },
    vertices,
  };
}
function warning(value: unknown): { warning: WeatherWarning; vertices: number } {
  const v = record(value);
  const warningID = requiredText(v.warning_id, 128);
  const geometry = polygon(v.polygon);
  const validFrom = timestamp(v.valid_from);
  const validTo = timestamp(v.valid_to);
  if (!(Date.parse(validTo) > Date.parse(validFrom))) throw new Error('Invalid warning window');
  return {
    warning: {
      warning_id: warningID,
      severity: severity(v.severity),
      level: level(v.level),
      type: text(v.type, 80),
      headline: requiredText(v.headline, 300),
      description: optionalText(v.description, 4000),
      instruction: optionalText(v.instruction, 2000),
      areas: areas(v.areas),
      valid_from: validFrom,
      valid_to: validTo,
      issued_at: optionalTimestamp(v.issued_at),
      modified_at: optionalTimestamp(v.modified_at),
      polygon: geometry.polygon,
    },
    vertices: geometry.vertices,
  };
}

/**
 * Validate a provider-independent alert resource before it enters shared state.
 * Checks warning bounds, status consistency, unique ids and bounded GeoJSON geometry.
 * @throws Error for malformed data; TypeScript types alone do not validate API input.
 */
export function parseWeatherAlerts(value: unknown): WeatherAlerts {
  const v = record(value);
  if (v.contract_version !== 'weather_alerts/v1') throw new Error('Invalid weather-alert version');
  for (const field of ['instance_id', 'resource_id', 'attribution']) requiredText(v[field], field === 'attribution' ? 512 : 128);
  if (typeof v.revision !== 'number' || !Number.isSafeInteger(v.revision) || v.revision < 1)
    throw new Error('Invalid weather-alert revision');
  for (const field of ['captured_at', 'last_checked_at', 'evaluated_at']) timestamp(v[field]);
  if (v.freshness !== 'fresh' && v.freshness !== 'stale') throw new Error('Invalid weather-alert freshness');
  const alertStatus = status(v.status);
  if (!Array.isArray(v.warnings) || v.warnings.length > 50) throw new Error('Invalid weather-alert list');
  const seen = new Set<string>();
  let totalVertices = 0;
  const warnings = v.warnings.map((entry) => {
    const parsed = warning(entry);
    if (seen.has(parsed.warning.warning_id)) throw new Error('Duplicate weather-alert id');
    seen.add(parsed.warning.warning_id);
    totalVertices += parsed.vertices;
    return parsed.warning;
  });
  if (totalVertices > 100000) throw new Error('Weather-alert geometry exceeds its bound');
  if ((alertStatus === 'none') !== (warnings.length === 0)) throw new Error('Inconsistent weather-alert status');
  return value as WeatherAlerts;
}
