/*
 * Presentation-only fusion of hub-derived presence across sensors. Occupancy is an OR; counts
 * are summed reported targets, not people. Nothing here feeds automations or persistence.
 */
export type FusedZone = { name: string; occupied: boolean | null };

export type FusedSensor = {
  id: string;
  name: string;
  /** Derived device availability, not the source device's connection. */
  online: boolean;
  occupied: boolean | null;
  targetCount: number | null;
  zones: readonly FusedZone[];
};

export type FusionResult = {
  occupied: boolean | null;
  /** Sum of known per-sensor counts; not a person count. */
  targetCount: number;
  online: number;
  total: number;
  /** Human-readable conflicts or caveats, empty when sensors agree. */
  notes: string[];
  /** Names of sensors whose derived device is offline. */
  offline: string[];
};

/** Combines sensors with explicit unknown semantics: offline or missing is never clear. */
export function fuseSensors(sensors: readonly FusedSensor[]): FusionResult {
  const known = sensors.filter((sensor) => sensor.occupied !== null);
  const knownOccupied = known.filter((sensor) => sensor.occupied === true).length;
  const occupied = known.length === 0 ? null : knownOccupied > 0;
  const counts = sensors
    .map((sensor) => sensor.targetCount)
    .filter((value): value is number => value !== null && Number.isFinite(value));
  const offline = sensors.filter((sensor) => !sensor.online).map((sensor) => sensor.name);
  const notes: string[] = [];
  if (knownOccupied > 0 && knownOccupied < known.length) {
    notes.push(`${knownOccupied} of ${known.length} reporting sensors are occupied while others report clear.`);
  }
  if (counts.length > 1) {
    const spread = Math.max(...counts) - Math.min(...counts);
    if (spread > 1) notes.push(`Reported target counts differ by ${spread} across sensors.`);
  }
  if (offline.length) notes.push(`${offline.length} sensor${offline.length === 1 ? '' : 's'} offline; they are not counted as clear.`);
  return {
    occupied,
    targetCount: counts.reduce((sum, value) => sum + value, 0),
    online: sensors.filter((sensor) => sensor.online).length,
    total: sensors.length,
    notes,
    offline,
  };
}
