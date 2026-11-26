/*
 * Optional ambient-light trigger for panel night mode. Stored per browser and validated at
 * the ingress; the panel re-evaluates the current lux value reactively.
 */
export type PanelAmbientLight = {
  deviceId: string;
  expose: string;
  /** Lux below this value requests night mode. */
  threshold: number;
};

/** Parses one bounded ambient-light preference; invalid/legacy data falls back to none. */
export function parsePanelAmbient(raw: unknown): PanelAmbientLight | null {
  if (typeof raw !== 'object' || raw === null) return null;
  const deviceId = Reflect.get(raw, 'deviceId');
  const expose = Reflect.get(raw, 'expose');
  const threshold = Reflect.get(raw, 'threshold');
  if (typeof deviceId !== 'string' || !deviceId) return null;
  if (typeof expose !== 'string' || !expose) return null;
  if (typeof threshold !== 'number' || !Number.isFinite(threshold) || threshold < 0 || threshold > 100000) return null;
  return { deviceId, expose, threshold };
}

/** True when a finite lux reading is below the configured threshold. */
export function isAmbientNight(lux: unknown, threshold: number): boolean {
  return typeof lux === 'number' && Number.isFinite(lux) && lux < threshold;
}
