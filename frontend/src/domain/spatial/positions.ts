import { CANVAS_SENSOR_FONT_RATIO, heatCellCorners, sensorProjectedPoint } from './canvas';
import { SPATIAL_DEFAULT_SCALE_MM } from './index';
import { centeredSpatialView, radarViewBox, type SpatialViewBounds } from './radar';
import { screenPoint } from './workspace';

/** Canvas view box plus symbol scaling for one Positions render. */
export type PositionViewMetrics = {
  viewBox: string;
  symbolScale: number;
  labelStroke: number;
};

const TARGET_LABEL_PX = 14;
const LABEL_STROKE_RATIO = 0.25;

/**
 * Fits recorded content around the sensor and derives the symbol scale for the Positions
 * canvas. The sensor remains the view centre; the fallback is the standard sensor view.
 */
export function positionViewMetrics(input: {
  bounds: SpatialViewBounds | null;
  center: { x: number; y: number };
  viewerWidth: number;
  viewerHeight: number;
  zoom: number;
}): PositionViewMetrics {
  const { bounds, center, viewerWidth, viewerHeight, zoom } = input;
  if (!bounds || viewerWidth <= 0 || viewerHeight <= 0) {
    return { viewBox: radarViewBox(SPATIAL_DEFAULT_SCALE_MM), symbolScale: zoom, labelStroke: 8 };
  }
  const view = centeredSpatialView(bounds, center, viewerWidth / viewerHeight);
  const fontMm = SPATIAL_DEFAULT_SCALE_MM * CANVAS_SENSOR_FONT_RATIO;
  const symbolScale = (fontMm * viewerHeight * zoom) / (TARGET_LABEL_PX * view.height);
  const labelStroke = (SPATIAL_DEFAULT_SCALE_MM * CANVAS_SENSOR_FONT_RATIO * LABEL_STROKE_RATIO) / symbolScale;
  return {
    viewBox: [view.x, view.y, view.width, view.height].join(' '),
    symbolScale,
    labelStroke,
  };
}

/** Axis-aligned box in millimetres, as the fitted view needs it. */
export type PointsBounds = { minX: number; minY: number; maxX: number; maxY: number };

/** Accumulates points into an axis-aligned box; ignores nothing and never fabricates a point. */
class BoundsAccumulator {
  private minX = Number.POSITIVE_INFINITY;
  private minY = Number.POSITIVE_INFINITY;
  private maxX = Number.NEGATIVE_INFINITY;
  private maxY = Number.NEGATIVE_INFINITY;

  add(point: { x: number; y: number }): void {
    this.minX = Math.min(this.minX, point.x);
    this.minY = Math.min(this.minY, point.y);
    this.maxX = Math.max(this.maxX, point.x);
    this.maxY = Math.max(this.maxY, point.y);
  }

  /** The box, or null when nothing was added. */
  finish(): PointsBounds | null {
    if (!Number.isFinite(this.minX)) return null;
    return { minX: this.minX, minY: this.minY, maxX: this.maxX, maxY: this.maxY };
  }

  /** Centre of the box, or null when nothing was added. */
  centre(): { x: number; y: number } | null {
    const bounds = this.finish();
    return bounds
      ? { x: (bounds.minX + bounds.maxX) / 2, y: (bounds.minY + bounds.maxY) / 2 }
      : null;
  }
}

/**
 * The fitted view box for the recorded Positions canvas.
 *
 * Contains the sensor, the room outline and the confirmed recording's projected points (plus the
 * heat cells when the heatmap is shown), then pads it. Zoom shrinks this around its centre in
 * `positionViewMetrics`, so the panel height never changes.
 *
 * @param input.workspace Room outline and sensor, in millimetres; projected here.
 * @param input.points Projected positions of the confirmed recording.
 * @param input.heatCells Heat cells to include; only when the heatmap view is shown.
 * @param input.orientation Display rotation applied to every input.
 * @returns The padded box, or null when there is nothing to contain.
 */
export function fittedPointsBounds(input: {
  workspace: {
    outline: { x: number; y: number }[];
    sensor: { x: number; y: number; rotation: number; fov: number };
  };
  points: readonly { x: number; y: number }[];
  heatCells: readonly { x: number; y: number }[];
  orientation: number;
}): PointsBounds | null {
  const accumulator = new BoundsAccumulator();
  // Every input is projected into the same rotated SVG space as the points: the room outline is
  // room-relative, and the heat cells are sensor-relative, so adding them unprojected mixed two
  // coordinate spaces and fitted the wrong box (worse with a flipped/rotated view).
  accumulator.add(screenPoint(input.workspace.sensor, input.orientation));
  for (const point of input.workspace.outline) {
    accumulator.add(screenPoint(point, input.orientation));
  }
  for (const point of input.points) accumulator.add(point);
  for (const cell of input.heatCells) {
    for (const corner of heatCellCorners(cell)) {
      accumulator.add(sensorProjectedPoint(corner, input.workspace.sensor, input.orientation));
    }
  }
  const bounds = accumulator.finish();
  if (!bounds) return null;
  const width = Math.max(500, bounds.maxX - bounds.minX);
  const height = Math.max(500, bounds.maxY - bounds.minY);
  const pad = Math.max(width, height) * 0.08 + 150;
  return { minX: bounds.minX - pad, minY: bounds.minY - pad, maxX: bounds.maxX + pad, maxY: bounds.maxY + pad };
}

/**
 * The screen point a zoom level should be anchored on: the centre of the heat activity when the
 * heatmap is shown, otherwise the sample current at the playhead, otherwise the sensor.
 *
 * Captured when the level changes or the playhead is moved explicitly, never on a playback tick,
 * so the target visibly travels across a fixed view.
 */
export function zoomAnchorPoint(input: {
  heatCells: readonly { x: number; y: number }[];
  projectedPoints: readonly { at: number; x: number; y: number }[];
  at: number;
  /** Raw workspace sensor; projected here so it shares the points' coordinate space. */
  sensor: { x: number; y: number; rotation: number; fov: number };
  orientation: number;
  heatmap: boolean;
}): { x: number; y: number } {
  if (input.heatmap) {
    const accumulator = new BoundsAccumulator();
    for (const cell of input.heatCells) {
      for (const corner of heatCellCorners(cell)) {
        accumulator.add(sensorProjectedPoint(corner, input.sensor, input.orientation));
      }
    }
    const centre = accumulator.centre();
    if (centre) return centre;
  }
  const current = input.projectedPoints.filter((point) => point.at <= input.at).at(-1);
  return current ? { x: current.x, y: current.y } : screenPoint(input.sensor, input.orientation);
}
