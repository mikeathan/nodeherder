import { describe, expect, test } from '@jest/globals';
import {
  RADAR_GRID_DIVISIONS,
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
});

test('known spatial quantities retain millimetre precision and shared missing markers', () => {
  expect(formatSpatialDistance(-1)).toBe('-0.001 m');
  expect(formatSpatialDistance(1234)).toBe('1.234 m');
  expect(formatSpatialDistance(null)).toBe('—');
  expect(formatSpatialDistance(NaN)).toBe('—');
});

describe('target distance labels', () => {
  test('leaves the label untouched when the option is off', () => {
    expect(spatialTargetLabel('Track 1', 2400, false)).toBe('Track 1');
  });

  test('appends a rounded distance when the option is on', () => {
    expect(spatialTargetLabel('Track 1', 2400, true)).toBe('Track 1 · 2.4 m');
    expect(spatialTargetLabel('Track 2', 250, true)).toBe('Track 2 · 0.3 m');
  });

  test('quantises to 0.1 m so filter noise does not rewrite the text', () => {
    const stable = spatialTargetLabel('Track 1', 2431, true);
    // Sub-centimetre filter wobble must not change the text.
    for (const wobble of [2400, 2412, 2431, 2444]) {
      expect(spatialTargetLabel('Track 1', wobble, true)).toBe(stable);
    }
    // Crossing a 0.1 m boundary is a real change and does update it.
    expect(spatialTargetLabel('Track 1', 2468, true)).toBe('Track 1 · 2.5 m');
    expect(spatialTargetLabel('Track 1', 2600, true)).not.toBe(stable);
  });

  test('never invents a distance for an unreadable position', () => {
    expect(spatialTargetLabel('Track 1', null, true)).toBe('Track 1');
    expect(spatialTargetLabel('Track 1', undefined, true)).toBe('Track 1');
    expect(spatialTargetLabel('Track 1', Number.NaN, true)).toBe('Track 1');
  });
});
