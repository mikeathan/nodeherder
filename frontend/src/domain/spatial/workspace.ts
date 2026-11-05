/*
 * Pure spatial-workspace domain rules: defaults, copying, geometry, admission
 * validation and coordinate transforms. Components and composables call these
 * helpers instead of carrying geometry or bounds inline.
 */
import type { SpatialPoint, SpatialWorkspace, SpatialZone, ZoneKind } from '@/types/spatial';
import {
  ENTRY_ROLES,
  FIRMWARE_ROLES,
  ROOM_OBJECT_KINDS,
  ZONE_COORDINATE_ROLES,
  ZONE_OFF_DELAY_ROLE,
  ZONE_POLYGON_ROLE,
  ZONE_SHAPES,
} from '@/types/spatial';
import { rotatePoint } from './radar';

/** Persisted workspace schema version. */
export const SPATIAL_WORKSPACE_VERSION = 1;
/** Sensor-origin coordinate bound in millimetres. */
export const SPATIAL_COORDINATE_LIMIT_MM = 20000;
/** Allowed absolute sensor/room rotation in degrees. */
export const SPATIAL_MAX_ROTATION = 360;
/** Allowed illustrative field of view in degrees. */
export const SPATIAL_MAX_FOV = 180;
/** Maximum number of overlays in one workspace. */
export const MAX_SPATIAL_ZONES = 8;
/** Maximum number of visual room items in one workspace. */
export const MAX_SPATIAL_OBJECTS = 40;
/** Maximum overlay vertices accepted from the numeric editor. */
export const MAX_SPATIAL_VERTICES = 16;
/** Minimum overlay vertices required for a polygon. */
export const MIN_SPATIAL_VERTICES = 3;
/** Maximum display characters for an overlay or room-item name. */
export const MAX_SPATIAL_NAME_LENGTH = 80;
/** Maximum bytes for an overlay or room-item identifier. */
export const MAX_SPATIAL_ID_LENGTH = 80;
/** Maximum bytes for a mapped firmware entity identifier. */
export const MAX_SPATIAL_ENTITY_LENGTH = 256;
/** The shoelace sum is twice the area in square millimetres. */
export const MIN_SPATIAL_TWICE_AREA = 1;
/** Tolerance for deciding that a pointer lies on a polygon edge. */
export const SPATIAL_EDGE_EPSILON = 0.001;
/** Offset in millimetres applied when adding a polygon corner. */
export const APPEND_SPATIAL_STEP_MM = 250;

