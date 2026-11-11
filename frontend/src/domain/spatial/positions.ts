import { CANVAS_SENSOR_FONT_RATIO } from './canvas';
import { SPATIAL_DEFAULT_SCALE_MM } from './index';
import { centeredSpatialView, radarViewBox, type SpatialViewBounds } from './radar';

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
