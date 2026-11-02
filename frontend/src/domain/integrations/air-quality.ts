import type {
  AirQualityCategory,
  AirQualityForecast,
  AirQualityIndexFamily,
  AirQualityPoint,
  AirQualityPollen,
  AirQualityValues,
} from '@/types/integrations/air-quality';

const CATEGORY_LABELS: Record<AirQualityCategory, string> = {
  unknown: 'Unknown',
  good: 'Good',
  fair: 'Fair',
  moderate: 'Moderate',
  poor: 'Poor',
  very_poor: 'Very poor',
  extremely_poor: 'Extremely poor',
  unhealthy_sensitive: 'Unhealthy for sensitive groups',
  unhealthy: 'Unhealthy',
  very_unhealthy: 'Very unhealthy',
  hazardous: 'Hazardous',
};

/** Present a server-derived category in readable text. */
export function airQualityCategoryLabel(category: AirQualityCategory | string): string {
  return CATEGORY_LABELS[category as AirQualityCategory] ?? 'Unknown';
}

/** Map a category to a small visual tone shared by the results and Home card. */
export function airQualityCategoryTone(category: AirQualityCategory | string): 'good' | 'fair' | 'moderate' | 'poor' | 'unknown' {
  switch (category) {
    case 'good':
      return 'good';
    case 'fair':
      return 'fair';
    case 'moderate':
    case 'unhealthy_sensitive':
      return 'moderate';
    case 'poor':
    case 'very_poor':
    case 'extremely_poor':
    case 'unhealthy':
    case 'very_unhealthy':
    case 'hazardous':
      return 'poor';
    default:
      return 'unknown';
  }
}

/** Present the declared index family. */
export function airQualityIndexFamilyLabel(family: AirQualityIndexFamily | string): string {
  return family === 'us' ? 'US AQI' : 'European AQI';
}

/** One present pollen species; entries with no provider value are omitted. */
export interface PollenReading {
  key: keyof AirQualityPollen;
  label: string;
  value: number;
}

const POLLEN_LABELS: Record<keyof AirQualityPollen, string> = {
  alder: 'Alder',
  birch: 'Birch',
  grass: 'Grass',
  mugwort: 'Mugwort',
  olive: 'Olive',
  ragweed: 'Ragweed',
};

/** List supplied pollen readings in a stable species order. */
export function pollenReadings(pollen: AirQualityPollen): PollenReading[] {
  return (Object.keys(POLLEN_LABELS) as (keyof AirQualityPollen)[])
    .flatMap((key) => {
      const value = pollen[key];
      return value == null ? [] : [{ key, label: POLLEN_LABELS[key], value }];
    });
}

/** Choose the provider current values or the latest elapsed hourly point. */
export function currentAirQualityPoint(forecast: AirQualityForecast): AirQualityValues | null {
  if (forecast.current) return forecast.current;
  const now = Date.parse(forecast.evaluated_at);
  let latest: AirQualityPoint | null = null;
  for (const point of forecast.hourly) {
    if (Date.parse(point.time) <= now) latest = point;
    else break;
  }
  return latest ?? forecast.hourly[0] ?? null;
}

/** Return the next 24 provider hours after the snapshot evaluation instant. */
export function upcomingAirQualityHours(forecast: AirQualityForecast): AirQualityPoint[] {
  const now = Date.parse(forecast.evaluated_at);
  return forecast.hourly.filter((point) => Date.parse(point.time) > now).slice(0, 24);
}

/** Pick the highest supplied pollen species for a compact summary. */
export function dominantPollen(pollen: AirQualityPollen): PollenReading | null {
  return pollenReadings(pollen).reduce<PollenReading | null>(
    (best, reading) => (best == null || reading.value > best.value ? reading : best),
    null
  );
}