/** Firmware-style slot ceilings; no model capability is inferred by these UI limits. */
export const ZONE_LIMITS: Record<ZoneKind, number> = { detection: 4, exclusion: 2, entry: 2 };
/** Human-readable zone kind used by editor labels and default names. */
export function zoneKindLabel(kind: ZoneKind): string {
  return kind.charAt(0).toUpperCase() + kind.slice(1);
}
/** Starter display name for the next zone of a kind. */
export function defaultZoneName(kind: ZoneKind, count: number): string {
  return `${zoneKindLabel(kind)} ${count + 1}`;
}
/** Starter rectangle used when a new zone is added, in sensor millimetres. */
export const STARTER_ZONE = { x: -500, y: 1000, width: 1000, height: 1000 } as const;
/** Starter polygon used when a new polygon zone is added, in sensor millimetres. */
export const STARTER_POLYGON: readonly SpatialPoint[] = [
  { x: -500, y: 1000 },
  { x: 500, y: 1000 },
  { x: 0, y: 2000 },
];
/** Starter room item used when a new object is added, in millimetres. */
export const STARTER_ROOM_OBJECT = { x: -500, y: 1000, width: 1000, height: 500, wallHeight: 100 } as const;
/** Starter room outline used when the user adds one, in millimetres. */
export const STARTER_ROOM_OUTLINE = { x: -2500, y: -500, width: 5000, height: 5000 } as const;
/** Minimum positive room-item dimension in millimetres. */
export const MIN_SPATIAL_DIMENSION_MM = 1;
/** Furniture names offered by the room builder. */
export const ROOM_FURNITURE_NAMES = [
  'Sofa',
  'Bed',
  'Desk',
  'Dining table',
  'Coffee table',
  'Chair',
  'Wardrobe',
  'TV',
  'Plant',
  'Custom rectangle',
] as const;
/** Editable room-item fields in display order. */
export const ROOM_OBJECT_FIELDS = [
  { key: 'x', label: 'Item X (mm)' },
  { key: 'y', label: 'Item Y (mm)' },
  { key: 'width', label: 'Item width (mm)' },
  { key: 'height', label: 'Item height (mm)' },
  { key: 'rotation', label: 'Item rotation (degrees)' },
] as const;
/** Rectangle bound fields in display order. */
export const ZONE_BOUND_FIELDS = [
  { key: 'x', label: 'Zone X' },
  { key: 'y', label: 'Zone Y' },
  { key: 'width', label: 'Zone width' },
  { key: 'height', label: 'Zone height' },
] as const;
/** Editable sensor-placement fields in display order. */
export const SENSOR_FIELDS = [
  { key: 'x', label: 'Sensor X (mm)' },
  { key: 'y', label: 'Sensor Y (mm)' },
  { key: 'rotation', label: 'Sensor direction (degrees)' },
  { key: 'fov', label: 'Illustrative field of view (degrees; 0 hides)' },
] as const;
/** Maximum independent snapshots retained by the workspace undo stack. */
export const MAX_UNDO_SNAPSHOTS = 20;
/** Confirmation shown after a correlated save acknowledgement. */
export const SPATIAL_SAVE_NOTICE = 'Spatial settings saved in NodeHerder. This does not apply zones to the sensor.';
/** Recovery copy when a save fails without a usable server message. */
export const SPATIAL_SAVE_FAILED = 'Unable to save spatial settings. Your draft and Undo are unchanged.';

/** Creates an independent, sensor-origin workspace for legacy configurations. */
export function createWorkspace(): SpatialWorkspace {
  return {
    version: SPATIAL_WORKSPACE_VERSION,
    zones: [],
    outline: [],
    objects: [],
    sensor: { x: 0, y: 0, rotation: 0, fov: 0 },
    entryEntities: { enabled: '', timeout: '', threshold: '', assumed: '', remaining: '' },
    firmwareEntities: Object.fromEntries(FIRMWARE_ROLES.map((role) => [role, ''])) as Record<string, string>,
  };
}

/** Copies editable settings so unsaved interactions never mutate Vuex state. */
export function copyWorkspace(value?: SpatialWorkspace): SpatialWorkspace {
  if (!value) return createWorkspace();
  const copy = JSON.parse(JSON.stringify(value)) as SpatialWorkspace;
  return {
    ...createWorkspace(),
    ...copy,
    zones: copy.zones ?? [],
    outline: copy.outline ?? [],
    objects: copy.objects ?? [],
    entryEntities: { ...createWorkspace().entryEntities, ...copy.entryEntities },
    firmwareEntities: { ...createWorkspace().firmwareEntities, ...copy.firmwareEntities },
  };
}

/**
 * Stable JSON key used to compare a browser draft with confirmed settings.
 * Go and browser serializers order object keys differently; sorting nested keys
 * prevents an equivalent snapshot from discarding selection or Undo history.
 */
export function workspaceDraftKey(draft: unknown): string {
  return JSON.stringify(draft, (_key, value: unknown) => {
    if (!value || typeof value !== 'object' || Array.isArray(value)) return value;
    return Object.fromEntries(Object.entries(value).sort(([left], [right]) => left.localeCompare(right)));
  });
}

function validCoordinate(value: number): boolean {
  return Number.isFinite(value) && Math.abs(value) <= SPATIAL_COORDINATE_LIMIT_MM;
}

