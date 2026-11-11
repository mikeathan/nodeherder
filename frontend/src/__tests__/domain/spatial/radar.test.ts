import { describe, expect, test } from '@jest/globals';
import {
  RADAR_GRID_DIVISIONS,
  centeredSpatialView,
  radarGridOffsets,
  radarGridSpacing,
  radarViewBox,
  rotatePoint,
  spatialTargetLabel,
  spatialTrackRows,
  formatSpatialDistance,
} from '@/domain/spatial/radar';

describe('spatial radar helpers', () => {
  test('rotates points clockwise', () => {
    const rotated = rotatePoint(1, 0, 90);
    expect(rotated.x).toBeCloseTo(0);
    expect(rotated.y).toBeCloseTo(1);
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
    expect(spatialTrackRows([{ label: 'Track 1', state: 'live', x: 100, y: 200 }])).toEqual([
      'Track 1 (live): X 0.1 m, Y 0.2 m',
    ]);
  });

  describe('sensor-centred views', () => {
    test('centres on the sensor and contains asymmetric content', () => {
      const view = centeredSpatialView({ minX: -500, minY: 0, maxX: 3000, maxY: 6000 }, { x: 0, y: 0 }, 1);
      expect(view).toEqual({ x: -6000, y: -6000, width: 12000, height: 12000 });
      expect(view.x + view.width / 2).toBe(0);
      expect(view.y + view.height / 2).toBe(0);
    });

    test('keeps the canvas aspect while containing the content', () => {
      const view = centeredSpatialView({ minX: -500, minY: 0, maxX: 3000, maxY: 6000 }, { x: 0, y: 0 }, 2);
      expect(view.width / view.height).toBeCloseTo(2, 5);
      expect(view.x).toBeLessThanOrEqual(-500);
      expect(view.x + view.width).toBeGreaterThanOrEqual(3000);
      expect(view.y).toBeLessThanOrEqual(0);
      expect(view.y + view.height).toBeGreaterThanOrEqual(6000);
    });

    test('centres on an offset sensor', () => {
      const view = centeredSpatialView({ minX: 0, minY: -1000, maxX: 2000, maxY: 0 }, { x: 1000, y: -500 }, 1);
      expect(view.x + view.width / 2).toBe(1000);
      expect(view.y + view.height / 2).toBe(-500);
    });

    test('falls back to a square for an invalid aspect', () => {
      const view = centeredSpatialView({ minX: -1000, minY: -1000, maxX: 1000, maxY: 1000 }, { x: 0, y: 0 }, 0);
      expect(view.width).toBe(view.height);
    });
  });
});

test('known spatial quantities retain millimetre precision and shared missing markers', () => {
  expect(formatSpatialDistance(-1)).toBe('-0.001 m');
  expect(formatSpatialDistance(1234)).toBe('1.234 m');
  expect(formatSpatialDistance(null)).toBe('—');
  expect(formatSpatialDistance(NaN)).toBe('—');
});

describe('target distance labels', () => {
  test('appends a rounded distance', () => {
    expect(spatialTargetLabel('Track 1', 2400)).toBe('Track 1 · 2.4 m');
    expect(spatialTargetLabel('Track 2', 250)).toBe('Track 2 · 0.3 m');
  });

  test('quantises to 0.1 m so filter noise does not rewrite the text', () => {
    const stable = spatialTargetLabel('Track 1', 2431);
    // Sub-centimetre filter wobble must not change the text.
    for (const wobble of [2400, 2412, 2431, 2444]) {
      expect(spatialTargetLabel('Track 1', wobble)).toBe(stable);
    }
    // Crossing a 0.1 m boundary is a real change and does update it.
    expect(spatialTargetLabel('Track 1', 2468)).toBe('Track 1 · 2.5 m');
    expect(spatialTargetLabel('Track 1', 2600)).not.toBe(stable);
  });

  test('never invents a distance for an unreadable position', () => {
    expect(spatialTargetLabel('Track 1', null)).toBe('Track 1');
    expect(spatialTargetLabel('Track 1', undefined)).toBe('Track 1');
    expect(spatialTargetLabel('Track 1', Number.NaN)).toBe('Track 1');
  });
});
