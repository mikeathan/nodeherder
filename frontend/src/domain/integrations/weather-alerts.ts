import type { AlertLevel, AlertSeverity } from '@/types/integrations/weather-alerts';

const SEVERITY_LABELS: Record<AlertSeverity, string> = {
  minor: 'Minor',
  moderate: 'Moderate',
  severe: 'Severe',
  extreme: 'Extreme',
  unknown: 'Unknown',
};

const LEVEL_LABELS: Record<AlertLevel, string> = {
  yellow: 'Yellow',
  amber: 'Amber',
  red: 'Red',
  unknown: 'Unknown',
};

/** Present the normalized warning severity. */
export function alertSeverityLabel(value: AlertSeverity | string): string {
  return SEVERITY_LABELS[value as AlertSeverity] ?? 'Unknown';
}

/** Present the provider warning level independently of the normalized severity. */
export function alertLevelLabel(value: AlertLevel | string): string {
  return LEVEL_LABELS[value as AlertLevel] ?? 'Unknown';
}

/** Map severity to a small visual tone shared by the results and Home card. */
export function alertSeverityTone(value: AlertSeverity | string): 'minor' | 'moderate' | 'severe' | 'unknown' {
  switch (value) {
    case 'minor':
      return 'minor';
    case 'moderate':
      return 'moderate';
    case 'severe':
    case 'extreme':
      return 'severe';
    default:
      return 'unknown';
  }
}

/** Label the number of active warnings without implying live monitoring. */
export function activeAlertCountLabel(count: number): string {
  if (count === 0) return 'No active warnings';
  return count === 1 ? '1 active warning' : `${count} active warnings`;
}