function cross(a: SpatialPoint, b: SpatialPoint, c: SpatialPoint): number {
  return (b.x - a.x) * (c.y - a.y) - (b.y - a.y) * (c.x - a.x);
}

function segmentsIntersect(a: SpatialPoint, b: SpatialPoint, c: SpatialPoint, d: SpatialPoint): boolean {
  // Opposite orientation signs include collinear/touching segments. Bounding boxes
  // distinguish a true collinear overlap from disjoint segments on the same line.
  return (
    cross(a, b, c) * cross(a, b, d) <= 0 &&
    cross(c, d, a) * cross(c, d, b) <= 0 &&
    Math.max(a.x, b.x) >= Math.min(c.x, d.x) &&
    Math.max(c.x, d.x) >= Math.min(a.x, b.x) &&
    Math.max(a.y, b.y) >= Math.min(c.y, d.y) &&
    Math.max(c.y, d.y) >= Math.min(a.y, b.y)
  );
}

/** Validates a simple polygon: finite bounded vertices, nonzero area and no crossing edges. */
export function validPolygon(points: readonly SpatialPoint[]): boolean {
  if (
    points.length < MIN_SPATIAL_VERTICES ||
    points.length > MAX_SPATIAL_VERTICES ||
    points.some((point) => !validCoordinate(point.x) || !validCoordinate(point.y))
  )
    return false;
  let area = 0;
  for (let i = 0; i < points.length; i++) {
    const a = points[i];
    const b = points[(i + 1) % points.length];
    if (a.x === b.x && a.y === b.y) return false;
    // Shoelace terms accumulate twice the signed area; winding does not affect validity.
    area += a.x * b.y - b.x * a.y;
    // Adjacent edges share a legal endpoint. Only non-adjacent pairs must be disjoint.
    for (let j = i + 2; j < points.length; j++) {
      if (i === 0 && j === points.length - 1) continue;
      if (segmentsIntersect(a, b, points[j], points[(j + 1) % points.length])) return false;
    }
  }
  return Math.abs(area) > MIN_SPATIAL_TWICE_AREA;
}

/** Whether four ordered vertices form an axis-aligned positive-size rectangle. */
export function isRectangle(points: readonly SpatialPoint[]): boolean {
  return (
    points.length === 4 &&
    points[0].y === points[1].y &&
    points[1].x === points[2].x &&
    points[2].y === points[3].y &&
    points[3].x === points[0].x
  );
}

function validEntityId(id: string): boolean {
  return typeof id === 'string' && id.length <= MAX_SPATIAL_ENTITY_LENGTH && !id.includes('\0');
}

function validEntityMap(entities: Partial<Record<string, string>> | undefined, roles: readonly string[]): boolean {
  return Object.entries(entities ?? {}).every(
    ([role, id]) => roles.includes(role) && typeof id === 'string' && validEntityId(id)
  );
}

function validItem(id: string, name: string, ids: Set<string>): boolean {
  if (!id || id.length > MAX_SPATIAL_ID_LENGTH || ids.has(id) || !name.trim() || name.length > MAX_SPATIAL_NAME_LENGTH) {
    return false;
  }
  ids.add(id);
  return true;
}

function validZone(zone: SpatialZone, counts: Record<string, number>, ids: Set<string>): boolean {
  const limit = ZONE_LIMITS[zone.kind];
  const currentCount = counts[zone.kind] ?? 0;
  if (!limit || currentCount >= limit || !ZONE_SHAPES.includes(zone.shape)) return false;
  counts[zone.kind] = currentCount + 1;
  if (!validPolygon(zone.points) || (zone.shape === 'rectangle' && !isRectangle(zone.points))) return false;
  const zoneRoles =
    zone.shape === 'polygon'
      ? [ZONE_POLYGON_ROLE, ZONE_OFF_DELAY_ROLE]
      : [...ZONE_COORDINATE_ROLES, ZONE_OFF_DELAY_ROLE];
  if (!validEntityMap(zone.entities, zoneRoles)) return false;
  if (zone.occupancyEntity && !validEntityId(zone.occupancyEntity)) return false;
  return validItem(zone.id, zone.name, ids);
}

