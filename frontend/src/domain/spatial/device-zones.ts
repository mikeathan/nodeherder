/*
 * Device-facing zone translation. Produces validated requests for explicitly
 * mapped firmware coordinate entities; never clamps, rounds or assumes firmware
 * support for a drawn overlay.
 */
import type { Device, Expose } from '@/types/device';
import type { SpatialZone, ZoneCoordinateRole } from '@/types/spatial';
import { ZONE_COORDINATE_ROLES } from '@/types/spatial';
import { exposeRange, invalidRange, numericStep } from '@/domain/controls';
import { exposeControlUnavailable, exposeDisplayLabel } from '@/domain/exposes';
import { formatWithUnit } from '@/utils/units';
import { LENGTH_TO_MM } from './index';
import { isRectangle, validPolygon } from './workspace';

export { ZONE_COORDINATE_ROLES } from '@/types/spatial';

/** Relative tolerance for checking that a requested value aligns to its entity step. */
export const ZONE_STEP_TOLERANCE = 1e-6;
/** Guidance shown by the physical-zone editor before and after submission. */
export const ZONE_APPLY_MESSAGES = {
  polygonUnsupported:
    'Polygon overlays cannot be applied through four coordinate entities. Use the manufacturer’s polygon configurator and compatible firmware.',
  partial:
    'Submission stopped. Some coordinates may have been requested; inspect reported values before retrying.',
  requested:
    'Coordinates requested. These writes are not atomic. Check reported values and firmware mode before using the zone; submission does not prove device persistence.',
} as const;
/** Blocking messages shared by the preview and command path. */
export const ZONE_MESSAGES = {
  rectangleOnly: 'Only valid rectangular zones can be applied through coordinate entities.',
  missingCoordinate: (role: ZoneCoordinateRole) => `Map ${role} to an available readable/writable length entity.`,
  duplicateCoordinate: 'Choose four distinct coordinate entities.',
  outOfRange: (role: ZoneCoordinateRole) =>
    `${role} is outside the entity’s advertised bounds, or bounds are unavailable.`,
  offStep: (role: ZoneCoordinateRole) => `${role} does not align with the entity’s advertised step.`,
} as const;

/** Only readable/writable numeric length entities support verified coordinate editing. */
export function isZoneCoordinate(expose: Expose): boolean {
  return (
    expose.type === 'numeric' &&
    expose.access_mode === 'readwrite' &&
    !exposeControlUnavailable(expose) &&
    !!LENGTH_TO_MM[expose.unit?.trim().toLowerCase()]
  );
}

/** Coordinate entities offered by the zone editor, ordered by display label. */
export function zoneCoordinateOptions(exposes: Record<string, Expose>): Expose[] {
  return Object.values(exposes)
    .filter(isZoneCoordinate)
    .sort((left, right) => exposeDisplayLabel(left).localeCompare(exposeDisplayLabel(right)));
}

/** Read-only binary entities offered as reported occupancy sources. */
export function occupancyFlagOptions(device: Device): Expose[] {
  return Object.values(device.exposes).filter((expose) => expose.type === 'binary' && expose.access_mode !== 'write');
}

/** One formatted request/report row for the physical-zone editor. */
export type ZoneCommandRow = { role: ZoneCoordinateRole; entity: string; request: string; reported: string };

/** Formats requested and currently reported values using each entity's own unit. */
export function zoneCommandRows(
  commands: readonly { role: ZoneCoordinateRole; entity: string; value: number }[],
  exposes: Record<string, Expose>
): ZoneCommandRow[] {
  return commands.map((command) => {
    const expose = exposes[command.entity];
    return {
      role: command.role,
      entity: command.entity,
      request: formatWithUnit(command.value, expose.unit),
      reported: formatWithUnit(expose.data, expose.unit),
    };
  });
}

/** Produces exact native-unit requests or a blocking error; never clamps or silently rounds geometry. */
export function zoneCommandPlan(
  zone: SpatialZone,
  exposes: Record<string, Expose>
): { error: string; commands: { role: ZoneCoordinateRole; entity: string; value: number }[] } {
  const fail = (error: string) => ({ error, commands: [] });
  if (zone.shape !== 'rectangle' || !validPolygon(zone.points) || !isRectangle(zone.points)) {
    return fail(ZONE_MESSAGES.rectangleOnly);
  }
  const values: Record<ZoneCoordinateRole, number> = {
    beginX: Math.min(...zone.points.map((point) => point.x)),
    beginY: Math.min(...zone.points.map((point) => point.y)),
    endX: Math.max(...zone.points.map((point) => point.x)),
    endY: Math.max(...zone.points.map((point) => point.y)),
  };
  const used = new Set<string>();
  const commands: { role: ZoneCoordinateRole; entity: string; value: number }[] = [];
  for (const role of ZONE_COORDINATE_ROLES) {
    const entity = zone.entities?.[role] ?? '';
    const expose = exposes[entity];
    if (!entity || !expose || !isZoneCoordinate(expose)) return fail(ZONE_MESSAGES.missingCoordinate(role));
    if (used.has(entity)) return fail(ZONE_MESSAGES.duplicateCoordinate);
    used.add(entity);
    const value = values[role] / LENGTH_TO_MM[expose.unit.trim().toLowerCase()];
    const range = exposeRange(expose);
    if (invalidRange(expose) || !range || value < range.min || value > range.max) {
      return fail(ZONE_MESSAGES.outOfRange(role));
    }
    const steps = (value - range.min) / numericStep(expose);
    if (Math.abs(steps - Math.round(steps)) > ZONE_STEP_TOLERANCE) return fail(ZONE_MESSAGES.offStep(role));
    commands.push({ role, entity, value });
  }
  return { error: '', commands };
}
