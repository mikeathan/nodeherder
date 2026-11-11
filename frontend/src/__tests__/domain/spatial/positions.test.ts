import { describe, expect, test } from '@jest/globals';
import { SPATIAL_DEFAULT_SCALE_MM } from '@/domain/spatial';
import { positionViewMetrics } from '@/domain/spatial/positions';
import { radarViewBox } from '@/domain/spatial/radar';

describe('positions view metrics', () => {
  test('falls back to the standard sensor view before layout is measured', () => {
    expect(
      positionViewMetrics({
        bounds: null,
        center: { x: 0, y: 0 },
        viewerWidth: 0,
        viewerHeight: 0,
        zoom: 2,
      })
    ).toEqual({ viewBox: radarViewBox(SPATIAL_DEFAULT_SCALE_MM), symbolScale: 2, labelStroke: 8 });
  });

  test('centres the sensor, keeps the canvas aspect and contains the content', () => {
    const metrics = positionViewMetrics({
      bounds: { minX: -500, minY: 0, maxX: 3000, maxY: 6000 },
      center: { x: 0, y: 0 },
      viewerWidth: 800,
      viewerHeight: 400,
      zoom: 1,
    });
    const [x, y, width, height] = metrics.viewBox.split(' ').map(Number);
    expect(width / height).toBeCloseTo(2, 5);
    expect(x + width / 2).toBeCloseTo(0, 1);
    expect(y + height / 2).toBeCloseTo(0, 1);
    expect(x).toBeLessThanOrEqual(-500);
    expect(x + width).toBeGreaterThanOrEqual(3000);
    expect(y).toBeLessThanOrEqual(0);
    expect(y + height).toBeGreaterThanOrEqual(6000);
    expect(metrics.symbolScale).toBeGreaterThan(0);
    expect(metrics.labelStroke).toBeGreaterThan(0);
  });

  test('zoom scales symbols without changing the base view box', () => {
    const input = {
      bounds: { minX: -500, minY: 0, maxX: 3000, maxY: 6000 },
      center: { x: 0, y: 0 },
      viewerWidth: 800,
      viewerHeight: 400,
      zoom: 1,
    };
    const one = positionViewMetrics(input);
    const two = positionViewMetrics({ ...input, zoom: 2 });
    expect(two.viewBox).toBe(one.viewBox);
    expect(two.symbolScale).toBeCloseTo(one.symbolScale * 2, 8);
    expect(two.labelStroke).toBeCloseTo(one.labelStroke / 2, 8);
  });
});