function validObject(object: SpatialWorkspace['objects'][number], ids: Set<string>): boolean {
  if (!ROOM_OBJECT_KINDS.includes(object.kind) || !validItem(object.id, object.name, ids)) return false;
  if (![object.x, object.y].every(validCoordinate)) return false;
  if (!Number.isFinite(object.width) || !Number.isFinite(object.height)) return false;
  if (object.width <= 0 || object.height <= 0) return false;
  if (object.width > SPATIAL_COORDINATE_LIMIT_MM || object.height > SPATIAL_COORDINATE_LIMIT_MM) return false;
  return Number.isFinite(object.rotation) && Math.abs(object.rotation) <= SPATIAL_MAX_ROTATION;
}

/** Admission feedback for versioned workspace settings; empty outlines are allowed. */
export function validateWorkspace(workspace: SpatialWorkspace): string {
  if (workspace.version !== SPATIAL_WORKSPACE_VERSION) return 'Unsupported workspace version.';
  if (workspace.zones.length > MAX_SPATIAL_ZONES || workspace.objects.length > MAX_SPATIAL_OBJECTS) {
    return 'Workspace limits exceeded.';
  }
  const { sensor } = workspace;
  if (
    !validCoordinate(sensor.x) ||
    !validCoordinate(sensor.y) ||
    !Number.isFinite(sensor.rotation) ||
    Math.abs(sensor.rotation) > SPATIAL_MAX_ROTATION ||
    !Number.isFinite(sensor.fov) ||
    sensor.fov < 0 ||
    sensor.fov > SPATIAL_MAX_FOV
  ) {
    return 'Check sensor position, rotation and field of view.';
  }
  if (workspace.outline.length && !validPolygon(workspace.outline)) {
    return `Room outline needs ${MIN_SPATIAL_VERTICES}–${MAX_SPATIAL_VERTICES} non-crossing corners.`;
  }
  const ids = new Set<string>();
  const counts: Record<string, number> = {};
  for (const zone of workspace.zones) {
    if (!validZone(zone, counts, ids)) {
      return `Zones need valid rectangles or simple polygons with ${MIN_SPATIAL_VERTICES}–${MAX_SPATIAL_VERTICES} corners, valid entities and free slots.`;
    }
  }
  for (const object of workspace.objects) {
    if (!validObject(object, ids)) return 'Check room object dimensions and rotation.';
  }
  if (!validEntityMap(workspace.entryEntities, ENTRY_ROLES)) {
    return 'Invalid entry entity mapping.';
  }
  if (!validEntityMap(workspace.firmwareEntities ?? {}, FIRMWARE_ROLES)) {
    return 'Invalid firmware entity mapping.';
  }
  return '';
}

/** Creates clockwise rectangle corners in sensor-relative millimetres. */
export function rectanglePoints(x: number, y: number, width: number, height: number): SpatialPoint[] {
  return [
    { x, y },
    { x: x + width, y },
    { x: x + width, y: y + height },
    { x, y: y + height },
  ];
}

/** Appends one starter corner without mutating the supplied list. */
export function appendSpatialPoint(points: readonly SpatialPoint[]): SpatialPoint[] {
  const last = points.at(-1) ?? { x: 0, y: 0 };
  return [...points, { x: last.x + APPEND_SPATIAL_STEP_MM, y: last.y + APPEND_SPATIAL_STEP_MM }];
}

