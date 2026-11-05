/*
 * Device-facing zone translation. Produces validated requests for explicitly
 * mapped firmware coordinate entities; never clamps, rounds or assumes firmware
 * support for a drawn overlay.
 */
import type { Device, Expose } from '@/types/device';
import type { FirmwareRole, SpatialPoint, SpatialWorkspace, SpatialZone, ZoneCoordinateRole, ZoneEntityRole } from '@/types/spatial';
import { ZONE_COORDINATE_ROLES, ZONE_OFF_DELAY_ROLE, ZONE_POLYGON_ROLE } from '@/types/spatial';
import { exposeRange, invalidRange, numericStep } from '@/domain/controls';
import { exposeControlUnavailable, exposeDisplayLabel, formatExposeValue } from '@/domain/exposes';
import { formatWithUnit } from '@/utils/units';
import { LENGTH_TO_MM } from './index';
import { isRectangle, validPolygon } from './workspace';

export { ZONE_COORDINATE_ROLES } from '@/types/spatial';

/** Relative tolerance for checking that a requested value aligns to its entity step. */
export const ZONE_STEP_TOLERANCE = 1e-6;
/** Guidance shown by the physical-zone editor before and after submission. */
export const ZONE_APPLY_MESSAGES = {
  polygonMode:
    'Polygon zones are written to the firmware polygon text entity. Enable the firmware polygon zone mode; while it is on, rectangle bounds do not apply.',
  polygonRequested:
    'Polygon text requested. The write is not atomic and the firmware may ignore points beyond the supported count; check the polygon mode switch and the reported value.',
  partial:
    'Submission stopped. Some coordinates may have been requested; inspect reported values before retrying.',
  requested:
    'Coordinates requested. These writes are not atomic. Check reported values and firmware mode before using the zone; submission does not prove device persistence.',
} as const;
/** Firmware polygon envelope in millimetres; outside it the firmware cannot represent the zone. */
export const POLYGON_X_LIMIT_MM = 6000;
export const POLYGON_Y_MAX_MM = 6000;
/** Firmware polygon parser bound: at most 20 points (40 floats). */
export const POLYGON_MAX_POINTS = 20;
export const POLYGON_MIN_POINTS = 3;

