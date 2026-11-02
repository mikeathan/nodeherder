export type WeatherAlertsContractVersion = 'weather_alerts/v1';
export type AlertStatus = 'active' | 'none';
export type AlertSeverity = 'minor' | 'moderate' | 'severe' | 'extreme' | 'unknown';
export type AlertLevel = 'yellow' | 'amber' | 'red' | 'unknown';
export type AlertFreshness = 'fresh' | 'stale';
export type AlertPolygonType = 'Polygon' | 'MultiPolygon';
export type AlertPolygonCoordinates = number[][][] | number[][][][];

/** Optional provider geometry retained for an opt-in geographic view. */
export interface AlertPolygon {
  type: AlertPolygonType;
  coordinates: AlertPolygonCoordinates;
}

/** One active warning with plain-text presentation fields. */
export interface WeatherWarning {
  warning_id: string;
  severity: AlertSeverity;
  level: AlertLevel;
  type: string;
  headline: string;
  description: string;
  instruction: string;
  areas: string[];
  valid_from: string;
  valid_to: string;
  issued_at: string | null;
  modified_at: string | null;
  polygon: AlertPolygon | null;
}

/** Validated alert snapshot owned by the existing integrations store. */
export interface WeatherAlerts {
  contract_version: WeatherAlertsContractVersion;
  instance_id: string;
  resource_id: string;
  revision: number;
  status: AlertStatus;
  warnings: WeatherWarning[];
  attribution: string;
  captured_at: string;
  last_checked_at: string;
  evaluated_at: string;
  freshness: AlertFreshness;
}

/** Declarative provider selectors for warning collections. */
export interface AlertWarningSeries {
  items: string;
  format: 'rfc3339' | 'unix';
  area_name?: string;
  fields: Record<string, string>;
}

/** Alerts-only mapping shared by Form/YAML ingress and backend validation. */
export interface AlertsMapping {
  attribution: string;
  warnings: AlertWarningSeries;
}