/** Axis-aligned bounds in millimetres for an existing nonempty polygon. */
export function polygonBounds(points: readonly SpatialPoint[]): {
  x: number;
  y: number;
  width: number;
  height: number;
} {
  const x = Math.min(...points.map((point) => point.x));
  const y = Math.min(...points.map((point) => point.y));
  return {
    x,
    y,
    width: Math.max(...points.map((point) => point.x)) - x,
    height: Math.max(...points.map((point) => point.y)) - y,
  };
}

/** Resizes one rectangle edge/corner against its current bounds. */
export function resizeZoneBounds(
  points: readonly SpatialPoint[],
  key: 'x' | 'y' | 'width' | 'height',
  value: number
): SpatialPoint[] {
  const bounds = { ...polygonBounds(points), [key]: value };
  return rectanglePoints(bounds.x, bounds.y, bounds.width, bounds.height);
}

/** Normalizes a dragged rectangle corner against its immutable opposite anchor, both in millimetres. */
export function resizeRectangle(anchor: SpatialPoint, corner: SpatialPoint): SpatialPoint[] {
  return rectanglePoints(
    Math.min(anchor.x, corner.x),
    Math.min(anchor.y, corner.y),
    Math.abs(corner.x - anchor.x),
    Math.abs(corner.y - anchor.y)
  );
}

/** Inclusive point-in-polygon test used only for explicitly labelled occupancy previews. */
export function containsPoint(points: readonly SpatialPoint[], point: SpatialPoint): boolean {
  let inside = false;
  for (let i = 0, j = points.length - 1; i < points.length; j = i++) {
    const a = points[j];
    const b = points[i];
    if (
      Math.abs(cross(a, b, point)) < SPATIAL_EDGE_EPSILON &&
      point.x >= Math.min(a.x, b.x) &&
      point.x <= Math.max(a.x, b.x) &&
      point.y >= Math.min(a.y, b.y) &&
      point.y <= Math.max(a.y, b.y)
    )
      return true;
    if (a.y > point.y !== b.y > point.y && point.x < ((b.x - a.x) * (point.y - a.y)) / (b.y - a.y) + a.x) {
      inside = !inside;
    }
  }
  return inside;
}

/**
 * Whether one point lies inside a zone, honouring exclusion masks. Detection zones are
 * masked by every exclusion overlay; other zone kinds are not. Pure geometry, so callers
 * own freshness and hysteresis decisions.
 */
export function targetInsideZone(
  zone: SpatialZone,
  zones: readonly SpatialZone[],
  point: SpatialPoint
): boolean {
  if (!containsPoint(zone.points, point)) return false;
  if (zone.kind !== 'detection') return true;
  return !zones.some((other) => other.kind === 'exclusion' && containsPoint(other.points, point));
}

/** Transforms one sensor-relative point into room coordinates without display rotation. */
export function roomPoint(point: SpatialPoint, sensor: SpatialWorkspace['sensor']): SpatialPoint {
  const local = rotatePoint(point.x, point.y, sensor.rotation);
  return { x: local.x + sensor.x, y: local.y + sensor.y };
}

/** Transforms sensor coordinates into room coordinates, then applies display rotation and SVG Y inversion. */
export function roomScreenPoint(
  point: SpatialPoint,
  sensor: SpatialWorkspace['sensor'],
  orientation: number
): SpatialPoint {
  return screenPoint(roomPoint(point, sensor), orientation);
}

/** Converts a room point to a rotated SVG point. */
export function screenPoint(point: SpatialPoint, orientation: number): SpatialPoint {
  const rotated = rotatePoint(point.x, point.y, orientation);
  return { x: rotated.x, y: -rotated.y };
}

/** Reverses SVG/view and sensor transforms for pointer editing in millimetres. */
export function sensorPoint(
  point: SpatialPoint,
  sensor: SpatialWorkspace['sensor'],
  orientation: number
): SpatialPoint {
  const room = rotatePoint(point.x, -point.y, -orientation);
  return rotatePoint(room.x - sensor.x, room.y - sensor.y, -sensor.rotation);
}