/** Blocking messages shared by the preview and command path. */
export const ZONE_MESSAGES = {
  rectangleOnly: 'Only valid rectangular zones can be applied through coordinate entities.',
  missingCoordinate: (role: ZoneCoordinateRole) => `Map ${role} to an available readable/writable length entity.`,
  duplicateCoordinate: 'Choose four distinct coordinate entities.',
  outOfRange: (role: ZoneCoordinateRole) =>
    `${role} is outside the entity’s advertised bounds, or bounds are unavailable.`,
  offStep: (role: ZoneCoordinateRole) => `${role} does not align with the entity’s advertised step.`,
  polygonShape: 'Only polygon zones can be applied through a polygon text entity.',
  missingPolygon: 'Map the polygon role to an available readable/writable text entity.',
  polygonPoints: (count: number) => `A polygon needs between ${POLYGON_MIN_POINTS} and ${POLYGON_MAX_POINTS} points; this one has ${count}.`,
  polygonInteger: 'Polygon coordinates must be whole millimetres; the firmware parser reads integers.',
  polygonEnvelope: `Polygon points must stay within X ±${POLYGON_X_LIMIT_MM} mm and Y 0–${POLYGON_Y_MAX_MM} mm.`,
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

/** Writable numeric entities usable for non-coordinate zone settings such as off delay. */
export function zoneNumericOptions(exposes: Record<string, Expose>): Expose[] {
  return Object.values(exposes)
    .filter(
      (expose) =>
        expose.type === 'numeric' && expose.access_mode === 'readwrite' && !exposeControlUnavailable(expose)
    )
    .sort((left, right) => exposeDisplayLabel(left).localeCompare(exposeDisplayLabel(right)));
}

/** Coordinate entities offered by the zone editor, ordered by display label. */
export function zoneCoordinateOptions(exposes: Record<string, Expose>): Expose[] {
  return Object.values(exposes)
    .filter(isZoneCoordinate)
    .sort((left, right) => exposeDisplayLabel(left).localeCompare(exposeDisplayLabel(right)));
}

/** Text entities that can carry a firmware polygon value. */
export function isZonePolygonEntity(expose: Expose): boolean {
  return expose.type === 'text' && expose.access_mode === 'readwrite' && !exposeControlUnavailable(expose);
}

/** Polygon text entities offered by the zone editor, ordered by display label. */
export function zonePolygonOptions(exposes: Record<string, Expose>): Expose[] {
  return Object.values(exposes)
    .filter(isZonePolygonEntity)
    .sort((left, right) => exposeDisplayLabel(left).localeCompare(exposeDisplayLabel(right)));
}

/**
 * Serialises a drawn polygon for the firmware parser: integer millimetres in draw order,
 * separated by semicolons. It never rounds or clamps; invalid geometry blocks the write.
 */
export function serializeZonePolygon(points: readonly SpatialPoint[]): { error: string; value: string } {
  if (points.length < POLYGON_MIN_POINTS || points.length > POLYGON_MAX_POINTS) {
    return { error: ZONE_MESSAGES.polygonPoints(points.length), value: '' };
  }
  for (const point of points) {
    if (!Number.isFinite(point.x) || !Number.isFinite(point.y) || !Number.isInteger(point.x) || !Number.isInteger(point.y)) {
      return { error: ZONE_MESSAGES.polygonInteger, value: '' };
    }
    if (Math.abs(point.x) > POLYGON_X_LIMIT_MM || point.y < 0 || point.y > POLYGON_Y_MAX_MM) {
      return { error: ZONE_MESSAGES.polygonEnvelope, value: '' };
    }
  }
  return { error: '', value: points.map((point) => `${point.x}:${point.y}`).join(';') };
}

/** Plans one polygon text write, or a blocking error; never clamps or truncates. */
export function zonePolygonPlan(
  zone: SpatialZone,
  exposes: Record<string, Expose>
): { error: string; commands: { role: ZoneEntityRole; entity: string; value: string }[] } {
  if (zone.shape !== 'polygon') return { error: ZONE_MESSAGES.polygonShape, commands: [] };
  const entity = exposes[zone.entities?.[ZONE_POLYGON_ROLE] ?? ''];
  if (!entity || !isZonePolygonEntity(entity)) return { error: ZONE_MESSAGES.missingPolygon, commands: [] };
  const serialized = serializeZonePolygon(zone.points);
  if (serialized.error) return { error: serialized.error, commands: [] };
  return { error: '', commands: [{ role: ZONE_POLYGON_ROLE, entity: entity.name, value: serialized.value }] };
}

/** Firmware settings the workspace can map, in display order, with the control they need. */
export const FIRMWARE_SETTINGS: readonly { role: FirmwareRole; label: string; kind: 'number' | 'switch' }[] = [
  { role: 'installationAngle', label: 'Installation angle', kind: 'number' },
  { role: 'inverseMounting', label: 'Inverse mounting', kind: 'switch' },
  { role: 'maxDistance', label: 'Maximum distance', kind: 'number' },
  { role: 'occupancyOffDelay', label: 'Occupancy off delay', kind: 'number' },
  { role: 'polygonMode', label: 'Polygon zone mode', kind: 'switch' },
];

/** Writable entities offered for one firmware setting role. */
export function firmwareEntityOptions(exposes: Record<string, Expose>, role: FirmwareRole): Expose[] {
  const kind = FIRMWARE_SETTINGS.find((setting) => setting.role === role)?.kind;
  return Object.values(exposes)
    .filter(
      (expose) =>
        expose.access_mode === 'readwrite' &&
        !exposeControlUnavailable(expose) &&
        (kind === 'switch' ? expose.type === 'binary' : expose.type === 'numeric')
    )
    .sort((left, right) => exposeDisplayLabel(left).localeCompare(exposeDisplayLabel(right)));
}

/** Validates one numeric firmware value against its entity's advertised range and step. */
function firmwareNumber(
  role: string,
  entity: Expose,
  value: number
): { error: string; command?: { role: string; entity: string; value: number } } {
  if (!Number.isFinite(value)) return { error: `Enter a finite value for ${role}.` };
  const range = exposeRange(entity);
  if (invalidRange(entity) || !range || value < range.min || value > range.max) {
    return { error: `${role} is outside the entity’s advertised bounds.` };
  }
  const steps = (value - range.min) / numericStep(entity);
  if (Math.abs(steps - Math.round(steps)) > ZONE_STEP_TOLERANCE) {
    return { error: `${role} does not align with the entity’s advertised step.` };
  }
  return { error: '', command: { role, entity: entity.name, value } };
}

/** Plans the mapped firmware settings from user-entered native-unit values. */
export function firmwareCommandPlan(
  workspace: SpatialWorkspace,
  exposes: Record<string, Expose>,
  values: Partial<Record<FirmwareRole, number | boolean>>
): { error: string; commands: { role: FirmwareRole; entity: string; value: number | boolean }[] } {
  const commands: { role: FirmwareRole; entity: string; value: number | boolean }[] = [];
  for (const setting of FIRMWARE_SETTINGS) {
    const entityName = workspace.firmwareEntities?.[setting.role];
    if (!entityName) continue;
    const entity = exposes[entityName];
    if (!entity || entity.access_mode !== 'readwrite' || exposeControlUnavailable(entity)) {
      return { error: `Map ${setting.label} to an available writable entity.`, commands: [] };
    }
    const value = values[setting.role];
    if (setting.kind === 'switch') {
      if (entity.type !== 'binary') return { error: `${setting.label} needs a binary switch entity.`, commands: [] };
      if (typeof value !== 'boolean') return { error: `Choose on or off for ${setting.label}.`, commands: [] };
      commands.push({ role: setting.role, entity: entityName, value });
      continue;
    }
    if (entity.type !== 'numeric') return { error: `${setting.label} needs a numeric entity.`, commands: [] };
    if (typeof value !== 'number') return { error: `Enter a value for ${setting.label}.`, commands: [] };
    const planned = firmwareNumber(setting.label, entity, value);
    if (planned.error) return { error: planned.error, commands: [] };
    commands.push({ role: setting.role, entity: entityName, value });
  }
  return { error: '', commands };
}

/** Plans the optional per-zone off-delay number. */
export function zoneOffDelayPlan(
  zone: SpatialZone,
  exposes: Record<string, Expose>,
  seconds: number
): { error: string; commands: { role: ZoneEntityRole; entity: string; value: number }[] } {
  const entityName = zone.entities?.[ZONE_OFF_DELAY_ROLE];
  if (!entityName) return { error: '', commands: [] };
  const entity = exposes[entityName];
  if (!entity || entity.type !== 'numeric' || entity.access_mode !== 'readwrite' || exposeControlUnavailable(entity)) {
    return { error: 'Map the off delay to an available writable numeric entity.', commands: [] };
  }
  const planned = firmwareNumber('Zone off delay', entity, seconds);
  if (planned.error) return { error: planned.error, commands: [] };
  return { error: '', commands: [{ role: ZONE_OFF_DELAY_ROLE, entity: entityName, value: seconds }] };
}

/** Read-only binary entities offered as reported occupancy sources. */
export function occupancyFlagOptions(device: Device): Expose[] {
  return Object.values(device.exposes).filter((expose) => expose.type === 'binary' && expose.access_mode !== 'write');
}

/** One formatted request/report row for the physical-zone editor. */
export type ZoneCommandRow = { role: string; entity: string; request: string; reported: string };

/** Formats requested and currently reported values; text is shown verbatim, numbers with units. */
export function zoneCommandRows(
  commands: readonly { role: string; entity: string; value: number | string | boolean }[],
  exposes: Record<string, Expose>
): ZoneCommandRow[] {
  return commands.map((command) => {
    const expose = exposes[command.entity];
    return {
      role: command.role,
      entity: command.entity,
      request:
        typeof command.value === 'number'
          ? formatWithUnit(command.value, expose.unit)
          : typeof command.value === 'boolean'
            ? command.value
              ? 'on'
              : 'off'
            : command.value,
      reported:
        typeof command.value === 'number'
          ? formatWithUnit(expose.data, expose.unit)
          : formatExposeValue(expose, expose.data),
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
