/*
 * Pure projection, snapping and SVG attribute helpers for the live spatial
 * canvas. The Vue component only binds events and renders these values.
 */
import type { HeatCell } from './analytics';
import type { SpatialPoint, SpatialWorkspace } from '@/types/spatial';
import { HEAT_CELL_MM } from './analytics';
import { roomScreenPoint, screenPoint } from './workspace';
import { DEGREES_TO_RADIANS, rotatePoint } from './radar';

/** Pointer movement snaps to this millimetre grid. */
export const CANVAS_SNAP_MM = 50;
/** Heat-cell fill opacity. */
export const HEAT_CELL_OPACITY = '0.65';
/** HSL hue range: the busiest cell is red, the coldest is blue. */
export const HEAT_HUE_START = 240;
export const HEAT_HUE_RANGE = 240;
export const HEAT_SATURATION = 80;
export const HEAT_LIGHTNESS = 50;
/** Radii and label offsets relative to the displayed range. */
export const CANVAS_SENSOR_RADIUS_RATIO = 0.02;
export const CANVAS_SENSOR_LABEL_X_RATIO = 0.03;
export const CANVAS_SENSOR_LABEL_Y_RATIO = 0.06;
export const CANVAS_SENSOR_FONT_RATIO = 0.04;
export const CANVAS_TARGET_RADIUS_RATIO = 0.023;
export const CANVAS_TARGET_LABEL_X_RATIO = 0.03;
export const CANVAS_TARGET_LABEL_Y_RATIO = -0.025;
export const CANVAS_TARGET_FONT_RATIO = 0.05;
export const CANVAS_HANDLE_RADIUS_RATIO = 0.025;
export const CANVAS_OBJECT_LABEL_OFFSET_RATIO = 0.06;
export const CANVAS_OBJECT_FONT_RATIO = 0.035;
export const CANVAS_ZONE_LABEL_OFFSET_RATIO = 0.035;
export const CANVAS_ZONE_FONT_RATIO = 0.04;

/** Heading arrow length in seconds of current velocity. */
export const CANVAS_HEADING_SECONDS = 0.25;
/** Speed below which no heading arrow is drawn, in millimetres per second. */
export const CANVAS_HEADING_MIN_SPEED_MM_S = 120;

/** Serializes points for an SVG polygon/polyline `points` attribute. */
export function pointsAttribute(points: readonly SpatialPoint[]): string {
  return points.map((point) => `${point.x},${point.y}`).join(' ');
}

/** Projects sensor-relative points into the rotated room view. */
export function sensorPolygon(
  points: readonly SpatialPoint[],
  sensor: SpatialWorkspace['sensor'],
  orientation: number
): string {
  return pointsAttribute(points.map((point) => roomScreenPoint(point, sensor, orientation)));
}

/** Projects room-relative points into the rotated SVG view. */
export function roomPolygon(points: readonly SpatialPoint[], orientation: number): string {
  return pointsAttribute(points.map((point) => screenPoint(point, orientation)));
}

/** Projects one sensor-relative point into the rotated room view. */
export function sensorProjectedPoint(
  point: SpatialPoint,
  sensor: SpatialWorkspace['sensor'],
  orientation: number
): SpatialPoint {
  return roomScreenPoint(point, sensor, orientation);
}

/** Builds the illustrative sensor field-of-view triangle. Zero FOV renders nothing. */
export function fovPolygon(scale: number, sensor: SpatialWorkspace['sensor'], orientation: number): string {
  if (!sensor.fov) return '';
  const angle = (sensor.fov / 2) * DEGREES_TO_RADIANS;
  const left = { x: -Math.sin(angle) * scale, y: Math.cos(angle) * scale };
  const right = { x: Math.sin(angle) * scale, y: Math.cos(angle) * scale };
  return sensorPolygon([{ x: 0, y: 0 }, left, right], sensor, orientation);
}

/** Corners of one rotated room item, in room coordinates. */
export function objectCorners(object: SpatialWorkspace['objects'][number]): SpatialPoint[] {
  const corners = [
    { x: 0, y: 0 },
    { x: object.width, y: 0 },
    { x: object.width, y: object.height },
    { x: 0, y: object.height },
  ];
  return corners.map((corner) => {
    const rotated = rotatePoint(corner.x, corner.y, object.rotation);
    return { x: rotated.x + object.x, y: rotated.y + object.y };
  });
}

/** Projects a room item into the rotated SVG view. */
export function objectPolygon(object: SpatialWorkspace['objects'][number], orientation: number): string {
  return roomPolygon(objectCorners(object), orientation);
}

/** Corners of one heat cell in sensor-relative millimetres. */
export function heatCellCorners(cell: HeatCell): SpatialPoint[] {
  return [
    { x: cell.x, y: cell.y },
    { x: cell.x + HEAT_CELL_MM, y: cell.y },
    { x: cell.x + HEAT_CELL_MM, y: cell.y + HEAT_CELL_MM },
    { x: cell.x, y: cell.y + HEAT_CELL_MM },
  ];
}

/** Projects one heat cell into the rotated room view. */
export function heatCellPolygon(cell: HeatCell, sensor: SpatialWorkspace['sensor'], orientation: number): string {
  return sensorPolygon(heatCellCorners(cell), sensor, orientation);
}

/** Colour for a normalized heat intensity, where 1 is the busiest cell. */
export function heatCellColor(intensity: number): string {
  return `hsl(${HEAT_HUE_START - intensity * HEAT_HUE_RANGE} ${HEAT_SATURATION}% ${HEAT_LIGHTNESS}%)`;
}

/** Rounds a millimetre value to the canvas snap grid. */
export function snapMillimetres(value: number): number {
  return Math.round(value / CANVAS_SNAP_MM) * CANVAS_SNAP_MM;
}

/** Rounds one axis-aligned point to the canvas snap grid. */
export function snapPoint(point: SpatialPoint): SpatialPoint {
  return { x: snapMillimetres(point.x), y: snapMillimetres(point.y) };
}

/** Snapped translation between a gesture start and its current pointer position. */
export function snapDelta(start: SpatialPoint, current: SpatialPoint): SpatialPoint {
  return { x: snapMillimetres(current.x - start.x), y: snapMillimetres(current.y - start.y) };
}
