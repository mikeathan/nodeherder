import { describe, expect, test } from '@jest/globals';
import {
  RADAR_GRID_DIVISIONS,
  radarGridOffsets,
  radarGridSpacing,
  radarTargetPositions,
  radarViewBox,
  rotatePoint,
  spatialTargetRows,
  formatSpatialDistance,
} from '@/domain/spatial/radar';

describe('spatial radar helpers', () => {
  test('rotates points clockwise and inverts SVG Y for screen positions', () => {
    const rotated = rotatePoint(1, 0, 90);
    expect(rotated.x).toBeCloseTo(0);
    expect(rotated.y).toBeCloseTo(1);
    const [position] = radarTargetPositions([{ id: 1, x: 1000, y: 0, distance: 1000 }], 90);
    expect(position.id).toBe(1);
    expect(position.x).toBeCloseTo(0);
    expect(position.y).toBeCloseTo(-1000);
  });

  test('builds bounded grid offsets around the sensor', () => {
    const offsets = radarGridOffsets(1000);
    expect(offsets).toHaveLength(RADAR_GRID_DIVISIONS * 2 - 1);
    expect(offsets[0]).toBe(-800);
    expect(offsets.at(-1)).toBe(800);
    expect(radarGridSpacing(1000)).toBe(200);
  });

  test('view box and accessible rows use the same units', () => {
    expect(radarViewBox(6000)).toBe('-6000 -6000 12000 12000');
    expect(spatialTargetRows([{ id: 1, x: 100, y: 200, distance: Math.hypot(100, 200) }])).toEqual([
      'Target 1: X 0.1 m, Y 0.2 m, distance 0.224 m',
    ]);
  });
});

test('known spatial quantities retain millimetre precision and shared missing markers', () => {
  expect(formatSpatialDistance(-1)).toBe('-0.001 m');
  expect(formatSpatialDistance(1234)).toBe('1.234 m');
  expect(formatSpatialDistance(null)).toBe('—');
  expect(formatSpatialDistance(NaN)).toBe('—');
});
