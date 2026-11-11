/*
 * Pure geometry and text helpers for the spatial radar drawing.
 * The SVG component consumes these values and contains no calculations.
 */
import { formatQuantity } from '@/utils/units';

/** Number of subdivisions between the centre and either edge. */
export const RADAR_GRID_DIVISIONS = 5;
/** Sensor dot radius relative to the displayed range. */
export const RADAR_SENSOR_RATIO = 1 / 45;
/** Target dot radius relative to the displayed range. */
export const RADAR_TARGET_RATIO = 1 / 35;
/** Label offsets and type size relative to the displayed range. */
export const RADAR_LABEL_X_RATIO = 1 / 25;
export const RADAR_LABEL_Y_RATIO = -1 / 40;
export const RADAR_LABEL_FONT_RATIO = 1 / 20;
/** Degrees-to-radians conversion shared by spatial geometry. */
export const DEGREES_TO_RADIANS = Math.PI / 180;
/** Millimetres per metre for shared distance formatting. */
export const MILLIMETRES_PER_METRE = 1000;

/** View box covering `scale` millimetres in every direction from the sensor. */
export function radarViewBox(scale: number): string {
  return `${-scale} ${-scale} ${scale * 2} ${scale * 2}`;
}

/** Axis-aligned bounds in projected canvas units (millimetres). */
export type SpatialViewBounds = { minX: number; minY: number; maxX: number; maxY: number };

/** Canvas view box in projected units. */
export type SpatialViewBox = { x: number; y: number; width: number; height: number };

/**
 * Builds an aspect-correct view box centred on `center` that contains `bounds`.
 *
 * The extent is symmetric around the centre, so a sensor-centred surface keeps the sensor
 * in the middle while the recorded content still fits. A non-positive or non-finite aspect
 * falls back to a square.
 */
export function centeredSpatialView(
  bounds: SpatialViewBounds,
  center: { x: number; y: number },
  aspect: number
): SpatialViewBox {
  const safeAspect = Number.isFinite(aspect) && aspect > 0 ? aspect : 1;
  const halfWidth = Math.max(Math.abs(bounds.minX - center.x), Math.abs(bounds.maxX - center.x));
  const halfHeight = Math.max(Math.abs(bounds.minY - center.y), Math.abs(bounds.maxY - center.y));
  const height = Math.max(2 * Math.max(halfHeight, halfWidth / safeAspect), 1);
  const width = height * safeAspect;
  const round = (value: number) => Number(value.toFixed(1));
  return {
    x: round(center.x - width / 2),
    y: round(center.y - height / 2),
    width: round(width),
    height: round(height),
  };
}

/** Grid line offsets from the centre, in millimetres. */
export function radarGridOffsets(scale: number): number[] {
  const step = scale / RADAR_GRID_DIVISIONS;
  return Array.from({ length: RADAR_GRID_DIVISIONS * 2 - 1 }, (_, index) => -scale + step * (index + 1));
}

/** Distance between two adjacent grid lines. */
export function radarGridSpacing(scale: number): number {
  return scale / RADAR_GRID_DIVISIONS;
}

/** Rotates a point clockwise by the given view orientation, in degrees. */
export function rotatePoint(x: number, y: number, degrees: number): { x: number; y: number } {
  const radians = degrees * DEGREES_TO_RADIANS;
  return {
    x: x * Math.cos(radians) - y * Math.sin(radians),
    y: x * Math.sin(radians) + y * Math.cos(radians),
  };
}

/** Accessible text list for the radar, using the same units as the drawing. */
export function spatialTrackRows(tracks: readonly { label: string; state: string; x: number; y: number }[]): string[] {
  return tracks.map(
    (track) =>
      `${track.label} (${track.state}): X ${formatSpatialDistance(track.x)}, Y ${formatSpatialDistance(track.y)}`
  );
}

/** Formats known millimetre coordinates as metres, retaining millimetre precision and shared missing-value behavior. */
export function formatSpatialDistance(millimetres: number | null | undefined): string {
  return formatQuantity(
    millimetres == null ? millimetres : millimetres / MILLIMETRES_PER_METRE,
    'distance',
    { digits: 3 }
  );
}

/**
 * Distance label step in millimetres: displayed distances are quantised to 0.1 m so a
 * filtered position hovering within centimetres does not rewrite the text every frame.
 */
export const SPATIAL_DISTANCE_LABEL_STEP_MM = 100;

/**
 * Label for one target, including its distance from the sensor.
 *
 * The caller passes the millimetres of the position actually drawn, so the text can never
 * disagree with the dot. A missing coordinate leaves the name alone rather than printing a
 * placeholder next to a target that has no readable position.
 */
export function spatialTargetLabel(
  label: string,
  millimetres: number | null | undefined
): string {
  if (millimetres == null || !Number.isFinite(millimetres)) return label;
  const stepped = Math.round(millimetres / SPATIAL_DISTANCE_LABEL_STEP_MM) * SPATIAL_DISTANCE_LABEL_STEP_MM;
  const metres = stepped / MILLIMETRES_PER_METRE;
  return `${label} · ${formatQuantity(metres, 'distance', { digits: 1 })}`;
}
